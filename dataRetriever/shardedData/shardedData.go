package shardedData

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var log = logger.DefaultLogger()

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
	//  data hashes that have that shard as destination
	shardedDataStore map[string]*shardStore
	cacherConfig     storageUnit.CacheConfig

	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func(key []byte)
}

type shardStore struct {
	CacheID   string
	DataStore storage.Cacher
}

// NewShardedData is responsible for creating an empty pool of data
func NewShardedData(cacherConfig storageUnit.CacheConfig) (*shardedData, error) {
	err := verifyCacherConfig(cacherConfig)
	if err != nil {
		return nil, err
	}

	return &shardedData{
		cacherConfig:         cacherConfig,
		mutShardedDataStore:  sync.RWMutex{},
		shardedDataStore:     make(map[string]*shardStore),
		mutAddedDataHandlers: sync.RWMutex{},
		addedDataHandlers:    make([]func(key []byte), 0),
	}, nil
}

func verifyCacherConfig(cacherConfig storageUnit.CacheConfig) error {
	_, err := newShardStore("", cacherConfig)
	return err
}

// newShardStore is responsible for creating an empty shardStore
func newShardStore(cacheId string, cacherConfig storageUnit.CacheConfig) (*shardStore, error) {
	cacher, err := storageUnit.NewCache(cacherConfig.Type, cacherConfig.Size, cacherConfig.Shards)
	if err != nil {
		return nil, err
	}
	return &shardStore{
		CacheID:   cacheId,
		DataStore: cacher,
	}, nil
}

// CreateShardStore is a ShardedData method that is responsible for creating
//  a new shardStore with cacheId index in the shardedDataStore map
func (sd *shardedData) CreateShardStore(cacheId string) {
	sd.mutShardedDataStore.Lock()
	sd.newShardStoreNoLock(cacheId)
	sd.mutShardedDataStore.Unlock()
}

func (sd *shardedData) newShardStoreNoLock(cacheId string) *shardStore {
	shardStore, err := newShardStore(cacheId, sd.cacherConfig)
	log.LogIfError(err)

	sd.shardedDataStore[cacheId] = shardStore
	return shardStore
}

// ShardStore returns a shard store of data associated with a given destination cacheId
func (sd *shardedData) ShardStore(cacheId string) *shardStore {
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
func (sd *shardedData) AddData(key []byte, data interface{}, cacheId string) {
	var mp *shardStore

	sd.mutShardedDataStore.Lock()
	mp = sd.shardedDataStore[cacheId]
	if mp == nil {
		mp = sd.newShardStoreNoLock(cacheId)
	}
	sd.mutShardedDataStore.Unlock()

	found, _ := mp.DataStore.HasOrAdd(key, data)

	if !found {
		sd.mutAddedDataHandlers.RLock()
		for _, handler := range sd.addedDataHandlers {
			go handler(key)
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
			val, _ := sourceStore.Get(key)
			sd.AddData(key, val, destCacheId)
		}
	}

	sd.mutShardedDataStore.Lock()
	delete(sd.shardedDataStore, sourceCacheId)
	sd.mutShardedDataStore.Unlock()
}

// MoveData will move all given data associated with the sourceCacheId to the destCacheId
func (sd *shardedData) MoveData(sourceCacheId, destCacheId string, key [][]byte) {
	sourceStore := sd.ShardDataStore(sourceCacheId)

	if sourceStore != nil {
		for _, key := range key {
			val, ok := sourceStore.Get(key)
			if ok {
				sd.AddData(key, val, destCacheId)
				sd.RemoveData(key, sourceCacheId)
			}
		}
	}
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
func (sd *shardedData) RegisterHandler(handler func(key []byte)) {
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
	if sd == nil {
		return true
	}
	return false
}
