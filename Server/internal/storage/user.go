package storage

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/marmyr/iagdbackup/internal/config"
	"math/rand"
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

// TODO: Test conflict on buddy id
func (*UserDb) Insert(entry UserEntry) (config.UserId, error) {
	db := config.GetDatabaseInstance()

	// Make up to 8 attempts to store the entry (may conflict on buddy id)
	for i := 0; i < 8; i++ {
		entry.BuddyId = generateBuddyId()
		result := db.Create(&entry)

		// Check if its a unique conflict, if so allow retries.
		retry := false
		if result.Error != nil {
			err := result.Error.(*mysql.MySQLError)
			if err.Number == UNIQUE_VIOLATION {
				retry = true // Then we're good..
			}
		}

		if !retry {
			if result.Error == nil && entry.UserId == config.UserId(0) {
				return config.UserId(0), errors.New("Userid not returned")
			}
			return entry.UserId, result.Error
		}
	}
	result := db.Create(entry)

	if result.Error == nil && entry.UserId == config.UserId(0) {
		return config.UserId(0), errors.New("Userid not returned")
	}
	return entry.UserId, result.Error
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
