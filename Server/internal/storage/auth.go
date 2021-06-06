package storage

import (
	"github.com/marmyr/iagdbackup/internal/config"
	"time"
)

type AuthDb struct {
}

type AuthEntry struct {
	UserId config.UserId `json:"-" gorm:"column:userid"`
	Email     string    `json:"-" gorm:"column:email"`
	Token  string        `json:"-"`
	Ts     time.Time     `json:"ts"`
}

func (AuthEntry) TableName() string {
	return "authentry"
}

type AuthAttempt struct {
	Key       string    `json:"key"`
	Email     string    `json:"-" gorm:"column:email"`
	Code      string    `json:"-"`
	CreatedAt time.Time `json:"created_at" sql:"-" gorm:"-"`
}

func (AuthAttempt) TableName() string {
	return "authattempt"
}

// IsValid checks if an access token is valid for a given user, returns 0 on invalid user/token combination
func (*AuthDb) GetUserId(email string, accessToken string) (config.UserId, error) {
	var session AuthEntry
	result := config.GetDatabaseInstance().Where("email = ? AND token = ?", email, accessToken).Take(&session)
	if IsNotFoundError(result.Error) {
		return 0, nil
	}

	return session.UserId, result.Error
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
	result := db.Where("created_at < NOW() - interval 1 day").Delete(AuthAttempt{})
	return result.Error
}

// GetAuthenticationAttempt fetches an auth attempt based on key and code
func (*AuthDb) GetAuthenticationAttempt(key string, code string) (*AuthAttempt, error) {
	var attempt AuthAttempt
	result := config.GetDatabaseInstance().Where("`key` = ? AND code = ? AND created_at > NOW() - INTERVAL 15 minute", key, code).Take(&attempt)
	if result.Error != nil {
		if IsNotFoundError(result.Error) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &attempt, result.Error
}

// StoreSuccessfulAuth stores an access token and deletes the login attempt entry
func (*AuthDb) StoreSuccessfulAuth(email string, userId config.UserId, key string, authToken string) error {
	db := config.GetDatabaseInstance()
	result := db.Create(&AuthEntry{UserId: userId, Token: authToken, Ts: time.Now(), Email: email})
	if result.Error != nil {
		return result.Error
	}

	if key != "" {
		result = db.Where("email = ? AND `key` = ?", email, key).Delete(AuthAttempt{})
	}
	return result.Error
}

// Purge will remove all access tokens and login attempts for the provided user
func (*AuthDb) Purge(user config.UserId, email string) error {
	db := config.GetDatabaseInstance()
	result1 := db.Where("email = ?", email).Delete(AuthAttempt{})
	result2 := db.Where("userid = ?", user).Delete(AuthEntry{})
	if result2.Error == nil {
		return result1.Error
	}
	return result2.Error
}

// Purge will remove all access tokens and login attempts for the provided user
func (*AuthDb) Logout(user config.UserId, accessToken string) error {
	db := config.GetDatabaseInstance()
	result := db.Where("userid = ? AND token = ?", user, accessToken).Delete(AuthEntry{})
	return result.Error
}
