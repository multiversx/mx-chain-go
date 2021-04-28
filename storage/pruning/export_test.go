package pruning

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ChangeEpoch -
func (ps *PruningStorer) ChangeEpoch(hdr data.HeaderHandler) error {
	return ps.changeEpoch(hdr)
}

// PrepareChangeEpoch -
func (ps *PruningStorer) PrepareChangeEpoch(metaBlock *block.MetaBlock) error {
	return ps.saveHeaderForEpochStartPrepare(metaBlock)
}

// ChangeEpochSimple -
func (ps *PruningStorer) ChangeEpochSimple(epochNum uint32) error {
	return ps.changeEpoch(&block.Header{Epoch: epochNum})
}

// ChangeEpochWithExisting -
func (ps *PruningStorer) ChangeEpochWithExisting(epoch uint32) error {
	return ps.changeEpochWithExisting(epoch)
}

// GetActivePersistersEpochs -
func (ps *PruningStorer) GetActivePersistersEpochs() []uint32 {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	sliceToRet := make([]uint32, 0)
	for _, actP := range ps.activePersisters {
		sliceToRet = append(sliceToRet, actP.epoch)
	}

	return sliceToRet
}

// GetOldEpochsActivePersisters -
func (fhps *FullHistoryPruningStorer) GetOldEpochsActivePersisters() storage.Cacher {
	return fhps.oldEpochsActivePersistersCache
}

// IsEpochActive -
func (fhps *FullHistoryPruningStorer) IsEpochActive(epoch uint32) bool {
	return fhps.isEpochActive(epoch)
}

// RemoveDirectoryIfEmpty -
func RemoveDirectoryIfEmpty(path string) {
	removeDirectoryIfEmpty(path)
}
