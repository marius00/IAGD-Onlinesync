-- Per-user database: items, deletion markers, character backups and access
-- tokens for a single user. userid is implicit (the file itself), so item ids
-- no longer need a composite (userid, id) key.

CREATE TABLE item (
    id                             TEXT PRIMARY KEY,
    id_baserecord                  INTEGER,
    id_prefixrecord                INTEGER,
    id_suffixrecord                INTEGER,
    id_modifierrecord              INTEGER,
    id_transmuterecord             INTEGER,
    id_reliccompletionbonusrecord  INTEGER,
    id_enchantmentrecord           INTEGER,
    id_materiarecord               INTEGER,
    id_ascendantaffixname          INTEGER,
    id_ascendantaffix2hname        INTEGER,
    seed                           INTEGER NOT NULL,
    relicseed                      INTEGER,
    enchantmentseed                INTEGER,
    materiacombines                INTEGER,
    stackcount                     INTEGER NOT NULL,
    rerollsused                    INTEGER,
    affixrerollsused               INTEGER,
    name                           TEXT DEFAULT '',
    namelowercase                  TEXT DEFAULT '',
    rarity                         TEXT DEFAULT '',
    mod                            TEXT DEFAULT '',
    levelrequirement               REAL DEFAULT 0,
    prefixrarity                   INTEGER,
    unknown                        INTEGER,
    ishardcore                     INTEGER NOT NULL DEFAULT 0,
    created_at                     INTEGER NOT NULL DEFAULT 0,
    ts                             INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_item_ts ON item(ts);

-- Items which have been deleted; ts drives sync-down of the deletion to other
-- clients.
CREATE TABLE deleteditem (
    id         TEXT PRIMARY KEY,
    ts         INTEGER NOT NULL
);
CREATE INDEX idx_deleteditem_ts ON deleteditem(ts);

-- Filename mapping for character backups (actual file bytes live in S3).
CREATE TABLE characters (
    name       TEXT PRIMARY KEY,
    filename   TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- Access tokens issued to this user.
CREATE TABLE authentry (
    token TEXT PRIMARY KEY,
    ts    INTEGER NOT NULL
);
