package download

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/download"
const Method = eventbus.GET

type responseType struct {
	Items   []storage.Item        `json:"items"`
	Deleted []storage.DeletedItem `json:"deleted"`
}

func ProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)
	u, _ := c.Get(eventbus.AuthUserKey)
	user := u.(string)

	partition, ok := c.GetQuery("partition")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "partition" is missing`})
		return
	}
	if !storage.IsValidFormat(partition) {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "partition" is invalid`})
		return
	}

	partitionDb := &storage.PartitionDb{}
	fetched, err := partitionDb.Get(user, partition)
	if err != nil {
		logger.Warn("Error fetching partition", zap.Error(err), zap.String("user", user), zap.String("partition", partition))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching partition"})
		return
	}

	if fetched == nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "Partition does not exist"}) // Client should delete the partition on their end
		return
	}

	itemDb := &storage.ItemDb{}
	items, err := itemDb.List(user, partition)
	if err != nil {
		logger.Warn("Error listing items", zap.Error(err), zap.String("user", user), zap.String("partition", partition))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching items"})
		return
	}

	deletedItemDb := &storage.DeletedItemDb{}
	deleted, err := deletedItemDb.List(user, partition)
	if err != nil {
		logger.Warn("Error listing deleted items", zap.Error(err), zap.String("user", user), zap.String("partition", partition))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching deleted items"})
		return
	}

	r := responseType{
		Items:   items,
		Deleted: deleted,
	}

	c.JSON(http.StatusOK, r)
}
