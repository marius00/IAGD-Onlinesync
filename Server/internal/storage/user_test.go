package storage

import "testing"

func TestUserDb(t *testing.T) {
	db := UserDb{}

	user := "user@example.com"
	entry := UserEntry{ UserId:  user }

	db.Purge(user)
	if err := db.Insert(entry); err != nil {
		t.Fatalf("Got error %v inserting user entry", err)
	}

	u, err := db.Get(user)
	if err != nil {
		t.Fatalf("Error fetching user, %v", err)
	}

	if u == nil {
		t.Fatal("Got nil user fetching user")
	}

	if u.UserId != user {
		t.Fatalf("Got userid %s expected userid %s", u.UserId, user)
	}

	if err := db.Purge(user); err != nil {
		t.Fatalf("Error purging user, %v", err)
	}
}