package delete

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/delete"
const Method = routing.DELETE

// Deletes an account and all its items
func ProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)
	u, _ := c.Get(routing.AuthUserKey)
	user := u.(string)
	var success = true
	
	itemdb := &storage.ItemDb{}
	err := itemdb.PurgeUser(user)
	if err != nil {
		logger.Warn("Error purging user items", zap.Error(err), zap.String("user", user))
		success = false
	}

	authDb := storage.AuthDb{}
	err = authDb.Purge(user)
	if err != nil {
		logger.Warn("Error purging user auth tokens", zap.Error(err), zap.String("user", user))
		success = false
	}

	userDb := storage.UserDb{}
	userDb.Purge(user)

	if success {
		c.JSON(http.StatusOK, gin.H{"msg": "Success"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Something went wrong, deletion may have partially succeeded"})
	}
}