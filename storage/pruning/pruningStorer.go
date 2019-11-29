package pruning

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var log = logger.GetOrCreate("storage/pruning")

// DefaultEpochDirectoryName represents the naming pattern for epoch directories
const DefaultEpochDirectoryName = "Epoch"

// PruningStorer represents a storer which creates a new persister for each epoch and removes older persisters
type PruningStorer struct {
	lock                  sync.RWMutex
	fullArchive           bool
	batcher               storage.Batcher
	persisters            []storage.Persister
	cacher                storage.Cacher
	bloomFilter           storage.BloomFilter
	dbConf                storageUnit.DBConfig
	numOfActivePersisters uint32
	identifier            string
}

// NewPruningStorer will return a new instance of PruningStorer without sharded directories' naming scheme
func NewPruningStorer(
	identifier string,
	fullArchive bool,
	cacheConf storageUnit.CacheConfig,
	dbConf storageUnit.DBConfig,
	bloomFilterConf storageUnit.BloomConfig,
	numOfActivePersisters uint32,
	notifier EpochStartNotifier,
) (*PruningStorer, error) {
	return initPruningStorer(identifier, fullArchive, cacheConf, dbConf, bloomFilterConf, numOfActivePersisters, "", notifier)
}

// NewShardedPruningStorer will return a new instance of PruningStorer with sharded directories' naming scheme
func NewShardedPruningStorer(
	identifier string,
	fullArchive bool,
	cacheConf storageUnit.CacheConfig,
	dbConf storageUnit.DBConfig,
	bloomFilterConf storageUnit.BloomConfig,
	numOfActivePersisters uint32,
	shardId uint32,
	notifier EpochStartNotifier,
) (*PruningStorer, error) {
	shardIdStr := fmt.Sprintf("%d", shardId)
	return initPruningStorer(identifier, fullArchive, cacheConf, dbConf, bloomFilterConf, numOfActivePersisters, shardIdStr, notifier)
}

// initPruningStorer will create a PruningStorer with or without sharded directories' naming scheme
func initPruningStorer(
	identifier string,
	fullArchive bool,
	cacheConf storageUnit.CacheConfig,
	dbConf storageUnit.DBConfig,
	bloomFilterConf storageUnit.BloomConfig,
	numOfActivePersisters uint32,
	shardId string,
	epochStartNotifier EpochStartNotifier,
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

	if numOfActivePersisters < 1 {
		return nil, storage.ErrInvalidNumberOfPersisters
	}
	if check.IfNil(epochStartNotifier) {
		return nil, storage.ErrNilEpochStartNotifier
	}

	cache, err = storageUnit.NewCache(cacheConf.Type, cacheConf.Size, cacheConf.Shards)
	if err != nil {
		return nil, err
	}

	filePath := dbConf.FilePath
	if len(shardId) > 0 {
		filePath = filePath + shardId
	}
	db, err = storageUnit.NewDB(dbConf.Type, filePath, dbConf.BatchDelaySeconds, dbConf.MaxBatchSize, dbConf.MaxOpenFiles)
	if err != nil {
		return nil, err
	}

	var persisters []storage.Persister
	persisters = append(persisters, db)

	if reflect.DeepEqual(bloomFilterConf, storageUnit.BloomConfig{}) {
		pdb := &PruningStorer{
			identifier:            identifier,
			fullArchive:           fullArchive,
			persisters:            persisters,
			cacher:                cache,
			bloomFilter:           nil,
			dbConf:                dbConf,
			numOfActivePersisters: numOfActivePersisters,
		}
		err = pdb.persisters[0].Init()
		if err != nil {
			return nil, err
		}

		pdb.registerHandler(epochStartNotifier)

		return pdb, nil
	}

	bf, err = storageUnit.NewBloomFilter(bloomFilterConf)
	if err != nil {
		return nil, err
	}

	pdb := &PruningStorer{
		identifier:            identifier,
		persisters:            persisters,
		cacher:                cache,
		bloomFilter:           bf,
		dbConf:                dbConf,
		numOfActivePersisters: numOfActivePersisters,
	}

	err = pdb.persisters[0].Init()
	if err != nil {
		return nil, err
	}

	pdb.registerHandler(epochStartNotifier)

	return pdb, nil
}

// Put adds data to both cache and persistence medium and updates the bloom filter
func (pd *PruningStorer) Put(key, data []byte) error {
	pd.lock.Lock()
	defer pd.lock.Unlock()

	pd.cacher.Put(key, data)

	err := pd.persisters[0].Put(key, data)
	if err != nil {
		pd.cacher.Remove(key)
		return err
	}

	if pd.bloomFilter != nil {
		pd.bloomFilter.Add(key)
	}

	return err
}

