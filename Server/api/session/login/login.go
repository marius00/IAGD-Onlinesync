package login

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/myservice/internal/eventbus"
	"github.com/marmyr/myservice/internal/logging"
	"github.com/marmyr/myservice/internal/storage"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"regexp"
	"time"
)

const Path = "/login"
const Method = eventbus.GET
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Input: GET /login?email=someone@example.com
// Output: JSON {"key": "somevalue"}
// Effect: Email to someone@example.com, pincode stored to DB.
func ProcessRequest(c *gin.Context) {
	email := c.Query("email")

	if !isEmailValid(email) {
		c.JSON(http.StatusBadRequest, gin.H{"msg": `Query parameter "email" does not appear to contain a valid e-mail address`})
		return
	}

	logger := logging.Logger(c)
	throttle := storage.ThrottleDb{}

	throttleKey := fmt.Sprintf("sendmail:%s", email)
	numRequests, err := throttle.GetNumEntries(throttleKey, c.Request.RemoteAddr)
	if err != nil {
		logger.Warn("Error fetching throttle entry", zap.String("user", email), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (throttle)`})
		return
	}

	if numRequests > 4 {
		logger.Warn("Error user throttled", zap.String("user", email), zap.Int("numRequests", numRequests))
		c.JSON(http.StatusTooManyRequests, gin.H{"msg": `Too many attempts, try again later. Much, much later.`})
		return
	}

	if throttle.Insert(throttleKey, c.Request.RemoteAddr) != nil {
		logger.Warn("Error inserting throttle entry", zap.String("user", email), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (throttle)`})
		return
	}

	db := storage.AuthDb{}

	attempt := storage.AuthAttempt{
		UserId: email,
		Key: uuid.NewV4().String(),
		Code: generateRandomCode(),
	}

	if err = db.InitiateAuthentication(attempt); err != nil {
		logger.Warn("Error inserting auth attempt entry", zap.String("user", email), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (throttle)`})
		return
	}

	// Send email with the login code
	if err := sendMail(logger, attempt.UserId, attempt.Code); err != nil {
		logger.Warn("Error sending email, initializing user authentication failed")
		c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (sendmail)`})
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": attempt.Key})
}

func isEmailValid(e string) bool {
	if len(e) < 6 || len(e) > 320 {
		return false
	}

	return emailRegex.MatchString(e)
}

// generateRandomCode generates a random 9 digit pincode
func generateRandomCode() string {
	return fmt.Sprintf("%d", 100000000 + rand.Intn(99999999))
}


// init ensures that the random function is seeded at startup, so the pin codes are not generated in a predictable sequence.
func init() {
	rand.Seed(time.Now().Unix())
}