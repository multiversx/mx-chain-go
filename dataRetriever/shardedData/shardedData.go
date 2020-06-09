package shardedData

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/immunitycache"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var log = logger.GetOrCreate("dataretriever/shardeddata")

var _ dataRetriever.ShardedDataCacherNotifier = (*shardedData)(nil)

// shardedData holds the list of data organised by destination shard
//
//  The shardStores field maps a cacher, containing data
//  hashes, to a corresponding identifier. It is able to add or remove
//  data given the shard id it is associated with. It can
//  also merge and split pools when required
type shardedData struct {
	name                string
	mutShardedDataStore sync.RWMutex
	// shardedDataStore is a key value store
	// Each key represents a destination shard id and the value will contain all
	// data hashes that have that shard as destination
	shardedDataStore map[string]*shardStore
	configPrototype  immunitycache.CacheConfig

	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func(key []byte, value interface{})
}

type shardStore struct {
	cacheID string
	cache   immunityCache
}

// NewShardedData is responsible for creating an empty pool of data
func NewShardedData(name string, config storageUnit.CacheConfig) (*shardedData, error) {
	log.Info("NewShardedData", "name", name, "config", config.String())

	configPrototype := immunitycache.CacheConfig{
		Name:                        "prototype",
		NumChunks:                   config.Shards,
		MaxNumItems:                 config.Capacity,
		MaxNumBytes:                 uint32(config.SizeInBytes),
		NumItemsToPreemptivelyEvict: dataRetriever.TxPoolNumTxsToPreemptivelyEvict,
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

func (sd *shardedData) newShardStore(cacheId string) (*shardStore, error) {
	config := sd.configPrototype
	config.Name = cacheId
	cache, err := immunitycache.NewImmunityCache(config)
	if err != nil {
		return nil, err
	}

	return &shardStore{
		cacheID: cacheId,
		cache:   cache,
	}, nil
}

func (sd *shardedData) newShardStoreNoLock(cacheId string) *shardStore {
	shardStoreObject, err := sd.newShardStore(cacheId)
	if err != nil {
		log.Error("newShardStoreNoLock", "error", err.Error())
	}

	sd.shardedDataStore[cacheId] = shardStoreObject
	return shardStoreObject
}

func (sd *shardedData) shardStore(cacheId string) *shardStore {
	sd.mutShardedDataStore.RLock()
	mp := sd.shardedDataStore[cacheId]
	sd.mutShardedDataStore.RUnlock()
	return mp
}

// ShardDataStore returns the shard data store containing data hashes
//  associated with a given destination cacheId
func (sd *shardedData) ShardDataStore(cacheId string) (c storage.Cacher) {
	mp := sd.shardStore(cacheId)
	if mp == nil {
		return nil
	}
	return mp.cache
}

// AddData will add data to the corresponding shard store
func (sd *shardedData) AddData(key []byte, value interface{}, sizeInBytes int, cacheID string) {
	log.Trace("shardedData.AddData()", "name", sd.name, "cacheID", cacheID)

	var mp *shardStore

	sd.mutShardedDataStore.Lock()
	mp = sd.shardedDataStore[cacheID]
	if mp == nil {
		mp = sd.newShardStoreNoLock(cacheID)
	}
	sd.mutShardedDataStore.Unlock()

	_, added := mp.cache.HasOrAdd(key, value, sizeInBytes)
	if added {
		sd.mutAddedDataHandlers.RLock()
		for _, handler := range sd.addedDataHandlers {
			go handler(key, value)
		}
		sd.mutAddedDataHandlers.RUnlock()
	}
}

// SearchFirstData searches the key against all shard data store, retrieving first value found
func (sd *shardedData) SearchFirstData(key []byte) (value interface{}, ok bool) {
	sd.mutShardedDataStore.RLock()
	for k := range sd.shardedDataStore {
		m := sd.shardedDataStore[k]
		if m == nil || m.cache == nil {
			continue
		}

		if m.cache.Has(key) {
			value, _ = m.cache.Peek(key)
			sd.mutShardedDataStore.RUnlock()
			return value, true
		}
	}
	sd.mutShardedDataStore.RUnlock()

	return nil, false
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

	log.Debug("shardedData.removeTxBulk()", "name", sd.name, "cacheID", cacheID, "numToRemove", len(keys), "numRemoved", numRemoved)
}

// ImmunizeSetOfDataAgainstEviction  marks the items as non-evictable (if the underlying cache supports this operation)
func (sd *shardedData) ImmunizeSetOfDataAgainstEviction(keys [][]byte, cacheID string) {
	store := sd.shardStore(cacheID)
	if store == nil {
		return
	}

	numNow, numFuture := store.cache.ImmunizeKeys(keys)
	log.Debug("shardedData.ImmunizeSetOfDataAgainstEviction()", "name", sd.name, "cacheID", cacheID, "len(keys)", len(keys), "numNow", numNow, "numFuture", numFuture)
}

// RemoveData will remove data hash from the corresponding shard store
func (sd *shardedData) RemoveData(key []byte, cacheId string) {
	mpdata := sd.ShardDataStore(cacheId)

	if mpdata != nil {
		mpdata.Remove(key)
	}
}

// RemoveDataFromAllShards will remove data from the store given only
//  the data hash. It will iterate over all shard store map and will remove it everywhere
func (sd *shardedData) RemoveDataFromAllShards(key []byte) {
	sd.mutShardedDataStore.RLock()
	for k := range sd.shardedDataStore {
		m := sd.shardedDataStore[k]
		if m == nil || m.cache == nil {
			continue
		}

		if m.cache.Has(key) {
			m.cache.Remove(key)
		}
	}
	sd.mutShardedDataStore.RUnlock()
}

// MergeShardStores will take all data associated with the sourceCacheId and move them
// to the destCacheId. It will then remove the sourceCacheId key from the store map
func (sd *shardedData) MergeShardStores(sourceCacheId, destCacheId string) {
	sourceStore := sd.ShardDataStore(sourceCacheId)

	if sourceStore != nil {
		for _, key := range sourceStore.Keys() {
			val, ok := sourceStore.Get(key)
			if !ok {
				log.Warn("programming error in shardedData: Keys() function reported a key that can not be retrieved")
				continue
			}

			valSizer, ok := val.(marshal.Sizer)
			if !ok {
				log.Warn("programming error in shardedData, objects contained are not of type marshal.Sizer")
				continue
			}

			sd.AddData(key, valSizer, valSizer.Size(), destCacheId)
		}
	}

	sd.mutShardedDataStore.Lock()
	delete(sd.shardedDataStore, sourceCacheId)
	sd.mutShardedDataStore.Unlock()
}

// Clear will delete all shard stores and associated data
func (sd *shardedData) Clear() {
	sd.mutShardedDataStore.Lock()
	for m := range sd.shardedDataStore {
		delete(sd.shardedDataStore, m)
	}
	sd.mutShardedDataStore.Unlock()
}

// ClearShardStore will delete all data associated with a given destination cacheId
func (sd *shardedData) ClearShardStore(cacheId string) {
	mp := sd.ShardDataStore(cacheId)
	if mp == nil {
		return
	}
	mp.Clear()
}

// RegisterHandler registers a new handler to be called when a new data is added
func (sd *shardedData) RegisterHandler(handler func(key []byte, value interface{})) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a ShardedData object")
		return
	}

	sd.mutAddedDataHandlers.Lock()
	sd.addedDataHandlers = append(sd.addedDataHandlers, handler)
	sd.mutAddedDataHandlers.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (sd *shardedData) IsInterfaceNil() bool {
	return sd == nil
}
