package save

import (
	"github.com/marmyr/myservice/internal/storage"
	"math/rand"
	"time"
)

type PartitionStorage interface {
	GetActivePartition(email string) (*storage.Partition, error)
	Insert(email string, partition string, numItems int) error
}

// getPartition will return the current active partition, or a new one if the current active one is too full.
func getPartition(db PartitionStorage, user string, numItems int) (string, error) {
	activePartition, err := db.GetActivePartition(user)
	if err != nil {
		return "", err
	}

	// We can continue using the existing partition
	if activePartition != nil && !storage.ExceedsThreshold(activePartition, numItems) {
		return activePartition.Partition, nil
	}

	// We need a new partition, for whatever reason.
	pkey := storage.GeneratePartitionKey(time.Now(), rand.Int()) // TODO: Was not really designed for rand..
	err = db.Insert(user, pkey, numItems)
	if err != nil {
		return "", err
	}

	return pkey, nil
}