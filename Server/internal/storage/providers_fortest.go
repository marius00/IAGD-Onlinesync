package storage

func combine(a string, b string) string {
	return a + "@@@" + b;
}

// TODO: StorageTest -> Providers -> Storage -> Cycle :explosion:
type InMemoryPartitionDb struct {
	Entries map[string]Partition
}

func (db *InMemoryPartitionDb) Get(user string, partition string) (*Partition, error) {
	k := combine(user, partition)
	if entry, ok := db.Entries[k]; ok {
		return &entry, nil
	}

	return nil, nil
}

type InMemoryDeletedItemDb struct {
	Entries map[string][]DeletedItem
}

func (db *InMemoryDeletedItemDb) List(user string, partition string) ([]DeletedItem, error) {
	k := combine(user, partition)
	if entry, ok := db.Entries[k]; ok {
		return entry, nil
	}

	return make([]DeletedItem, 0), nil
}

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