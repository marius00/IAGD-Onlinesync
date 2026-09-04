package character

// Upload outcomes, returned as "status" on a 200 so the client can tell an
// accepted backup apart from one that was skipped. A client must only record a
// character as backed up for Stored and Unchanged; Throttled means the archive
// was discarded and should be offered again later.
const (
	StatusStored    = "stored"
	StatusUnchanged = "unchanged"
	StatusThrottled = "throttled"
)
