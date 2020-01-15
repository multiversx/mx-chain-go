package pruning

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var log = logger.GetOrCreate("storage/pruning")

// persisterData structure is used so the persister and its path can be kept in the same place
type persisterData struct {
	persister storage.Persister
	path      string
	isClosed  bool
}

// PruningStorer represents a storer which creates a new persister for each epoch and removes older activePersisters
type PruningStorer struct {
	pruningEnabled        bool
	lock                  sync.RWMutex
	shardCoordinator      sharding.Coordinator
	fullArchive           bool
	activePersisters      []*persisterData
	persistersMapByEpoch  map[uint32]*persisterData
	cacher                storage.Cacher
	bloomFilter           storage.BloomFilter
	pathManager           storage.PathManagerHandler
	dbPath                string
	persisterFactory      DbFactoryHandler
	numOfEpochsToKeep     uint32
	numOfActivePersisters uint32
	identifier            string
}

// NewPruningStorer will return a new instance of PruningStorer without sharded directories' naming scheme
func NewPruningStorer(args *StorerArgs) (*PruningStorer, error) {
	return initPruningStorer(args, "")
}

// NewShardedPruningStorer will return a new instance of PruningStorer with sharded directories' naming scheme
func NewShardedPruningStorer(
	args *StorerArgs,
	shardId uint32,
) (*PruningStorer, error) {
	shardIdStr := fmt.Sprintf("%d", shardId)
	return initPruningStorer(args, shardIdStr)
}

// initPruningStorer will create a PruningStorer with or without sharded directories' naming scheme
func initPruningStorer(
	args *StorerArgs,
	shardIdStr string,
) (*PruningStorer, error) {
	var cache storage.Cacher
	var db storage.Persister
	var bf storage.BloomFilter
	var err error

	defer func() {
		if err != nil && db != nil {
			_ = db.Destroy()
		}
	}()

	if args.NumOfActivePersisters < 1 {
		return nil, storage.ErrInvalidNumberOfPersisters
	}
	if check.IfNil(args.Notifier) {
		return nil, storage.ErrNilEpochStartNotifier
	}
	if check.IfNil(args.PersisterFactory) {
		return nil, storage.ErrNilPersisterFactory
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, storage.ErrNilShardCoordinator
	}
	if check.IfNil(args.PathManager) {
		return nil, storage.ErrNilPathManager
	}

	cache, err = storageUnit.NewCache(args.CacheConf.Type, args.CacheConf.Size, args.CacheConf.Shards)
	if err != nil {
		return nil, err
	}

	filePath := args.PathManager.PathForEpoch(core.GetShardIdString(args.ShardCoordinator.SelfId()), args.StartingEpoch, args.Identifier)
	if len(shardIdStr) > 0 {
		filePath = filePath + shardIdStr
		args.Identifier += shardIdStr
	}
	db, err = args.PersisterFactory.Create(filePath)
	if err != nil {
		return nil, err
	}

	var persisters []*persisterData
	persisters = append(persisters, &persisterData{
		persister: db,
		path:      filePath,
		isClosed:  false,
	})

	persistersMapByEpoch := make(map[uint32]*persisterData)
	persistersMapByEpoch[args.StartingEpoch] = persisters[0]

	pdb := &PruningStorer{
		pruningEnabled:        args.PruningEnabled,
		identifier:            args.Identifier,
		fullArchive:           args.FullArchive,
		activePersisters:      persisters,
		persisterFactory:      args.PersisterFactory,
		shardCoordinator:      args.ShardCoordinator,
		persistersMapByEpoch:  persistersMapByEpoch,
		cacher:                cache,
		bloomFilter:           nil,
		pathManager:           args.PathManager,
		dbPath:                args.DbPath,
		numOfEpochsToKeep:     args.NumOfEpochsToKeep,
		numOfActivePersisters: args.NumOfActivePersisters,
	}

	if args.BloomFilterConf.Size != 0 { // if size is 0, that means an empty config was used so bloom filter will be nil
		bf, err = storageUnit.NewBloomFilter(args.BloomFilterConf)
		if err != nil {
			return nil, err
		}

		pdb.bloomFilter = bf
	}

	err = pdb.activePersisters[0].persister.Init()
	if err != nil {
		return nil, err
	}

	pdb.registerHandler(args.Notifier)

	return pdb, nil
}

// Put adds data to both cache and persistence medium and updates the bloom filter
func (ps *PruningStorer) Put(key, data []byte) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	ps.cacher.Put(key, data)

	err := ps.activePersisters[0].persister.Put(key, data)
	if err != nil {
		ps.cacher.Remove(key)
		return err
	}

	if ps.bloomFilter != nil {
		ps.bloomFilter.Add(key)
	}

	return nil
}

