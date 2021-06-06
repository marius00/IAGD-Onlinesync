package storage

import (
	"fmt"
	"github.com/satori/go.uuid"
	"math/rand"
	"testing"
)

func TestEntireLoginFlow(t *testing.T) {
	db := AuthDb{}

	userDb := UserDb{}
	userId, _ := userDb.Insert(UserEntry{
		Email: fmt.Sprintf("%s@example.com", uuid.NewV4().String()),
	})
	defer userDb.Purge(*userId)

	code := 10000000 + rand.Intn(9999999)
	attempt := AuthAttempt {
		Key: uuid.NewV4().String(),
		Code: fmt.Sprintf("%d", code),
		Email: "auth@example.com",
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
	err = db.StoreSuccessfulAuth(attempt.Email, *userId, attempt.Key, accessToken)
	if err != nil {
		t.Fatalf("Error storing access token: %v", err)
	}

	fetchedUserId, err := db.GetUserId(attempt.Email, accessToken)
	if err != nil {
		t.Fatalf("Error validating access token: %v", err)
	}
	if fetchedUserId <= 0 {
		t.Fatal("Expected access token to be valid")
	}

	if db.Purge(*userId, attempt.Email) != nil {
		t.Fatal("Failed to purge user info")
	}
}

func TestValidateInexistingAccessToken(t *testing.T) {
	db := AuthDb{}

	userId, err := db.GetUserId("noauth@example.com", "invalidtoken")
	if err != nil {
		t.Fatalf("Error validating access token: %v", err)
	}
	if userId > 0 {
		t.Fatal("Expected access token to be invalid")
	}
}

func TestAuthFlowWrongPincode(t *testing.T) {
	userDb := UserDb{}
	email := fmt.Sprintf("%s@example.com", uuid.NewV4().String())
	userId, _ := userDb.Insert(UserEntry{
		Email: email,
	})
	defer userDb.Purge(*userId)

	db := AuthDb{}

	code := "12341234"
	attempt := AuthAttempt {
		Key: uuid.NewV4().String(),
		Code: code,
		Email: email,
	}
	err := db.InitiateAuthentication(attempt)
	if err != nil {
		t.Fatalf("Error initiating attempt: %v", err)
	}

	// API would send an e-mail here

	// User enters wrong code
	fetched, err := db.GetAuthenticationAttempt(attempt.Key, "55554444")
	if err != nil {
		t.Fatalf("Error fetching attempt: %v", err)
	}

	if fetched != nil {
		t.Fatal("Expected no attempt, got one.")
	}

	if db.Purge(*userId, email) != nil {
		t.Fatal("Failed to purge user info")
	}
}

func TestAuthFlowWrongKey(t *testing.T) {
	email := fmt.Sprintf("%s@example.com", uuid.NewV4().String())

	userDb := UserDb{}
	userId, _ := userDb.Insert(UserEntry{
		Email: email,
	})
	defer userDb.Purge(*userId)


	db := AuthDb{}

	code := "12341234"
	attempt := AuthAttempt {
		Key: uuid.NewV4().String(),
		Code: code,
		Email: email,
	}
	err := db.InitiateAuthentication(attempt)
	if err != nil {
		t.Fatalf("Error initiating attempt: %v", err)
	}

	// API would send an e-mail here

	// User sends in wrong key
	fetched, err := db.GetAuthenticationAttempt("not my key", code)
	if err != nil {
		t.Fatalf("Error fetching attempt: %v", err)
	}

	if fetched != nil {
		t.Fatal("Expected no attempt, got one.")
	}

	if db.Purge(*userId, email) != nil {
		t.Fatal("Failed to purge user info")
	}
}

// TODO: Test validate expired?