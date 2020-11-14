package storage

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type PersistentStorage struct {
}

const (
	TableEntries = "Entries"
	ColumnId = "id"
	ColumnPartition = "partition"
)

func (*PersistentStorage) StoreSlice(data []map[string]interface{}, tableName string) error {
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
}

// TODO: Should maybe take owner+partition as input, and be responsible for ApplyOwner and Sanitize?
// Store will store arbitrary key:value data as JSON to the specified DynamoDB table
func (*PersistentStorage) Store(data map[string]interface{}, tableName string) error {
	input := toPutItemInput(data, tableName)

	output, err := sess.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		return err
	}

	output.String()

	fmt.Println("Successfully added input to table " + tableName)
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

func toPut(data map[string]interface{}, tableName string) *dynamodb.Put {
	params := &dynamodb.Put{
		Item:      convertData(data),
		TableName: aws.String(tableName),
	}

	return params
}

func toPutItemInput(data map[string]interface{}, tableName string) *dynamodb.PutItemInput {
	params := &dynamodb.PutItemInput{ // TODO: Would be nice to validate that the partition matches the format "owner:Year:week:it"
		Item:      convertData(data),
		TableName: aws.String(tableName),
	}

	return params
}

// SanitizePartition will remove the "owner:" prefix from a provided partition
func SanitizePartition(partition string) string {
	idx := strings.Index(partition, ":")
	return partition[idx+1:]
}

// ApplyOwner will append a prefix to the partition-entry to be used for Item Insertions
func ApplyOwner(partition Partition, owner string) string {
	return ApplyOwnerS(partition.Partition, owner)
}
func ApplyOwnerS(partition string, owner string) string {
	return fmt.Sprintf("%s:%s", owner, partition)
}