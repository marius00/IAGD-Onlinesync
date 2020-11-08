package utils

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
)

func GetJsonData(c *gin.Context) (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	if e := json.Unmarshal(data, &jsonMap); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
		return nil, e
	}

	return jsonMap, nil
}

func GetJsonDataSlice(c *gin.Context) ([]map[string]interface{}, error) {
	jsonMap := make([]map[string]interface{}, 0)
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	if e := json.Unmarshal(data, &jsonMap); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
		return nil, e
	}

	return jsonMap, nil
}

type ErrorMessage struct {
	Message string `json:"message"`
}

func WriteErrorMessage(c *gin.Context, message string) {
	r, err := json.Marshal(&ErrorMessage{message})
	if err != nil {
		log.Printf("Error serializing error message %v", err)
	} else {
		c.Writer.Write(r)
	}
}