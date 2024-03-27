package healthcheck

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/routing"
	"net/http"
)

const Path = "/health"
const Method = routing.GET

func ProcessRequest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}
