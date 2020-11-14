package save

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/endpoints/utils"
	"github.com/marmyr/myservice/internal/eventbus"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/save"
const Method = eventbus.POST


func ProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)

	u, exists := c.Get(eventbus.AuthUserKey)
	if !exists {
		logger.Warn("Error parsing user credentials")
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error retrieving user credentials"})
		return
	}
	user := u.(string)

	// Parse JSON
	data, err := utils.GetJsonDataSlice(c.Request.Body)
	if err != nil {
		logger.Info("Error parsing JSON body", zap.Error(err), zap.String("user", user))
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	// Validate JSON
	if validationError := validate(data); validationError != "" {
		logger.Info("Error validating JSON body", zap.String("validation", validationError), zap.String("user", user))
		c.JSON(http.StatusBadRequest, gin.H{"msg": validationError})
		return
	}


	// Store to DB
	db := &storage.PersistentStorage{}
	partitionDb := &storage.PartitionDb{}

	partition, err := getPartition(partitionDb, user, len(data))
	if err != nil {
		logger.Error("Error validating JSON body", zap.Error(err), zap.String("user", user))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching a valid upload partition"})
		return
	}

	// Item table expects partitions to be prefixed with "user:" to avoid needing globally unique partitions across players.
	partitionWithOwner := storage.ApplyOwnerS(user, partition)
	for _, entry := range data {
		entry[storage.ColumnPartition] = partitionWithOwner // TODO: We're not setting Timestamp!
		err = db.Store(entry, storage.TableEntries)

		// TODO: Better error handling, what if 1/30 fails?
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
			return
		}
	}

	// TODO: Return the partition they were stored to -- for each item prolly.
	c.JSON(http.StatusOK, nil)
}


/*
func ProcessRequest(c *gin.Context) {
	data, err := utils.GetJsonData(c)
	if err != nil {
		c.Status(http.StatusBadRequest)
		utils.WriteErrorMessage(c, err.Error())
		return
	}

	storage := &storage.PersistentStorage{}
	err = storage.Store(data, storage.TableEntries)
	if err != nil {
		c.Status(http.StatusBadRequest)
		utils.WriteErrorMessage(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, nil)
}*/

// validate ensures that the input data is valid-ish
func validate(data []map[string]interface{}) string {
	for _, m := range data {
		if _, ok := m[storage.ColumnId]; !ok {
			return `One or more items is missing the property "id"`
		}

		if _, ok := m[storage.ColumnPartition]; ok {
			return fmt.Sprintf(`Item with id="%s" contains invalid property "partition"`, m["id"].(string))
		}
	}

	if len(data) == 0 {
		return "Input array is empty, no items provided"
	}

	// TODO: Reevaluate this once its been decided if batches should be limit to 30 (split jobs for example)
	// DynamoDB max batch size
	if len(data) > 30 {
		return fmt.Sprintf("Input array contains %d items, maximum is 30", len(data))
	}

	return ""
}