package storage

import (
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/coredb"
	"github.com/rs/zerolog/log"
)

// BootstrapFromMySQL seeds core.db with the central tables that must be
// available for every user before their per-user data is drained: the user
// directory (email -> db filename, buddy id) and the shared records (string
// dedup) table. Per-user data (items, deleted items, characters, auth tokens)
// is NOT copied here — that happens lazily per user via EnsureMigrated.
//
// It is idempotent (ON CONFLICT DO NOTHING) and is skipped entirely once MySQL
// is decommissioned. It preserves userid and id_record values so that existing
// item foreign keys keep resolving.
func BootstrapFromMySQL() error {
	if !config.MySQLConfigured() {
		log.Info().Msg("MySQL not configured, skipping bootstrap")
		return nil
	}

	mysql := config.GetDatabaseInstance()
	core, err := coredb.Get()
	if err != nil {
		return err
	}

	// --- Users ---
	type userRow struct {
		UserId  config.UserId `db:"userid"`
		Email   string        `db:"email"`
		BuddyId *int32        `db:"buddy_id"`
	}
	var users []userRow
	if err := mysql.Select(&users, "SELECT userid, email, buddy_id FROM users"); err != nil {
		return err
	}

	for _, u := range users {
		_, err := core.Exec("INSERT INTO users(userid, email, buddy_id, db_filename) VALUES (?, ?, ?, ?) ON CONFLICT DO NOTHING",
			u.UserId, u.Email, u.BuddyId, config.UserDbFilename(u.Email))
		if err != nil {
			return err
		}
	}
	log.Info().Msgf("Bootstrap: %d users present in core.db", len(users))

	// --- Records (string dedup) ---
	type recordRow struct {
		Id     int64  `db:"id_record"`
		Record string `db:"record"`
	}
	var records []recordRow
	if err := mysql.Select(&records, "SELECT id_record, record FROM records"); err != nil {
		return err
	}

	for _, r := range records {
		_, err := core.Exec("INSERT INTO records(id_record, record) VALUES (?, ?) ON CONFLICT DO NOTHING", r.Id, r.Record)
		if err != nil {
			return err
		}
	}
	log.Info().Msgf("Bootstrap: %d records present in core.db", len(records))

	return nil
}
