package storage

import (
	"github.com/marmyr/myservice/internal/testutils"
	"log"
	"testing"
)

func TestCreateThrottleEntries(t *testing.T) {
	if !testutils.RunAgainstRealDatabase() {
		log.Println("Skipping DB test")
		return
	}

	user := "throttle@example.com"

	db := ThrottleDb{}
	if err := db.Purge(user, "no ip"); err != nil {
		t.Fatal("Failed to purge throttle")
	}

	if err := db.Insert(user, "my ip"); err != nil {
		t.Fatalf("Error inserting throttle %v", err)
	}

	{ // Test lookup on user
		numEntries, err := db.GetNumEntries(user, "not my ip")
		if err != nil {
			t.Fatalf("Error fetching throttle %v", err)
		}
		if numEntries != 1 {
			t.Fatalf("Expected %d entries, got %d", 1, numEntries)
		}
	}

	{ // Test lookup on IP
		numEntries, err := db.GetNumEntries("notuser@example.com", "my ip")
		if err != nil {
			t.Fatalf("Error fetching throttle %v", err)
		}
		if numEntries != 1 {
			t.Fatalf("Expected %d entries, got %d", 1, numEntries)
		}
	}

	{ // Test lookup on no hits
		numEntries, err := db.GetNumEntries("notuser@example.com", "not my ip")
		if err != nil {
			t.Fatalf("Error fetching throttle %v", err)
		}
		if numEntries != 0 {
			t.Fatalf("Expected %d entries, got %d", 0, numEntries)
		}
	}



	if err := db.Insert(user, "my ip"); err != nil {
		t.Fatalf("Error inserting throttle %v", err)
	}

	if err := db.Insert(user, "my ip"); err != nil {
		t.Fatalf("Error inserting throttle %v", err)
	}

	{ // Test lookup on user
		numEntries, err := db.GetNumEntries(user, "not my ip")
		if err != nil {
			t.Fatalf("Error fetching throttle %v", err)
		}
		if numEntries != 3 {
			t.Fatalf("Expected %d entries, got %d", 3, numEntries)
		}
	}
}
