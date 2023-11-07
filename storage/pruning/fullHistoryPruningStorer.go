package pruning

import (
	"fmt"
	"math"

	storageCore "github.com/multiversx/mx-chain-core-go/storage"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
)

// FullHistoryPruningStorer represents a storer for full history nodes
// which creates a new persister for each epoch and removes older activePersisters
type FullHistoryPruningStorer struct {
	*PruningStorer
	args                           StorerArgs
	shardId                        string
	oldEpochsActivePersistersCache storage.Cacher
}

// NewFullHistoryPruningStorer will return a new instance of PruningStorer without sharded directories' naming scheme
func NewFullHistoryPruningStorer(args FullHistoryStorerArgs) (*FullHistoryPruningStorer, error) {
	return initFullHistoryPruningStorer(args, "")
}

// NewShardedFullHistoryPruningStorer will return a new instance of PruningStorer with sharded directories' naming scheme
func NewShardedFullHistoryPruningStorer(
	args FullHistoryStorerArgs,
	shardID uint32,
) (*FullHistoryPruningStorer, error) {
	shardStr := fmt.Sprintf("%d", shardID)
	return initFullHistoryPruningStorer(args, shardStr)
}

func initFullHistoryPruningStorer(args FullHistoryStorerArgs, shardId string) (*FullHistoryPruningStorer, error) {
	err := checkArgs(args.StorerArgs)
	if err != nil {
		return nil, err
	}

	activePersisters, persistersMapByEpoch, err := initPersistersInEpoch(args.StorerArgs, shardId)
	if err != nil {
		return nil, err
	}

	ps, err := initPruningStorer(args.StorerArgs, shardId, activePersisters, persistersMapByEpoch)
	if err != nil {
		return nil, err
	}

	ps.registerHandler(args.Notifier)

	if args.NumOfOldActivePersisters < 1 || args.NumOfOldActivePersisters > math.MaxInt32 {
		return nil, storage.ErrInvalidNumberOfOldPersisters
	}

	fhps := &FullHistoryPruningStorer{
		PruningStorer: ps,
		args:          args.StorerArgs,
		shardId:       shardId,
	}
	fhps.oldEpochsActivePersistersCache, err = cache.NewLRUCacheWithEviction(int(args.NumOfOldActivePersisters), fhps.onEvicted)
	if err != nil {
		return nil, err
	}

	return fhps, nil
}

// GetFromEpoch will search a key only in the persister for the given epoch
func (fhps *FullHistoryPruningStorer) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	data, err := fhps.searchInEpoch(key, epoch)
	if err == nil && data != nil {
		return data, nil
	}

	return fhps.searchInEpoch(key, epoch+1)
}

// GetBulkFromEpoch will search a bulk of keys in the persister for the given epoch
// doesn't return an error if a key or any isn't found
func (fhps *FullHistoryPruningStorer) GetBulkFromEpoch(keys [][]byte, epoch uint32) ([]storageCore.KeyValuePair, error) {
	persister, err := fhps.getOrOpenPersister(epoch)
	if err != nil {
		return nil, err
	}

	results := make([]storageCore.KeyValuePair, 0, len(keys))
	for _, key := range keys {
		dataInCache, found := fhps.cacher.Get(key)
		if found {
			keyValue := storageCore.KeyValuePair{Key: key, Value: dataInCache.([]byte)}
			results = append(results, keyValue)
			continue
		}
		data, errGet := persister.Get(key)
		if errGet == nil && data != nil {
			keyValue := storageCore.KeyValuePair{Key: key, Value: data}
			results = append(results, keyValue)
		}
	}

	return results, nil
}

// PutInEpoch will set the key-value pair in the given epoch
func (fhps *FullHistoryPruningStorer) PutInEpoch(key []byte, data []byte, epoch uint32) error {
	fhps.cacher.Put(key, data, len(data))

	persister, err := fhps.getOrOpenPersister(epoch)
	if err != nil {
		return err
	}

	return fhps.doPutInPersister(key, data, persister)
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

	return nil, fmt.Errorf("%w for key %x in %s",
		storage.ErrKeyNotFound, key, fhps.identifier)
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
	persister, _, err := fhps.createAndInitPersisterIfClosedUnprotected(pdata)
	if err != nil {
		return nil, err
	}

	_, ok := fhps.oldEpochsActivePersistersCache.Get([]byte(epochString))
	if !ok {
		log.Debug("fhps - getOrOpenPersister - put in cache", "epoch", epochString)
		fhps.oldEpochsActivePersistersCache.Put([]byte(epochString), pdata, 0)
	}
	return persister, nil
}

func (fhps *FullHistoryPruningStorer) getPersisterData(epochString string, epoch uint32) (*persisterData, bool) {
	pdata, exists := fhps.oldEpochsActivePersistersCache.Get([]byte(epochString))
	if exists {
		return pdata.(*persisterData), true
	}

	var pDataObj *persisterData
	pDataObj, exists = fhps.persistersMapByEpoch[epoch]
	if exists {
		return pDataObj, true
	}

	return nil, false
}

// Close will try to close all opened persisters, including the ones in the LRU cache
func (fhps *FullHistoryPruningStorer) Close() error {
	fhps.oldEpochsActivePersistersCache.Clear()

	return fhps.PruningStorer.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (fhps *FullHistoryPruningStorer) IsInterfaceNil() bool {
	return fhps == nil
}