// Get searches the key in the cache. In case it is not found, it searches
// for the key in bloom filter first and if found
// it further searches it in the associated databases.
// In case it is found in the database, the cache is updated with the value as well.
func (pd *PruningStorer) Get(key []byte) ([]byte, error) {
	pd.lock.Lock()
	defer pd.lock.Unlock()

	v, ok := pd.cacher.Get(key)
	var err error

	if !ok {
		// not found in cache
		// search it in second persistence medium
		found := false
		for _, persister := range pd.persisters {
			log.Info("pr db", "num persisters", len(pd.persisters))
			if pd.bloomFilter == nil || pd.bloomFilter.MayContain(key) == true {
				v, err = persister.Get(key)

				if err != nil {
					log.Debug(pd.identifier+" pruning db - get",
						"error", err.Error())
					continue
				}

				found = true
				// if found in persistence unit, add it in cache
				pd.cacher.Put(key, v)
				break
			}
		}
		if !found {
			return nil, errors.New(fmt.Sprintf("%s: key %s not found", pd.identifier, base64.StdEncoding.EncodeToString(key)))
		}
	}

	return v.([]byte), nil
}

// Has checks if the key is in the Unit.
// It first checks the cache. If it is not found, it checks the bloom filter
// and if present it checks the db
func (pd *PruningStorer) Has(key []byte) error {
	pd.lock.RLock()
	defer pd.lock.RUnlock()

	has := pd.cacher.Has(key)
	if has {
		return nil
	}

	if pd.bloomFilter == nil || pd.bloomFilter.MayContain(key) == true {
		for _, persister := range pd.persisters {
			if persister.Has(key) != nil {
				continue
			}

			return nil
		}
	}

	return storage.ErrKeyNotFound
}

// Remove removes the data associated to the given key from both cache and persistence medium
func (pd *PruningStorer) Remove(key []byte) error {
	pd.lock.Lock()
	defer pd.lock.Unlock()

	pd.cacher.Remove(key)
	return pd.persisters[0].Remove(key)
}

// ClearCache cleans up the entire cache
func (pd *PruningStorer) ClearCache() {
	pd.cacher.Clear()
}

// DestroyUnit cleans up the bloom filter, the cache, and the dbs
func (pd *PruningStorer) DestroyUnit() error {
	pd.lock.Lock()
	defer pd.lock.Unlock()

	if pd.bloomFilter != nil {
		pd.bloomFilter.Clear()
	}

	pd.cacher.Clear()

	var err error
	numOfPersistersRemoved := 0
	totalNumOfPersisters := len(pd.persisters)
	for _, persister := range pd.persisters {
		err = persister.Destroy()
		if err != nil {
			log.Debug("pruning db: destroy",
				"error", err.Error())
			continue
		}
		numOfPersistersRemoved++
	}

	if numOfPersistersRemoved != totalNumOfPersisters {
		return errors.New(fmt.Sprintf("couldn't destroy all persisters. %d/%d destroyed",
			numOfPersistersRemoved,
			totalNumOfPersisters,
		))
	}
	return pd.persisters[0].Destroy()
}

// registerHandler will register a new function to the epoch start notifier
func (pd *PruningStorer) registerHandler(handler EpochStartNotifier) {
	subscribeHandler := notifier.MakeHandlerForEpochStart(func(hdr data.HeaderHandler) {
		err := pd.changeEpoch(hdr.GetEpoch())
		if err != nil {
			log.Warn("change epoch in storer", "error", err.Error())
		}
	})

	handler.RegisterHandler(subscribeHandler)
}

// changeEpoch will handle creating a new persister and removing of the older ones
func (pd *PruningStorer) changeEpoch(epoch uint32) error {
	filePath := pd.getNewFilePath(epoch)
	db, err := storageUnit.NewDB(pd.dbConf.Type, filePath, pd.dbConf.BatchDelaySeconds, pd.dbConf.MaxBatchSize, pd.dbConf.MaxOpenFiles)
	if err != nil {
		return err
	}

	pd.lock.Lock()
	singleItemPersisters := []storage.Persister{db}
	pd.persisters = append(singleItemPersisters, pd.persisters...)
	err = pd.persisters[0].Init()
	if err != nil {
		return err
	}

	if !pd.fullArchive {
		// remove the older persisters
		for idx := pd.numOfActivePersisters; idx < uint32(len(pd.persisters)); idx++ {
			err = pd.persisters[int(idx)].Destroy()
			if err != nil {
				log.Debug("pruning db: remove old dbs",
					"error", err.Error())
			} else {
				//essh.epochStartHandlers = append(essh.epochStartHandlers[:idx], essh.epochStartHandlers[idx+1:]...)
				pd.persisters = append(pd.persisters[:idx], pd.persisters[idx+1:]...)
			}
		}
	}

	pd.lock.Unlock()

	return nil
}

// getNewFilePath will return the file path for the new epoch. It uses regex to change the default path
func (pd *PruningStorer) getNewFilePath(epoch uint32) string {
	// using a regex to match the epoch directory name as placeholder followed by at least one digit
	rg := regexp.MustCompile("Epoch_\\d+")
	newEpochDirectoryName := fmt.Sprintf("%s_%d", DefaultEpochDirectoryName, epoch)
	return rg.ReplaceAllString(pd.dbConf.FilePath, newEpochDirectoryName)
}

// IsInterfaceNil -
func (pd *PruningStorer) IsInterfaceNil() bool {
	return pd == nil
}
