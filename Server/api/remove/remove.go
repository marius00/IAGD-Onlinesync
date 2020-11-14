package remove

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"net/http"
)

const Path = "/remove"
const Method = eventbus.POST

func ProcessRequest(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"msg": "Not implemented"})
}