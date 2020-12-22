package migrate

import (
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"net/http"
)

const Path = "/migrate"
const Method = eventbus.POST

// Migrate a token from Azure to AWS
func ProcessRequest(c *gin.Context) {
	// TODO: Accept token
	// TODO: Verify token (length etc)
	// TODO: Throttle
	// TODO: Ask azure endpoint

	// TODO: Make azure endpoint :D

	// TODO: Return new token + email
	c.JSON(http.StatusInternalServerError, gin.H{"msg": "Not implemented"})
}
// TODO: Endpoint which takes EMAIL and returns a TOKEN (stores token+pin + sends email)
// https://github.com/marius00/IAGD-Onlinesync/blob/master/ItemSync/Items/ValidateEmail.cs

// TODO: Endpoint which verifies pin for token (pin+token => secret, if new player, insert buddy id)
// https://github.com/marius00/IAGD-Onlinesync/blob/master/ItemSync/Items/VerifyEmailToken.cs

// TODO: