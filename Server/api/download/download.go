package download

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

const Path = "/download"
const Method = eventbus.GET

var ProcessRequest = processRequest(&storage.ItemDb{})

type responseType struct {
	Items   []storage.Item        `json:"items"`
	Deleted []storage.DeletedItem `json:"deleted"`
}

type ItemProvider interface {
	List(user string, lastTimestamp int64) ([]storage.Item, error)
	ListDeletedItems(user string, lastTimestamp int64) ([]storage.DeletedItem, error)
}

func processRequest(itemDb ItemProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logging.Logger(c)
		u, _ := c.Get(eventbus.AuthUserKey)
		user := u.(string)

		lastTimestampStr, ok := c.GetQuery("ts")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "ts" is missing`})
			return
		}

		lastTimestamp, err := strconv.ParseInt(lastTimestampStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "ts" is not a valid timestamp`})
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
			Items:   items,
			Deleted: deleted,
		}

		c.JSON(http.StatusOK, r)
	}
}
