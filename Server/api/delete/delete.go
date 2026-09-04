package delete

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/config"
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

	// The character archives themselves live in S3 under a per-user prefix.
	// Removing them can take a while for a user with many characters, and the
	// caller does not need to wait for it, so it runs detached.
	go purgeCharacterArchives(userId)

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

// purgeCharacterArchives deletes a user's character backups from S3. Runs
// detached from the request, so it logs failures rather than reporting them.
func purgeCharacterArchives(userId config.UserId) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic purging character archives for user %v, %v", userId, r)
		}
	}()

	if err := storage.DeleteS3Prefix(storage.CharacterPrefix(userId)); err != nil {
		log.Printf("Error purging character archives for user %v, %v", userId, err)
	}
}
