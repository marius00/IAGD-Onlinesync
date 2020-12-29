package storage

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"math/rand"
	"testing"
)

func TestEntireLoginFlow(t *testing.T) {
	db := AuthDb{}

	code := 10000000 + rand.Intn(9999999)
	attempt := AuthAttempt {
		Key: uuid.NewV4().String(),
		Code: fmt.Sprintf("%d", code),
		UserId: "auth@example.com",
	}
	err := db.InitiateAuthentication(attempt)
	if err != nil {
		t.Fatalf("Error initiating attempt: %v", err)
	}

	// API would send an e-mail here

	fetched, err := db.GetAuthenticationAttempt(attempt.Key, attempt.Code)
	if err != nil {
		t.Fatalf("Error fetching attempt: %v", err)
	}

	if fetched == nil {
		t.Fatal("Error fetching attempt, returned nil")
	}

	if fetched.Key != attempt.Key || fetched.Code != attempt.Code {
		t.Fatal("Fetched entry is not the one stored")
	}

	accessToken := uuid.NewV4().String()
	err = db.StoreSuccessfulAuth(attempt.UserId, attempt.Key, accessToken)
	if err != nil {
		t.Fatalf("Error storing access token: %v", err)
	}

	isValid, err := db.IsValid(attempt.UserId, accessToken)
	if err != nil {
		t.Fatalf("Error validating access token: %v", err)
	}
	if !isValid {
		t.Fatal("Expected access token to be valid")
	}

	if db.Purge(attempt.UserId) != nil {
		t.Fatal("Failed to purge user info")
	}
}

func TestValidateInexistingAccessToken(t *testing.T) {
	db := AuthDb{}

	isValid, err := db.IsValid("noauth@example.com", "invalidtoken")
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("Error validating access token: %v", err)
	}
	if isValid {
		t.Fatal("Expected access token to be invalid")
	}
}

func TestAuthFlowWrongPincode(t *testing.T) {
	db := AuthDb{}

	code := "12341234"
	attempt := AuthAttempt {
		Key: uuid.NewV4().String(),
		Code: code,
		UserId: "wrongcode@example.com",
	}
	err := db.InitiateAuthentication(attempt)
	if err != nil {
		t.Fatalf("Error initiating attempt: %v", err)
	}

	// API would send an e-mail here

	// User enters wrong code
	fetched, err := db.GetAuthenticationAttempt(attempt.Key, "55554444")
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("Error fetching attempt: %v", err)
	}

	if fetched != nil {
		t.Fatal("Expected no attempt, got one.")
	}

	if db.Purge(attempt.UserId) != nil {
		t.Fatal("Failed to purge user info")
	}
}

func TestAuthFlowWrongKey(t *testing.T) {
	db := AuthDb{}

	code := "12341234"
	attempt := AuthAttempt {
		Key: uuid.NewV4().String(),
		Code: code,
		UserId: "wrongkey@example.com",
	}
	err := db.InitiateAuthentication(attempt)
	if err != nil {
		t.Fatalf("Error initiating attempt: %v", err)
	}

	// API would send an e-mail here

	// User sends in wrong key
	fetched, err := db.GetAuthenticationAttempt("not my key", code)
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("Error fetching attempt: %v", err)
	}

	if fetched != nil {
		t.Fatal("Expected no attempt, got one.")
	}

	if db.Purge(attempt.UserId) != nil {
		t.Fatal("Failed to purge user info")
	}
}

// TODO: Test validate expired?