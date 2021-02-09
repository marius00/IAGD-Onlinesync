package upload

import (
	"encoding/json"
	"fmt"
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

const Path = "/upload"
const Method = routing.POST

type responseType struct {
	Unprocessed []string `json:"unprocessed"` // Items which remains unprocessed due to errors
}

// Accepts a POST request with a JSON body of format [{}, {}, {}] -- Any fields containing numbers should be sent in as strings
func ProcessRequest(c *gin.Context) {
	timeOfUpload := util.GetCurrentTimestamp()
	logger := logging.Logger(c)
	user := routing.GetUser(c)

	// Parse JSON
	data, err := decode(c.Request.Body)
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

	var unprocessed []string
	numErrors := 0 // If everything is failing, just give up.
	for _, item := range data {
		if numErrors < 5 {
			item.Ts = timeOfUpload
			err = db.Insert(user, item)
		}

		if err != nil || numErrors >= 5 {
			unprocessed = append(unprocessed, item.Id)
			numErrors = numErrors + 1
			logger.Warn("Unable to store new item", zap.Error(err), zap.String("user", user), zap.String("id", item.Id)) // TODO: May get some log spam if this happens.. since err continues to be !=nil
		}
	}

	r := responseType{
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
func validate(data []storage.Item) string {
	for _, m := range data {
		if m.Id == "" {
			return `One or more items is missing the property "id"`
		}

		if len(m.Id) < 32 {
			return `The field "id" must be of length 32 or longer.`
		}

		if m.Ts > 0 {
			return fmt.Sprintf(`Item with id="%s" contains invalid property "_timestamp"`, m.Id)
		}
		if m.UserId != "" {
			return fmt.Sprintf(`Item with id="%s" contains invalid property "User"`, m.Id)
		}
		if len(m.BaseRecord) < 6 {
			return fmt.Sprintf(`Item with id="%s" contains is missing the field "baseRecord"`, m.Id)
		}

		if m.CachedStats == "" {
			return fmt.Sprintf(`Item with id="%s" contains is missing the field "cachedStats"`, m.Id)
		}
		if m.StackCount <= 0 {
			return fmt.Sprintf(`Item with id="%s" has a non-positive stack count`, m.Id)
		}
	}

	if len(data) == 0 {
		return "Input array is empty, no items provided"
	}

	if len(data) > 100 {
		return fmt.Sprintf("Input array contains %d items, maximum is 100", len(data))
	}

	return ""
}


func decode(body io.Reader) ([]storage.Item, error) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var entries []storage.Item
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}
