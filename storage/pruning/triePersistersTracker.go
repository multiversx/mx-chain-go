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

// This component is not safe to use in a concurrent manner
func newTriePersisterTracker(args *EpochArgs) *triePersistersTracker {
	oldestEpochActive, oldestEpochKeep := computeOldestEpochActiveAndToKeep(args)

	return &triePersistersTracker{
		oldestEpochKeep:      oldestEpochKeep,
		oldestEpochActive:    oldestEpochActive,
		numDbsMarkedAsActive: 0,
		numDbsMarkedAsSynced: 0,
	}
}

// hasInitializedEnoughPersisters returns true if enough persisters have been initialized
func (tpi *triePersistersTracker) hasInitializedEnoughPersisters(epoch int64) bool {
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

// shouldClosePersister returns true if the given persister needs to be closed
func (tpi *triePersistersTracker) shouldClosePersister(epoch int64) bool {
	return epoch < tpi.oldestEpochActive && tpi.hasActiveDbsNecessary()
}

func (tpi *triePersistersTracker) collectPersisterData(p storage.Persister) {
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
