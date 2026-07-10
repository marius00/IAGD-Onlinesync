// Package sqlitedb provides a shared SQLite connection + migration runner used by
// both core.db (shared state) and the per-user .db files.
package sqlitedb

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// Open opens (creating if necessary) a SQLite database at path, applies the
// standard PRAGMAs, and runs any pending migrations found in migrationsFS.
func Open(path string, migrationsFS embed.FS) (*sqlx.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating directory for %s: %w", path, err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", path)
	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}

	// SQLite only supports a single writer at a time; avoid pool contention errors.
	db.SetMaxOpenConns(1)

	if err := migrate(db, migrationsFS); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrating %s: %w", path, err)
	}

	return db, nil
}

// migrate applies any migration files not yet reflected in PRAGMA user_version.
// Migration files must be named NNNN_description.sql (e.g. 0001_init.sql) and are
// applied in ascending numeric order, each inside its own transaction.
func migrate(db *sqlx.DB, migrationsFS embed.FS) error {
	files, err := migrationFiles(migrationsFS)
	if err != nil {
		return err
	}

	var currentVersion int
	if err := db.Get(&currentVersion, "PRAGMA user_version"); err != nil {
		return fmt.Errorf("reading user_version: %w", err)
	}

	for _, f := range files {
		if f.version <= currentVersion {
			continue
		}

		contents, err := fs.ReadFile(migrationsFS, f.path)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", f.path, err)
		}

		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("beginning transaction for migration %s: %w", f.path, err)
		}

		if _, err := tx.Exec(string(contents)); err != nil {
			tx.Rollback()
			return fmt.Errorf("applying migration %s: %w", f.path, err)
		}

		// PRAGMA user_version does not support bind parameters.
		if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", f.version)); err != nil {
			tx.Rollback()
			return fmt.Errorf("bumping user_version for migration %s: %w", f.path, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %s: %w", f.path, err)
		}

		currentVersion = f.version
	}

	return nil
}

type migrationFile struct {
	version int
	path    string
}

// migrationFiles walks the embedded FS (expected to contain files directly under
// "migrations/") and returns them sorted ascending by their numeric prefix.
func migrationFiles(migrationsFS embed.FS) ([]migrationFile, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("reading migrations directory: %w", err)
	}

	var files []migrationFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		parts := strings.SplitN(entry.Name(), "_", 2)
		var version int
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			return nil, fmt.Errorf("migration file %q does not start with a numeric version: %w", entry.Name(), err)
		}

		files = append(files, migrationFile{version: version, path: "migrations/" + entry.Name()})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })
	return files, nil
}
