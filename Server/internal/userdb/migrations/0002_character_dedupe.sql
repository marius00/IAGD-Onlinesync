-- Character backups are deduplicated server-side: content_hash is a digest of
-- the uploaded archive's contents, used to skip the S3 upload when a client
-- re-sends a character that has not actually changed.
ALTER TABLE characters ADD COLUMN content_hash TEXT NOT NULL DEFAULT '';
