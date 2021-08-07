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
	jsonItems, err := decode(c.Request.Body)
	if err != nil {
		logger.Info("Error parsing JSON body", zap.Error(err), zap.Any("user", user))
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	// Validate JSON
	if validationError := validate(jsonItems); validationError != "" {
		logger.Info("Error validating JSON body", zap.String("validation", validationError), zap.Any("user", user))
		c.JSON(http.StatusBadRequest, gin.H{"msg": validationError})
		return
	}

	// Store to DB
	db := &storage.ItemDb{}
	inputItems, err := db.ToInputItems(user, jsonItems)
	if err != nil {
		logger.Warn("Unable to fetch item records", zap.Any("user", user), zap.Int("numItems", len(jsonItems)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching item records"})
		return
	}


	for idx := range inputItems {
		inputItems[idx].Ts = timeOfUpload
	}

	if err := db.Insert(user, inputItems); err != nil {
		logger.Warn("Unable to store new item(s)", zap.Error(err), zap.Any("user", user))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Some error storing items, may have partially succeeded"})
		return
	}

	r := responseType{
		Unprocessed: []string{},
	}

	c.JSON(http.StatusOK, r)
}

// validate ensures that the input data is valid-ish
func validate(data []storage.JsonItem) string {
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

		if len(m.BaseRecord) < 6 {
			return fmt.Sprintf(`Item with id="%s" is missing the field "baseRecord"`, m.Id)
		}

		if !HasValidRecords(m) {
			return fmt.Sprintf(`Item with id="%s" has a one or more invalid records`, m.Id)
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


// HasValidRecords will verify that all records on an item are valid ascii (not garbled crap)
func HasValidRecords(item storage.JsonItem) bool {
	records := []string{
		item.BaseRecord, item.PrefixRecord, item.SuffixRecord,
		item.ModifierRecord, item.TransmuteRecord, item.TransmuteRecord,
		item.EnchantmentRecord, item.MateriaRecord,
	}

	for _, record := range records {
		if record != "" {
			if !util.IsASCII(record) || len(record) > 255 {
				return false
			}
		}
	}

	return true
}

func decode(body io.Reader) ([]storage.JsonItem, error) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var entries []storage.JsonItem
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}
