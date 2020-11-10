package storage

import (
	"fmt"
	"time"
)

type PartitionDb struct {
}

type Partition struct {
	Email     string `json:"-"`         // User/Owner/Email
	Partition string `json:"partition"` // Partition key
	IsActive  bool   `json:"isActive"`  // If this partition is active and accepts new items
	NumItems  int    `json:"numItems"`  // The _estimated_ number of items in this partition, consumer is responsible for updating the value and does not account for race conditions.
}

func (*PartitionDb) Insert(email string, partition string, numItems int) error {
	// TODO: Insert entry into partition table
	// TODO: Loop all active partitions and set to inactive
	return nil
}

// SetNumItems will update the estimated number of items in a given partition
func (*PartitionDb) SetNumItems(email string, partition string, numItems int) error {
	// TODO: Update number of items in partition table [partition, metadata]
	return nil
}

// Delete will delete a given partition entry for a user
func (*PartitionDb) Delete(email string, partition string) error {
	// TODO: Delete partition entry
	// TODO: Delete entire partition from item table [or delegate to item db? -- delegating may simplify testing]
	return nil
}

// Will get the active partition for a given user, may return nil
func (*PartitionDb) GetActivePartition(email string) (*Partition, error) {
	// TODO: Get partition where IsActive=true
	// If itemCount > Permitted, caller should close and create new. [simplifies testing]
	// TODO: What if this is the first partition?
	return nil, nil
}

// Fetch all partitions for a given user
func (*PartitionDb) List(email string) ([]Partition, error) {
	// TODO: Get all partitions for user
	return nil, nil
}

// GeneratePartitionKey will generate a partition key for the provided time period and iteration. (Iteration is arbitrary, allowing multiple partitions for a given time period, to prevent them growing too large)
func GeneratePartitionKey(time time.Time, iteration int) string {
	y, w := time.ISOWeek()
	return fmt.Sprintf("%04d:%02d:%02d", y, w, iteration)
}
