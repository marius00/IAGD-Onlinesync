package delete

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/delete"
const Method = routing.DELETE

// Deletes an account and all its items
func ProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)
	userId := routing.GetUser(c)
	userDb := storage.UserDb{}
	var success = true
	
	itemdb := &storage.ItemDb{}
	err := itemdb.Purge(userId)
	if err != nil {
		logger.Warn("Error purging user items", zap.Error(err), zap.Any("user", userId))
		success = false
	}

	userEntry, err := userDb.Get(userId) // TODO: Eating err
	if err != nil {
		logger.Error("Error fetching user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Something went wrong, deletion may have partially succeeded"})
		return
	}

	authDb := storage.AuthDb{}
	err = authDb.Purge(userId, userEntry.Email)
	if err != nil {
		logger.Warn("Error purging user auth tokens", zap.Error(err), zap.Any("user", userId))
		success = false
	}


	characterDb := storage.CharacterDb{}

	characterDb.Purge(userId)



	userDb.Purge(userId)

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