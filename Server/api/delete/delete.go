package delete

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"github.com/marmyr/iagdbackup/internal/userdb"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/delete"
const Method = routing.DELETE

// Deletes an account and all its items
func ProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)
	userId := routing.GetUser(c)
	email := routing.GetEmail(c)
	userDb := storage.UserDb{}
	var success = true

	itemdb := &storage.ItemDb{}
	err := itemdb.Purge(email)
	if err != nil {
		logger.Warn("Error purging user items", zap.Error(err), zap.Any("user", userId))
		success = false
	}

	characterDb := storage.CharacterDb{}
	characterDb.Purge(email)

	authDb := storage.AuthDb{}
	err = authDb.Purge(userId, email)
	if err != nil {
		logger.Warn("Error purging user auth tokens", zap.Error(err), zap.Any("user", userId))
		success = false
	}

	userDb.Purge(userId)

	// Everything for this user lives in their single .db file; remove it entirely.
	if err := userdb.Remove(email); err != nil {
		logger.Warn("Error removing user database file", zap.Error(err), zap.Any("user", userId))
		success = false
	}

	if success {
		c.JSON(http.StatusOK, gin.H{"msg": "Success"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Something went wrong, deletion may have partially succeeded"})
	}
}

/*
func abc() {

	sess := storage.ConnectAws()
	uploader := s3manager.NewBatchDelete(sess)
	uploader.Delete()
	up, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: &contentType,
	})
}*/
