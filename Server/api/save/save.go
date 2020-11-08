package save

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/endpoints/utils"
	"github.com/marmyr/myservice/internal/config"
	"github.com/marmyr/myservice/internal/eventbus"
	"net/http"
)

const Path = "/save"
const Method = eventbus.POST


func ProcessRequest(c *gin.Context) {
	data, err := utils.GetJsonData(c)
	if err != nil {
		c.Status(http.StatusBadRequest)
		utils.WriteErrorMessage(c, err.Error())
		return
	}

	storage := &config.PersistentStorage{}
	err = storage.Store(data, config.TableEntries)
	if err != nil {
		c.Status(http.StatusBadRequest)
		utils.WriteErrorMessage(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, nil)
}

func validate(data map[string]interface{}) {

}