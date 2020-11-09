package logincheck

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"net/http"
)

const Path = "/logincheck"
const Method = eventbus.GET

func ProcessRequest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "Everything went OK"})
}