package storage

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"time"
)

const (
	tablePartitions = "Partitionss" // TODO: Fix this when AWS stops being anal about just having deleted a table
	columnEmail     = "email"
	columnPartition = "partition"
	columnIsActive  = "isActive"
	columnNumItems  = "numItems"
)

type PartitionDb struct {
}

type Partition struct {
	Email     string `json:"email"`         // User/Owner/Email
	Partition string `json:"partition"` // Partition key
	IsActive  bool   `json:"isActive"`  // If this partition is active and accepts new items
	NumItems  int    `json:"numItems"`  // The _estimated_ number of items in this partition, consumer is responsible for updating the value and does not account for race conditions.
}

// TODO: Test
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

	// TODO: Loop all active partitions and set to inactive
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
			// TODO: Deactivate
		}
	}

	return nil
}

// TODO: Test
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

// TODO: Test
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

// TODO: Test
// Delete will delete a given partition entry for a user
func (*PartitionDb) Delete(email string, partition string) error {
	// TODO: Delete entire partition from item table [or delegate to item db? -- delegating may simplify testing]

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
	// TODO: On caller: If itemCount > Permitted, caller should close and create new. [simplifies testing]

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
