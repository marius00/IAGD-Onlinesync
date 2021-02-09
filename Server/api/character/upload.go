package character

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"os"
)

const UploadPath = "/character/upload"
const UploadMethod = routing.POST

var region = os.Getenv(config.Region)
var bucket = os.Getenv(config.BucketName)

// Returns an URL where a character backup can be uploaded
func UploadProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)
	user := routing.GetUser(c)

	name, ok := c.GetQuery("name")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "name" is missing. Please provide the character name.`})
		return
	}

	hash :=  md5.Sum([]byte(name))
	key := fmt.Sprintf("characters/%s/%s.zip", user, hex.EncodeToString(hash[:]))
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		logger.Warn("Error receiving/reading uploaded file", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"msg": `Forgot to attach the file? Param: "file"`,})
		return
	}

	// Store to db
	db := storage.CharacterDb{}
	entry := storage.CharacterEntry{
		UserId:   user,
		Name:     name,
		Filename: key,
	}

	if err := db.Insert(entry); err != nil {
		logger.Warn("Error storing character entry to db", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error storing character entry",})
		return
	}

	// Upload to s3
	logger.Info("Uploading", zap.String("filename", key))
	contentType := "application/zip"

	sess := storage.ConnectAws()
	uploader := s3manager.NewUploader(sess)
	up, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: &contentType,
	})

	if err != nil {
		logger.Warn("Failed to upload to S3", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to upload file", "uploader": up,})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "All good, move along."})
}
