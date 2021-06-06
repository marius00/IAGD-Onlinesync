package storage

import (
	"github.com/marmyr/iagdbackup/internal/testutils"
	"github.com/marmyr/iagdbackup/internal/util"
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

	if err := db.Purge(user); err != nil {
		t.Fatal("Failed to purge items for user")
	}
	if err := userDb.Purge(user); err != nil {
		t.Fatal("Failed to purge user entry")
	}


	ts := util.GetCurrentTimestamp()
	if err := userDb.Insert(UserEntry{
		UserId: user,
	}); err != nil {
		t.Fatal("Failed to create user")
	}

	defer db.Purge(user)
	defer userDb.Purge(user)


	expected := JsonItem {
		Id: "C11A9D5D-F92F-4079-AC68-C44ED2D36B10",
		UserId: user,
		Ts: ts,
		BaseRecord: "my base record",
	}

	inputItems, _ := db.ToInputItems([]JsonItem {expected})
	if err := db.Insert(user, inputItems[0]); err != nil {
		t.Fatalf("Error inserting item %v", err)
	}

	items, err := db.List(user, ts-1)
	if err != nil {
		t.Fatalf("Error fetching items %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}

	if items[0].Id != expected.Id || items[0].BaseRecord != expected.BaseRecord || items[0].Ts != expected.Ts {
		t.Fatal("The returned item is not the same as stored to DB")
	}

	if err := db.Delete(user, expected.Id, ts); err != nil {
		t.Fatalf("Error deleting item %v", err)
	}

	deletedItems, err := db.ListDeletedItems(user, ts-1)
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
	user := "past-item@example.com"
	userDb.Insert(UserEntry{UserId:user})

	item := JsonItem {
		Id: "C11A9D5D-F92F-4079-AC68-AAAAAAAAAAAA",
		Ts: ts,
		BaseRecord: "my base record",
	}

	db := ItemDb{}
	if err := db.Purge(user); err != nil {
		t.Fatal("Failed to purge user")
	}

	inputItems, _ := db.ToInputItems([]JsonItem {item})
	if err := db.Insert(user, inputItems[0]); err != nil {
		t.Fatalf("Error inserting item %v", err)
	}

	// Same timestamp
	items, err := db.List(user, ts)
	if err != nil {
		t.Fatalf("Error fetching items %v", err)
	}

	if len(items) != 0 {
		t.Fatalf("Expected 0 item, got %d", len(items))
	}

	// Newer timestamp
	items, err = db.List(user, ts+1)
	if err != nil {
		t.Fatalf("Error fetching items %v", err)
	}

	if len(items) != 0 {
		t.Fatalf("Expected 0 item, got %d", len(items))
	}

	if err := db.Purge(user); err != nil {
		t.Fatal("Failed to purge user")
	}
}

func TestInsertSameItemTwice(t *testing.T) {
	db := ItemDb{}
	ts := util.GetCurrentTimestamp()
	user := "insert-twice@example.com"
	item := JsonItem {
		Id: "C11A9D5D-F92F-4079-AC68-C44ED2D36B10",
		Ts: ts,
		BaseRecord: "base recordddddsssssss",
	}
	defer db.Purge(user)

	userDb := UserDb{}
	userDb.Insert(UserEntry{
		UserId:user,
	})
	if err := db.Purge(user); err != nil {
		t.Fatal("Failed to purge user")
	}

	inputItems, _ := db.ToInputItems([]JsonItem {item})
	if err := db.Insert(user, inputItems[0]); err != nil {
		t.Fatalf("Error inserting item %v", err)
	}

	if err := db.Insert(user, inputItems[0]); err != nil {
		t.Fatalf("Error inserting item %v", err)
	}

	items, err := db.List(user, ts-1)
	if err != nil {
		t.Fatalf("Error fetching items %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}

	db.Purge(user)
	userDb.Purge(user)
}

