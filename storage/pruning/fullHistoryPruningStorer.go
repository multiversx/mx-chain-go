package pruning

import (
	"encoding/hex"
	"fmt"
	"math"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

// FullHistoryPruningStorer represents a storer for full history nodes
// which creates a new persister for each epoch and removes older activePersisters
type FullHistoryPruningStorer struct {
	*PruningStorer
	args                           *StorerArgs
	shardId                        string
	oldEpochsActivePersistersCache storage.Cacher
}

// NewFullHistoryPruningStorer will return a new instance of PruningStorer without sharded directories' naming scheme
func NewFullHistoryPruningStorer(args *FullHistoryStorerArgs) (*FullHistoryPruningStorer, error) {
	return initFullHistoryPruningStorer(args, "")
}

// NewShardedFullHistoryPruningStorer will return a new instance of PruningStorer with sharded directories' naming scheme
func NewShardedFullHistoryPruningStorer(
	args *FullHistoryStorerArgs,
	shardID uint32,
) (*FullHistoryPruningStorer, error) {
	shardStr := fmt.Sprintf("%d", shardID)
	return initFullHistoryPruningStorer(args, shardStr)
}

func initFullHistoryPruningStorer(args *FullHistoryStorerArgs, shardId string) (*FullHistoryPruningStorer, error) {
	ps, err := initPruningStorer(args.StorerArgs, shardId)
	if err != nil {
		return nil, err
	}

	if args.NumOfOldActivePersisters < 1 || args.NumOfOldActivePersisters > math.MaxInt32 {
		return nil, storage.ErrInvalidNumberOfOldPersisters
	}

	fhps := &FullHistoryPruningStorer{
		PruningStorer: ps,
		args:          args.StorerArgs,
		shardId:       shardId,
	}
	fhps.oldEpochsActivePersistersCache, err = lrucache.NewCacheWithEviction(int(args.NumOfOldActivePersisters), fhps.onEvicted)
	if err != nil {
		return nil, err
	}

	return fhps, nil
}

func (fhps *FullHistoryPruningStorer) onEvicted(key interface{}, value interface{}) {
	pd, ok := value.(*persisterData)
	if ok {
		//since the put operation on oldEpochsActivePersistersCache is already done under the mutex we shall not lock
		// the same mutex again here. It is safe to proceed without lock.

		for _, active := range fhps.activePersisters {
			if active.epoch == pd.epoch {
				return
			}
		}

		if pd.getIsClosed() {
			return
		}

		err := pd.Close()
		if err != nil {
			log.Warn("initFullHistoryPruningStorer - onEvicted", "key", key, "err", err.Error())
		}
	}
}

// GetFromEpoch will search a key only in the persister for the given epoch
func (fhps *FullHistoryPruningStorer) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	data, err := fhps.searchInEpoch(key, epoch)
	if err == nil && data != nil {
		return data, nil
	}

	return fhps.searchInEpoch(key, epoch+1)
}

func (fhps *FullHistoryPruningStorer) searchInEpoch(key []byte, epoch uint32) ([]byte, error) {
	if fhps.isEpochActive(epoch) {
		return fhps.PruningStorer.SearchFirst(key)
	}

	data, err := fhps.getFromOldEpoch(key, epoch)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (fhps *FullHistoryPruningStorer) isEpochActive(epoch uint32) bool {
	fhps.lock.RLock()
	oldestEpochInCurrentSetting := fhps.activePersisters[len(fhps.activePersisters)-1].epoch
	newestEpochInCurrentSetting := fhps.activePersisters[0].epoch
	fhps.lock.RUnlock()

	return epoch >= oldestEpochInCurrentSetting && epoch <= newestEpochInCurrentSetting
}

func (fhps *FullHistoryPruningStorer) getFromOldEpoch(key []byte, epoch uint32) ([]byte, error) {
	persister, err := fhps.getOrOpenPersister(epoch)
	if err != nil {
		return nil, err
	}

	res, err := persister.Get(key)
	if err == nil {
		return res, nil
	}

	log.Trace("FullHistoryPruningStorer.getFromOldEpoch",
		"id", fhps.identifier,
		"epoch", epoch,
		"key", key,
		"error", err.Error())

	return nil, fmt.Errorf("key %s not found in %s",
		hex.EncodeToString(key), fhps.identifier)
}

func (fhps *FullHistoryPruningStorer) getOrOpenPersister(epoch uint32) (storage.Persister, error) {
	epochString := fmt.Sprintf("%d", epoch)

	fhps.lock.RLock()
	pdata, exists := fhps.getPersisterData(epochString, epoch)
	fhps.lock.RUnlock()

	if exists {
		isClosed := pdata.getIsClosed()
		if !isClosed {
			return pdata.getPersister(), nil
		}
	}

	fhps.lock.Lock()
	defer fhps.lock.Unlock()

	pdata, exists = fhps.getPersisterData(epochString, epoch)
	if !exists {
		newPdata, errPersisterData := createPersisterDataForEpoch(fhps.args, epoch, fhps.shardId)
		if errPersisterData != nil {
			return nil, errPersisterData
		}

		fhps.oldEpochsActivePersistersCache.Put([]byte(epochString), newPdata, 0)
		fhps.persistersMapByEpoch[epoch] = newPdata

		return newPdata.getPersister(), nil
	}
	persister, _, err := fhps.createAndInitPersisterIfClosed(pdata)
	if err != nil {
		return nil, err
	}

	return persister, nil
}

func (fhps *FullHistoryPruningStorer) getPersisterData(epochString string, epoch uint32) (*persisterData, bool) {
	pdata, exists := fhps.oldEpochsActivePersistersCache.Get([]byte(epochString))
	if exists {
		return pdata.(*persisterData), true
	}

	pdata, exists = fhps.persistersMapByEpoch[epoch]
	if exists {
		return pdata.(*persisterData), true
	}

	return nil, false
}
