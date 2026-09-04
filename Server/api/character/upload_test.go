package character

import (
	"archive/zip"
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/marmyr/iagdbackup/internal/storage"
	"github.com/marmyr/iagdbackup/internal/testutils"
)

type zipEntry struct {
	name string
	body string
}

// buildArchive produces a zip the way the client does, including the
// timestamped comment that makes two archives of the same save differ.
func buildArchive(t *testing.T, comment string, entries []zipEntry) []byte {
	t.Helper()

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	if err := w.SetComment(comment); err != nil {
		t.Fatalf("Error setting comment: %v", err)
	}

	for _, e := range entries {
		f, err := w.Create(e.name)
		if err != nil {
			t.Fatalf("Error creating entry %s: %v", e.name, err)
		}
		if _, err := f.Write([]byte(e.body)); err != nil {
			t.Fatalf("Error writing entry %s: %v", e.name, err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Error closing archive: %v", err)
	}

	return buf.Bytes()
}

func hashOf(t *testing.T, archive []byte) string {
	t.Helper()

	r, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatalf("Error reading archive: %v", err)
	}

	return hashArchiveContents(r)
}

func TestContentHashIgnoresArchiveMetadata(t *testing.T) {
	entries := []zipEntry{
		{"main/_Bob/player.gdc", "save data"},
		{"main/_Bob/levels_world001.map/map.fow", "fog"},
	}

	first := hashOf(t, buildArchive(t, "Created at Monday 09:00", entries))
	second := hashOf(t, buildArchive(t, "Created at Monday 17:42", entries))

	if first != second {
		t.Fatalf("Expected the same hash for identical contents, got %s and %s", first, second)
	}
}

func TestContentHashIgnoresEntryOrder(t *testing.T) {
	a := hashOf(t, buildArchive(t, "c", []zipEntry{
		{"main/_Bob/player.gdc", "save data"},
		{"main/_Bob/map.fow", "fog"},
	}))
	b := hashOf(t, buildArchive(t, "c", []zipEntry{
		{"main/_Bob/map.fow", "fog"},
		{"main/_Bob/player.gdc", "save data"},
	}))

	if a != b {
		t.Fatalf("Expected entry order not to affect the hash, got %s and %s", a, b)
	}
}

func TestContentHashChangesWithContents(t *testing.T) {
	base := []zipEntry{{"main/_Bob/player.gdc", "level 40"}}

	unchanged := hashOf(t, buildArchive(t, "c", base))
	levelled := hashOf(t, buildArchive(t, "c", []zipEntry{{"main/_Bob/player.gdc", "level 41"}}))
	renamed := hashOf(t, buildArchive(t, "c", []zipEntry{{"main/_Rob/player.gdc", "level 40"}}))
	extra := hashOf(t, buildArchive(t, "c", append(append([]zipEntry{}, base...), zipEntry{"main/_Bob/map.fow", "fog"})))

	for name, other := range map[string]string{"contents": levelled, "entry name": renamed, "extra entry": extra} {
		if unchanged == other {
			t.Fatalf("Expected a different hash when the %s changed", name)
		}
	}
}

func TestLegacyUploadIsGoneAndDoesNoWork(t *testing.T) {
	w := testutils.HostEndpoint(LegacyUploadProcessRequest, "", nil)

	if w.Code != http.StatusGone {
		t.Fatalf("Expected %d, got %d", http.StatusGone, w.Code)
	}
}

func TestSyncStateTracksUploadsForRateLimiting(t *testing.T) {
	const email = "ratelimit@example.com"
	const name = "_Bob"

	db := storage.CharacterDb{}

	state, err := db.GetSyncState(email, name)
	if err != nil {
		t.Fatalf("Error reading sync state: %v", err)
	}
	if state != nil {
		t.Fatal("Expected no sync state for a character that was never backed up")
	}

	if err := db.Upsert(email, storage.CharacterEntry{Name: name, Filename: "characters/1/abc.zip"}, "hash-1"); err != nil {
		t.Fatalf("Error storing character: %v", err)
	}

	state, err = db.GetSyncState(email, name)
	if err != nil {
		t.Fatalf("Error reading sync state: %v", err)
	}
	if state == nil {
		t.Fatal("Expected a sync state after storing a character")
	}
	if state.ContentHash != "hash-1" {
		t.Fatalf("Expected hash-1, got %s", state.ContentHash)
	}
	if age := state.Age(time.Now()); age >= MinUploadInterval {
		t.Fatalf("Expected a freshly stored character to be inside the rate limit window, age was %v", age)
	}

	// A second upload of changed contents replaces the stored digest.
	if err := db.Upsert(email, storage.CharacterEntry{Name: name, Filename: "characters/1/abc.zip"}, "hash-2"); err != nil {
		t.Fatalf("Error updating character: %v", err)
	}

	state, _ = db.GetSyncState(email, name)
	if state.ContentHash != "hash-2" {
		t.Fatalf("Expected hash-2, got %s", state.ContentHash)
	}

	entries, err := db.List(email)
	if err != nil {
		t.Fatalf("Error listing characters: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected re-uploading a character to update in place, got %d rows", len(entries))
	}
}

func TestTouchKeepsStoredArchiveButResetsTheWindow(t *testing.T) {
	const email = "touch@example.com"
	const name = "_Rob"

	db := storage.CharacterDb{}
	if err := db.Upsert(email, storage.CharacterEntry{Name: name, Filename: "characters/1/def.zip"}, "hash-1"); err != nil {
		t.Fatalf("Error storing character: %v", err)
	}

	if err := db.Touch(email, name); err != nil {
		t.Fatalf("Error touching character: %v", err)
	}

	state, err := db.GetSyncState(email, name)
	if err != nil {
		t.Fatalf("Error reading sync state: %v", err)
	}
	if state.ContentHash != "hash-1" {
		t.Fatalf("Expected Touch to leave the stored digest alone, got %s", state.ContentHash)
	}

	entry, err := db.Get(email, name)
	if err != nil || entry == nil {
		t.Fatalf("Expected the character to still resolve to its archive, got %v (err=%v)", entry, err)
	}
	if entry.Filename != "characters/1/def.zip" {
		t.Fatalf("Unexpected filename %s", entry.Filename)
	}
}
