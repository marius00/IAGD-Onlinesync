package storage

import (
	"fmt"
	"github.com/marmyr/myservice/endpoints/testutils"
	"log"
	"testing"
	"time"
)

func TestSanitizePartition(t *testing.T) {
	testutils.ExpectEquals(t, "b:c", SanitizePartition("a:b:c"))
	testutils.ExpectEquals(t, "c", SanitizePartition("c"))
	testutils.ExpectEquals(t, "c", SanitizePartition("b:c"))
	testutils.ExpectEquals(t, "a:b:c", SanitizePartition("x:a:b:c"))
}

func TestApplyOwner(t *testing.T) {
	testutils.ExpectEquals(t, "a:b:c", ApplyOwner("a", Partition{Partition: "b:c"}))

	initial := "b:c"
	owner := "owner@example.com"
	combined := ApplyOwner(owner, Partition{Partition: initial})
	testutils.ExpectEquals(t, owner+":"+initial, combined)

	sanitized := SanitizePartition(combined)
	testutils.ExpectEquals(t, initial, sanitized)
}

func TestCreateListDeleteItem(t *testing.T) {
	if !testutils.RunAgainstRealDatabase() {
		log.Println("Skipping DB test again DynamoDb")
		return
	}

	user := "item@example.com"
	p := "2020:15:1"
	id := "C11A9D5D-F92F-4079-AC68-C44ED2D36B10"
	item := map[string]interface{}{
		ColumnId:        id,
		ColumnTimestamp: fmt.Sprintf("%d", time.Now().UnixNano()),
		"stuff":         "fun stuff here",
	}

	db := ItemDb{}
	if err := db.Insert(user, p, item); err != nil {
		t.Fatalf("Error inserting item %v", err)
	}

	items, err := db.List(user, p)
	if err != nil {
		t.Fatalf("Error fetching items %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}

	if items[0][ColumnId] != item[ColumnId] || items[0][ColumnTimestamp] != item[ColumnTimestamp] || items[0]["stuff"] != item["stuff"] {
		t.Fatal("The returned item is not the same as stored to DB")
	}

	if err := db.Delete(user, p, id); err != nil {
		t.Fatalf("Error deleting item %v", err)
	}
}

// TODO: A test which stores an array of real items to determine total size.