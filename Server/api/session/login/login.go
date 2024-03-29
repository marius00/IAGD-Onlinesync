package login

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const Path = "/login"
const Method = routing.GET

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type MailProvider func(logger zap.Logger, recipient string, code string) error

var ProcessRequest = ProcessRequestInternal(sendMail)

// Input: GET /login?email=someone@example.com
// Output: JSON {"key": "somevalue"}
// Effect: Email to someone@example.com, pincode stored to DB.
func ProcessRequestInternal(mailProvider MailProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := strings.ToLower(c.Query("email"))

		if !isEmailValid(email) {
			c.JSON(http.StatusBadRequest, gin.H{"msg": `Query parameter "email" does not appear to contain a valid e-mail address`})
			return
		}

		logger := logging.Logger(c)
		throttleDb := storage.ThrottleDb{}

		throttleKey := fmt.Sprintf("sendmail:%s", email)
		throttleIp := fmt.Sprintf("sendmail:%s", c.ClientIP())
		throttled, err := throttleDb.Throttle(throttleKey, throttleIp, 4)
		if err != nil {
			logger.Warn("Error fetching throttle entry", zap.String("user", email), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (throttle)`})
			return
		}
		if throttled {
			logger.Warn("Error user throttled", zap.String("user", email))
			c.JSON(http.StatusTooManyRequests, gin.H{"msg": `Too many attempts, try again later. Much, much later.`})
			return
		}

		db := storage.AuthDb{}

		attempt := storage.AuthAttempt{
			Email: email,
			Key:   uuid.NewV4().String(),
			Code:  generateRandomCode(),
		}

		if err = db.InitiateAuthentication(attempt); err != nil {
			logger.Warn("Error inserting auth attempt entry", zap.String("user", email), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (throttle)`})
			return
		}

		// Send email with the login code
		if err := mailProvider(logger, attempt.Email, attempt.Code); err != nil {
			logger.Warn("Error sending email, initializing user authentication failed")
			c.JSON(http.StatusInternalServerError, gin.H{"msg": `Internal server error (sendmail)`})
			return
		}

		c.JSON(http.StatusOK, gin.H{"key": attempt.Key})
	}
}

func isEmailValid(e string) bool {
	if len(e) < 6 || len(e) > 320 {
		return false
	}

	return emailRegex.MatchString(e)
}

// generateRandomCode generates a random 9 digit pincode
func generateRandomCode() string {
	return fmt.Sprintf("%d", 100000000+rand.Intn(99999999))
}

// init ensures that the random function is seeded at startup, so the pin codes are not generated in a predictable sequence.
func init() {
	rand.Seed(time.Now().Unix())
}
