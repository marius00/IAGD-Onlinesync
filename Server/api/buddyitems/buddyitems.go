package buddyitems

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/routing"
	"github.com/marmyr/myservice/internal/storage"
	"github.com/marmyr/myservice/internal/util"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/buddyitems"
const Method = routing.GET

type responseType struct {
	Items     []storage.OutputItem  `json:"items"`
	Removed   []storage.DeletedItem `json:"removed"`
	Timestamp int64                 `json:"timestamp"`
}

// Download buddy items
func ProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)

	// Id for buddy
	buddyId, ok := c.GetQuery("id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "id" is missing`})
		return
	}

	currentTimestamp := util.GetCurrentTimestamp()
	lastTimestamp, ok := util.GetTimestamp(c)
	if !ok {
		return
	}

	// Fetch user from buddy id
	userDb := storage.UserDb{}
	user, err := userDb.GetFromBuddyId(buddyId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `Could not find buddy with this id`})
		return
	}

	itemDb := storage.ItemDb{}
	items, err := itemDb.List(user.UserId, lastTimestamp)
	if err != nil {
		logger.Warn("Error listing items", zap.Error(err), zap.String("user", user.UserId), zap.Int64("lastTimestamp", lastTimestamp))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching items"})
		return
	}

	deleted, err := itemDb.ListDeletedItems(user.UserId, lastTimestamp)
	if err != nil {
		logger.Warn("Error listing deleted items", zap.Error(err), zap.String("user", user.UserId), zap.Int64("lastTimestamp", lastTimestamp))
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
