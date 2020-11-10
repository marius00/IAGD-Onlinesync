package partitions

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/partitions"
const Method = eventbus.GET

type DB interface {
	List(email string) ([]storage.Partition, error)
}

// Fetch all partitions for a given user
func processRequest(partitionDb DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, exists := c.Get(eventbus.AuthUserKey); exists {
			partitions, err := partitionDb.List(user.(string))
			if err != nil {
				logger := logging.Logger(c)
				logger.Warn("Error fetching partitions", zap.Error(err), zap.String("user", user.(string)))
				c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching partitions"})

			}
			c.JSON(http.StatusOK, partitions)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error retrieving user credentials"})
		}
	}
}

var ProcessRequest = processRequest(&storage.PartitionDb{})