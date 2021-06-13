package storage

import (
	"fmt"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/testutils"
	"github.com/marmyr/iagdbackup/internal/util"
	"github.com/satori/go.uuid"
	"log"
	"testing"
)

var itemDb = ItemDb{}
var userDb = UserDb{}

func TestCreateListDeleteItem(t *testing.T) {
	if !testutils.RunAgainstRealDatabase() {
		log.Println("Skipping DB test")
		return
	}
	user := "item@example.com"

	ts := util.GetCurrentTimestamp()
	userId := CreateTestUser(t, user)
	defer itemDb.Purge(userId)
	defer userDb.Purge(userId)

	expected := JsonItem{
		Id:         "C11A9D5D-F92F-4079-AC68-C44ED2D36B10",
		UserId:     userId,
		Ts:         ts,
		BaseRecord: "my base record",
	}

	inputItems, _ := itemDb.ToInputItems([]JsonItem{expected})
	FailOnError(t, itemDb.Insert(userId, inputItems), "Error inserting item")

	items, err := itemDb.List(userId, ts-1)
	FailOnError(t, err, "Error fetching items")
	testutils.ExpectEquals(t, len(items), 1, "Number of items")
	testutils.ExpectEquals(t, items[0].Id, expected.Id, "The returned item is not the same as stored to DB")
	testutils.ExpectEquals(t, items[0].BaseRecord, expected.BaseRecord, "The returned item is not the same as stored to DB")
	testutils.ExpectEquals(t, items[0].Ts, expected.Ts, "The returned item is not the same as stored to DB")

	FailOnError(t, itemDb.Delete(userId, []string{expected.Id}, ts), "Error deleting item")
	FailOnError(t, itemDb.Delete(userId, []string{expected.Id, "definitely not my id"}, ts), "Error deleting item")

	deletedItems, err := itemDb.ListDeletedItems(userId, ts-1)
	FailOnError(t, err, "Error fetching deleted items")

	testutils.ExpectEquals(t, len(deletedItems), 2, "Number of deleted items")
	testutils.ExpectEquals(t, deletedItems[0].Id, expected.Id, "Deleted item ID")
}

func TestDoesNotFetchItemInThePast(t *testing.T) {
	if !testutils.RunAgainstRealDatabase() {
		log.Println("Skipping DB test")
		return
	}

	ts := util.GetCurrentTimestamp()
	user := fmt.Sprintf("past-item-%s@example.com", uuid.NewV4().String())
	userId := CreateTestUser(t, user)

	item := JsonItem{
		Id:         "C11A9D5D-F92F-4079-AC68-AAAAAAAAAAAA",
		Ts:         ts,
		BaseRecord: "my base record",
	}

	inputItems, _ := itemDb.ToInputItems([]JsonItem{item})
	FailOnError(t, itemDb.Insert(userId, inputItems), "Error inserting item")

	// Same timestamp
	items, err := itemDb.List(userId, ts)
	FailOnError(t, err, "Error fetching items")

	if len(items) != 0 {
		t.Fatalf("Expected 0 item, got %d", len(items))
	}

	// Newer timestamp
	items, err = itemDb.List(userId, ts+1)
	FailOnError(t, err, "Error fetching items")

	if len(items) != 0 {
		t.Fatalf("Expected 0 item, got %d", len(items))
	}
}

func TestInsertSameItemTwice(t *testing.T) {
	ts := util.GetCurrentTimestamp()
	user := fmt.Sprintf("insert-twice-%s@example.com", uuid.NewV4().String())
	item := JsonItem{
		Id:         "C11A9D5D-F92F-4079-AC68-C44ED2D36B10",
		Ts:         ts,
		BaseRecord: "base recordddddsssssss",
	}

	userId := CreateTestUser(t, user)
	defer userDb.Purge(userId)
	defer itemDb.Purge(userId)

	inputItems, _ := itemDb.ToInputItems([]JsonItem{item})
	FailOnError(t, itemDb.Insert(userId, inputItems), "Error inserting item")
	FailOnError(t, itemDb.Insert(userId, inputItems), "Error inserting item")

	items, err := itemDb.List(userId, ts-1)
	FailOnError(t, err, "Error fetching items")

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
}

// TODO: A lot of overhead for this, might be better represented as a cucumber/godog test
func TestInsertSameItemTwiceDifferentBatches(t *testing.T) {
	ts := util.GetCurrentTimestamp()
	user := fmt.Sprintf("insert-twice-mixed-%s@example.com", uuid.NewV4().String())
	itemA := JsonItem{
		Id:         "AAAAAAAA-F92F-4079-AC68-C44ED2D36B10",
		Ts:         ts,
		BaseRecord: "base recordddddsssssss",
	}

	itemB := JsonItem{
		Id:         "BBBBBBBB-F92F-4079-AC68-C44ED2D36B10",
		Ts:         ts,
		BaseRecord: "base recordddddsssssss",
	}

	userId := CreateTestUser(t, user)
	defer userDb.Purge(userId)
	defer itemDb.Purge(userId)

	inputItems, _ := itemDb.ToInputItems([]JsonItem{itemA, itemB})
	FailOnError(t, itemDb.Insert(userId, []InputItem{inputItems[0]}), "Error inserting item")
	FailOnError(t, itemDb.Insert(userId, inputItems), "Error inserting item")

	items, err := itemDb.List(userId, ts-1)
	FailOnError(t, err, "Error fetching items")

	if len(items) != 2 {
		t.Fatalf("Expected 2 item, got %d", len(items))
	}
}

// Create a clean user for tests
func CreateTestUser(t *testing.T, email string) config.UserId {
	userId, err := userDb.Insert(UserEntry{
		Email: email,
	})
	if err != nil {
		t.Fatalf("Error inserting user, %v", err)
	}

	// Ensure we have no left-over data for this user
	if err := itemDb.Purge(userId); err != nil {
		t.Fatal("Failed to purge user")
	}

	return userId
}

func FailOnError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s, %v", message, err)
	}
}
