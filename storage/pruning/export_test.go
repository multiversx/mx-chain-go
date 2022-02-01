package pruning

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// NewEmptyPruningStorer -
func NewEmptyPruningStorer() *PruningStorer {
	return &PruningStorer{
		persistersMapByEpoch: make(map[uint32]*persisterData),
	}
}

// AddMockActivePersister -
func (ps *PruningStorer) AddMockActivePersisters(epochs []uint32) {
	for _, e := range epochs {
		pd := &persisterData{
			epoch: e,
		}

		ps.activePersisters = append(ps.activePersisters, pd)
	}
}

// PersistersMapByEpochToSlice -
func (ps *PruningStorer) PersistersMapByEpochToSlice() []uint32 {
	slice := make([]uint32, 0)
	for epoch := range ps.persistersMapByEpoch {
		slice = append(slice, epoch)
	}

	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})

	return slice
}

// ProcessPersistersToClose -
func (ps *PruningStorer) ProcessPersistersToClose() []uint32 {
	persistersToClose := ps.processPersistersToClose()
	slice := make([]uint32, 0)
	for _, p := range persistersToClose {
		slice = append(slice, p.epoch)
	}

	return slice
}

func (ps *PruningStorer) SetNumActivePersistersParameter(num uint32) {
	ps.numOfActivePersisters = num
}

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

// GetNumActivePersisters -
func (ps *PruningStorer) GetNumActivePersisters() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return len(ps.activePersisters)
}

// GetOldEpochsActivePersisters -
func (fhps *FullHistoryPruningStorer) GetOldEpochsActivePersisters() storage.Cacher {
	return fhps.oldEpochsActivePersistersCache
}

// IsEpochActive -
func (fhps *FullHistoryPruningStorer) IsEpochActive(epoch uint32) bool {
	return fhps.isEpochActive(epoch)
}
