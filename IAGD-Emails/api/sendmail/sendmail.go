package sendmail

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/gin-gonic/gin"
	"github.com/marmyr/iagdsendmail/internal/eventbus"
	"net/http"
	"strconv"
)

const Path = "/sendmail"
const Method = eventbus.POST


type Input struct {
	Email string `form:"email" json:"email" binding:"required"`
	Code string `form:"code" json:"code" binding:"required"`
}

const (
	Sender = "itemassistant@evilsoft.net"
	Subject = "GD Item Assistant - Backups"
	CharSet = "UTF-8"

	// The HTML body for the email.
	HtmlBody =  "<h1>GD Item Assistant - Access token</h1><p>Your access token for logging into online backups is: %s</p>" +
		"<br><br><small>If you did not request an access token you can safely ignore this e-mail. This token was manually requested by the end-user.</small>"

	//The email body for recipients with non-HTML email clients.
	TextBody = "Your access token for logging into online backups is: %s.\n If you did not request an access token you can safely ignore this e-mail. This token was manually requested by the end-user."
)

func validateInputCode(code string) string {
	if len(code) > 9 {
		return "Invalid token length, maxlen 9"
	}

	if _, err := strconv.Atoi(code); err != nil {
		return "Invalid token format, not a number"
	}

	return ""
}

func ProcessRequest(c *gin.Context) {
	var input Input
	if err := c.BindJSON(&input); err != nil {
		c.Writer.WriteString(err.Error())
	} else {
		if errorMessage := validateInputCode(input.Code); errorMessage != "" {
			c.Status(http.StatusBadRequest)
			c.Writer.WriteString(errorMessage)
		} else {
			fmt.Printf("Sending access token to %s\n", input.Email)
			if sendMail(input.Email, input.Code) != nil {
				c.Status(http.StatusInternalServerError)
			} else {
				c.Status(http.StatusOK)
			}
		}
	}
}


// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/ses-example-send-email.html
func sendMail(recipient string, code string) error {
	// Create a new session in the us-east-1 region.
	sess, err := session.NewSession(&aws.Config{
		Region:aws.String("us-east-1")}, // TODO: Auto detect?
	)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	// Create an SES session.
	svc := ses.New(sess) // TODO: Can this be persisted?

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{ aws.String(recipient), },
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(fmt.Sprintf(HtmlBody, code)),
				},
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(fmt.Sprintf(TextBody, code)),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(Subject),
			},
		},
		Source: aws.String(Sender),
	}

	// Attempt to send the email.
	_, err = svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

		return err
	}

	return nil
}