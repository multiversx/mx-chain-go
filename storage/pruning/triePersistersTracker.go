package pruning

import (
	"bytes"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
)

type triePersistersTracker struct {
	oldestEpochKeep      int64
	oldestEpochActive    int64
	numDbsMarkedAsActive int
	numDbsMarkedAsSynced int
}

// NewTriePersisterTracker creates a new instance of triePersistersTracker.
// It is not safe to use in a concurrent manner
func NewTriePersisterTracker(args EpochArgs) *triePersistersTracker {
	oldestEpochActive, oldestEpochKeep := computeOldestEpochActiveAndToKeep(args)

	return &triePersistersTracker{
		oldestEpochKeep:      oldestEpochKeep,
		oldestEpochActive:    oldestEpochActive,
		numDbsMarkedAsActive: 0,
		numDbsMarkedAsSynced: 0,
	}
}

// HasInitializedEnoughPersisters returns true if enough persisters have been initialized
func (tpi *triePersistersTracker) HasInitializedEnoughPersisters(epoch int64) bool {
	shouldKeepEpoch := epoch >= tpi.oldestEpochKeep
	if shouldKeepEpoch {
		return false
	}

	storageHasBeenSynced := tpi.numDbsMarkedAsActive >= 1 && tpi.numDbsMarkedAsSynced >= 1
	if storageHasBeenSynced {
		return true
	}

	return tpi.hasActiveDbsNecessary()
}

func (tpi *triePersistersTracker) hasActiveDbsNecessary() bool {
	return tpi.numDbsMarkedAsActive >= minNumOfActiveDBsNecessary
}

// ShouldClosePersister returns true if the given persister needs to be closed
func (tpi *triePersistersTracker) ShouldClosePersister(epoch int64) bool {
	return epoch < tpi.oldestEpochActive && tpi.hasActiveDbsNecessary()
}

// CollectPersisterData gathers data about the persisters
func (tpi *triePersistersTracker) CollectPersisterData(p storage.Persister) {
	if isDbMarkedAsActive(p) {
		tpi.numDbsMarkedAsActive++
	}

	if isDbSynced(p) {
		tpi.numDbsMarkedAsSynced++
	}
}

func isDbSynced(p storage.Persister) bool {
	val, err := p.Get([]byte(common.TrieSyncedKey))
	if err != nil {
		return false
	}

	return bytes.Equal(val, []byte(common.TrieSyncedVal))
}

func isDbMarkedAsActive(p storage.Persister) bool {
	val, err := p.Get([]byte(common.ActiveDBKey))
	if err != nil {
		return false
	}

	return bytes.Equal(val, []byte(common.ActiveDBVal))
}

// IsInterfaceNil returns true if there is no value under the interface
func (tpi *triePersistersTracker) IsInterfaceNil() bool {
	return tpi == nil
}
