package storage

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/marmyr/iagdbackup/internal/config"
	"log"
	"math/rand"
	"strings"
	"time"
)

type UserDb struct {
}

type UserEntry struct {
	UserId    config.UserId `json:"-" gorm:"primaryKey; column:userid"`
	Email     string        `json:"-" gorm:"column:email"`
	BuddyId   int32         `json:"buddyId" gorm:"column:buddy_id"`
	CreatedAt time.Time     `json:"created_at" sql:"-" gorm:"-"`
}

func (UserEntry) TableName() string {
	return "users"
}

func (*UserDb) Get(user config.UserId) (*UserEntry, error) {
	var userEntry UserEntry
	result := config.GetDatabaseInstance().Where("userid = ?", user).Take(&userEntry)
	if result.Error != nil {
		if IsNotFoundError(result.Error) {
			return nil, nil
		}

		return nil, result.Error
	}

	return &userEntry, result.Error
}

func (*UserDb) GetByEmail(email string) (*UserEntry, error) {
	if !strings.Contains(email, "@") {
		log.Fatalf("Attempted to fetch user `%s` which is not a valid email", email)
	}

	var userEntry UserEntry
	result := config.GetDatabaseInstance().Where("email = ?", email).Take(&userEntry)
	if result.Error != nil {
		if IsNotFoundError(result.Error) {
			return nil, nil
		}

		return nil, result.Error
	}

	return &userEntry, result.Error
}

func (*UserDb) GetFromBuddyId(buddyId string) (*UserEntry, error) {
	var userEntry UserEntry
	result := config.GetDatabaseInstance().Where("buddy_id = ?", buddyId).Take(&userEntry)
	if result.Error != nil {
		if IsNotFoundError(result.Error) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &userEntry, result.Error
}

func setBuddyId(entry UserEntry) error {
	db := config.GetDatabaseInstance()

	// Make up to N attempts to store the entry (may conflict on buddy id)
	for i := 0; i < 100; i++ {
		entry.BuddyId = generateBuddyId()
		result := db.Model(&entry).Update("buddy_id", entry.BuddyId)

		// Check if its a unique conflict, if so allow retries.
		if result.Error != nil {
			err := result.Error.(*mysql.MySQLError)
			if err.Number != UNIQUE_VIOLATION {
				return err
			}
			// Unique violation, loop re-runs
		} else {
			return nil
		}
	}

	return errors.New("could not produce buddy id for user")
}

// TODO: Test conflict on buddy id
func (*UserDb) Insert(entry UserEntry) (config.UserId, error) {
	db := config.GetDatabaseInstance()

	//https://stackoverflow.com/questions/39333102/how-to-create-or-update-a-record-with-gorm
	//result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&entry)
	entry.BuddyId = generateBuddyId()
	result := db.FirstOrCreate(&entry, UserEntry{
		UserId:  entry.UserId,
		Email:   entry.Email,
	})
	// TODO: Create a conflict resolve or something to insert buddyId != 0
	// https://stackoverflow.com/questions/46321243/how-to-generate-a-unique-random-number-when-insert-in-mysql/46321328
	// Creating the buddy-id random in sql might also be an option? ON INSERT UPDATE, default value, something.

	if result.Error != nil {
		return config.UserId(0), result.Error
	}

	if entry.UserId == config.UserId(0) {
		return config.UserId(0), errors.New("Userid not returned")
	}

	// Attempt to set a buddy id. If this fails, immediately delete the user.
	// If a buddy with id=0 is allowed to exist, bad things will happen.
	// TODO: This will create login errors if two users login at the exact same time
	/*if err := setBuddyId(entry); err != nil {
		db.Where("userid = ?", entry.UserId).Delete(UserEntry{})
		return config.UserId(0), err
	}*/

	return entry.UserId, nil
}

func (*UserDb) Purge(user config.UserId) error {
	db := config.GetDatabaseInstance()
	result := db.Where("userid = ?", user).Delete(UserEntry{})
	return result.Error
}

// init ensures that the random function is seeded at startup, so the pin codes are not generated in a predictable sequence.
func init() {
	rand.Seed(time.Now().Unix())
}

func generateBuddyId() int32 {
	return 100000 + rand.Int31n(99999)
}
