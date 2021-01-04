package download

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"github.com/marmyr/iagdbackup/internal/util"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/download"
const Method = routing.GET

var ProcessRequest = processRequest(&storage.ItemDb{})

type responseType struct {
	Items     []storage.OutputItem  `json:"items"`
	Removed   []storage.DeletedItem `json:"removed"`
	Timestamp int64                 `json:"timestamp"`
}

type ItemProvider interface {
	List(user string, lastTimestamp int64) ([]storage.OutputItem, error)
	ListDeletedItems(user string, lastTimestamp int64) ([]storage.DeletedItem, error)
}

func processRequest(itemDb ItemProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logging.Logger(c)
		u, _ := c.Get(routing.AuthUserKey)
		user := u.(string)

		currentTimestamp := util.GetCurrentTimestamp()
		lastTimestamp, ok := util.GetTimestamp(c)
		if !ok {
			return
		}

		items, err := itemDb.List(user, lastTimestamp)
		if err != nil {
			logger.Warn("Error listing items", zap.Error(err), zap.String("user", user), zap.Int64("lastTimestamp", lastTimestamp))
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching items"})
			return
		}

		deleted, err := itemDb.ListDeletedItems(user, lastTimestamp)
		if err != nil {
			logger.Warn("Error listing deleted items", zap.Error(err), zap.String("user", user), zap.Int64("lastTimestamp", lastTimestamp))
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching deleted items"})
			return
		}

		r := responseType{
			Items:     items,
			Removed:   deleted,
			Timestamp: currentTimestamp,
		}

		c.JSON(http.StatusOK, r)
	}
}
