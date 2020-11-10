package eventbus

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/logging"
	"go.uber.org/zap"
	"net/http"
)

const AuthUserKey = "AuthUserKey"
type Authorizer interface {
	IsValid(email string, token string) (bool, error)
}

// authorizedHandler ensures that all requests has a valid access token, rejected requests are aborted.
func authorizedHandler(authDb Authorizer) gin.HandlerFunc {
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

		isValid, err := authDb.IsValid(user, token)
		if err != nil {
			logger := logging.Logger(c)
			logger.Info("Error validating auth token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "API: Error validating authorization token"})
			c.Abort()
		} else if !isValid {
			logger := logging.Logger(c)
			logger.Warn("Received an invalid auth token", zap.String("user", user))
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "API: Authorization token invalid"})
			c.Abort()
		} else {
			c.Set(AuthUserKey, user)
			c.Next()
		}
	}
}