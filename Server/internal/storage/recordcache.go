package storage

import (
	"context"
	"database/sql"
	"github.com/marmyr/iagdbackup/internal/config"
	"sync"
	"time"
)

// https://stackoverflow.com/questions/36167200/how-safe-are-golang-maps-for-concurrent-read-write-operations

var m = map[string]int64{}
var lock = sync.RWMutex{}

func Write(record string) error {
	lock.Lock()
	defer lock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sql := "INSERT INTO records(record) VALUES (:record)"

	DB := config.GetDatabaseInstance()
	ret, err := DB.NamedExecContext(ctx, sql, map[string]any{"record": record})
	if err != nil {
		return err
	}

	id, err := ret.LastInsertId()
	if err != nil {
		return err
	}

	m[record] = id

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

	DB := config.GetDatabaseInstance()
	it, err := DB.QueryContext(ctx, "SELECT id_record, record FROM records")
	if err != nil {
		return err
	}

	for it.Next() {
		var id int64
		var record string
		if err = it.Scan(&id, &record); err != nil {
			return err
		}

		m[record] = id
	}

	return nil
}
