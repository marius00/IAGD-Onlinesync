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


func TestShouldThrottleWhenExceeding(t *testing.T) {
	db := ThrottleDb{}
	user := "throttle-exceed@example.com"
	ip := "127.0.0.1"
	db.Purge(user, ip)
	defer db.Purge(user, ip)

	throttled, err := db.Throttle(user, ip, 1)
	if throttled || err != nil {
		t.Fatalf("Expected success=true and err=nil, got success=%v, err=%v", throttled, err)
	}

	throttled, err = db.Throttle(user, ip, 1)
	if throttled || err != nil {
		t.Fatalf("Expected success=true and err=nil, got success=%v, err=%v", throttled, err)
	}

	throttled, err = db.Throttle(user, ip, 1)
	if !throttled || err != nil {
		t.Fatalf("Expected success=false and err=nil, got success=%v, err=%v", throttled, err)
	}
}

func TestShouldNotThrottleWhenBelowLimit(t *testing.T) {
	db := ThrottleDb{}
	user := "throttle-below@example.com"
	ip := "127.0.0.1"
	db.Purge(user, ip)
	defer db.Purge(user, ip)

	throttled, err := db.Throttle(user, ip, 3)
	if throttled || err != nil {
		t.Fatalf("Expected success=true and err=nil, got success=%v, err=%v", throttled, err)
	}

	throttled, err = db.Throttle(user, ip, 3)
	if throttled || err != nil {
		t.Fatalf("Expected success=true and err=nil, got success=%v, err=%v", throttled, err)
	}

	throttled, err = db.Throttle(user, ip, 3)
	if throttled || err != nil {
		t.Fatalf("Expected success=true and err=nil, got success=%v, err=%v", throttled, err)
	}
}