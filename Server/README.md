# IAGD Backup API

Backup service for Grim Dawn items, consumed exclusively by the IAGD client.

## Storage architecture

Data is stored in SQLite under a persistent volume (`STORAGE_PATH`, default `/storage`):

* **`core.db`** — shared state: the user directory (`email → db_filename`, buddy id),
  the `records` string-dedup table, login attempts, throttle entries, and per-user
  migration bookkeeping.
* **`<STORAGE_PATH>/users/<sha256(email)>.db`** — one database per user, holding that
  user's items, deletion markers, character-backup filenames, and access tokens.

Items store numeric record ids; the strings are resolved from an in-memory cache
preloaded from `core.db` (no cross-database JOIN on the download path).

### MySQL migration (transitional)

The service is migrating off a single large MySQL database onto the per-user SQLite
model above. During the transition MySQL is a **strictly read-only source**:

* At startup, `BootstrapFromMySQL` seeds `core.db` with `users` and `records`
  (preserving ids).
* Each user's items/characters/tokens are drained into their `.db` **lazily** on their
  first authenticated request, and by a throttled background sweep for the long tail.
  Drains are validated (item counts) and idempotent.
* Migration state lives in `core.db`; MySQL is never mutated. Once every user is
  migrated, unset the `DATABASE_*` env vars and MySQL can be decommissioned — the
  bootstrap/drain paths are then skipped.

Schema changes to the per-user databases are applied lazily per file via numbered
migrations under `internal/userdb/migrations` (tracked with `PRAGMA user_version`);
`core.db` migrates the same way from `internal/coredb/migrations`.

## Build & deploy

Deployed as a Docker image on Coolify with `/storage` mounted as a persistent volume.

    make test      # runs against a local MySQL mirror (see below)
    make build     # builds the linux monolith binary into bin/
    make docker    # builds the Docker image

### Environment variables

| Var | Purpose |
| --- | --- |
| `STORAGE_PATH` | Persistent volume for SQLite databases (default `/storage`) |
| `DATABASE_HOST` / `DATABASE_NAME` / `DATABASE_USER` / `DATABASE_PASSWORD` | Read-only MySQL source, only needed until the drain is complete |
| `ALLOWED_ORIGIN` | CORS allowed origin |
| `REGION` / `BUCKETNAME` | S3 for character backups |

### Backups

The SQLite files under `STORAGE_PATH` are the source of truth. Do **not** copy a live
`.db` naively (WAL mode makes that inconsistent). Use one of:

* filesystem/volume snapshots of `STORAGE_PATH`, or
* `VACUUM INTO '<dest>'` per database for a consistent point-in-time copy.

## Project structure

Shared code lives under `internal`, the logic behind each endpoint under `api`, and the
process entrypoint (route wiring, scheduled maintenance/migration) under
`endpoints/monolith.go`.

### Adding a new endpoint

* Create the logic under `api/myEndpointName`, exporting `Path`, `Method` and `ProcessRequest`.
* Register it in `endpoints/monolith.go`.

# How does it work?

Data is partitioned per user (one SQLite database each).

### Authentication

The API expects the following headers to be set:

* `X-Api-User: user@example.com`
* `Authorization: AccessTokenWithoutBearerPrefix`

### Upload

Accepts an array of items. Each item has an `id` field with a GUID value. Returns any
items that were not processed due to errors.

### Download

Returns the current server timestamp, all items stored since a given timestamp, and all
items queued for removal (so other clients sync down deletions).

### Remove

Accepts a list of item ids to delete. Items are removed and added as deletion entries so
all clients are notified of the pending deletion.

### Testing locally

Bring up the local MySQL mirror (see `docker-compose.yml` in `D:\Dev\item`), then run
`make test`. Tests point `STORAGE_PATH` at a throwaway temp directory, so the production
volume is never touched, and seed `core.db` from the mirror as needed.

In IAGD, under persistent settings:

    "cloudAuthToken": "58f8e362-15a9-4872-b98c-f8438e299e8a",
    "cloudUser": "pincode@example.com",

Ensure that user exists in the (read-only) MySQL mirror so it gets bootstrapped into
`core.db`, or register directly via the login flow. Set `EnvLocalDev` URL in
`IAGrim.Backup.Cloud.Uris` to `http://localhost:8080`.
