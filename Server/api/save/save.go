package save

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/endpoints/utils"
	"github.com/marmyr/myservice/internal/eventbus"
	"github.com/marmyr/myservice/internal/storage"
	"net/http"
)

const Path = "/save"
const Method = eventbus.POST


func ProcessRequest(c *gin.Context) {
	// Parse JSON
	data, err := utils.GetJsonDataSlice(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	// Validate JSON
	if validationError := validate(data); validationError != "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": validationError})
		return
	}

	// Store to DB
	db := &storage.PersistentStorage{}
	for _, entry := range data {

		err = db.Store(entry, storage.TableEntries)

		// TODO: Better error handling, what if 1/30 fails?
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
			return
		}
	}

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
		if _, ok := m["id"]; !ok {
			return `One or more items is missing the property "id"`
		}

		if _, ok := m["partition"]; ok {
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