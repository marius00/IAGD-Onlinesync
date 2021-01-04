package util

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetTimestamp(c *gin.Context) (int64, bool) {
	lastTimestampStr, ok := c.GetQuery("ts")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "ts" is missing`})
		return -1, false
	}

	lastTimestamp, err := strconv.ParseInt(lastTimestampStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query parameter "ts" is not a valid timestamp`})
		return -1, false
	}

	return lastTimestamp, true
}