package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	TableAccessToken = "AccessTokens"
	ColumnEmail = "Email"
	ColumnToken = "Token"
)


type AuthDb struct {
}

// TODO: Have tables be created automagically?
// IsValid checks if an access token is valid for a given user
func (*AuthDb) IsValid(email string, token string) (bool, error) {
	result, err := sess.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(TableAccessToken),
		Key: map[string]*dynamodb.AttributeValue{
			ColumnEmail: {
				S: aws.String(email),
			},
			ColumnToken: {
				S: aws.String(token),
			},
		},
	})

	if err != nil {
		return false, err
	}

	return result.Item != nil, nil
}
