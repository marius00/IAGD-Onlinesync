package storage

import (
	"fmt"
	"github.com/marmyr/myservice/endpoints/testutils"
	"log"
	"os"
	"testing"
	"time"
)

func TestGeneratePartitionKeyFirstOfMonth(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	when := time.Date(2018, time.April, 1, 12, 0, 0, 0, loc)
	testutils.ExpectEquals(t, "2018:13:01", GeneratePartitionKey(when, 1))
}

func TestGeneratePartitionKeyStartOfWeek(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	when := time.Date(2018, time.April, 2, 12, 0, 0, 0, loc)
	testutils.ExpectEquals(t, "2018:14:15", GeneratePartitionKey(when, 15))
}

func TestGeneratePartitionKeyExceedingIterations(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	when := time.Date(2018, time.April, 2, 12, 0, 0, 0, loc)
	testutils.ExpectEquals(t, "2018:14:1015", GeneratePartitionKey(when, 1015))
}

func TestExtractIteration(t *testing.T) {
	p := Partition{Partition: "2018:14:1015"}
	it, err := GetIteration(p)
	if err != nil {
		t.Fatal("Expected err to be nil")
	}
	testutils.ExpectEquals(t, "1015", fmt.Sprintf("%d", it))
}

func TestExtractIterationInvalid(t *testing.T) {
	p := Partition{Partition: "2018:14:stuff"}
	it, err := GetIteration(p)
	if err == nil {
		t.Fatal("Expected err to be returned")
	}
	testutils.ExpectEquals(t, "0", fmt.Sprintf("%d", it))
}

func TestEntirePartitionIntegration(t *testing.T) {
	if os.Getenv("WINDIR") != "C:\\WINDOWS" {
		log.Println("Skipping DB test again DynamoDb") // TODO: Get a CI instance up and running
		return
	}
	db := &PartitionDb{}
	email := "testerson@example.com"
	tm := time.Now()
	p := GeneratePartitionKey(tm, 1)
	if err := db.Insert(email, p, 50); err != nil {
		t.Fatalf("%v", err)
	}

	// Test fetch created partition
	{
		fetchedPartition, err := db.GetActivePartition(email)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if fetchedPartition.Partition != p || !fetchedPartition.IsActive || fetchedPartition.NumItems != 50 {
			t.Fatal("Stuff aint right")
		}
	}

	// Test update NumItems
	err := db.SetNumItems(email, p, 123)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Test result of updated NumItems
	{
		fetchedPartition, err := db.GetActivePartition(email)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if fetchedPartition.Partition != p || !fetchedPartition.IsActive || fetchedPartition.NumItems != 123 {
			t.Fatal("Stuff aint right")
		}
	}

	// Test insert new partition [old deactivates]
	p2 := GeneratePartitionKey(tm, 2)
	err = db.Insert(email, p2, 2)
	if err != nil {
		t.Fatalf("%v", err)
	}

	{ // Ensure we have two partitions now
		partitions, err := db.List(email)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if len(partitions) != 2 {
			t.Fatal("Expected 2 partitions")
		}
	}

	// Ensure the new partition is the active one
	{
		activePartition, err := db.GetActivePartition(email)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if activePartition.Partition != p2 {
			t.Fatal("Expected new active partition")
		}
	}

	// Delete the newly created partition
	err = db.Delete(email, p2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	{ // Should not have any active partitions now
		activePartition, err := db.GetActivePartition(email)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if activePartition != nil {
			t.Fatal("Expected no active partition")
		}
	}


	{ // Ensure we have one partitions now that we deleted one
		partitions, err := db.List(email)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if len(partitions) != 1 {
			t.Fatal("Expected 1 partition only")
		}
	}


	err = db.Delete(email, p)
	if err != nil {
		t.Fatalf("%v -- Could not clean up after test", err)
	}
}
