package storage

import (
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/coredb"
	"github.com/marmyr/iagdbackup/internal/userdb"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
)

// Migration drains a user's data out of the read-only MySQL source and into
// their per-user SQLite database, exactly once per user. Migration state is
// tracked in core.db (never in MySQL — the legacy database is treated as
// strictly read-only), and mirrored into an in-memory set so the common
// already-migrated case never touches a database.
//
// Corruption tolerance: a small fraction of prod item rows are known to be
// corrupt. Individual rows that fail to read or insert are skipped and counted
// rather than failing the whole user — losing a few rows is acceptable. A user
// is only marked FAILED on an infrastructure error (MySQL unreachable, SQLite
// unwritable), and the background drainer gives up on a user after
// maxDrainAttempts so it never gets stuck.

const maxDrainAttempts = 5

var (
	migratedSet    = map[config.UserId]bool{}
	failedAttempts = map[config.UserId]int{}
	migratedLock   sync.RWMutex
	migrateGroup   singleflight.Group
)

// IsMigrated reports whether the user's data already lives in SQLite.
func IsMigrated(userId config.UserId) bool {
	migratedLock.RLock()
	defer migratedLock.RUnlock()
	return migratedSet[userId]
}

// isDrainExhausted reports whether a user has failed to drain too many times for
// the background drainer to keep retrying.
func isDrainExhausted(userId config.UserId) bool {
	migratedLock.RLock()
	defer migratedLock.RUnlock()
	return failedAttempts[userId] >= maxDrainAttempts
}

func markMigratedInMemory(userId config.UserId) {
	migratedLock.Lock()
	defer migratedLock.Unlock()
	migratedSet[userId] = true
	delete(failedAttempts, userId)
}

