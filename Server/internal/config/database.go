package config

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"reflect"
	"strconv"
)

type PersistentStorage struct {
}

var Session *dynamodb.DynamoDB // TODO: Don't export?
const (
	TableEntries = "Entries"
)

func init() {
	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	Session = dynamodb.New(s)

	fmt.Println("Initialized session towards DynamoDB")
}



func (*PersistentStorage) StoreSlice(data []map[string]interface{}, tableName string) error {
	items := make([]*dynamodb.TransactWriteItem, len(data))
	for idx, item := range data {
		input := convertToPut(item, tableName)
		items[idx] = &dynamodb.TransactWriteItem {
			Put: input,
		}
	}


	transactionItems := &dynamodb.TransactWriteItemsInput{
		TransactItems: items,
	}

	output, err := Session.TransactWriteItems(transactionItems)

	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		return err
	}

	output.String()

	fmt.Println("Successfully added input to table " + tableName)
	return nil
}

func (*PersistentStorage) Store(data map[string]interface{}, tableName string) error {
	input := convert(data, tableName)

	output, err := Session.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		return err
	}

	output.String()

	fmt.Println("Successfully added input to table " + tableName)
	return nil
}

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
			log.Printf("Unknown type: %s", reflect.ValueOf(v).Kind().String())
		}
	}

	return vv
}

func convertToPut(data map[string]interface{}, tableName string) *dynamodb.Put {
	params := &dynamodb.Put{
		Item:      convertData(data),
		TableName: aws.String(tableName),
	}

	return params
}

func convert(data map[string]interface{}, tableName string) *dynamodb.PutItemInput {
	params := &dynamodb.PutItemInput{
		Item:      convertData(data),
		TableName: aws.String(tableName),
	}

	return params
}