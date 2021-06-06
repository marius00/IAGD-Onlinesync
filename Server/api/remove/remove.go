package remove

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/storage"
	"github.com/marmyr/iagdbackup/internal/util"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
)

const Path = "/remove"
const Method = routing.POST

type DeleteItemEntry struct {
	ID        string `json:"id"`        // Item GUID
}

func ProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)
	user := routing.GetUser(c)

	entries, err := decode(c.Request.Body)
	if err != nil {
		logger.Info("Error parsing JSON body", zap.Error(err), zap.Any("user", user))
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	if validationError := validate(entries); validationError != "" {
		logger.Info("Error validating JSON body", zap.String("validation", validationError), zap.Any("user", user))
		c.JSON(http.StatusBadRequest, gin.H{"msg": validationError})
		return
	}

	itemDb := storage.ItemDb{}
	var successfulDeletes []DeleteItemEntry

	timeOfRemove := util.GetCurrentTimestamp()
	for _, entry := range entries {
		// The deletion entry is used to ensure other clients deletes it
		if err := itemDb.Delete(user, entry.ID, timeOfRemove); err != nil {
			logger.Warn("Failed to delete item", zap.Error(err), zap.Any("user", user), zap.String("id", entry.ID))
		} else {
			successfulDeletes = append(successfulDeletes, entry)
		}
	}

	c.JSON(http.StatusOK, successfulDeletes)
}

func validate(entries []DeleteItemEntry) string {
	for _, entry := range entries {
		if len(entry.ID) < 32 {
			return `The field "id" must be of length 32 or longer.`
		}
	}

	return ""
}

func decode(body io.Reader) ([]DeleteItemEntry, error) {
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
