package shardedData

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/counting"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("dataretriever/shardeddata")

var _ dataRetriever.ShardedDataCacherNotifier = (*shardedData)(nil)

const untitledCacheName = "untitled"

// shardedData holds the list of data organised by destination shard
//
//	The shardStores field maps a cacher, containing data
//	hashes, to a corresponding identifier. It is able to add or remove
//	data given the shard id it is associated with. It can
//	also merge and split pools when required
type shardedData struct {
	name                string
	mutShardedDataStore sync.RWMutex
	// shardedDataStore is a key value store
	// Each key represents a destination shard id and the value will contain all
	// data hashes that have that shard as destination
	shardedDataStore map[string]*shardStore
	configPrototype  cache.CacheConfig

	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func(key []byte, value interface{})
}

type shardStore struct {
	cacheID string
	cache   immunityCache
}

// NewShardedData is responsible for creating an empty pool of data
func NewShardedData(name string, config storageunit.CacheConfig) (*shardedData, error) {
	log.Debug("NewShardedData", "name", name, "config", config.String())

	configPrototype := cache.CacheConfig{
		Name:                        untitledCacheName,
		NumChunks:                   config.Shards,
		MaxNumItems:                 config.Capacity,
		MaxNumBytes:                 uint32(config.SizeInBytes),
		NumItemsToPreemptivelyEvict: storage.TxPoolNumTxsToPreemptivelyEvict,
	}

	err := configPrototype.Verify()
	if err != nil {
		return nil, err
	}

	return &shardedData{
		name:              name,
		configPrototype:   configPrototype,
		shardedDataStore:  make(map[string]*shardStore),
		addedDataHandlers: make([]func(key []byte, value interface{}), 0),
	}, nil
}

// ShardDataStore returns the shard data store containing data hashes
// associated with a given destination cacheID
func (sd *shardedData) ShardDataStore(cacheID string) (c storage.Cacher) {
	store := sd.shardStore(cacheID)
	if store == nil {
		return nil
	}

	return store.cache
}

func (sd *shardedData) shardStore(cacheID string) *shardStore {
	sd.mutShardedDataStore.RLock()
	store := sd.shardedDataStore[cacheID]
	sd.mutShardedDataStore.RUnlock()
	return store
}

// AddData will add data to the corresponding shard store
func (sd *shardedData) AddData(key []byte, value interface{}, sizeInBytes int, cacheID string) {
	log.Trace("shardedData.AddData()", "name", sd.name, "cacheID", cacheID, "key", key, "size", sizeInBytes)

	store := sd.getOrCreateShardStoreWithLock(cacheID)

	_, added := store.cache.HasOrAdd(key, value, sizeInBytes)
	if added {
		sd.mutAddedDataHandlers.RLock()
		for _, handler := range sd.addedDataHandlers {
			handler(key, value)
		}
		sd.mutAddedDataHandlers.RUnlock()
	}
}

func (sd *shardedData) getOrCreateShardStoreWithLock(cacheID string) *shardStore {
	sd.mutShardedDataStore.Lock()
	defer sd.mutShardedDataStore.Unlock()

	store, ok := sd.shardedDataStore[cacheID]
	if !ok {
		store = sd.addShardStoreNoLock(cacheID)
	}

	return store
}

func (sd *shardedData) addShardStoreNoLock(cacheID string) *shardStore {
	store, err := sd.newShardStore(cacheID)
	if err != nil {
		log.Error("addShardStoreNoLock", "error", err.Error())
		return nil
	}

	sd.shardedDataStore[cacheID] = store
	return store
}

func (sd *shardedData) newShardStore(cacheID string) (*shardStore, error) {
	config := sd.configPrototype
	config.Name = fmt.Sprintf("%s:%s", sd.name, cacheID)
	newImmunityCache, err := cache.NewImmunityCache(config)
	if err != nil {
		return nil, err
	}

	return &shardStore{
		cacheID: cacheID,
		cache:   newImmunityCache,
	}, nil
}

// SearchFirstData searches the key against all shard data store, retrieving first value found
func (sd *shardedData) SearchFirstData(key []byte) (value interface{}, ok bool) {
	sd.mutShardedDataStore.RLock()
	defer sd.mutShardedDataStore.RUnlock()

	for _, store := range sd.shardedDataStore {
		value, ok = store.cache.Peek(key)
		if ok {
			return
		}
	}

	return
}

// RemoveSetOfDataFromPool removes a list of keys from the corresponding pool
func (sd *shardedData) RemoveSetOfDataFromPool(keys [][]byte, cacheID string) {
	store := sd.shardStore(cacheID)
	if store == nil {
		return
	}

	numRemoved := 0
	for _, key := range keys {
		if store.cache.RemoveWithResult(key) {
			numRemoved++
		}
	}

	log.Trace("shardedData.removeTxBulk()", "name", sd.name, "cacheID", cacheID, "numToRemove", len(keys), "numRemoved", numRemoved)
}

