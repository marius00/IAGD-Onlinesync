package storage

import (
	"context"
	"fmt"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/testutils"
	"github.com/marmyr/iagdbackup/internal/util"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

var itemDb = ItemDb{}
var userDb = UserDb{}

func TestCreateListDeleteItemWithoutAscendantStuff(t *testing.T) {
	email := fmt.Sprintf("%s@example.com", uuid.NewV4().String())

	ts := util.GetCurrentTimestamp()
	userId := CreateTestUser(t, email)
	defer userDb.Purge(userId)
	defer itemDb.Purge(email)
	Preload()

	expected := JsonItem{
		Id:         uuid.NewV4().String(),
		Ts:         ts,
		BaseRecord: "my base record",
	}

	inputItems, _ := itemDb.ToInputItems([]JsonItem{expected})
	err := itemDb.Insert(email, inputItems)
	assert.NoErrorf(t, err, "Error inserting item")

	items, err := itemDb.List(context.Background(), email, ts-1)
	assert.NoErrorf(t, err, "Error listing items")
	assert.Len(t, items, 1, "Expected to list 1 item")
	assert.Equalf(t, expected.Id, items[0].Id, "Expected items to be equal")
	assert.Equalf(t, expected.BaseRecord, items[0].BaseRecord, "Expected items to be equal")
	assert.Equalf(t, "", items[0].AscendantAffixNameRecord, "Expected items to be equal")
	assert.Equalf(t, "", items[0].AscendantAffix2hNameRecord, "Expected items to be equal")
	assert.Equalf(t, int64(0), items[0].RerollsUsed, "Expected items to be equal")
	assert.Equalf(t, int64(0), items[0].AffixRerollsUsed, "Expected items to be equal")
	assert.Equalf(t, expected.Ts, items[0].Ts, "Expected items to be equal")
	assert.Equalf(t, "", items[0].Mod, "Expected no mod to be set")

	FailOnError(t, itemDb.Delete(context.Background(), email, []string{expected.Id}, ts), "Error deleting item")
	FailOnError(t, itemDb.Delete(context.Background(), email, []string{expected.Id, "definitely not my id"}, ts), "Error deleting item")

	deletedItems, err := itemDb.ListDeletedItems(email, ts-1)
	FailOnError(t, err, "Error fetching deleted items")

	assert.Len(t, deletedItems, 1, "Expected 1 item to have been deleted")
	testutils.ExpectEquals(t, deletedItems[0].Id, expected.Id, "Deleted item ID")
}

func TestCreateListDeleteItem(t *testing.T) {
	email := fmt.Sprintf("%s@example.com", uuid.NewV4().String())

	ts := util.GetCurrentTimestamp()
	userId := CreateTestUser(t, email)
	defer userDb.Purge(userId)
	defer itemDb.Purge(email)
	Preload()

	expected := JsonItem{
		Id:                         uuid.NewV4().String(),
		Ts:                         ts,
		BaseRecord:                 "my base record",
		AscendantAffixNameRecord:   "something",
		AscendantAffix2hNameRecord: "something else",
		RerollsUsed:                55,
		AffixRerollsUsed:           9,
	}

	inputItems, _ := itemDb.ToInputItems([]JsonItem{expected})
	err := itemDb.Insert(email, inputItems)
	assert.NoErrorf(t, err, "Error inserting item")

	items, err := itemDb.List(context.Background(), email, ts-1)
	assert.NoErrorf(t, err, "Error listing items")
	assert.Len(t, items, 1, "Expected to list 1 item")
	assert.Equalf(t, expected.Id, items[0].Id, "Expected items to be equal")
	assert.Equalf(t, expected.BaseRecord, items[0].BaseRecord, "Expected items to be equal")
	assert.Equalf(t, expected.AscendantAffixNameRecord, items[0].AscendantAffixNameRecord, "Expected items to be equal")
	assert.Equalf(t, expected.AscendantAffix2hNameRecord, items[0].AscendantAffix2hNameRecord, "Expected items to be equal")
	assert.Equalf(t, expected.RerollsUsed, items[0].RerollsUsed, "Expected items to be equal")
	assert.Equalf(t, expected.AffixRerollsUsed, items[0].AffixRerollsUsed, "Expected items to be equal")
	assert.Equalf(t, expected.Ts, items[0].Ts, "Expected items to be equal")
	assert.Equalf(t, "", items[0].Mod, "Expected no mod to be set")

	FailOnError(t, itemDb.Delete(context.Background(), email, []string{expected.Id}, ts), "Error deleting item")
	FailOnError(t, itemDb.Delete(context.Background(), email, []string{expected.Id, "definitely not my id"}, ts), "Error deleting item")

	deletedItems, err := itemDb.ListDeletedItems(email, ts-1)
	FailOnError(t, err, "Error fetching deleted items")

	assert.Len(t, deletedItems, 1, "Expected 1 item to have been deleted")
	testutils.ExpectEquals(t, deletedItems[0].Id, expected.Id, "Deleted item ID")
}

func TestDoesNotFetchItemInThePast(t *testing.T) {
	if !testutils.RunAgainstRealDatabase() {
		log.Println("Skipping DB test")
		return
	}

	ts := util.GetCurrentTimestamp()
	email := fmt.Sprintf("past-item-%s@example.com", uuid.NewV4().String())
	CreateTestUser(t, email)

	item := JsonItem{
		Id:         "C11A9D5D-F92F-4079-AC68-AAAAAAAAAAAA",
		Ts:         ts,
		BaseRecord: "my base record",
	}

	inputItems, _ := itemDb.ToInputItems([]JsonItem{item})
	FailOnError(t, itemDb.Insert(email, inputItems), "Error inserting item")

	// Same timestamp
	items, err := itemDb.List(context.Background(), email, ts)
	FailOnError(t, err, "Error fetching items")

	if len(items) != 0 {
		t.Fatalf("Expected 0 item, got %d", len(items))
	}

	// Newer timestamp
	items, err = itemDb.List(context.Background(), email, ts+1)
	FailOnError(t, err, "Error fetching items")

	if len(items) != 0 {
		t.Fatalf("Expected 0 item, got %d", len(items))
	}
}

func TestInsertSameItemTwice(t *testing.T) {
	ts := util.GetCurrentTimestamp()
	email := fmt.Sprintf("insert-twice-%s@example.com", uuid.NewV4().String())
	item := JsonItem{
		Id:         "C11A9D5D-F92F-4079-AC68-C44ED2D36B10",
		Ts:         ts,
		BaseRecord: "base recordddddsssssss",
	}

	userId := CreateTestUser(t, email)
	defer userDb.Purge(userId)
	defer itemDb.Purge(email)

	inputItems, _ := itemDb.ToInputItems([]JsonItem{item})
	FailOnError(t, itemDb.Insert(email, inputItems), "Error inserting item")
	FailOnError(t, itemDb.Insert(email, inputItems), "Error inserting item")

	items, err := itemDb.List(context.Background(), email, ts-1)
	FailOnError(t, err, "Error fetching items")

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
}

func TestInsertSameItemTwiceDifferentBatches(t *testing.T) {
	ts := util.GetCurrentTimestamp()
	email := fmt.Sprintf("insert-twice-mixed-%s@example.com", uuid.NewV4().String())
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

	userId := CreateTestUser(t, email)
	defer userDb.Purge(userId)
	defer itemDb.Purge(email)

	inputItems, _ := itemDb.ToInputItems([]JsonItem{itemA, itemB})
	FailOnError(t, itemDb.Insert(email, []InputItem{inputItems[0]}), "Error inserting item")
	FailOnError(t, itemDb.Insert(email, inputItems), "Error inserting item")

	items, err := itemDb.List(context.Background(), email, ts-1)
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
	if err := itemDb.Purge(email); err != nil {
		t.Fatal("Failed to purge user")
	}

	return userId
}

func FailOnError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s, %v", message, err)
	}
}
