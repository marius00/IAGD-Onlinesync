package storage


import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/marmyr/iagdbackup/internal/config"
	"os"
)

func ConnectAws() *session.Session {
	Region := os.Getenv(config.Region)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(Region)},
	)

	if err != nil {
		panic(err)
	}

	return sess
}