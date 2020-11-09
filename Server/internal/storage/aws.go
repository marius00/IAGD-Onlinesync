package storage

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var sess *dynamodb.DynamoDB
func ConnectAws() *dynamodb.DynamoDB {
	if sess == nil {
		s := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))
		sess = dynamodb.New(s)

		fmt.Println("Initialized session towards DynamoDB")
	}

	return sess
}

func init() {
	sess = ConnectAws()
}
