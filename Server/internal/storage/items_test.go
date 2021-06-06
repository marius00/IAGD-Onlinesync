package storage

import (
	"fmt"
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


	expected := JsonItem {
		Id: "C11A9D5D-F92F-4079-AC68-C44ED2D36B10",
		UserId: userId,
		Ts: ts,
		BaseRecord: "my base record",
	}

	inputItems, _ := db.ToInputItems([]JsonItem {expected})
	if err := db.Insert(userId, inputItems[0]); err != nil {
		t.Fatalf("Error inserting item %v", err)
	}

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

	if err := db.Delete(userId, expected.Id, ts); err != nil {
		t.Fatalf("Error deleting item %v", err)
	}

	deletedItems, err := db.ListDeletedItems(userId, ts-1)
	if err != nil {
		t.Fatalf("Error fetching deleted items %v", err)
	}

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

	userDb := UserDb{}

	ts := util.GetCurrentTimestamp()
	user := fmt.Sprintf("past-item-%s@example.com",  uuid.NewV4().String())
	userId, err := userDb.Insert(UserEntry{Email:user})
	if err != nil {
		t.Fatalf("Error creating user, %v", err)
	}

	defer userDb.Purge(userId)

	item := JsonItem {
		Id: "C11A9D5D-F92F-4079-AC68-AAAAAAAAAAAA",
		Ts: ts,
		BaseRecord: "my base record",
	}

	db := ItemDb{}
	if err := db.Purge(userId); err != nil {
		t.Fatal("Failed to purge user")
	}

	inputItems, _ := db.ToInputItems([]JsonItem {item})
	if err := db.Insert(userId, inputItems[0]); err != nil {
		t.Fatalf("Error inserting item %v", err)
	}

	// Same timestamp
	items, err := db.List(userId, ts)
	if err != nil {
		t.Fatalf("Error fetching items %v", err)
	}

	if len(items) != 0 {
		t.Fatalf("Expected 0 item, got %d", len(items))
	}

	// Newer timestamp
	items, err = db.List(userId, ts+1)
	if err != nil {
		t.Fatalf("Error fetching items %v", err)
	}

	if len(items) != 0 {
		t.Fatalf("Expected 0 item, got %d", len(items))
	}

	if err := db.Purge(userId); err != nil {
		t.Fatal("Failed to purge user")
	}
}

func TestInsertSameItemTwice(t *testing.T) {
	db := ItemDb{}
	ts := util.GetCurrentTimestamp()
	user := fmt.Sprintf("insert-twice-%s@example.com", uuid.NewV4().String())
	item := JsonItem {
		Id: "C11A9D5D-F92F-4079-AC68-C44ED2D36B10",
		Ts: ts,
		BaseRecord: "base recordddddsssssss",
	}

	userDb := UserDb{}
	userId, err := userDb.Insert(UserEntry{
		Email:user,
	})
	if err != nil {
		t.Fatalf("Error inserting user, %v", err)
	}


	defer db.Purge(userId)
	if err := db.Purge(userId); err != nil {
		t.Fatal("Failed to purge user")
	}

	inputItems, _ := db.ToInputItems([]JsonItem {item})
	if err := db.Insert(userId, inputItems[0]); err != nil {
		t.Fatalf("Error inserting item %v", err)
	}

	if err := db.Insert(userId, inputItems[0]); err != nil {
		t.Fatalf("Error inserting item %v", err)
	}

	items, err := db.List(userId, ts-1)
	if err != nil {
		t.Fatalf("Error fetching items %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}

	db.Purge(userId)
	userDb.Purge(userId)
}

