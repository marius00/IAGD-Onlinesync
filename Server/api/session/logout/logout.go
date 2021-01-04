package logout

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/logout"
const Method = routing.POST

func ProcessRequest(c *gin.Context) {
	u, _ := c.Get(routing.AuthUserKey)
	user := u.(string)
	token := c.GetHeader("Authorization")

	authDb := storage.AuthDb{}
	err := authDb.Logout(user, token)
	if err != nil {
		logger := logging.Logger(c)
		logger.Warn("Error logging out user", zap.String("user", user), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Internal server error"})
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "Logged out."})
	}
}
