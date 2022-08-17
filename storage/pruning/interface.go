package pruning

import (
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

// PersistersTracker defines what a persisters tracker should do
type PersistersTracker interface {
	HasInitializedEnoughPersisters(epoch int64) bool
	ShouldClosePersister(p storage.Persister, epoch int64) bool
}
