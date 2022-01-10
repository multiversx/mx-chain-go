package pruning

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/common"
)

type triePruningStorer struct {
	*PruningStorer
}

// NewTriePruningStorer will return a new instance of NewTriePruningStorer
func NewTriePruningStorer(args *StorerArgs) (*triePruningStorer, error) {
	ps, err := initPruningStorer(args, "")
	if err != nil {
		return nil, err
	}

	return &triePruningStorer{ps}, nil
}

// PutInEpochWithoutCache adds data to persistence medium related to the specified epoch and updates the bloom filter
func (ps *triePruningStorer) PutInEpochWithoutCache(key []byte, data []byte, epoch uint32) error {
	ps.lock.RLock()
	pd, exists := ps.persistersMapByEpoch[epoch]
	ps.lock.RUnlock()
	if !exists {
		return fmt.Errorf("put in epoch: persister for epoch %d not found", epoch)
	}

	persister, closePersister, err := ps.createAndInitPersisterIfClosedProtected(pd)
	if err != nil {
		return err
	}
	defer closePersister()

	err = persister.Put(key, data)
	if err != nil {
		return err
	}

	if ps.bloomFilter != nil {
		ps.bloomFilter.Add(key)
	}

	return nil
}

// GetFromOldEpochsWithoutCache searches the old epochs for the given key without updating the cache
func (ps *triePruningStorer) GetFromOldEpochsWithoutCache(key []byte) ([]byte, error) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	if ps.bloomFilter != nil && !ps.bloomFilter.MayContain(key) {
		return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
	}

	for idx := uint32(1); (idx < ps.numOfActivePersisters) && (idx < uint32(len(ps.activePersisters))); idx++ {
		val, err := ps.activePersisters[idx].persister.Get(key)
		if err != nil {
			continue
		}

		return val, nil
	}

	return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
}

// GetFromLastEpoch searches only the last epoch storer for the given key
func (ps *triePruningStorer) GetFromLastEpoch(key []byte) ([]byte, error) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	if ps.bloomFilter != nil && !ps.bloomFilter.MayContain(key) {
		return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
	}

	if len(ps.activePersisters) < 2 {
		return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
	}

	return ps.activePersisters[1].persister.Get(key)
}

// GetLatestStorageEpoch returns the epoch for the latest opened persister
func (ps *triePruningStorer) GetLatestStorageEpoch() (uint32, error) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	if len(ps.activePersisters) == 0 {
		return 0, fmt.Errorf("there are no active persisters")
	}

	return ps.activePersisters[0].epoch, nil
}

func (ps *triePruningStorer) persisterNeeded() bool {
	for i := 0; i < int(ps.numOfActivePersisters); i++ {
		val, err := ps.activePersisters[i].persister.Get([]byte(common.ActiveDBKey))
		if err != nil {
			continue
		}

		if bytes.Equal(val, []byte(common.ActiveDBVal)) {
			return false
		}
	}

	return true
}

func (ps *triePruningStorer) shouldCloseOldPersisters(pd *persisterData) bool {
	val, err := pd.persister.Get([]byte(common.ActiveDBKey))
	if err != nil {
		return false
	}

	if bytes.Equal(val, []byte(common.ActiveDBVal)) {
		return true
	}

	return false
}
