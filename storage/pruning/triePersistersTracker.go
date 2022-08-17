package pruning

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type triePersistersTracker struct {
	oldestEpochKeep      int64
	oldestEpochActive    int64
	numDbsMarkedAsActive int
	numDbsMarkedAsSynced int
}

// NewTriePersisterTracker creates a new instance of triePersistersTracker
func NewTriePersisterTracker(args *StorerArgs) *triePersistersTracker {
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
func (tpi *triePersistersTracker) ShouldClosePersister(p storage.Persister, epoch int64) bool {
	shouldClosePersister := epoch < tpi.oldestEpochActive && tpi.hasActiveDbsNecessary()

	if isDbMarkedAsActive(p) {
		tpi.numDbsMarkedAsActive++
	}

	if isDbSynced(p) {
		tpi.numDbsMarkedAsSynced++
	}

	return shouldClosePersister
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