// Get searches the key in the cache. In case it is not found, it verifies with the bloom filter
// if the key may be in the db. If bloom filter confirms then it further searches in the databases.
func (ps *PruningStorer) Get(key []byte) ([]byte, error) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	v, ok := ps.cacher.Get(key)
	var err error

	if !ok {
		// not found in cache
		// search it in active persisters
		found := false
		for idx := uint32(0); (idx < ps.numOfActivePersisters) && (idx < uint32(len(ps.activePersisters))); idx++ {
			if ps.bloomFilter == nil || ps.bloomFilter.MayContain(key) {
				v, err = ps.activePersisters[idx].persister.Get(key)
				if err != nil {
					continue
				}

				found = true
				// if found in persistence unit, add it to cache
				ps.cacher.Put(key, v)
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("key %s not found in %s",
				base64.StdEncoding.EncodeToString(key), ps.identifier)
		}
	}

	return v.([]byte), nil
}

// Close will close PruningStorer
func (ps *PruningStorer) Close() error {
	closedSuccessfully := true
	for _, persister := range ps.activePersisters {
		err := persister.persister.Close()

		if err != nil {
			log.Error("cannot close persister", err)
			closedSuccessfully = false
		}
	}

	if closedSuccessfully {
		return nil
	}

	return storage.ErrClosingPersisters
}

// GetFromEpoch will search a key only in the persister for the given epoch
func (ps *PruningStorer) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	// TODO: this will be used when requesting from resolvers
	ps.lock.Lock()
	defer ps.lock.Unlock()

	v, ok := ps.cacher.Get(key)
	if ok {
		return v.([]byte), nil
	}

	persisterData, exists := ps.persistersMapByEpoch[epoch]
	if !exists {
		return nil, fmt.Errorf("key %s not found in %s",
			base64.StdEncoding.EncodeToString(key), ps.identifier)
	}

	if !persisterData.isClosed {
		return persisterData.persister.Get(key)
	}

	persister, err := ps.persisterFactory.Create(persisterData.path)
	if err != nil {
		log.Debug("open old persister", "error", err.Error())
		return nil, err
	}

	defer func() {
		err = persister.Close()
		if err != nil {
			log.Debug("persister.Close()", "error", err.Error())
		}
	}()

	err = persister.Init()
	if err != nil {
		log.Debug("init old persister", "error", err.Error())
		return nil, err
	}

	res, err := persister.Get(key)
	if err == nil {
		return res, nil
	}

	log.Warn("get from closed persister",
		"id", ps.identifier,
		"epoch", epoch,
		"key", key,
		"error", err.Error())

	return nil, fmt.Errorf("key %s not found in %s",
		base64.StdEncoding.EncodeToString(key), ps.identifier)

}

// SearchFirst will search a given key in all the active persisters, from the newest to the oldest
func (ps *PruningStorer) SearchFirst(key []byte) ([]byte, error) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	v, ok := ps.cacher.Get(key)
	if ok {
		return v.([]byte), nil
	}

	var res []byte
	var err error
	for _, pd := range ps.activePersisters {
		res, err = pd.persister.Get(key)
		if err == nil {
			return res, nil
		}
	}

	log.Debug("SearchFirst error",
		"unit", ps.identifier,
		"last err", err,
		"num active persisters", len(ps.activePersisters),
		"key", key)

	return nil, fmt.Errorf("%w - SearchFirst, key = %s, num active persisters = %d",
		storage.ErrKeyNotFound,
		base64.StdEncoding.EncodeToString(key),
		len(ps.activePersisters),
	)
}

// Has checks if the key is in the Unit.
// It first checks the cache. If it is not found, it checks the bloom filter
// and if present it checks the db
func (ps *PruningStorer) Has(key []byte) error {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	has := ps.cacher.Has(key)
	if has {
		return nil
	}

	if ps.bloomFilter == nil || ps.bloomFilter.MayContain(key) {
		for _, persister := range ps.activePersisters {
			if persister.persister.Has(key) != nil {
				continue
			}

			return nil
		}
	}

	return storage.ErrKeyNotFound
}

// HasInEpoch checks if the key is in the Unit in a given epoch.
// It first checks the cache. If it is not found, it checks the bloom filter
// and if present it checks the db
func (ps *PruningStorer) HasInEpoch(key []byte, epoch uint32) error {
	// TODO: this will be used when requesting from resolvers
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	has := ps.cacher.Has(key)
	if has {
		return nil
	}

	if ps.bloomFilter == nil || ps.bloomFilter.MayContain(key) {
		persisterData, ok := ps.persistersMapByEpoch[epoch]
		if !ok {
			return storage.ErrKeyNotFound
		}

		if !persisterData.isClosed {
			return persisterData.persister.Has(key)
		}

		persister, err := ps.persisterFactory.Create(persisterData.path)
		if err != nil {
			log.Debug("open old persister", "error", err.Error())
			return err
		}

		defer func() {
			err = persister.Close()
			if err != nil {
				log.Debug("persister.Close()", "error", err.Error())
			}
		}()

		err = persister.Init()
		if err != nil {
			log.Debug("init old persister", "error", err.Error())
			return err
		}

		return persister.Has(key)
	}

	return storage.ErrKeyNotFound
}

