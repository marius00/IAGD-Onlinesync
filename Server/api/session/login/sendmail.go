package login

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"go.uber.org/zap"
)

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

// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/ses-example-send-email.html
func sendMail(logger zap.Logger, recipient string, code string) error {
	// Create a new session in the us-east-1 region.
	sess, err := session.NewSession(&aws.Config{
		Region:aws.String("us-east-1")}, // TODO: Auto detect?
	)

	if err != nil {
		logger.Warn("Error connecting to SES", zap.Error(err))
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
			logger.Warn("Error sending email", zap.String("errorCode", aerr.Code()), zap.String("error", aerr.Error()), zap.Error(err))
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			logger.Warn("Error sending email", zap.Error(err))
		}

		return err
	}

	return nil
}