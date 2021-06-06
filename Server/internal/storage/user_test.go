package storage

import (
	"fmt"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/satori/go.uuid"
	"testing"
)

func TestUserDb(t *testing.T) {
	db := UserDb{}

	user := fmt.Sprintf("%v-user@example.com", uuid.NewV4().String())
	entry := UserEntry{ Email:  user }

	userId, err := db.Insert(entry)
	if  err != nil {
		t.Fatalf("Got error %v inserting user entry", err)
	}
	if userId == config.UserId(0) {
		// Again not terribly obvious, but when we fix the "gorm:stuff" above, we may fail getting the user id.
		t.Fatalf("Error creating user, userId is 0")
	}
	defer db.Purge(userId)

	u, err := db.Get(userId)
	if err != nil {
		t.Fatalf("Error fetching user, %v", err)
	}

	if u == nil {
		t.Fatal("Got nil user fetching user")
	}

	if u.Email != user {
		t.Fatalf("Got email %s expected email %s", u.Email, user)
	}

	if err := db.Purge(userId); err != nil {
		t.Fatalf("Error purging user, %v", err)
	}
}

func TestObtuseGormStuff(t *testing.T) {
	DB := config.GetDatabaseInstance()

	var users []UserEntry
	result := DB.Find(&users)

	// We seemingly do nothing, but when we change up some "gorm:stuff" on UserEntry, we get: reflect.Value.SetInt using unaddressable value
	if result.Error != nil {
		t.Fatalf("Error fetching users, %v", result.Error)
	}


	userDb := UserDb{}
	userId, err := userDb.Insert(UserEntry{
		Email: fmt.Sprintf("%s@example.com", uuid.NewV4().String()),
	})
	if err != nil {
		t.Fatalf("Error creating user.. %v", err)
	}
	if userId == config.UserId(0) {
		// Again not terribly obvious, but when we fix the "gorm:stuff" above, we may fail getting the user id.
		t.Fatalf("Error creating user, userId is 0")
	}
}

func TestStuff(t *testing.T) {
	userDb := UserDb{}

	email := fmt.Sprintf("%s@example.com", uuid.NewV4().String())
	userId, err := userDb.Insert(UserEntry{
		Email: email,
	})
	if err != nil {
		t.Fatalf("Error creating user, %v", err)
	}
	if userId == config.UserId(0) {
		t.Fatalf("Error creating user, id == 0")
	}

	fetched, err := userDb.GetByEmail(email)
	if err != nil {
		t.Fatalf("Error fetching user, %v", err)
	}

	if fetched.UserId != userId {
		t.Fatalf("Expected userid %v, got userid %v", userId, fetched.UserId)
	}
}