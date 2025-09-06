package authstatus

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/status"
const Method = routing.POST

// Input: token=userDefinedUUID
// Output: {"status": "CREATED"}
func ProcessRequest(c *gin.Context) {
	token := c.PostForm("token")
	logger := logging.Logger(c)

	// Verify input args
	if len(token) != 36 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `POST parameter "token" does not appear to contain a valid token`})
		return
	}

	// Verify that the code is correct
	db := storage.AuthDb{}
	attempt, err := db.GetAuthenticationAttemptStatus(token)
	if err != nil {
		logger.Warn("Error fetching auth attempt", zap.String("token", token), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error`})
		return
	}
	if attempt == nil {
		logger.Info("Auth attempt not found", zap.Any("token", token), zap.String("status", "CREATED"))
		c.JSON(http.StatusOK, gin.H{"status": "CREATED", "token": nil, "email": nil})
	} else if attempt.Status != "COMPLETED" {
		logger.Info("Auth attempt in state CREATED", zap.Any("token", token), zap.String("status", "CREATED"))
		c.JSON(http.StatusOK, gin.H{"status": "CREATED", "token": nil, "email": nil})
	} else {
		logger.Info("Auth attempt in state COMPLETED", zap.Any("token", token), zap.String("status", "COMPLETED"))
		c.JSON(http.StatusOK, gin.H{"status": attempt.Status, "token": db.GetLatestAuthToken(attempt.Email), "email": attempt.Email})
	}
}
