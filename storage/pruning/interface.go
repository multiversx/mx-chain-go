package pruning

import (
	storageCore "github.com/ElrondNetwork/elrond-go-core/storage"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// EpochStartNotifier defines what a component which will handle registration to epoch start event should do
type EpochStartNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	IsInterfaceNil() bool
}

// DbFactoryHandler defines what a db factory implementation should do
type DbFactoryHandler interface {
	Create(filePath string) (storage.Persister, error)
	CreateDisabled() storage.Persister
	IsInterfaceNil() bool
}

type storerWithEpochOperations interface {
	GetFromEpoch(key []byte, epoch uint32) ([]byte, error)
	GetBulkFromEpoch(keys [][]byte, epoch uint32) ([]storageCore.KeyValuePair, error)
	PutInEpoch(key []byte, data []byte, epoch uint32) error
	Close() error
}
