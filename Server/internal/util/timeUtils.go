package util

import "time"

// GetCurrentTimestamp returns the current unix time. Exists primarily to avoid accidentally calling UnixNano() or similar
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}
