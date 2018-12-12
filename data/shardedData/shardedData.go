package shardedData

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.NewDefaultLogger()

var defaultCacherConfig = &storage.CacheConfig{
	Size: 1000,
	Type: storage.LRUCache,
}

// ShardedData holds the list of data organised by destination shard
//
// The shardStores field maps a cacher, containing data
//  hashes, to a corresponding shard id. It is able to add or remove
//  data given the shard id it is associated with. It can
//  also merge and split pools when required
type ShardedData struct {
	// shardStores is a key value store
	// Each key represents a destination shard id and the value will contain all
	//  data hashes that have that shard as destination
	mutShardedDataStore sync.RWMutex
	shardedDataStore    map[uint32]*shardStore
	cacherConfig        *storage.CacheConfig

	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func(key []byte)
}

type shardStore struct {
	ShardID   uint32
	DataStore storage.Cacher
}

// NewShardedData is responsible for creating an empty pool of data
func NewShardedData(cacherConfig *storage.CacheConfig) *ShardedData {
	if cacherConfig == nil {
		cacherConfig = defaultCacherConfig
	}
	return &ShardedData{
		cacherConfig:         cacherConfig,
		mutShardedDataStore:  sync.RWMutex{},
		shardedDataStore:     make(map[uint32]*shardStore),
		mutAddedDataHandlers: sync.RWMutex{},
		addedDataHandlers:    make([]func(key []byte), 0),
	}
}

// newShardStore is responsible for creating an empty shardStore
func newShardStore(destShardID uint32, cacherConfig *storage.CacheConfig) *shardStore {
	if cacherConfig == nil {
		cacherConfig = defaultCacherConfig
	}
	cacher, err := storage.NewCache(cacherConfig.Type, cacherConfig.Size)
	if err != nil {
		// TODO: This should be replaced with the correct log panic
		panic("Could not create cache storage for pools")
	}
	return &shardStore{
		destShardID,
		cacher,
	}
}

// NewShardStore is a ShardedData method that is responsible for creating
//  a new shardStore at the destShardID index in the shardedDataStore map
func (sd *ShardedData) NewShardStore(destShardID uint32) {
	sd.mutShardedDataStore.Lock()
	sd.shardedDataStore[destShardID] = newShardStore(destShardID, sd.cacherConfig)
	sd.mutShardedDataStore.Unlock()
}

// ShardStore returns a shard store of data associated with a given destination shardID
func (sd *ShardedData) ShardStore(shardID uint32) *shardStore {
	sd.mutShardedDataStore.RLock()
	mp := sd.shardedDataStore[shardID]
	sd.mutShardedDataStore.RUnlock()
	return mp
}

// ShardDataStore returns the shard data store containing data hashes
//  associated with a given destination shardID
func (sd *ShardedData) ShardDataStore(shardID uint32) (c storage.Cacher) {
	mp := sd.ShardStore(shardID)
	if mp == nil {
		return nil
	}
	return mp.DataStore
}

// AddData will add data to the corresponding shard store
func (sd *ShardedData) AddData(key []byte, data interface{}, destShardID uint32) {
	if sd.ShardStore(destShardID) == nil {
		sd.NewShardStore(destShardID)
	}
	mp := sd.ShardDataStore(destShardID)
	found, _ := mp.HasOrAdd(key, data)

	if !found {
		sd.mutAddedDataHandlers.RLock()
		for _, handler := range sd.addedDataHandlers {
			go handler(key)
		}
		sd.mutAddedDataHandlers.RUnlock()
	}
}

// SearchData searches the key against all shard data store, retrieving found data in a map
func (sd *ShardedData) SearchData(key []byte) (shardValuesPairs map[uint32]interface{}) {
	shardValuesPairs = make(map[uint32]interface{})

	sd.mutShardedDataStore.RLock()
	for k := range sd.shardedDataStore {
		m := sd.ShardDataStore(k)
		if m.Has(key) {
			shardValuesPairs[k], _ = m.Get(key)
		}
	}
	sd.mutShardedDataStore.RUnlock()

	return shardValuesPairs
}

// RemoveData will remove data hash from the corresponding shard store
func (sd *ShardedData) RemoveData(key []byte, destShardID uint32) {
	sd.mutShardedDataStore.RLock()
	mpdata := sd.ShardDataStore(destShardID)
	sd.mutShardedDataStore.RUnlock()

	if mpdata != nil {
		mpdata.Remove(key)
	}
}

// RemoveDataFromAllShards will remove data from the store given only
//  the data hash. It will iterate over all shard store map and will remove it everywhere
func (sd *ShardedData) RemoveDataFromAllShards(key []byte) {
	sd.mutShardedDataStore.RLock()
	for k := range sd.shardedDataStore {
		m := sd.ShardDataStore(k)
		if m.Has(key) {
			m.Remove(key)
		}
	}
	sd.mutShardedDataStore.RUnlock()
}

// MergeShardStores will take all data associated with the sourceShardId and move them
// to the destShardID. It will then remove the sourceShardID key from the store map
func (sd *ShardedData) MergeShardStores(sourceShardID, destShardID uint32) {
	sourceStore := sd.ShardDataStore(sourceShardID)

	if sourceStore != nil {
		for _, key := range sourceStore.Keys() {
			data, _ := sourceStore.Get(key)
			sd.AddData(key, data, destShardID)
		}
	}

	sd.mutShardedDataStore.Lock()
	delete(sd.shardedDataStore, sourceShardID)
	sd.mutShardedDataStore.Unlock()
}

// MoveData will move all given data associated with the sourceShardId to the destShardId
func (sd *ShardedData) MoveData(sourceShardID, destShardID uint32, key [][]byte) {
	sourceStore := sd.ShardDataStore(sourceShardID)

	if sourceStore != nil {
		for _, key := range key {
			data, _ := sourceStore.Get(key)
			sd.AddData(key, data, destShardID)
			sd.RemoveData(key, sourceShardID)
		}
	}
}

// Clear will delete all shard stores and associated data
func (sd *ShardedData) Clear() {
	sd.mutShardedDataStore.Lock()
	for m := range sd.shardedDataStore {
		delete(sd.shardedDataStore, m)
	}
	sd.mutShardedDataStore.Unlock()
}

// ClearMiniPool will delete all data associated with a given destination shardID
func (sd *ShardedData) ClearMiniPool(shardID uint32) {
	mp := sd.ShardDataStore(shardID)
	if mp == nil {
		return
	}
	mp.Clear()
}

// RegisterAddedDataHandler registers a new handler to be called when a new data is added
func (sd *ShardedData) RegisterAddedDataHandler(handler func(key []byte)) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a ShardedData object")
		return
	}

	sd.mutAddedDataHandlers.Lock()
	sd.addedDataHandlers = append(sd.addedDataHandlers, handler)
	sd.mutAddedDataHandlers.Unlock()
}
