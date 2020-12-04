package pruning

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/storage"
)

func (ps *PruningStorer) ChangeEpoch(hdr data.HeaderHandler) error {
	return ps.changeEpoch(hdr)
}

func (ps *PruningStorer) PrepareChangeEpoch(metaBlock *block.MetaBlock) error {
	return ps.saveHeaderForEpochStartPrepare(metaBlock)
}

func (ps *PruningStorer) ChangeEpochSimple(epochNum uint32) error {
	return ps.changeEpoch(&block.Header{Epoch: epochNum})
}

func (ps *PruningStorer) ChangeEpochWithExisting(epoch uint32) error {
	return ps.changeEpochWithExisting(epoch)
}

func (ps *PruningStorer) GetActivePersistersEpochs() []uint32 {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	sliceToRet := make([]uint32, 0)
	for _, actP := range ps.activePersisters {
		sliceToRet = append(sliceToRet, actP.epoch)
	}

	return sliceToRet
}

func (ps *FullHistoryPruningStorer) GetOldEpochsActivePersisters() storage.Cacher {
	return ps.oldEpochsActivePersistersCache
}

func RemoveDirectoryIfEmpty(path string) {
	removeDirectoryIfEmpty(path)
}
