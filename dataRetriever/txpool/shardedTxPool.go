package txpool

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

var log = logger.GetOrCreate("dataretriever/txpool")

// shardedTxPool holds transaction caches organised by destination shard
type shardedTxPool struct {
	mutex             sync.RWMutex
	backingMap        map[string]*txPoolShard
	cacherConfig      storageUnit.CacheConfig
	mutexAddCallbacks sync.RWMutex
	onAddCallbacks    []func(key []byte)
}

type txPoolShard struct {
	CacheID string
	Cache   *txcache.TxCache
}

// NewShardedTxPool creates a new sharded tx pool
// Implements "dataRetriever.TxPool"
func NewShardedTxPool(config storageUnit.CacheConfig) dataRetriever.TxPool {
	return &shardedTxPool{
		cacherConfig:      cacherConfig,
		mutex:             sync.RWMutex{},
		backingMap:        make(map[string]*txPoolShard),
		mutexAddCallbacks: sync.RWMutex{},
		onAddCallbacks:    make([]func(key []byte), 0),
	}
}

// GetTxCache returns the requested cache
func (txPool *shardedTxPool) GetTxCache(cacheID string) *txcache.TxCache {
	cache := txPool.getOrCreateCache(cacheID)
	return cache
}

func (txPool *shardedTxPool) getOrCreateCache(cacheID string) *txcache.TxCache {
	txPool.mutex.RLock()
	shard, ok := txPool.backingMap[cacheID]
	txPool.mutex.RUnlock()

	if ok {
		return shard.Cache
	}

	// The cache not yet created, we'll create in a critical section
	txPool.mutex.Lock()

	// We have to check again if not created (concurrency issue)
	shard, ok = txPool.backingMap[cacheID]
	if !ok {
		nChunksHint := txPool.cacherConfig.Shards
		cache := txcache.NewTxCache(nChunksHint)
		shard := &txPoolShard{
			CacheID: cacheID,
			Cache:   cache,
		}

		txPool.backingMap[cacheID] = shard
	}

	txPool.mutex.Unlock()

	return shard.Cache
}

// AddTx will add the transaction to the cache
func (txPool *shardedTxPool) AddTx(txHash []byte, tx data.TransactionHandler, cacheID string) {
	cache := txPool.getOrCreateCache(cacheID)
	// TODO: HasOrAdd()
	cache.AddTx(txHash, tx)

	// TODO: if added (did not exist, call handlers)
	txPool.onAdded(txHash)
}

func (txPool *shardedTxPool) onAdded(txHash []byte) {
	txPool.mutexAddCallbacks.RLock()
	defer txPool.mutexAddCallbacks.RUnlock()

	for _, handler := range txPool.onAddCallbacks {
		go handler(txHash)
	}
}

// SearchFirstTx searches the transaction against all shard data store, retrieving the first found
func (txPool *shardedTxPool) SearchFirstTx(txHash []byte) (tx data.TransactionHandler, ok bool) {
	txPool.mutex.RLock()
	defer txPool.mutex.RUnlock()

	for key := range txPool.backingMap {
		shard := txPool.backingMap[key]
		tx, ok := shard.Cache.GetByTxHash(txHash)
		if ok {
			return tx, ok
		}
	}

	return nil, false
}

// RemoveTx will remove data hash from the corresponding shard store
func (txPool *shardedTxPool) RemoveTx(txHash []byte, cacheID string) {
	cache := txPool.getOrCreateCache(cacheID)
	cache.RemoveTxByHash(txHash)
}

// RemoveSetOfDataFromPool removes a list of keys from the corresponding pool
func (sd *shardedData) RemoveSetOfDataFromPool(keys [][]byte, cacheId string) {
	for _, key := range keys {
		sd.RemoveData(key, cacheId)
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
