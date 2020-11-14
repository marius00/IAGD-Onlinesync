package logincheck

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"net/http"
)

const Path = "/logincheck"
const Method = eventbus.GET

func ProcessRequest(c *gin.Context) {
	// TODO: Should this maybe return limits?
	c.JSON(http.StatusOK, gin.H{"msg": "Logged in and all that good stuff."})
}