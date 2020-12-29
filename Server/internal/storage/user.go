package storage

import (
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"github.com/marmyr/myservice/internal/config"
	"math/rand"
	"time"
)

type UserDb struct {
}

type UserEntry struct {
	UserId    string    `json:"userid" gorm:"column:userid"`
	BuddyId   int32     `json:"buddyId" gorm:"column:buddy_id"`
	CreatedAt time.Time `json:"created_at" sql:"-" gorm:"-"`
}

func (UserEntry) TableName() string {
	return "users"
}

func (*UserDb) Get(user string) (*UserEntry, error) {
	var userEntry UserEntry
	result := config.GetDatabaseInstance().Where("userid = ?", user).Take(&userEntry)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, result.Error
	}

	return &userEntry, result.Error
}

func (*UserDb) GetFromBuddyId(user string) (*UserEntry, error) {
	var userEntry UserEntry
	result := config.GetDatabaseInstance().Where("buddy_id = ?", user).Take(&userEntry)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return &userEntry, result.Error
}

func (*UserDb) Insert(entry UserEntry) error {
	db := config.GetDatabaseInstance()

	// Make up to 8 attempts to store the entry (may conflict on buddy id)
	for i := 0; i < 8; i++ {
		entry.BuddyId = generateBuddyId()
		result := db.Create(entry)

		// Check if its a unique conflict, if so allow retries.
		retry := false
		if result.Error != nil {
			err := result.Error.(*pq.Error)
			if err.Code == UNIQUE_VIOLATION {
				retry = true // Then we're good..
			}
		}

		if !retry {
			return result.Error
		}
	}
	result := db.Create(entry)

	return result.Error
}

func (*UserDb) Purge(user string) error {
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
