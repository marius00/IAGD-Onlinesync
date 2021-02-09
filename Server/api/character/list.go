package character

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

const ListPath = "/character"
const ListMethod = routing.GET

func ListProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)
	u, _ := c.Get(routing.AuthUserKey)
	user := u.(string)

	db := storage.CharacterDb{}
	entries, err := db.List(user)
	if err != nil {
		logger.Warn("Failed fetching character entries", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching characters"})
		return
	}

	c.JSON(http.StatusOK, entries)
}
