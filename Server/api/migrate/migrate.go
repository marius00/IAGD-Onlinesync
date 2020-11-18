package migrate

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"net/http"
)

const Path = "/migrate"
const Method = eventbus.POST

// Migrate a token from Azure to AWS
func ProcessRequest(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"msg": "Not implemented"})
}