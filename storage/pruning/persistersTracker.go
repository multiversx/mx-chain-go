package pruning

import "github.com/ElrondNetwork/elrond-go/storage"

type persistersTracker struct {
	oldestEpochKeep   int64
	oldestEpochActive int64
}

// NewPersistersTracker creates a new instance of persistersTracker
func NewPersistersTracker(args *StorerArgs) *persistersTracker {
	oldestEpochActive, oldestEpochKeep := computeOldestEpochActiveAndToKeep(args)

	return &persistersTracker{
		oldestEpochKeep:   oldestEpochKeep,
		oldestEpochActive: oldestEpochActive,
	}
}

// HasInitializedEnoughPersisters returns true if enough persisters have been initialized
func (pi *persistersTracker) HasInitializedEnoughPersisters(epoch int64) bool {
	return epoch < pi.oldestEpochKeep
}

// ShouldClosePersister returns true if the given persister needs to be closed
func (pi *persistersTracker) ShouldClosePersister(_ storage.Persister, epoch int64) bool {
	return epoch < pi.oldestEpochActive
}
