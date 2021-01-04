package search

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

const Path = "/search"
const Method = routing.GET

// Download buddy items
func ProcessRequest(c *gin.Context) {
	logger := logging.Logger(c)

	// Id for buddy
	buddyId, ok := c.GetQuery("id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "id" is missing`})
		return
	}

	// Offset for scrolling
	offsetStr, ok := c.GetQuery("offset")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "offset" is missing`})
		return
	}
	offset, _ := strconv.ParseInt(offsetStr, 10, 64)


	// Offset for scrolling
	searchText, ok := c.GetQuery("search")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "search" is missing`})
		return
	}



	// Fetch user from buddy id
	userDb := storage.UserDb{}
	user, err := userDb.GetFromBuddyId(buddyId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `Could not find buddy with this id`})
		return
	}

	// Search for items
	itemDb := storage.ItemDb{}
	items, err := itemDb.ListBuddyItems(user.UserId, split(searchText), offset)
	if err != nil {
		logger.Warn("Error listing items", zap.Error(err), zap.String("user", user.UserId))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error fetching items"})
		return
	}

	// Stream to json
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

func split(text string) []string {

	if text == "" {
		return make([]string, 0)
	}
	return strings.SplitN(strings.ReplaceAll(text, "%", "%%"), " ", 8)
}