package buddy


import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/routing"
	"github.com/marmyr/myservice/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

const Path = "/buddy"
const Method = routing.GET

var ProcessRequest = processRequest(&storage.ItemDb{})

type ItemProvider interface {
	ListBuddyItems(user string) ([]storage.BuddyItem, error)
}

// Download buddy items
func processRequest(itemDb ItemProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logging.Logger(c)

		buddyId, ok := c.GetQuery("id")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "id" is missing`})
			return
		}

		userDb := storage.UserDb{}
		user, err := userDb.GetFromBuddyId(buddyId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": `Could not find buddy with this id`})
			return
		}

		items, err := itemDb.ListBuddyItems(user.UserId)
		if err != nil {
			logger.Warn("Error listing items", zap.Error(err), zap.String("user", user.UserId))
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching items"})
			return
		}

		c.Status(200)
		c.Writer.WriteString("[")
		for idx, item := range items {
			c.Writer.WriteString(item.CachedStats)
			if idx < len(items)-1 {
				c.Writer.WriteString(",")
			}
		}
		c.Writer.WriteString("]")
	}
}
