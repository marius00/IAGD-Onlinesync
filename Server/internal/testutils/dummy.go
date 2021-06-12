package testutils

import "github.com/marmyr/iagdbackup/internal/config"

type DummyAuthorizer struct {}
func (*DummyAuthorizer) GetUserId(email string, token string) (config.UserId, error) {
	if email == "test@example.com" && token == "123456" {
		return 1, nil
	}

	return 0, nil
}

type DummyThrottler struct {}
func (*DummyThrottler) GetNumEntries(user string, ip string) (int, error) {
	return 1, nil
}
func (*DummyThrottler) Insert(user string, ip string) error {
	return nil
}