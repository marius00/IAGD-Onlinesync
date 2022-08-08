package logincheck

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/routing"
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
		Delete:   3240000,
		Download: 3240000,
		Upload:   3240000,
	}
	multiUsage := LimitEntry{
		Delete:   10000,
		Download: 10000,
		Upload:   1000,
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":        "Logged in and all that good stuff.",
		"regular":    regular,
		"multiUsage": multiUsage,
	})
}
