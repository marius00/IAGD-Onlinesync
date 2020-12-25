package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/routing"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/storage"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/auth"
const Method = routing.POST

// Input: key=yourToken&code=123123
// Output: {"token": "yourAccessToken", "type": "usertype"}
// Usertype is either NEW or EXISTING (NEW for newly created users)
func ProcessRequest(c *gin.Context) {
	key := c.PostForm("key")
	code := c.PostForm("code")
	logger := logging.Logger(c)
	throttleDb := storage.ThrottleDb{}

	// Verify input args
	if len(key) != 36 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `POST parameter "key" does not appear to contain a valid key`})
		return
	}

	if len(code) != 9 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `POST parameter "code" does not appear to contain a valid pin code`})
		return
	}

	// Handle throttling
	throttleKey := fmt.Sprintf("verifyKey:%s", key)
	authIp := fmt.Sprintf("verifyKey:%s", c.ClientIP())
	throttled, err := throttleDb.Throttle(throttleKey, authIp, 4)
	if err != nil {
		logger.Warn("Error fetching throttle entry", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (throttle)`})
		return
	}
	if throttled {
		logger.Warn("Error user throttled", zap.String("key", key))
		c.JSON(http.StatusTooManyRequests, gin.H{"msg": `Too many attempts, try again later. Much, much later.`})
		return
	}

	// Verify that the code is correct
	db := storage.AuthDb{}
	fetched, err := db.GetAuthenticationAttempt(key, code)
	if err != nil {
		logger.Warn("Error fetching auth attempt", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error`})
		return
	}

	if fetched == nil {
		logger.Warn("Attempted to validate inexisting access key", zap.String("key", key))
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid access token"})
		return
	}

	// Store auth entry
	accessToken := uuid.NewV4().String()
	err = db.StoreSuccessfulAuth(fetched.UserId, fetched.Key, accessToken)
	if err != nil {
		logger.Warn("Error storing auth token", zap.String("key", key), zap.String("user", fetched.UserId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error`})
		return
	}

	// Create a user entry for the user, if one does not exist.
	var existing = "EXISTING"
	userDb := storage.UserDb{}
	u, err := userDb.Get(fetched.UserId)
	if err != nil {
		logger.Warn("Error fetching user entry", zap.String("user", fetched.UserId), zap.Error(err))
	}
	if u == nil {
		// TODO: Check if items exists in Azure?
		existing = "NEW"
		if err := userDb.Insert(storage.UserEntry{UserId: fetched.UserId}); err != nil {
			logger.Warn("Error inserting user entry", zap.String("user", fetched.UserId), zap.Error(err))
		}
	}

	logger.Debug("Login succeeded", zap.String("user", fetched.UserId))
	c.JSON(http.StatusOK, gin.H{"token": accessToken, "usertype": existing}) // TODO: Real usertype
}
