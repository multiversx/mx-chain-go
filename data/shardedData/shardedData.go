package shardedData

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.NewDefaultLogger()

// shardedData holds the list of data organised by destination shard
//
// The shardStores field maps a cacher, containing data
//  hashes, to a corresponding shard id. It is able to add or remove
//  data given the shard id it is associated with. It can
//  also merge and split pools when required
type shardedData struct {
	mutShardedDataStore sync.RWMutex
	// shardedDataStore is a key value store
	// Each key represents a destination shard id and the value will contain all
	//  data hashes that have that shard as destination
	shardedDataStore map[uint32]*shardStore
	cacherConfig     storage.CacheConfig

	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func(key []byte)
}

type shardStore struct {
	ShardID   uint32
	DataStore storage.Cacher
}

// NewShardedData is responsible for creating an empty pool of data
func NewShardedData(cacherConfig storage.CacheConfig) (*shardedData, error) {
	err := verifyCacherConfig(cacherConfig)
	if err != nil {
		return nil, err
	}

	return &shardedData{
		cacherConfig:         cacherConfig,
		mutShardedDataStore:  sync.RWMutex{},
		shardedDataStore:     make(map[uint32]*shardStore),
		mutAddedDataHandlers: sync.RWMutex{},
		addedDataHandlers:    make([]func(key []byte), 0),
	}, nil
}

func verifyCacherConfig(cacherConfig storage.CacheConfig) error {
	_, err := newShardStore(0, cacherConfig)
	return err
}

// newShardStore is responsible for creating an empty shardStore
func newShardStore(destShardID uint32, cacherConfig storage.CacheConfig) (*shardStore, error) {
	cacher, err := storage.NewCache(cacherConfig.Type, cacherConfig.Size)
	if err != nil {
		return nil, err
	}
	return &shardStore{
		destShardID,
		cacher,
	}, nil
}

// CreateShardStore is a ShardedData method that is responsible for creating
//  a new shardStore at the destShardID index in the shardedDataStore map
func (sd *shardedData) CreateShardStore(destShardID uint32) {
	sd.mutShardedDataStore.Lock()
	sd.newShardStoreNoLock(destShardID)
	sd.mutShardedDataStore.Unlock()
}

func (sd *shardedData) newShardStoreNoLock(destShardID uint32) *shardStore {
	shardStore, err := newShardStore(destShardID, sd.cacherConfig)
	log.LogIfError(err)

	sd.shardedDataStore[destShardID] = shardStore
	return shardStore
}

// ShardStore returns a shard store of data associated with a given destination shardID
func (sd *shardedData) ShardStore(shardID uint32) *shardStore {
	sd.mutShardedDataStore.RLock()
	mp := sd.shardedDataStore[shardID]
	sd.mutShardedDataStore.RUnlock()
	return mp
}

// ShardDataStore returns the shard data store containing data hashes
//  associated with a given destination shardID
func (sd *shardedData) ShardDataStore(shardID uint32) (c storage.Cacher) {
	mp := sd.ShardStore(shardID)
	if mp == nil {
		return nil
	}
	return mp.DataStore
}

// AddData will add data to the corresponding shard store
func (sd *shardedData) AddData(key []byte, data interface{}, destShardID uint32) {
	var mp *shardStore

	sd.mutShardedDataStore.Lock()
	mp = sd.shardedDataStore[destShardID]
	if mp == nil {
		mp = sd.newShardStoreNoLock(destShardID)
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
func (sd *shardedData) RemoveSetOfDataFromPool(keys [][]byte, destShardID uint32) {
	for _, key := range keys {
		sd.RemoveData(key, destShardID)
	}
}

// RemoveData will remove data hash from the corresponding shard store
func (sd *shardedData) RemoveData(key []byte, destShardID uint32) {
	mpdata := sd.ShardDataStore(destShardID)

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

// MergeShardStores will take all data associated with the sourceShardId and move them
// to the destShardID. It will then remove the sourceShardID key from the store map
func (sd *shardedData) MergeShardStores(sourceShardID, destShardID uint32) {
	sourceStore := sd.ShardDataStore(sourceShardID)

	if sourceStore != nil {
		for _, key := range sourceStore.Keys() {
			val, _ := sourceStore.Get(key)
			sd.AddData(key, val, destShardID)
		}
	}

	sd.mutShardedDataStore.Lock()
	delete(sd.shardedDataStore, sourceShardID)
	sd.mutShardedDataStore.Unlock()
}

// MoveData will move all given data associated with the sourceShardId to the destShardId
func (sd *shardedData) MoveData(sourceShardID, destShardID uint32, key [][]byte) {
	sourceStore := sd.ShardDataStore(sourceShardID)

	if sourceStore != nil {
		for _, key := range key {
			val, ok := sourceStore.Get(key)
			if ok {
				sd.AddData(key, val, destShardID)
				sd.RemoveData(key, sourceShardID)
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

// ClearShardStore will delete all data associated with a given destination shardID
func (sd *shardedData) ClearShardStore(shardID uint32) {
	mp := sd.ShardDataStore(shardID)
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
