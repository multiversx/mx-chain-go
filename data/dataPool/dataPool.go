package dataPool

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// DataPool holds the list of structures organised by destination shard
//
// The miniPools field maps a cacher, containing data
//  hashes, to a corresponding shard id. It is able to add or remove
//  interface{} given the shard id it is associated with. It can
//  also merge and split pools when required
type DataPool struct {
	lock sync.RWMutex
	// MiniPoolsStore is a key value store
	// Each key represents a destination shard id and the value will contain all
	//  data hashes that have that shard as destination
	miniPoolsStore map[uint32]*miniPool
	cacheConfig    *config.CacheConfig

	AddedData func(dataHash []byte)
}

type miniPool struct {
	ShardID   uint32
	DataStore storage.Cacher
}

// NewDataPool is responsible for creating an empty pool of data
func NewDataPool(cacheConfig *config.CacheConfig) *DataPool {
	return &DataPool{
		miniPoolsStore: make(map[uint32]*miniPool),
		cacheConfig:    cacheConfig,
	}
}

// NewMiniPool is responsible for creating an empty mini pool
func NewMiniPool(destShardID uint32, config *config.CacheConfig) *miniPool {
	cacher, err := storage.CreateCacheFromConf(config)
	if err != nil {
		// TODO: This should be replaced with the correct log panic
		panic("Could not create cache storage for pools")
	}
	return &miniPool{
		destShardID,
		cacher,
	}
}

// NewMiniPool is a DataPool method that is responsible for creating
//  a new mini pool at the destShardID index in the MiniPoolsStore map
func (dp *DataPool) newMiniPool(destShardID uint32) {
	dp.lock.Lock()
	dp.miniPoolsStore[destShardID] = NewMiniPool(destShardID, dp.cacheConfig)
	dp.lock.Unlock()
}

// MiniPool returns a minipool of data associated with a given destination shardID
func (dp *DataPool) MiniPool(shardID uint32) *miniPool {
	dp.lock.RLock()
	mp := dp.miniPoolsStore[shardID]
	dp.lock.RUnlock()
	return mp
}

// MiniPoolDataStore returns the minipool data store containing data hashes
//  associated with a given destination shardID
func (dp *DataPool) MiniPoolDataStore(shardID uint32) (c storage.Cacher) {
	mp := dp.MiniPool(shardID)
	if mp == nil {
		return nil
	}
	return mp.DataStore
}

// AddData will add a data to the corresponding pool
func (dp *DataPool) AddData(dataHash []byte, data interface{}, destShardID uint32) {
	if dp.MiniPool(destShardID) == nil {
		dp.newMiniPool(destShardID)
	}
	mp := dp.MiniPoolDataStore(destShardID)
	found, _ := mp.HasOrAdd(dataHash, data)

	if dp.AddedData != nil && !found {
		dp.AddedData(dataHash)
	}
}

// RemoveData will remove a data from the corresponding pool
func (dp *DataPool) RemoveData(dataHash []byte, destShardID uint32) {
	mpData := dp.MiniPoolDataStore(destShardID)
	if mpData != nil {
		mpData.Remove(dataHash)
	}
}

// RemoveDataFromAllShards will remove a data from the pool given only
//  the data hash. It will iterate over all mini pools and will remove it everywhere
func (dp *DataPool) RemoveDataFromAllShards(dataHash []byte) {
	for k := range dp.miniPoolsStore {
		m := dp.MiniPoolDataStore(k)
		if m.Has(dataHash) {
			m.Remove(dataHash)
		}
	}
}

// MergeMiniPools will take all data associated with the sourceShardId and move it
// to the destShardID. It will then remove the sourceShardID key from the store map
func (dp *DataPool) MergeMiniPools(sourceShardID, destShardID uint32) {
	sourceStore := dp.MiniPoolDataStore(sourceShardID)

	if sourceStore != nil {
		for _, dataHash := range sourceStore.Keys() {
			data, _ := sourceStore.Get(dataHash)
			dp.AddData(dataHash, data, destShardID)
		}
	}

	dp.lock.Lock()
	delete(dp.miniPoolsStore, sourceShardID)
	dp.lock.Unlock()
}

// MoveData will move all given data associated with the sourceShardId to the destShardId
func (dp *DataPool) MoveData(sourceShardID, destShardID uint32, dataHashes [][]byte) {
	sourceStore := dp.MiniPoolDataStore(sourceShardID)

	if sourceStore != nil {
		for _, dataHash := range dataHashes {
			data, _ := sourceStore.Get(dataHash)
			dp.AddData(dataHash, data, destShardID)
			dp.RemoveData(dataHash, sourceShardID)
		}
	}
}

// Clear will delete all minipools and associated data
func (dp *DataPool) Clear() {
	for m := range dp.miniPoolsStore {
		dp.lock.Lock()
		delete(dp.miniPoolsStore, m)
		dp.lock.Unlock()
	}
}

// ClearMiniPool will delete all data associated with a given destination shardID
func (dp *DataPool) ClearMiniPool(shardID uint32) {
	mp := dp.MiniPoolDataStore(shardID)
	if mp == nil {
		return
	}
	mp.Clear()
}