// Remove removes the data associated to the given key from both cache and persistence medium
func (ps *PruningStorer) Remove(key []byte) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	var err error
	ps.cacher.Remove(key)
	for _, persisterData := range ps.activePersisters {
		err = persisterData.persister.Remove(key)
		if err == nil {
			return nil
		}
	}

	return err
}

// ClearCache cleans up the entire cache
func (ps *PruningStorer) ClearCache() {
	ps.cacher.Clear()
}

// DestroyUnit cleans up the bloom filter, the cache, and the dbs
func (ps *PruningStorer) DestroyUnit() error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if ps.bloomFilter != nil {
		ps.bloomFilter.Clear()
	}

	ps.cacher.Clear()

	var err error
	numOfPersistersRemoved := 0
	totalNumOfPersisters := len(ps.persistersMapByEpoch)
	for _, persisterData := range ps.persistersMapByEpoch {
		if persisterData.isClosed {
			err = persisterData.persister.DestroyClosed()
		} else {
			err = persisterData.persister.Destroy()
		}

		if err != nil {
			log.Debug("pruning db: destroy",
				"error", err.Error())
			continue
		}
		numOfPersistersRemoved++
	}

	if numOfPersistersRemoved != totalNumOfPersisters {
		log.Debug("error destroying pruning db",
			"identifier", ps.identifier,
			"destroyed", numOfPersistersRemoved,
			"total", totalNumOfPersisters)
		return storage.ErrDestroyingUnit
	}

	return nil
}

// registerHandler will register a new function to the epoch start notifier
func (ps *PruningStorer) registerHandler(handler EpochStartNotifier) {
	subscribeHandler := notifier.MakeHandlerForEpochStart(func(hdr data.HeaderHandler) {
		err := ps.changeEpoch(hdr.GetEpoch())
		if err != nil {
			log.Warn("change epoch in storer", "error", err.Error())
		}
	})

	handler.RegisterHandler(subscribeHandler)
}

// changeEpoch will handle creating a new persister and removing of the older ones
func (ps *PruningStorer) changeEpoch(epoch uint32) error {
	log.Debug("PruningStorer - change epoch", "epoch", epoch)
	// if pruning is not enabled, don't create new persisters, but use the same one instead
	if !ps.pruningEnabled {
		log.Debug("PruningStorer - change epoch - pruning is disabled")
		return nil
	}

	ps.lock.Lock()
	defer ps.lock.Unlock()

	shardId := core.GetShardIdString(ps.shardCoordinator.SelfId())
	filePath := ps.pathManager.PathForEpoch(shardId, epoch, ps.identifier)
	db, err := ps.persisterFactory.Create(filePath)
	if err != nil {
		log.Warn("change epoch error", "error - "+ps.identifier, err.Error())
		return err
	}

	newPersister := &persisterData{
		persister: db,
		path:      filePath,
		isClosed:  false,
	}

	singleItemPersisters := []*persisterData{newPersister}
	ps.activePersisters = append(singleItemPersisters, ps.activePersisters...)
	ps.persistersMapByEpoch[epoch] = newPersister

	err = ps.activePersisters[0].persister.Init()
	if err != nil {
		return err
	}

	err = ps.closeAndDestroyPersisters(epoch)
	if err != nil {
		log.Debug("closing and destroying old persister", "error", err.Error())
		return err
	}

	return nil
}

func (ps *PruningStorer) closeAndDestroyPersisters(epoch uint32) error {
	// recent activePersisters have to he closed for both scenarios: full archive or not
	if ps.numOfActivePersisters < uint32(len(ps.activePersisters)) {
		persisterToClose := ps.activePersisters[ps.numOfActivePersisters]
		err := persisterToClose.persister.Close()
		if err != nil {
			log.Error("error closing persister", "error", err.Error(), "id", ps.identifier)
			return err
		}
		// remove it from the active persisters slice
		ps.activePersisters = ps.activePersisters[:ps.numOfActivePersisters]
		persisterToClose.isClosed = true
		epochToClose := epoch - ps.numOfActivePersisters
		ps.persistersMapByEpoch[epochToClose] = persisterToClose
	}

	if !ps.fullArchive && uint32(len(ps.persistersMapByEpoch)) > ps.numOfEpochsToKeep {
		epochToRemove := epoch - ps.numOfEpochsToKeep
		persisterToDestroy, ok := ps.persistersMapByEpoch[epochToRemove]
		if !ok {
			return errors.New("persister to destroy not found")
		}
		delete(ps.persistersMapByEpoch, epochToRemove)

		err := persisterToDestroy.persister.DestroyClosed()
		if err != nil {
			return err
		}
		removeDirectoryIfEmpty(persisterToDestroy.path)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ps *PruningStorer) IsInterfaceNil() bool {
	return ps == nil
}
