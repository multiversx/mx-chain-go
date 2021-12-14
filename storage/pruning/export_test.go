package pruning

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ChangeEpoch -
func (ps *PruningStorer) ChangeEpoch(hdr data.HeaderHandler) error {
	return ps.changeEpoch(hdr)
}

// PrepareChangeEpoch -
func (ps *PruningStorer) PrepareChangeEpoch(metaBlock data.HeaderHandler) error {
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

// GetActivePersistersEpochs -
func (ps *PruningStorer) SetCacher(cacher storage.Cacher) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	ps.cacher = cacher
}

// GetOldEpochsActivePersisters -
func (fhps *FullHistoryPruningStorer) GetOldEpochsActivePersisters() storage.Cacher {
	return fhps.oldEpochsActivePersistersCache
}

// IsEpochActive -
func (fhps *FullHistoryPruningStorer) IsEpochActive(epoch uint32) bool {
	return fhps.isEpochActive(epoch)
}
