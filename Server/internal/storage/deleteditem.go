package storage

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

const (
	tableDeletedItem           = "DeletedItem"
	columnDeletedItemPartition = "partition"
	columnDeletedItemId        = "id"
	columnDeletedItemTimestamp = "timestamp"
)

type DeletedItemDb struct {
}

type DeletedItem struct {
	Partition string
	Id        string
	Timestamp int64
}

// Marks a new item to be deleted upon sync
func (x *DeletedItemDb) Insert(partition Partition, item DeletedItem) error {
	m := map[string]string{
		columnDeletedItemPartition: partition.Partition, // Includes owner/email
		columnDeletedItemId:        item.Id,
		columnDeletedItemTimestamp: fmt.Sprintf("%d", item.Timestamp),
	}

	av, err := dynamodbattribute.MarshalMap(m)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableDeletedItem),
	}

	_, err = sess.PutItem(input)
	if err != nil {
		return err
	}

	return nil
}

// Fetch all items queued to be deleted for a given partition
func (*DeletedItemDb) List(user string, partition string) ([]DeletedItem, error) {
	userPrimaryKeyExpr := expression.Key(columnDeletedItemPartition).Equal(expression.Value(ApplyOwnerS(user, partition)))

	expr, err := expression.NewBuilder().WithKeyCondition(userPrimaryKeyExpr).Build()
	if err != nil {
		return nil, err
	}

	resp, err := sess.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(tableDeletedItem),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})

	if err != nil {
		return nil, err
	}

	var partitionArr []DeletedItem
	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &partitionArr)

	if err != nil {
		return nil, err
	}

	return partitionArr, nil
}
