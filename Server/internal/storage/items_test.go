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

	ts := util.GetCurrentTimestamp()
	user := "item@example.com"
	item := JsonItem {
		Id: "C11A9D5D-F92F-4079-AC68-C44ED2D36B10",
		Ts: ts,
		BaseRecord: "my base record",
	}

	db := ItemDb{}
	if err := db.PurgeUser(user); err != nil {
		t.Fatal("Failed to purge user")
	}

	inputItems, _ := db.ToInputItems([]JsonItem {item})
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

	if items[0].Id != item.Id || items[0].BaseRecord != item.BaseRecord || items[0].Ts != item.Ts {
		t.Fatal("The returned item is not the same as stored to DB")
	}

	if err := db.Delete(user, item.Id, ts); err != nil {
		t.Fatalf("Error deleting item %v", err)
	}

	deletedItems, err := db.ListDeletedItems(user, ts-1)
	if err != nil {
		t.Fatalf("Error fetching deleted items %v", err)
	}

	if len(deletedItems) != 1 {
		t.Fatalf("Expected 1 deleted item, got %d", len(items))
	}

	if deletedItems[0].Id != item.Id {
		t.Fatalf("Expected deleted item id %s, got id %s", item.Id, deletedItems[0].Id)
	}
}

func TestDoesNotFetchItemInThePast(t *testing.T) {
	if !testutils.RunAgainstRealDatabase() {
		log.Println("Skipping DB test")
		return
	}

	ts := util.GetCurrentTimestamp()
	user := "past-item@example.com"
	item := JsonItem {
		Id: "C11A9D5D-F92F-4079-AC68-AAAAAAAAAAAA",
		Ts: ts,
		BaseRecord: "my base record",
	}

	db := ItemDb{}
	if err := db.PurgeUser(user); err != nil {
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

	if err := db.PurgeUser(user); err != nil {
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
	defer db.PurgeUser(user)

	if err := db.PurgeUser(user); err != nil {
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
}

