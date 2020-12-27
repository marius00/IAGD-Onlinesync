package storage

import (
	"github.com/marmyr/myservice/internal/config"
	"time"
)

type AuthDb struct {
}

type AuthEntry struct {
	UserId string    `json:"-" gorm:"column:userid"`
	Token  string    `json:"-"`
	Ts     time.Time `json:"ts"`
}

func (AuthEntry) TableName() string {
	return "authentry"
}

type AuthAttempt struct {
	Key       string    `json:"key"`
	UserId    string    `json:"-" gorm:"column:userid"`
	Code      string    `json:"-"`
	CreatedAt time.Time `json:"created_at" sql:"-" gorm:"-"`
}

func (AuthAttempt) TableName() string {
	return "authattempt"
}

// IsValid checks if an access token is valid for a given user
func (*AuthDb) IsValid(user string, accessToken string) (bool, error) {
	var sessions []AuthEntry
	result := config.GetDatabaseInstance().Where("userid = ? AND token = ?", user, accessToken).Find(&sessions)

	return len(sessions) > 0, result.Error
}

// InitiateAuthentication initializes an authentication with key/code
func (*AuthDb) InitiateAuthentication(entry AuthAttempt) error {
	db := config.GetDatabaseInstance()
	result := db.Create(&entry)
	return result.Error
}

// Maintenance performs maintenance work such as deleting expired entries
func (*AuthDb) Maintenance() error {
	db := config.GetDatabaseInstance()
	result := db.Where("created_at < NOW() - interval '1 day'").Delete(AuthAttempt{})
	return result.Error
}

// GetAuthenticationAttempt fetches an auth attempt based on key and code
func (*AuthDb) GetAuthenticationAttempt(key string, code string) (*AuthAttempt, error) {
	var attempts []AuthAttempt
	result := config.GetDatabaseInstance().Where("key = ? AND code = ? AND created_at > NOW() - INTERVAL '15 minutes'", key, code).Find(&attempts)

	if len(attempts) > 0 {
		return &attempts[0], result.Error
	}

	return nil, result.Error
}

// StoreSuccessfulAuth stores an access token and deletes the login attempt entry
func (*AuthDb) StoreSuccessfulAuth(user string, key string, authToken string) error {
	db := config.GetDatabaseInstance()
	result := db.Create(&AuthEntry{UserId: user, Token: authToken, Ts: time.Now()})
	if result.Error != nil {
		return result.Error
	}

	if key != "" {
		result = db.Where("userid = ? AND key = ?", user, key).Delete(AuthAttempt{})
	}
	return result.Error
}

// Purge will remove all access tokens and login attempts for the provided user
func (*AuthDb) Purge(user string) error {
	db := config.GetDatabaseInstance()
	result1 := db.Where("userid = ?", user).Delete(AuthAttempt{})
	result2 := db.Where("userid = ?", user).Delete(AuthEntry{})
	if result2.Error == nil {
		return result1.Error
	}
	return result2.Error
}

// Purge will remove all access tokens and login attempts for the provided user
func (*AuthDb) Logout(user string, accessToken string) error {
	db := config.GetDatabaseInstance()
	result := db.Where("userid = ? AND token = ?", user, accessToken).Delete(AuthEntry{})
	return result.Error
}
