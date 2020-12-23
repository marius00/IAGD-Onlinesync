package logincheck

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/routing"
	"net/http"
)

const Path = "/logincheck"
const Method = routing.GET

type LimitEntry struct {
	Delete   int64 `json:"delete"`
	Download int64 `json:"download"`
	Upload   int64 `json:"upload"`
}

func ProcessRequest(c *gin.Context) {
	regular := LimitEntry{
		Delete:   32400000,
		Download: 32400000,
		Upload:   32400000,
	}
	multiUsage := LimitEntry{
		Delete:   3600000,
		Download: 7200000,
		Upload:   10800000,
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":        "Logged in and all that good stuff.",
		"regular":    regular,
		"multiUsage": multiUsage,
	})
}