// ImmunizeSetOfDataAgainstEviction  marks the items as non-evictable
func (sd *shardedData) ImmunizeSetOfDataAgainstEviction(keys [][]byte, cacheID string) {
	store := sd.getOrCreateShardStoreWithLock(cacheID)
	numNow, numFuture := store.cache.ImmunizeKeys(keys)
	log.Trace("shardedData.ImmunizeSetOfDataAgainstEviction()", "name", sd.name, "cacheID", cacheID, "len(keys)", len(keys), "numNow", numNow, "numFuture", numFuture)
}

// RemoveData will remove data hash from the corresponding shard store
func (sd *shardedData) RemoveData(key []byte, cacheID string) {
	store := sd.shardStore(cacheID)
	if store == nil {
		return
	}

	store.cache.Remove(key)
}

// RemoveDataFromAllShards will remove data from the store given only
//
//	the data hash. It will iterate over all shard store map and will remove it everywhere
func (sd *shardedData) RemoveDataFromAllShards(key []byte) {
	sd.mutShardedDataStore.RLock()
	defer sd.mutShardedDataStore.RUnlock()

	for _, store := range sd.shardedDataStore {
		store.cache.Remove(key)
	}
}

// MergeShardStores will take all data associated with the sourceCacheId and move them
// to the destCacheId. It will then remove the sourceCacheId key from the store map
func (sd *shardedData) MergeShardStores(sourceCacheID, destCacheID string) {
	sourceStore := sd.shardStore(sourceCacheID)

	if sourceStore != nil {
		for _, key := range sourceStore.cache.Keys() {
			val, ok := sourceStore.cache.Get(key)
			if !ok {
				log.Warn("programming error in shardedData: Keys() function reported a key that can not be retrieved")
				continue
			}

			valSizer, ok := val.(marshal.Sizer)
			if !ok {
				log.Warn("programming error in shardedData, objects contained are not of type marshal.Sizer")
				continue
			}

			sd.AddData(key, valSizer, valSizer.Size(), destCacheID)
		}
	}

	sd.mutShardedDataStore.Lock()
	delete(sd.shardedDataStore, sourceCacheID)
	sd.mutShardedDataStore.Unlock()
}

// Clear will delete all shard stores and associated data
func (sd *shardedData) Clear() {
	sd.mutShardedDataStore.Lock()
	sd.shardedDataStore = make(map[string]*shardStore)
	sd.mutShardedDataStore.Unlock()
}

// ClearShardStore will delete all data associated with a given destination cacheID
func (sd *shardedData) ClearShardStore(cacheID string) {
	store := sd.shardStore(cacheID)
	if store == nil {
		return
	}

	store.cache.Clear()
}

// RegisterOnAdded registers a new handler to be called when a new data is added
func (sd *shardedData) RegisterOnAdded(handler func(key []byte, value interface{})) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a ShardedData object")
		return
	}

	sd.mutAddedDataHandlers.Lock()
	sd.addedDataHandlers = append(sd.addedDataHandlers, handler)
	sd.mutAddedDataHandlers.Unlock()
}

// Keys returns all the keys contained in shard caches
func (sd *shardedData) Keys() [][]byte {
	sd.mutShardedDataStore.RLock()
	defer sd.mutShardedDataStore.RUnlock()

	keys := make([][]byte, 0)
	for _, shard := range sd.shardedDataStore {
		c := shard.cache
		keys = append(keys, c.Keys()...)
	}

	return keys
}

// GetCounts returns the total number of transactions in the pool
func (sd *shardedData) GetCounts() counting.CountsWithSize {
	sd.mutShardedDataStore.RLock()
	defer sd.mutShardedDataStore.RUnlock()

	counts := counting.NewConcurrentShardedCountsWithSize()

	for cacheID, shard := range sd.shardedDataStore {
		c := shard.cache
		counts.PutCounts(cacheID, int64(c.Len()), int64(c.NumBytes()))
	}

	return counts
}

// Diagnose diagnoses the internal caches
func (sd *shardedData) Diagnose(deep bool) {
	log.Trace("shardedData.Diagnose()", "counts", sd.GetCounts().String())

	sd.mutShardedDataStore.RLock()
	defer sd.mutShardedDataStore.RUnlock()

	for _, shard := range sd.shardedDataStore {
		shard.cache.Diagnose(deep)
	}
}

// Close returns nil on this implementation
func (sd *shardedData) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sd *shardedData) IsInterfaceNil() bool {
	return sd == nil
}
