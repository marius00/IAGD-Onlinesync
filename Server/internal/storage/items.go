package storage

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type ItemDb struct {
}

const (
	tableItems = "Items"
	ColumnId = "id"
	ColumnPartition = "partition"
	ColumnTimestamp = "_timestamp"
)

type Item = map[string]interface{}


// Delete will delete a an item for a user
func (*ItemDb) Delete(user string, partition string, id string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			ColumnPartition: {
				S: aws.String(ApplyOwnerS(user, partition)),
			},
			ColumnId: {
				S: aws.String(id),
			},
		},
		TableName: aws.String(tableItems),
	}

	_, err := sess.DeleteItem(input)
	if err != nil {
		return err
	}

	return nil
}


// Fetch all items in a partition for a given user
func (*ItemDb) List(user string, partition string) ([]Item, error) {
	pKey := ApplyOwnerS(user, partition)
	userPrimaryKeyExpr := expression.Key(ColumnPartition).Equal(expression.Value(pKey))

	expr, err := expression.NewBuilder().WithKeyCondition(userPrimaryKeyExpr).Build()
	if err != nil {
		return nil, err
	}

	resp, err := sess.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(tableItems),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})

	if err != nil {
		return nil, err
	}

	var items []Item
	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &items)

	if err != nil {
		return nil, err
	}

	return items, nil
}

/*
func (*ItemDb) StoreSlice(data []map[string]interface{}, tableName string) error {
	items := make([]*dynamodb.TransactWriteItem, len(data))
	for idx, item := range data {
		input := toPut(item, tableName)
		items[idx] = &dynamodb.TransactWriteItem {
			Put: input,
		}
	}

	transactionItems := &dynamodb.TransactWriteItemsInput{
		TransactItems: items,
	}

	output, err := sess.TransactWriteItems(transactionItems)

	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		return err
	}

	output.String()

	fmt.Println("Successfully added input to table " + tableName)
	return nil
}*/

// Store will store arbitrary key:value data as JSON to the specified DynamoDB table
func (*ItemDb) Insert(user string, partition string, data Item) error {
	// Convert the arbitrary data and override partition
	cnv := convertData(data)
	p := ApplyOwnerS(user, partition)
	cnv[ColumnPartition] = &dynamodb.AttributeValue{S: &p,}

	input := &dynamodb.PutItemInput{ // TODO: Would be nice to validate that the partition matches the format "owner:Year:week:it"
		Item:      cnv,
		TableName: aws.String(tableItems),
	}

	_, err := sess.PutItem(input)
	if err != nil {
		return err
	}

	return nil
}

// convertData converts key:value data to a typesafe format DynamoDB can understand
func convertData(data map[string]interface{}) map[string]*dynamodb.AttributeValue {
	var vv = make(map[string]*dynamodb.AttributeValue)
	for k, v := range data {
		if reflect.ValueOf(v).Kind() == reflect.String {
			x := v.(string)
			xx := &(x)
			vv[k] = &dynamodb.AttributeValue{S: xx,}
		} else if reflect.ValueOf(v).Kind() == reflect.Float64 {
			x := fmt.Sprintf("%f", v.(float64))
			xx := &(x)
			vv[k] = &dynamodb.AttributeValue{N: xx,}
		} else if reflect.ValueOf(v).Kind() == reflect.Float32 {
			x := fmt.Sprintf("%f", v.(float32))
			xx := &(x)
			vv[k] = &dynamodb.AttributeValue{N: xx,}
		} else if reflect.ValueOf(v).Kind() == reflect.Int64 || reflect.ValueOf(v).Kind() == reflect.Int32 { // Is this really a use-case? Can JSON props into an int64?
			x := strconv.Itoa(v.(int))
			xx := &(x)
			vv[k] = &dynamodb.AttributeValue{N: xx,}
		} else {
			// TODO: Tests for this
			log.Printf("Unknown type: %s", reflect.ValueOf(v).Kind().String())
		}
	}

	return vv
}
/*
func toPut(data map[string]interface{}, tableName string) *dynamodb.Put {
	params := &dynamodb.Put{
		Item:      convertData(data),
		TableName: aws.String(tableName),
	}

	return params
}
*/

// SanitizePartition will remove the "owner:" prefix from a provided partition
func SanitizePartition(partition string) string {
	idx := strings.Index(partition, ":")
	return partition[idx+1:]
}

// ApplyOwner will append a prefix to the partition-entry to be used for Item Insertions
func ApplyOwner(user string, partition Partition) string {
	return ApplyOwnerS(user, partition.Partition)
}

// ApplyOwner will append a prefix to the partition-entry to be used for Item Insertions
func ApplyOwnerS(user string, partition string) string {
	return fmt.Sprintf("%s:%s", user, partition)
}