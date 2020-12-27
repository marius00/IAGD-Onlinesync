package migrate

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/routing"
	"github.com/marmyr/myservice/internal/storage"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
)

const Path = "/migrate"
const Method = routing.GET

type responseType struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

// Migrate a token from Azure to the new backup system
// Input: GET /migrate?token=tokenInAzure
// Output: JSON {"token": "somevalue", "email":"email@example.com"}
func ProcessRequest(c *gin.Context) {
	token, ok := c.GetQuery("token")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query paramter "token" is missing`})
		return
	}
	if len(token) < 3 || len(token) > 80 { // Len 64 in azure for ~99.99% of the tokens, but [3..80] should be fine.
		c.JSON(http.StatusBadRequest, gin.H{"msg": `The query paramter "token" does not appear to contain a valid token`})
		return
	}

	logger := logging.Logger(c)
	throttleDb := storage.ThrottleDb{}
	tokenEntry := fmt.Sprintf("token:%s", token)
	ipEntry := fmt.Sprintf("token:%s", c.ClientIP())
	throttled, err := throttleDb.Throttle(tokenEntry, ipEntry, 3)
	if err != nil {
		logger.Warn("Error verifying throttle entry for migration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (throttle)`})
		return
	}
	if throttled {
		logger.Warn("Too many migration attempts", zap.Error(err))
		c.JSON(http.StatusTooManyRequests, gin.H{"msg": `Too many attempts, try again later. Much, much later.`})
		return
	}

	req, err := http.NewRequest("POST", "https://iagd.azurewebsites.net/api/Migrate", nil)
	if err != nil {
		logger.Warn("Error creating request against old system", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error`})
		return
	}

	client := &http.Client{}
	req.Header.Set("Simple-Auth", token)
	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("Error executing request against old system", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error`})
		return
	}
	if resp.StatusCode != 200 {
		logger.Warn("Error executing request against old system", zap.Int("statusCode", resp.StatusCode))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error`})
		return
	}

	decoded, err := decode(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		logger.Warn("Error parsing json reply from old system", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error`})
		return
	}

	user := decoded.User

	authDb := storage.AuthDb{}
	userDb := storage.UserDb{}
	if err := userDb.Insert(storage.UserEntry{UserId: user}); err != nil {
		logger.Warn("Error inserting user entry", zap.String("user", user), zap.Error(err))
	}

	accessToken := uuid.NewV4().String()
	err = authDb.StoreSuccessfulAuth(user, "", accessToken)
	if err != nil {
		logger.Warn("Error storing auth token", zap.String("user", user), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error`})
		return
	}

	// Success!
	r := responseType{
		Email: user,
		Token: accessToken,
	}
	c.JSON(http.StatusOK, r)
}

type AzureResponse struct {
	User string `json:"user"`
}

func decode(body io.Reader) (AzureResponse, error) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return AzureResponse{}, err
	}

	var resp AzureResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return AzureResponse{}, err
	}

	return resp, nil
}
