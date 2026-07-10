package storage

import (
	"time"

	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/coredb"
	"github.com/rs/zerolog/log"
)

// MigrateAllUsers drains every not-yet-migrated user from MySQL into SQLite,
// pausing pauseBetween between users to spare the (slow) host. Users that have
// failed to drain too many times (isDrainExhausted) are skipped so the drainer
// never gets stuck on a chronically broken user. Returns counts for this run.
func MigrateAllUsers(pauseBetween time.Duration) (migrated int, failed int, skipped int) {
	if !config.MySQLConfigured() {
		log.Info().Msg("MySQL not configured, skipping migration sweep")
		return 0, 0, 0
	}

	userDb := UserDb{}
	emails, err := userDb.AllEmails()
	if err != nil {
		log.Warn().Msgf("Migration sweep: could not list users: %v", err)
		return 0, 0, 0
	}

	for _, email := range emails {
		entry, err := userDb.GetByEmail(email)
		if err != nil || entry == nil {
			continue
		}

		if IsMigrated(entry.UserId) {
			continue
		}
		if isDrainExhausted(entry.UserId) {
			skipped++
			continue
		}

		if err := EnsureMigrated(email, entry.UserId); err != nil {
			log.Warn().Msgf("Sweep: failed to migrate user %d (%s): %v", entry.UserId, email, err)
			failed++
			continue
		}

		migrated++
		if pauseBetween > 0 {
			time.Sleep(pauseBetween)
		}
	}

	return migrated, failed, skipped
}

// RunBackgroundDrain continuously and slowly drains users off MySQL so the
// legacy database can eventually be decommissioned without waiting for every
// user to log in (some may not for years). It trickles through users with
// pauseBetween between each, then sleeps passInterval before re-scanning (to
// pick up newly-added users and retry not-yet-exhausted failures). Intended to
// be run in its own goroutine for the lifetime of the process.
func RunBackgroundDrain(pauseBetween, passInterval time.Duration) {
	if !config.MySQLConfigured() {
		log.Info().Msg("MySQL not configured, background drain disabled")
		return
	}

	log.Info().Msgf("Background MySQL drain started (%.0fs between users, %.0fs between passes)",
		pauseBetween.Seconds(), passInterval.Seconds())

	for {
		migrated, failed, skipped := MigrateAllUsers(pauseBetween)

		done, failedTotal, pending, err := MigrationProgress()
		if err != nil {
			log.Warn().Msgf("Background drain: could not read progress: %v", err)
		} else {
			log.Info().Msgf("Background drain pass complete: +%d migrated, %d failed, %d exhausted-skipped this pass; totals: %d done, %d failed, %d pending",
				migrated, failed, skipped, done, failedTotal, pending)
			if pending == 0 {
				log.Info().Msg("All users drained off MySQL; the legacy database can now be decommissioned (unset DATABASE_* env vars).")
			}
		}

		time.Sleep(passInterval)
	}
}

// MigrationProgress reports how many users are fully migrated (done), have
// exhausted their drain attempts (failed), and remain to be drained (pending).
// When pending reaches 0, MySQL can be decommissioned.
func MigrationProgress() (done int, failed int, pending int, err error) {
	db, err := coredb.Get()
	if err != nil {
		return 0, 0, 0, err
	}

	var total int
	if err = db.Get(&total, "SELECT COUNT(*) FROM users"); err != nil {
		return 0, 0, 0, err
	}
	if err = db.Get(&done, "SELECT COUNT(*) FROM migration_state WHERE status = 'DONE'"); err != nil {
		return 0, 0, 0, err
	}
	if err = db.Get(&failed, "SELECT COUNT(*) FROM migration_state WHERE status = 'FAILED' AND attempts >= ?", maxDrainAttempts); err != nil {
		return 0, 0, 0, err
	}

	pending = total - done - failed
	if pending < 0 {
		pending = 0
	}
	return done, failed, pending, nil
}
