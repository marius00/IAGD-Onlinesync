-- Track drain attempts so the background drainer can give up on a chronically
-- failing user (e.g. infrastructure errors) instead of retrying forever.
ALTER TABLE migration_state ADD COLUMN attempts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE migration_state ADD COLUMN last_error TEXT;
