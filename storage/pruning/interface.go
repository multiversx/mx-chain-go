package pruning

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/storage"
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

// PersistersTracker defines what a persisters tracker should do
type PersistersTracker interface {
	HasInitializedEnoughPersisters(epoch int64) bool
	ShouldClosePersister(epoch int64) bool
	CollectPersisterData(p storage.Persister)
	IsInterfaceNil() bool
}

type storerWithEpochOperations interface {
	GetFromEpoch(key []byte, epoch uint32) ([]byte, error)
	GetBulkFromEpoch(keys [][]byte, epoch uint32) ([]data.KeyValuePair, error)
	PutInEpoch(key []byte, data []byte, epoch uint32) error
	Close() error
}
