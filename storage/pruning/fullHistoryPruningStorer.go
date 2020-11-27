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
	args                      *StorerArgs
	shardId                   string
	oldEpochsActivePersisters storage.Cacher
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

	oldEpochsActivePersisters, err := lrucache.NewCacheWithEviction(int(args.NumOfOldActivePersisters), onEvicted)

	if err != nil {
		return nil, err
	}

	return &FullHistoryPruningStorer{
		PruningStorer:             ps,
		args:                      args.StorerArgs,
		shardId:                   shardId,
		oldEpochsActivePersisters: oldEpochsActivePersisters,
	}, nil
}

func onEvicted(key interface{}, value interface{}) {
	pd, ok := value.(*persisterData)
	if ok {
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
	fhps.lock.RLock()
	oldestEpochInCurrentSetting := fhps.activePersisters[len(fhps.activePersisters)-1].epoch
	fhps.lock.RUnlock()

	isActiveEpoch := epoch > oldestEpochInCurrentSetting-fhps.numOfActivePersisters
	if isActiveEpoch {
		return fhps.PruningStorer.Get(key)
	}

	data, err := fhps.getFromOldEpoch(key, epoch)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (fhps *FullHistoryPruningStorer) getFromOldEpoch(key []byte, epoch uint32) ([]byte, error) {
	persister, err := fhps.getPersister(epoch)
	if err != nil {
		return nil, err
	}

	res, err := persister.Get(key)
	if err == nil {
		return res, nil
	}

	log.Warn("GetFromEpoch persister",
		"id", fhps.identifier,
		"epoch", epoch,
		"key", key,
		"error", err.Error())

	return nil, fmt.Errorf("key %s not found in %s",
		hex.EncodeToString(key), fhps.identifier)
}

func (fhps *FullHistoryPruningStorer) getPersister(epoch uint32) (storage.Persister, error) {
	epochString := fmt.Sprintf("%d", epoch)

	fhps.lock.RLock()
	pdata, exists := fhps.oldEpochsActivePersisters.Get([]byte(epochString))
	fhps.lock.RUnlock()

	var pd *persisterData
	if exists {
		pd = pdata.(*persisterData)
		isClosed := pd.getIsClosed()
		if !isClosed {
			return pd.persister, nil
		}
	}

	fhps.lock.Lock()
	defer fhps.lock.Unlock()

	pdata, exists = fhps.oldEpochsActivePersisters.Get([]byte(epochString))
	if !exists {
		newPdata, errPersisterData := createPersisterDataForEpoch(fhps.args, epoch, fhps.shardId)
		if errPersisterData != nil {
			return nil, errPersisterData
		}

		fhps.oldEpochsActivePersisters.Put([]byte(epochString), newPdata, 0)
		fhps.persistersMapByEpoch[epoch] = newPdata
		pdata = newPdata
	}
	pd = pdata.(*persisterData)
	persister, _, err := fhps.createAndInitPersisterIfClosed(pd)
	if err != nil {
		return nil, err
	}

	return persister, nil
}
