package storage

import (
	"github.com/marmyr/myservice/internal/testutils"
	"log"
	"testing"
	"time"
)

func TestCreateListDeleteItem(t *testing.T) {
	if !testutils.RunAgainstRealDatabase() {
		log.Println("Skipping DB test")
		return
	}

	ts := time.Now().Unix()
	user := "item@example.com"
	item := Item {
		Id: "C11A9D5D-F92F-4079-AC68-C44ED2D36B10",
		Ts: ts,
		BaseRecord: "my base record",
	}

	db := ItemDb{}
	if err := db.PurgeUser(user); err != nil {
		t.Fatal("Failed to purge user")
	}

	if err := db.Insert(user, item); err != nil {
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

	ts := time.Now().Unix()
	user := "past-item@example.com"
	item := Item {
		Id: "C11A9D5D-F92F-4079-AC68-AAAAAAAAAAAA",
		Ts: ts,
		BaseRecord: "my base record",
	}

	db := ItemDb{}
	if err := db.PurgeUser(user); err != nil {
		t.Fatal("Failed to purge user")
	}

	if err := db.Insert(user, item); err != nil {
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
