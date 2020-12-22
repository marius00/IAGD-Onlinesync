package util

import "time"

// GetTimestamp returns the current unix time. Exists primarily to avoid accidentally calling UnixNano() or similar
func GetTimestamp() int64 {
	return time.Now().Unix()
}
