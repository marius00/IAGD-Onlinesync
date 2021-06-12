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

func TestCreateListDeleteItem(t *testing.T) {
	if !testutils.RunAgainstRealDatabase() {
		log.Println("Skipping DB test")
		return
	}
	user := "item@example.com"
	db := ItemDb{}
	userDb := UserDb{}

	existingId, err := userDb.GetByEmail(user)
	if existingId != nil {
		if err := db.Purge(existingId.UserId); err != nil {
			t.Fatal("Failed to purge items for user")
		}
		if err := userDb.Purge(existingId.UserId); err != nil {
			t.Fatal("Failed to purge user entry")
		}
	}

	ts := util.GetCurrentTimestamp()
	userId, err := userDb.Insert(UserEntry{Email: user,})
	if err != nil {
		t.Fatal("Failed to create user")
	}

	defer db.Purge(userId)
	defer userDb.Purge(userId)

	expected := JsonItem{
		Id:         "C11A9D5D-F92F-4079-AC68-C44ED2D36B10",
		UserId:     userId,
		Ts:         ts,
		BaseRecord: "my base record",
	}

	inputItems, _ := db.ToInputItems([]JsonItem{expected})
	FailOnError(t, db.Insert(userId, inputItems), "Error inserting item")

	items, err := db.List(userId, ts-1)
	if err != nil {
		t.Fatalf("Error fetching items %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}

	if items[0].Id != expected.Id || items[0].BaseRecord != expected.BaseRecord || items[0].Ts != expected.Ts {
		t.Fatal("The returned item is not the same as stored to DB")
	}

	FailOnError(t, db.Delete(userId, expected.Id, ts),"Error deleting item")

	deletedItems, err := db.ListDeletedItems(userId, ts-1)
	FailOnError(t, err, "Error fetching deleted items")

	if len(deletedItems) != 1 {
		t.Fatalf("Expected 1 deleted item, got %d", len(items))
	}

	if deletedItems[0].Id != expected.Id {
		t.Fatalf("Expected deleted item id %s, got id %s", expected.Id, deletedItems[0].Id)
	}

}

func TestDoesNotFetchItemInThePast(t *testing.T) {
	if !testutils.RunAgainstRealDatabase() {
		log.Println("Skipping DB test")
		return
	}

	itemDb := ItemDb{}
	userDb := UserDb{}

	ts := util.GetCurrentTimestamp()
	user := fmt.Sprintf("past-item-%s@example.com", uuid.NewV4().String())
	userId := CreateTestUser(t, user)
	defer userDb.Purge(userId)
	defer itemDb.Purge(userId)

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
	itemDb := ItemDb{}
	ts := util.GetCurrentTimestamp()
	user := fmt.Sprintf("insert-twice-%s@example.com", uuid.NewV4().String())
	item := JsonItem{
		Id:         "C11A9D5D-F92F-4079-AC68-C44ED2D36B10",
		Ts:         ts,
		BaseRecord: "base recordddddsssssss",
	}

	userDb := UserDb{}
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

	itemDb.Purge(userId)
}

// TODO: A lot of overhead for this, might be better represented as a cucumber/godog test
func TestInsertSameItemTwiceDifferentBatches(t *testing.T) {
	itemDb := ItemDb{}
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

	userDb := UserDb{}
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

	itemDb.Purge(userId)
}




// Create a clean user for tests
func CreateTestUser(t *testing.T, email string) config.UserId {

	userDb := UserDb{}
	userId, err := userDb.Insert(UserEntry{
		Email:email,
	})
	if err != nil {
		t.Fatalf("Error inserting user, %v", err)
	}

	// Ensure we have no left-over data for this user
	itemDb := ItemDb{}
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
