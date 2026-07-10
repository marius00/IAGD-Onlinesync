-- Shared state used across all users: the user directory, the string dedup
-- table ("records"), login/throttle bookkeeping, and per-user migration state.

CREATE TABLE users (
    userid     INTEGER PRIMARY KEY AUTOINCREMENT,
    email      TEXT NOT NULL UNIQUE,
    buddy_id   INTEGER UNIQUE,
    db_filename TEXT NOT NULL UNIQUE,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- String dedup table. id_record values are preserved from MySQL on bootstrap so
-- that already-issued item foreign keys keep resolving to the same string.
CREATE TABLE records (
    id_record INTEGER PRIMARY KEY,
    record    TEXT NOT NULL UNIQUE
);

-- Login codes, valid for a short-lived window before a user is authenticated.
-- Kept centrally (rather than per-user) because it's written for e-mails that
-- may not have a user/db yet.
CREATE TABLE authattempt (
    key        TEXT NOT NULL,
    code       TEXT NOT NULL DEFAULT '',
    email      TEXT NOT NULL,
    status     TEXT NOT NULL CHECK (status IN ('CREATED', 'COMPLETED')),
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (key, code)
);

-- Brute-force / spam throttling. Kept centrally since lookups span both a
-- per-user dimension and a per-IP dimension (an unauthenticated caller has no
-- per-user db to write to yet).
CREATE TABLE throttleentry (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    userid     TEXT,
    ip         TEXT,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX idx_throttleentry_userid ON throttleentry(userid);
CREATE INDEX idx_throttleentry_ip ON throttleentry(ip);

-- Tracks the MySQL -> SQLite drain per user. A row only exists once migration
-- has started; status='done' once the per-user .db is authoritative.
CREATE TABLE migration_state (
    userid            INTEGER PRIMARY KEY,
    status            TEXT NOT NULL CHECK (status IN ('IN_PROGRESS', 'DONE', 'FAILED')),
    mysql_item_count  INTEGER,
    sqlite_item_count INTEGER,
    migrated_at       INTEGER,
    swept_at          INTEGER
);
