package pruning

import "github.com/multiversx/mx-chain-go/storage"

type normalPersistersTracker struct {
	oldestEpochKeep   int64
	oldestEpochActive int64
}

// NewPersistersTracker creates a new instance of normalPersistersTracker.
// It is not safe to use in a concurrent manner
func NewPersistersTracker(args EpochArgs) *normalPersistersTracker {
	oldestEpochActive, oldestEpochKeep := computeOldestEpochActiveAndToKeep(args)

	return &normalPersistersTracker{
		oldestEpochKeep:   oldestEpochKeep,
		oldestEpochActive: oldestEpochActive,
	}
}

// HasInitializedEnoughPersisters returns true if enough persisters have been initialized
func (pi *normalPersistersTracker) HasInitializedEnoughPersisters(epoch int64) bool {
	return epoch < pi.oldestEpochKeep
}

// CollectPersisterData gathers data about the persisters
func (pi *normalPersistersTracker) CollectPersisterData(_ storage.Persister) {
}

// ShouldClosePersister returns true if the given persister needs to be closed
func (pi *normalPersistersTracker) ShouldClosePersister(epoch int64) bool {
	return epoch < pi.oldestEpochActive
}

// IsInterfaceNil returns true if there is no value under the interface
func (pi *normalPersistersTracker) IsInterfaceNil() bool {
	return pi == nil
}
