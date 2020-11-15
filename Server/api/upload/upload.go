package upload

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/endpoints/utils"
	"github.com/marmyr/myservice/internal/eventbus"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const Path = "/upload"
const Method = eventbus.POST

type responseType struct {
	Partition   string   `json:"partition"`   // Partition items were stored to
	Unprocessed []string `json:"unprocessed"` // Items which remains unprocessed due to errors
}

// Accepts a POST request with a JSON body of format [{}, {}, {}] -- Any fields containing numbers should be sent in as strings
func ProcessRequest(c *gin.Context) {
	t := time.Now().UnixNano()
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
	db := &storage.ItemDb{}
	partitionDb := &storage.PartitionDb{}

	partitionNoPrefix, err := getPartition(partitionDb, user, len(data))
	if err != nil {
		logger.Error("Error validating JSON body", zap.Error(err), zap.String("user", user))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching a valid upload partition"})
		return
	}

	var unprocessed []string
	numErrors := 0 // If everything is failing, just give up.
	for _, entry := range data {
		if numErrors < 5 {
			entry[storage.ColumnTimestamp] = fmt.Sprintf("%d", t)
			err = db.Insert(user, partitionNoPrefix, entry)
		}

		if err != nil || numErrors >= 5 {
			unprocessed = append(unprocessed, entry[storage.ColumnId].(string))
			numErrors = numErrors + 1
			logger.Warn("Unable to store new item", zap.Error(err), zap.String("user", user), zap.String("id", entry[storage.ColumnId].(string)), zap.String("partition", partitionNoPrefix)) // TODO: May get some log spam if this happens.. since err continues to be !=nil
		}
	}

	r := responseType{
		Partition:   partitionNoPrefix,
		Unprocessed: unprocessed,
	}

	if len(unprocessed) == len(data) {
		logger.Warn("Returning 500 internal server error, failed to process all items", zap.String("user", user), zap.Int("numItems", len(data)))
		c.JSON(http.StatusInternalServerError, r)
	} else {
		c.JSON(http.StatusOK, r)
	}
}

// validate ensures that the input data is valid-ish
func validate(data []map[string]interface{}) string {
	for _, m := range data {
		if _, ok := m[storage.ColumnId]; !ok {
			return `One or more items is missing the property "id"`
		}

		if len(m[storage.ColumnId].(string)) < 32 {
			return `The field "id" must be of length 32 or longer.`
		}

		if _, ok := m[storage.ColumnPartition]; ok {
			return fmt.Sprintf(`Item with id="%s" contains invalid property "partition"`, m["id"].(string))
		}

		if _, ok := m[storage.ColumnTimestamp]; ok {
			return fmt.Sprintf(`Item with id="%s" contains invalid property "_timestamp"`, m["id"].(string))
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
