package getbuddyid

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/buddyId"
const Method = routing.GET

// Get buddy id for the current logged in user
func ProcessRequest(c *gin.Context) {
	user := routing.GetUser(c)

	userDb := storage.UserDb{}
	userEntry, err := userDb.Get(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Could not fetch user`})
		return
	}

	if userEntry == nil {
		_, err := userDb.Insert(storage.UserEntry{UserId: user})
		if err != nil { // TODO: Email?
			logger := logging.Logger(c)
			logger.Warn("Error inserting user entry", zap.Any("user", user), zap.Error(err))
		}

		userEntry, err := userDb.Get(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": `Could not fetch user`})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": userEntry.BuddyId})
	} else {
		c.JSON(http.StatusOK, gin.H{"id": userEntry.BuddyId})
	}
}