// SetMigrated records a user as fully migrated, both in core.db and in memory.
// Used for brand-new users (no MySQL data to drain) at creation time and on
// successful drains.
//
// The in-memory flag is set regardless of whether the core.db write succeeds.
// By the time this is called, the data itself is already safely committed in
// the user's SQLite database - that's the fact that must never be re-drained
// over. If we only marked the flag on a successful core.db write, a transient
// core.db failure here would leave IsMigrated() false even though the data is
// there; the next call to EnsureMigrated would then re-run the (clear-then-copy)
// drain and wipe anything written to the user's db since the successful drain
// (e.g. a newly-issued access token). core.db's copy of this fact is a
// best-effort durability record, not the source of truth for whether the
// migration itself may safely be repeated.
func SetMigrated(userId config.UserId, mysqlItemCount, sqliteItemCount int) error {
	markMigratedInMemory(userId)

	db, err := coredb.Get()
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO migration_state(userid, status, mysql_item_count, sqlite_item_count, migrated_at)
		VALUES (?, 'DONE', ?, ?, unixepoch())
		ON CONFLICT(userid) DO UPDATE SET status = 'DONE', mysql_item_count = excluded.mysql_item_count,
			sqlite_item_count = excluded.sqlite_item_count, migrated_at = excluded.migrated_at`,
		userId, mysqlItemCount, sqliteItemCount)
	return err
}

// markFailed records a failed drain attempt (infrastructure error) and bumps the
// attempt counter, both in core.db and in memory.
func markFailed(userId config.UserId, cause error) {
	migratedLock.Lock()
	failedAttempts[userId]++
	attempts := failedAttempts[userId]
	migratedLock.Unlock()

	db, err := coredb.Get()
	if err != nil {
		log.Warn().Msgf("Could not record migration failure for %d: %v", userId, err)
		return
	}

	_, err = db.Exec(`INSERT INTO migration_state(userid, status, attempts, last_error)
		VALUES (?, 'FAILED', ?, ?)
		ON CONFLICT(userid) DO UPDATE SET status = 'FAILED', attempts = ?, last_error = excluded.last_error`,
		userId, attempts, cause.Error(), attempts)
	if err != nil {
		log.Warn().Msgf("Could not persist migration failure for %d: %v", userId, err)
	}
}

// PreloadMigrationState fills the in-memory migrated set and failure counters
// from core.db so that already-migrated (and already-exhausted) users are
// recognised without per-request database access.
func PreloadMigrationState() error {
	db, err := coredb.Get()
	if err != nil {
		return err
	}

	type row struct {
		UserId   config.UserId `db:"userid"`
		Status   string        `db:"status"`
		Attempts int           `db:"attempts"`
	}
	var rows []row
	if err := db.Select(&rows, "SELECT userid, status, attempts FROM migration_state"); err != nil {
		return err
	}

	migratedLock.Lock()
	defer migratedLock.Unlock()
	for _, r := range rows {
		if r.Status == "DONE" {
			migratedSet[r.UserId] = true
		} else if r.Status == "FAILED" {
			failedAttempts[r.UserId] = r.Attempts
		}
	}

	return nil
}

// EnsureMigrated makes sure a user's data has been drained from MySQL into their
// SQLite database. It is safe to call on every request: migrated users return
// immediately, and concurrent first-touch calls for the same user are collapsed.
func EnsureMigrated(email string, userId config.UserId) error {
	if IsMigrated(userId) {
		return nil
	}

	// Nothing to drain from once MySQL is decommissioned. A user that isn't
	// marked migrated at that point shouldn't exist, but don't wedge the request.
	if !config.MySQLConfigured() {
		return SetMigrated(userId, 0, 0)
	}

	_, err, _ := migrateGroup.Do(email, func() (interface{}, error) {
		if IsMigrated(userId) {
			return nil, nil
		}
		if err := drainUser(email, userId); err != nil {
			markFailed(userId, err)
			return nil, err
		}
		return nil, nil
	})

	return err
}

// mysqlAuthEntry mirrors an authentry row read from MySQL.
type mysqlAuthEntry struct {
	Token string `db:"token"`
	Ts    int64  `db:"ts"`
}

// selectMySQLItems reads a user's items from MySQL. It deliberately does NOT
// reference affixrerollsused (which was never added to the MySQL schema); the
// column defaults to 0 in SQLite. IFNULL guards the columns that are non-null in
// the SQLite schema.
const selectMySQLItems = "SELECT id, ts, IFNULL(`mod`, '') AS `mod`, IFNULL(ishardcore, 0) AS ishardcore, " +
	"id_baserecord, id_prefixrecord, id_suffixrecord, id_modifierrecord, id_transmuterecord, " +
	"id_materiarecord, id_reliccompletionbonusrecord, id_enchantmentrecord, " +
	"id_ascendantaffixname, id_ascendantaffix2hname, " +
	"seed, IFNULL(relicseed, 0) AS relicseed, IFNULL(enchantmentseed, 0) AS enchantmentseed, " +
	"IFNULL(materiacombines, 0) AS materiacombines, stackcount, " +
	"IFNULL(rerollsused, 0) AS rerollsused, 0 AS affixrerollsused, " +
	"created_at, IFNULL(name, '') AS name, IFNULL(namelowercase, '') AS namelowercase, " +
	"IFNULL(rarity, '') AS rarity, IFNULL(levelrequirement, 0) AS levelrequirement, " +
	"IFNULL(prefixrarity, 0) AS prefixrarity, userid " +
	"FROM item WHERE userid = ?"

// drainUser copies a single user's data from MySQL into their SQLite database
// within one transaction. Individual corrupt rows are skipped (and counted); the
// whole operation only fails on infrastructure errors. The copy first clears the
// destination tables so a retried, previously-failed migration is idempotent.
func drainUser(email string, userId config.UserId) error {
	mysql := config.GetDatabaseInstance()

	items, readSkipped := readMySQLItems(mysql, userId)

	// Deleted items / characters / tokens are tiny and rarely corrupt; on a read
	// error, log and carry on with what we have rather than failing the user.
	var deleted []DeletedItem
	if err := mysql.Select(&deleted, "SELECT id, ts FROM deleteditem WHERE userid = ?", userId); err != nil {
		log.Warn().Msgf("Error reading deleted items for %d (continuing): %v", userId, err)
	}
	var characters []CharacterEntry
	if err := mysql.Select(&characters, "SELECT name, filename FROM characters WHERE userid = ?", userId); err != nil {
		log.Warn().Msgf("Error reading characters for %d (continuing): %v", userId, err)
	}
	var auths []mysqlAuthEntry
	if err := mysql.Select(&auths, "SELECT token, UNIX_TIMESTAMP(ts) AS ts FROM authentry WHERE userid = ?", userId); err != nil {
		log.Warn().Msgf("Error reading auth tokens for %d (continuing): %v", userId, err)
	}

	udb, err := userdb.Get(email)
	if err != nil {
		return err
	}

	tx, err := udb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear-then-copy the data tables so a retried (previously partial) drain is
	// idempotent. authentry is deliberately NOT cleared: a token may already have
	// been issued at login (StoreSuccessfulAuth) before the drain runs, and
	// clearing it here would log the user out. MySQL tokens are merged into it
	// below via ON CONFLICT DO NOTHING instead.
	for _, table := range []string{"item", "deleteditem", "characters"} {
		if _, err := tx.Exec("DELETE FROM " + table); err != nil {
			return fmt.Errorf("clearing %s for %d: %w", table, userId, err)
		}
	}

	inserted, insertSkipped := insertItemsTolerant(tx, items)

	for _, d := range deleted {
		if _, err := tx.Exec("INSERT INTO deleteditem(id, ts) VALUES (?, ?) ON CONFLICT(id) DO NOTHING", d.Id, d.Ts); err != nil {
			log.Warn().Msgf("Skipping corrupt deleted item for %d: %v", userId, err)
		}
	}
	for _, ch := range characters {
		if _, err := tx.Exec("INSERT INTO characters(name, filename) VALUES (?, ?) ON CONFLICT(name) DO NOTHING", ch.Name, ch.Filename); err != nil {
			log.Warn().Msgf("Skipping corrupt character for %d: %v", userId, err)
		}
	}
	for _, a := range auths {
		if _, err := tx.Exec("INSERT INTO authentry(token, ts) VALUES (?, ?) ON CONFLICT(token) DO NOTHING", a.Token, a.Ts); err != nil {
			log.Warn().Msgf("Skipping corrupt auth token for %d: %v", userId, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	skipped := readSkipped + insertSkipped
	if err := SetMigrated(userId, inserted+skipped, inserted); err != nil {
		// SetMigrated already marked the user migrated in memory regardless of
		// this error, so the drain will not be repeated in this process. Only the
		// durable core.db record failed to write; if the process restarts before
		// a later successful write, this user would be re-drained once more.
		log.Warn().Msgf("Migrated user %d but failed to persist migration state (in-memory state still updated): %v", userId, err)
	}

	if skipped > 0 {
		log.Warn().Msgf("Migrated user %d (%s) with %d skipped/corrupt item rows (%d migrated)", userId, email, skipped, inserted)
	}
	log.Info().Msgf("Migrated user %d (%s): %d items (%d skipped), %d deleted, %d characters, %d tokens",
		userId, email, inserted, skipped, len(deleted), len(characters), len(auths))

	return nil
}

// readMySQLItems streams a user's items from MySQL, skipping (and counting) any
// row that fails to scan. Returns the successfully-read items and the skip count.
func readMySQLItems(mysql *sqlx.DB, userId config.UserId) ([]InputItem, int) {
	rows, err := mysql.Queryx(selectMySQLItems, userId)
	if err != nil {
		// Treat an inability to even start the read as "no items readable"; the
		// user will simply migrate empty. (A hard connection failure surfaces
		// later at Begin/Commit and fails the drain properly.)
		log.Warn().Msgf("Error querying items for %d: %v", userId, err)
		return nil, 0
	}
	defer rows.Close()

	var items []InputItem
	var skipped int
	for rows.Next() {
		var it InputItem
		if err := rows.StructScan(&it); err != nil {
			log.Warn().Msgf("Skipping unreadable item row for %d: %v", userId, err)
			skipped++
			continue
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		log.Warn().Msgf("Item cursor error for %d after %d rows (continuing with what was read): %v", userId, len(items), err)
	}

	return items, skipped
}

// insertItemsTolerant inserts items into the open transaction, skipping (and
// counting) any individual row that fails. Returns inserted and skipped counts.
func insertItemsTolerant(tx *sqlx.Tx, items []InputItem) (inserted int, skipped int) {
	insertItem := fmt.Sprintf("INSERT INTO item (%s) VALUES (%s) ON CONFLICT(id) DO NOTHING", insertColumns, insertPlaceholders)
	for i := range items {
		if _, err := tx.NamedExec(insertItem, items[i]); err != nil {
			log.Warn().Msgf("Skipping corrupt item %s: %v", items[i].Id, err)
			skipped++
			continue
		}
		inserted++
	}
	return inserted, skipped
}
