package pruning

import "github.com/ElrondNetwork/elrond-go/storage"

type persistersTracker struct {
	oldestEpochKeep   int64
	oldestEpochActive int64
}

// This component is not safe to use in a concurrent manner
func newPersistersTracker(args *StorerArgs) *persistersTracker {
	oldestEpochActive, oldestEpochKeep := computeOldestEpochActiveAndToKeep(args)

	return &persistersTracker{
		oldestEpochKeep:   oldestEpochKeep,
		oldestEpochActive: oldestEpochActive,
	}
}

// hasInitializedEnoughPersisters returns true if enough persisters have been initialized
func (pi *persistersTracker) hasInitializedEnoughPersisters(epoch int64) bool {
	return epoch < pi.oldestEpochKeep
}

// shouldClosePersister returns true if the given persister needs to be closed
func (pi *persistersTracker) shouldClosePersister(_ storage.Persister, epoch int64) bool {
	return epoch < pi.oldestEpochActive
}
