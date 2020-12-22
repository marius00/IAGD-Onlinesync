package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/storage"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/auth"
const Method = eventbus.POST

// Input: key=yourToken&code=123123
// Output: {"token": "yourAccessToken"}
func ProcessRequest(c *gin.Context) {
	key := c.Query("key")
	code := c.Query("code")
	logger := logging.Logger(c)
	throttle := storage.ThrottleDb{}

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
	numRequests, err := throttle.GetNumEntries(throttleKey, c.Request.RemoteAddr)
	if err != nil {
		logger.Warn("Error fetching throttle entry", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (throttle)`})
		return
	}

	if numRequests > 4 {
		logger.Warn("Error user throttled", zap.String("key", key), zap.Int("numRequests", numRequests))
		c.JSON(http.StatusTooManyRequests, gin.H{"msg": `Too many attempts, try again later. Much, much later.`})
		return
	}

	if err = throttle.Insert(throttleKey, c.Request.RemoteAddr); err != nil {
		logger.Warn("Error inserting throttle entry", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (throttle)`})
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

	logger.Debug("Login succeeded", zap.String("user", fetched.UserId))
	c.JSON(http.StatusOK, gin.H{"token": accessToken})
}
