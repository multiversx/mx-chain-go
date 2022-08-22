package pruning

import "github.com/ElrondNetwork/elrond-go/storage"

type normalPersistersTracker struct {
	oldestEpochKeep   int64
	oldestEpochActive int64
}

// This component is not safe to use in a concurrent manner
func newPersistersTracker(args *EpochArgs) *normalPersistersTracker {
	oldestEpochActive, oldestEpochKeep := computeOldestEpochActiveAndToKeep(args)

	return &normalPersistersTracker{
		oldestEpochKeep:   oldestEpochKeep,
		oldestEpochActive: oldestEpochActive,
	}
}

// hasInitializedEnoughPersisters returns true if enough persisters have been initialized
func (pi *normalPersistersTracker) hasInitializedEnoughPersisters(epoch int64) bool {
	return epoch < pi.oldestEpochKeep
}

func (pi *normalPersistersTracker) collectPersisterData(_ storage.Persister) {
}

// shouldClosePersister returns true if the given persister needs to be closed
func (pi *normalPersistersTracker) shouldClosePersister(epoch int64) bool {
	return epoch < pi.oldestEpochActive
}
