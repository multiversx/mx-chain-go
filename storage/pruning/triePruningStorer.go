package pruning

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
)

const (
	lastEpochIndex    = 1
	currentEpochIndex = 0
	// leave this at 2, because in order to have a complete state at a certain moment, 2 dbs need to be opened
	minNumOfActiveDBsNecessary = 2
)

type triePruningStorer struct {
	*PruningStorer
}

// NewTriePruningStorer will return a new instance of NewTriePruningStorer
func NewTriePruningStorer(args StorerArgs) (*triePruningStorer, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	activePersisters, persistersMapByEpoch, err := initPersistersInEpoch(args, "")
	if err != nil {
		return nil, err
	}

	ps, err := initPruningStorer(args, "", activePersisters, persistersMapByEpoch)
	if err != nil {
		return nil, err
	}

	tps := &triePruningStorer{ps}
	ps.lastEpochNeededHandler = tps.lastEpochNeeded
	tps.registerHandler(args.Notifier)

	return tps, nil
}

func (ps *triePruningStorer) lastEpochNeeded() uint32 {
	numActiveDBs := 0
	lastEpochNeeded := uint32(0)
	for i := 0; i < len(ps.activePersisters); i++ {
		lastEpochNeeded = ps.activePersisters[i].epoch
		val, err := ps.activePersisters[i].persister.Get([]byte(common.ActiveDBKey))
		if err != nil {
			continue
		}

		if bytes.Equal(val, []byte(common.ActiveDBVal)) {
			numActiveDBs++
		}

		if numActiveDBs == minNumOfActiveDBsNecessary {
			break
		}
	}

	return lastEpochNeeded
}

// PutInEpochWithoutCache adds data to persistence medium related to the specified epoch
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

	return nil
}

// GetFromOldEpochsWithoutAddingToCache searches the old epochs for the given key without adding to the cache
func (ps *triePruningStorer) GetFromOldEpochsWithoutAddingToCache(key []byte) ([]byte, core.OptionalUint32, error) {
	v, ok := ps.cacher.Get(key)
	if ok && !bytes.Equal([]byte(common.ActiveDBKey), key) {
		return v.([]byte), core.OptionalUint32{}, nil
	}

	ps.lock.RLock()
	defer ps.lock.RUnlock()

	numClosedDbs := 0
	for idx := 1; idx < len(ps.activePersisters); idx++ {
		val, err := ps.activePersisters[idx].persister.Get(key)
		if err != nil {
			if err == storage.ErrDBIsClosed {
				numClosedDbs++
			}

			continue
		}

		epoch := core.OptionalUint32{
			Value:    ps.activePersisters[idx].epoch,
			HasValue: true,
		}
		return val, epoch, nil
	}

	if numClosedDbs+1 == len(ps.activePersisters) && len(ps.activePersisters) > 1 {
		return nil, core.OptionalUint32{}, storage.ErrDBIsClosed
	}

	return nil, core.OptionalUint32{}, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
}

// GetFromLastEpoch searches only the last epoch storer for the given key
func (ps *triePruningStorer) GetFromLastEpoch(key []byte) ([]byte, error) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	if len(ps.activePersisters) < 2 {
		return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
	}

	return ps.activePersisters[lastEpochIndex].persister.Get(key)
}

// GetFromCurrentEpoch searches only the current epoch storer for the given key
func (ps *triePruningStorer) GetFromCurrentEpoch(key []byte) ([]byte, error) {
	ps.lock.RLock()

	if len(ps.activePersisters) == 0 {
		ps.lock.RUnlock()
		return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
	}

	persister := ps.activePersisters[currentEpochIndex].persister
	ps.lock.RUnlock()

	return persister.Get(key)
}

// GetWithStats searches the key in the cache. In case it is not found, the key may be in the db.
// it will return true if the key has been found in cache
func (ps *triePruningStorer) GetWithStats(key []byte) ([]byte, bool, error) {
	v, foundInCache, err := ps.getWithCacheStatus(key)
	return v, foundInCache, err
}

// GetLatestStorageEpoch returns the epoch for the latest opened persister
func (ps *triePruningStorer) GetLatestStorageEpoch() (uint32, error) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	if len(ps.activePersisters) == 0 {
		return 0, fmt.Errorf("there are no active persisters")
	}

	return ps.activePersisters[currentEpochIndex].epoch, nil
}

// RemoveFromAllActiveEpochs removes the data associated to the given key from both cache and epochs storers
func (ps *triePruningStorer) RemoveFromAllActiveEpochs(key []byte) error {
	var err error
	ps.cacher.Remove(key)

	ps.lock.RLock()
	defer ps.lock.RUnlock()
	for _, pd := range ps.activePersisters {
		err = pd.persister.Remove(key)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ps *triePruningStorer) IsInterfaceNil() bool {
	return ps == nil
}
