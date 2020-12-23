package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/logging"
	"go.uber.org/zap"
	"net/http"
)

const MaxAttempts int = 15
const AuthUserKey = "AuthUserKey"
type Authorizer interface {
	IsValid(email string, token string) (bool, error)
}
type Throttler interface {
	GetNumEntries(user string, ip string) (int, error)
	Insert(user string, ip string) error
}

// authorizedHandler ensures that all requests has a valid access token, rejected requests are aborted.
func authorizedHandler(authDb Authorizer, throttleDb Throttler) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "API: Authorization header missing"})
			c.Abort()
			return
		}

		user := c.GetHeader("X-Api-User")
		if user == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "API: X-Api-User header missing"})
			c.Abort()
			return
		}

		ip := c.ClientIP()
		numAttempts, err := throttleDb.GetNumEntries(user, ip)
		if numAttempts >= MaxAttempts || err != nil {
			logger := logging.Logger(c)
			logger.Info("Throttling request due to excess attempts", zap.Error(err), zap.Int("numAttempts", numAttempts), zap.String("user", user), zap.String("ip", ip))
			c.JSON(http.StatusTooManyRequests, gin.H{"msg": "API: Throttled"})
			c.Abort()
			return
		}

		isValid, err := authDb.IsValid(user, token)
		if err != nil {
			logger := logging.Logger(c)
			logger.Info("Error validating auth token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "API: Error validating authorization token"})
			c.Abort()
		} else if !isValid {
			logger := logging.Logger(c)
			logger.Warn("Received an invalid auth token", zap.String("user", user))
			throttleDb.Insert(user, ip)
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "API: Authorization token invalid"})
			c.Abort()
		} else {
			c.Set(AuthUserKey, user)
			c.Next()
		}
	}
}