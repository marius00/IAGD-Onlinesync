package storage

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"strconv"
	"strings"
	"time"
)

const (
	tablePartitions = "Partitionss" // TODO: Fix this when AWS stops being anal about just having deleted a table
	columnEmail     = "email"
	columnPartition = "partition"
	columnIsActive  = "isActive"
	columnNumItems  = "numItems"
)

// Using a struct for namespacing, using a different package name would create a folder nightmare.
type PartitionDb struct {
}

type Partition struct {
	Email     string `json:"email"`         // User/Owner/Email
	Partition string `json:"partition"` // Partition key
	IsActive  bool   `json:"isActive"`  // If this partition is active and accepts new items
	NumItems  int    `json:"numItems"`  // The _estimated_ number of items in this partition, consumer is responsible for updating the value and does not account for race conditions.
}

// Inserts a new partition and marks other partitions as inactive
func (x *PartitionDb) Insert(email string, partition string, numItems int) error {
	p := Partition{
		Email:     email,
		Partition: partition,
		IsActive:  true,
		NumItems:  numItems,
	}

	av, err := dynamodbattribute.MarshalMap(p)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tablePartitions),
	}

	_, err = sess.PutItem(input)
	if err != nil {
		return err
	}

	x.deactivateAllPartitionsExcept(email, partition)
	return nil
}

func (x *PartitionDb) deactivateAllPartitionsExcept(email string, partition string) error {
	partitions, err := x.List(email)
	if err != nil {
		return err
	}

	for _, p := range partitions {
		if p.IsActive && p.Partition != partition {
			x.markInactive(p.Email, p.Partition)
		}
	}

	return nil
}

// markInactive will update a given partition with IsActive=false
func (*PartitionDb) markInactive(email string, partition string) error {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":a": {
				BOOL: aws.Bool(false),
			},
		},
		TableName: aws.String(tablePartitions),
		Key: map[string]*dynamodb.AttributeValue{
			columnEmail: {
				S: aws.String(email),
			},
			columnPartition: {
				S: aws.String(partition),
			},
		},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String(fmt.Sprintf("set %s = :a", columnIsActive)),
	}

	_, err := sess.UpdateItem(input)
	if err != nil {
		return err
	}

	return nil
}

// SetNumItems will update the estimated number of items in a given partition
func (*PartitionDb) SetNumItems(email string, partition string, numItems int) error {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":n": {
				N: aws.String(fmt.Sprintf("%d", numItems)),
			},
		},
		TableName: aws.String(tablePartitions),
		Key: map[string]*dynamodb.AttributeValue{
			columnEmail: {
				S: aws.String(email),
			},
			columnPartition: {
				S: aws.String(partition),
			},
		},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String(fmt.Sprintf("set %s = :n", columnNumItems)),
	}

	_, err := sess.UpdateItem(input)
	if err != nil {
		return err
	}


	return nil
}

// Delete will delete a given partition entry for a user
func (*PartitionDb) Delete(email string, partition string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			columnEmail: {
				S: aws.String(email),
			},
			columnPartition: {
				S: aws.String(partition),
			},
		},
		TableName: aws.String(tablePartitions),
	}

	_, err := sess.DeleteItem(input)
	if err != nil {
		return err
	}

	return nil
}

// Will get the first active partition for a given user, may return nil
func (x *PartitionDb) GetActivePartition(email string) (*Partition, error) {
	partitionArr, err := x.List(email)
	if err != nil {
		return nil, err
	}

	// Return the active one
	for _, partition := range partitionArr {
		if partition.IsActive {
			return &partition, nil
		}
	}

	return nil, nil
}

// Fetch all partitions for a given user
func (*PartitionDb) List(email string) ([]Partition, error) {
	userPrimaryKeyExpr := expression.Key(columnEmail).Equal(expression.Value(email))

	expr, err := expression.NewBuilder().WithKeyCondition(userPrimaryKeyExpr).Build()
	if err != nil {
		return nil, err
	}

	resp, err := sess.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(tablePartitions),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		// FilterExpression: isActiveExpr, // TODO: Consider using this to filter. Zero impact on read units though, so somewhat pointless.
	})

	if err != nil {
		return nil, err
	}

	var partitionArr []Partition
	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &partitionArr)

	if err != nil {
		return nil, err
	}

	return partitionArr, nil
}

// GeneratePartitionKey will generate a partition key for the provided time period and iteration. (Iteration is arbitrary, allowing multiple partitions for a given time period, to prevent them growing too large)
func GeneratePartitionKey(time time.Time, iteration int) string {
	y, w := time.ISOWeek()
	return fmt.Sprintf("%04d:%02d:%02d", y, w, iteration)
}

func ExceedsThreshold(partition *Partition, numItemsToInsert int) bool {
	return partition == nil || partition.NumItems + numItemsToInsert > 1000 // TODO: Should not be hardcoded, figure out a good number. Test to ensure its <1MB
}

func GetIteration(partition Partition) (int, error) {
	idx := strings.LastIndex(partition.Partition, ":")
	if idx != -1 {
		return strconv.Atoi(partition.Partition[idx+1:])
	} else {
		return 0, errors.New("invalid format")
	}
}

// IsValidFormat verifies that an externally provided partition is valid. This will reject internal partitions of user:year:week:idx format
func IsValidFormat(partition string) bool {
	s := strings.Split(partition, ":")
	if len(s) != 3 {
		return false
	}

	year, err := strconv.Atoi(s[0])
	if err != nil || year < 2020 || year > 2050 {
		return false
	}

	week, err := strconv.Atoi(s[1])
	if err != nil || week < 0 || week > 52 {
		return false
	}

	idx, err := strconv.Atoi(s[2])
	if err != nil || idx < 0 {
		return false
	}

	return true
}