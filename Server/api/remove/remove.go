package remove

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/storage"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const Path = "/remove"
const Method = eventbus.POST

type DeleteItemEntry struct {
	Id        string `json:"id"`        // Item GUID
	Partition string `json:"partition"` // Partition key, without owner prefix
}


func ProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)
	u, _ := c.Get(eventbus.AuthUserKey)
	user := u.(string)

	entries, err := decode(c.Request.Body)
	if err != nil {
		logger.Info("Error parsing JSON body", zap.Error(err), zap.String("user", user))
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	if validationError := validate(entries); validationError != "" {
		logger.Info("Error validating JSON body", zap.String("validation", validationError), zap.String("user", user))
		c.JSON(http.StatusBadRequest, gin.H{"msg": validationError})
		return
	}

	// +++
	itemDb := storage.ItemDb{}
	partitionDb := storage.PartitionDb{}
	deletedItemDb := storage.DeletedItemDb{}

	activePartition, err := partitionDb.GetActivePartition(user)
	if err != nil {
		logger.Warn("Error fetching active partition", zap.Error(err), zap.String("user", user))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching active partition"})
		return
	}

	var successfulDeletes []DeleteItemEntry

	t := time.Now().UnixNano()
	for _, entry := range entries {
		// The deletion entry goes in the active partition, to ensure other clients syncs it down.
		var success = true
		if err := deletedItemDb.Insert(*activePartition, storage.DeletedItem{Partition: entry.Partition, Id: entry.Id, Timestamp: t}); err != nil {
			logger.Warn("Failed to insert deletion entry", zap.Error(err), zap.String("user", user), zap.String("id", entry.Id), zap.String("partition", entry.Partition))
			success = false
		}

		// Delete the item from the ItemDB
		if err := itemDb.Delete(user, storage.ApplyOwnerS(user, entry.Partition), entry.Id); err != nil {
			logger.Warn("Failed to delete item", zap.Error(err), zap.String("user", user), zap.String("id", entry.Id), zap.String("partition", entry.Partition))
			success = false
		}

		// If we actually managed to delete it..
		if success {
			successfulDeletes = append(successfulDeletes, entry)
		}
	}

	// TODO: Insert deletion entry
	// TODO: Remove items from itemDb
	// TODO: Update partition size?
	// TODO: If partitionEmpty => Delete partition?


	// TODO: Return
	c.JSON(http.StatusOK, successfulDeletes)
}

func validate(entries []DeleteItemEntry) string {
	for _, entry := range entries {
		if len(entry.Id) < 32 {
			return `The field "id" must be of length 32 or longer.`
		}

		if !storage.IsValidFormat(entry.Partition) {
			return `The field "partition" is of an invalid format"`
		}

		// TODO: Validate that the entries are sorted?
	}

	return ""
}

func decode(body io.ReadCloser) ([]DeleteItemEntry, error) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var entries []DeleteItemEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}
