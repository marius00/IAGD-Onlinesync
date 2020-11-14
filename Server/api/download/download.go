package download

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"net/http"
)

const Path = "/download"
const Method = eventbus.GET

func ProcessRequest(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"msg": "Not implemented"})
}