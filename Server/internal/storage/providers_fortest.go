package storage

func combine(a string, b string) string {
	return a + "@@@" + b
}

// TODO: StorageTest -> Providers -> Storage -> Cycle :explosion:
type InMemoryItemDb struct {
	Entries map[string][]Item
}

func (db *InMemoryItemDb) List(user string, partition string) ([]Item, error) {
	k := combine(user, partition)
	if entry, ok := db.Entries[k]; ok {
		return entry, nil
	}

	return make([]Item, 0), nil
}