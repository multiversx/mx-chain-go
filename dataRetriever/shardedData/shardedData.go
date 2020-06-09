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
// The shardStores field maps a cacher, containing data
//  hashes, to a corresponding identifier. It is able to add or remove
//  data given the shard id it is associated with. It can
//  also merge and split pools when required
type shardedData struct {
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
	CacheID   string
	DataStore storage.Cacher
}

// NewShardedData is responsible for creating an empty pool of data
func NewShardedData(config storageUnit.CacheConfig) (*shardedData, error) {
	log.Info("NewShardedData", "config", config.String())

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
		configPrototype:   configPrototype,
		shardedDataStore:  make(map[string]*shardStore),
		addedDataHandlers: make([]func(key []byte, value interface{}), 0),
	}, nil
}

func (sd *shardedData) newShardStore(cacheId string) (*shardStore, error) {
	config := sd.configPrototype
	config.Name = cacheId
	cacher, err := immunitycache.NewImmunityCache(config)
	if err != nil {
		return nil, err
	}

	return &shardStore{
		CacheID:   cacheId,
		DataStore: cacher,
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

// ShardStore returns a shard store of data associated with a given destination cacheId
func (sd *shardedData) ShardStore(cacheId string) *shardStore {
	// TODO: remove from interface
	sd.mutShardedDataStore.RLock()
	mp := sd.shardedDataStore[cacheId]
	sd.mutShardedDataStore.RUnlock()
	return mp
}

// ShardDataStore returns the shard data store containing data hashes
//  associated with a given destination cacheId
func (sd *shardedData) ShardDataStore(cacheId string) (c storage.Cacher) {
	mp := sd.ShardStore(cacheId)
	if mp == nil {
		return nil
	}
	return mp.DataStore
}

// AddData will add data to the corresponding shard store
func (sd *shardedData) AddData(key []byte, value interface{}, sizeInBytes int, cacheId string) {
	var mp *shardStore

	sd.mutShardedDataStore.Lock()
	mp = sd.shardedDataStore[cacheId]
	if mp == nil {
		mp = sd.newShardStoreNoLock(cacheId)
	}
	sd.mutShardedDataStore.Unlock()

	_, added := mp.DataStore.HasOrAdd(key, value, sizeInBytes)
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
		if m == nil || m.DataStore == nil {
			continue
		}

		if m.DataStore.Has(key) {
			value, _ = m.DataStore.Peek(key)
			sd.mutShardedDataStore.RUnlock()
			return value, true
		}
	}
	sd.mutShardedDataStore.RUnlock()

	return nil, false
}

// RemoveSetOfDataFromPool removes a list of keys from the corresponding pool
func (sd *shardedData) RemoveSetOfDataFromPool(keys [][]byte, cacheId string) {
	for _, key := range keys {
		sd.RemoveData(key, cacheId)
	}
}

// ImmunizeSetOfDataAgainstEviction  marks the items as non-evictable (if the underlying cache supports this operation)
func (sd *shardedData) ImmunizeSetOfDataAgainstEviction(keys [][]byte, cacheID string) {
	cache := sd.ShardDataStore(cacheID)
	if cache == nil {
		return
	}

	cacheAsImmunityCache, ok := cache.(immunityCache)
	if !ok {
		log.Error("shardedData.ImmunizeSetOfDataAgainstEviction(): programming error, cache is not of expected type")
		return
	}

	cacheAsImmunityCache.ImmunizeTxsAgainstEviction(keys)
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
		if m == nil || m.DataStore == nil {
			continue
		}

		if m.DataStore.Has(key) {
			m.DataStore.Remove(key)
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
