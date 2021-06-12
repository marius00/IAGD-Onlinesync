package character

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const DownloadPath = "/character/download"
const DownloadMethod = routing.GET

// Requests a download URL for the provided character name (?name=myCharName)
func DownloadProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)
	user := routing.GetUser(c)

	name, ok := c.GetQuery("name")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "name" is missing. Please provide the character name.`})
		return
	}

	db := storage.CharacterDb{}
	entry, err := db.Get(user, name)
	if err != nil {
		logger.Warn("Failed fetching character entry", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching character"})
		return
	} else if entry == nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "Character not found"})
		return
	}

	sess := storage.ConnectAws()
	svc := s3.New(sess)
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(entry.Filename),
	})
	urlStr, err := req.Presign(5 * time.Minute)

	if err != nil {
		logger.Warn("Failed to sign download URL", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Signing error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "", "url": urlStr})
}
