package storage

import (
	"context"
	"database/sql"
	"github.com/marmyr/iagdbackup/internal/coredb"
	"sync"
	"time"
)

// https://stackoverflow.com/questions/36167200/how-safe-are-golang-maps-for-concurrent-read-write-operations

var m = map[string]int64{}
var mReverse = map[int64]string{}
var lock = sync.RWMutex{}

func Write(record string) error {
	lock.Lock()
	defer lock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	DB, err := coredb.Get()
	if err != nil {
		return err
	}

	ret, err := DB.ExecContext(ctx, "INSERT INTO records(record) VALUES (?)", record)
	if err != nil {
		return err
	}

	id, err := ret.LastInsertId()
	if err != nil {
		return err
	}

	m[record] = id
	mReverse[id] = record

	return nil
}

func ReadRecordId(record string) sql.NullInt64 {
	lock.RLock()
	defer lock.RUnlock()

	v := m[record]
	if v != 0 {
		return sql.NullInt64{Int64: v, Valid: true}
	} else {
		return sql.NullInt64{Valid: false}
	}
}

// ReadRecord resolves a numeric record id back to its string via the in-memory
// cache. Returns "" for a null/zero id or an unknown id. This replaces the
// per-download JOIN against the records table.
func ReadRecord(id sql.NullInt64) string {
	if !id.Valid || id.Int64 == 0 {
		return ""
	}

	lock.RLock()
	defer lock.RUnlock()
	return mReverse[id.Int64]
}

func RecordExists(record string) bool {
	lock.RLock()
	defer lock.RUnlock()
	_, ok := m[record]
	return ok
}

func Preload() error {
	lock.Lock()
	defer lock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	DB, err := coredb.Get()
	if err != nil {
		return err
	}

	it, err := DB.QueryContext(ctx, "SELECT id_record, record FROM records")
	if err != nil {
		return err
	}
	defer it.Close()

	for it.Next() {
		var id int64
		var record string
		if err = it.Scan(&id, &record); err != nil {
			return err
		}

		m[record] = id
		mReverse[id] = record
	}

	return nil
}
