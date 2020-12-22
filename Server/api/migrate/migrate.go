package migrate

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/migrate"
const Method = eventbus.GET

// Migrate a token from Azure to the new backup system
// Input: GET /migrate?token=tokenInAzure
// Output: JSON {"token": "somevalue", "email":"email@example.com"}
func ProcessRequest(c *gin.Context) {
	token, ok := c.GetQuery("token")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query paramter "token" is missing`})
		return
	}
	if len(token) < 3 || len(token) > 80 { // Len 64 in azure for ~99.99% of the tokens, but [3..80] should be fine.
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query paramter "token" does not appear to contain a valid token`})
		return
	}

	logger := logging.Logger(c)
	throttleDb := storage.ThrottleDb{}
	tokenEntry := fmt.Sprintf("token:%s", token)
	numAttempts, err := throttleDb.GetNumEntries(tokenEntry, c.Request.RemoteAddr)
	if err != nil {
		logger.Warn("Error verifying throttle entry for migration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (throttle)`})
		return
	}
	if numAttempts > 3 {
		logger.Warn("Too many migration attempts", zap.Error(err))
		c.JSON(http.StatusTooManyRequests, gin.H{"msg": `Too many attempts, try again later. Much, much later.`})
		return
	}

	if err := throttleDb.Insert(tokenEntry, c.Request.RemoteAddr); err != nil {
		logger.Warn("Error inserting throttle entry for migration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (throttle)`})
		return
	}


	// TODO: Ask azure endpoint

	// TODO: Make azure endpoint :D

	// TODO: Return new token + email
	c.JSON(http.StatusInternalServerError, gin.H{"msg": "Not implemented"})
}
