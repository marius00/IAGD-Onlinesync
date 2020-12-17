package storage

import (
	"github.com/marmyr/myservice/internal/config"
)

type AuthDb struct {
}

type AuthEntry struct {
	UserId  string `json:"userid"`
	Token string `json:"-"`
	Ts    int64  `json:"ts"`
}

// IsValid checks if an access token is valid for a given user
func (*AuthDb) IsValid(user string, token string) (bool, error) {
	var sessions []AuthEntry
	result := config.GetDatabaseInstance().Where("userid = ? AND token = ?", user, token).Find(&sessions)

	return len(sessions) > 0, result.Error
}
