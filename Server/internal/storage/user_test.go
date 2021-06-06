package storage

import "testing"

func TestUserDb(t *testing.T) {
	db := UserDb{}

	user := "user@example.com"
	entry := UserEntry{ Email:  user }

	userId, err := db.Insert(entry)
	if  err != nil {
		t.Fatalf("Got error %v inserting user entry", err)
	}
	defer db.Purge(*userId)

	u, err := db.Get(*userId)
	if err != nil {
		t.Fatalf("Error fetching user, %v", err)
	}

	if u == nil {
		t.Fatal("Got nil user fetching user")
	}

	if u.Email != user {
		t.Fatalf("Got email %s expected email %s", u.Email, user)
	}

	if err := db.Purge(*userId); err != nil {
		t.Fatalf("Error purging user, %v", err)
	}
}