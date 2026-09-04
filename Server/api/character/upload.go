package character

import (
	"archive/zip"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"go.uber.org/zap"
)

// v2 exists to lock out clients older than the fix for the upload loop in
// IAGD 1.5.9732-1.5.9735: those clients re-upload every character once a
// second for the lifetime of the session, which cost more in S3 PUTs than the
// rest of the service combined. LegacyUploadPath answers them without doing
// any work. Item sync is unaffected; character backup is a courtesy service
// and old clients simply lose it until they update.
const UploadPath = "/character/upload/v2"
const UploadMethod = routing.POST

const LegacyUploadPath = "/character/upload"
const LegacyUploadMethod = routing.POST

// MinUploadInterval caps how often a single character can be written to S3.
// Backups are best-effort; a character that changes several times a day is
// stored once a day.
const MinUploadInterval = 24 * time.Hour

var bucket = os.Getenv(config.BucketName)

// acceptedExtensions are the file types a Grim Dawn save folder is allowed to
// contain. Anything else means the archive is not a save folder.
var acceptedExtensions = map[string]bool{
	".gdc": true,
	".gdd": true,
	".fow": true,
	".dat": true,
	".bin": true,
	".cpn": true,
	".gst": true,
	".gsh": true,
}

// LegacyUploadProcessRequest retires the unversioned endpoint. Deliberately
// does no auth, no database work and no reading of the request body: clients
// stuck in the upload loop hit this constantly and must stay cheap.
func LegacyUploadProcessRequest(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{
		"msg": "This version of Item Assistant is no longer able to back up characters. Please update to the latest version.",
	})
}

// UploadProcessRequest stores a character backup, unless we already hold it.
func UploadProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)
	user := routing.GetUser(c)
	email := routing.GetEmail(c)

	name, ok := c.GetQuery("name")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "name" is missing. Please provide the character name.`})
		return
	}

	db := storage.CharacterDb{}
	state, err := db.GetSyncState(email, name)
	if err != nil {
		logger.Warn("Error reading character sync state", zap.Error(err), zap.Any("user", user))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error reading character state"})
		return
	}

	// Rate limit before touching S3. Answered as a success so that a
	// well-behaved client records the character as handled and waits, rather
	// than treating it as a failure and retrying in a tight loop.
	if state != nil && state.Age(time.Now()) < MinUploadInterval {
		c.JSON(http.StatusOK, gin.H{"msg": "Already backed up today, skipping.", "status": StatusThrottled})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logger.Warn("Error receiving/reading uploaded file", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"msg": `Forgot to attach the file? Param: "file"`})
		return
	}

	zipReader, err := zip.NewReader(file, header.Size)
	if err != nil {
		logger.Warn("Error reading uploaded zip file", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The provided file does not appear to be a valid zip file`})
		return
	}

	for _, entry := range zipReader.File {
		if !acceptedExtensions[filepath.Ext(entry.Name)] {
			logger.Warn("Attempt at uploading a file with the wrong extension", zap.String("filename", entry.Name))
			c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf(`The filename "%s" does not appear to belong to Grim Dawn`, entry.Name)})
			return
		}
	}

	// Skip the upload when the archive holds exactly what we already stored.
	// The client re-zips on every attempt and the archive carries a timestamped
	// comment, so the bytes differ even when the save does not; hashing the
	// contents rather than the file is what makes this reliable.
	contentHash := hashArchiveContents(zipReader)
	if state != nil && state.ContentHash == contentHash {
		if err := db.Touch(email, name); err != nil {
			logger.Warn("Error updating character entry", zap.Error(err), zap.Any("user", user))
		}
		c.JSON(http.StatusOK, gin.H{"msg": "Unchanged since last backup, skipping.", "status": StatusUnchanged})
		return
	}

	hash := md5.Sum([]byte(name))
	key := fmt.Sprintf("characters/%v/%s.zip", user, hex.EncodeToString(hash[:]))

	logger.Info("Uploading", zap.String("filename", key))
	contentType := "application/zip"

	// Reading the archive above leaves the offset wherever the zip reader left
	// it; the upload must start from the beginning of the file.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logger.Warn("Error rewinding uploaded file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to upload file"})
		return
	}

	uploader := s3manager.NewUploader(storage.ConnectAws())
	up, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: &contentType,
	})

	if err != nil {
		logger.Warn("Failed to upload to S3", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to upload file", "uploader": up})
		return
	}

	// Recorded only after the object exists, so the database never points at a
	// key we failed to write.
	entry := storage.CharacterEntry{Name: name, Filename: key}
	if err := db.Upsert(email, entry, contentHash); err != nil {
		logger.Warn("Error storing character entry to db", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error storing character entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "All good, move along.", "status": StatusStored})
}

// hashArchiveContents digests what an archive holds rather than its bytes,
// using the per-entry CRC32 already present in the zip headers. No
// decompression, and unaffected by archive comments, entry timestamps or
// compression level.
func hashArchiveContents(r *zip.Reader) string {
	lines := make([]string, 0, len(r.File))
	for _, entry := range r.File {
		lines = append(lines, fmt.Sprintf("%s:%d:%d", entry.Name, entry.CRC32, entry.UncompressedSize64))
	}
	sort.Strings(lines)

	h := sha256.New()
	for _, line := range lines {
		fmt.Fprintln(h, line)
	}

	return hex.EncodeToString(h.Sum(nil))
}
