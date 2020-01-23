package txpool

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

var log = logger.GetOrCreate("dataretriever/txpool")

// shardedTxPool holds transaction caches organised by destination shard
type shardedTxPool struct {
	mutex             sync.RWMutex
	backingMap        map[string]*txPoolShard
	mutexAddCallbacks sync.RWMutex
	onAddCallbacks    []func(key []byte)
	cacheConfig       storageUnit.CacheConfig
	evictionConfig    txcache.EvictionConfig
}

type txPoolShard struct {
	CacheID string
	Cache   *txcache.TxCache
}

// NewShardedTxPool creates a new sharded tx pool
// Implements "dataRetriever.TxPool"
func NewShardedTxPool(config storageUnit.CacheConfig) (dataRetriever.ShardedDataCacherNotifier, error) {
	err := verifyConfig(config)
	if err != nil {
		return nil, err
	}

	size := config.Size
	evictionConfig := txcache.EvictionConfig{
		Enabled:                         true,
		CountThreshold:                  size,
		ThresholdEvictSenders:           process.TxPoolThresholdEvictSenders,
		NumOldestSendersToEvict:         process.TxPoolNumOldestSendersToEvict,
		ALotOfTransactionsForASender:    process.TxPoolALotOfTransactionsForASender,
		NumTxsToEvictForASenderWithALot: process.TxPoolNumTxsToEvictForASenderWithALot,
	}

	shardedTxPoolObject := &shardedTxPool{
		mutex:             sync.RWMutex{},
		backingMap:        make(map[string]*txPoolShard),
		mutexAddCallbacks: sync.RWMutex{},
		onAddCallbacks:    make([]func(key []byte), 0),
		cacheConfig:       config,
		evictionConfig:    evictionConfig,
	}

	return shardedTxPoolObject, nil
}

func verifyConfig(config storageUnit.CacheConfig) error {
	if config.Size < 1 {
		return dataRetriever.ErrCacheConfigInvalidSize
	}
	if config.Shards < 1 {
		return dataRetriever.ErrCacheConfigInvalidShards
	}

	return nil
}

// ShardDataStore returns the requested cache, as the generic Cacher interface
func (txPool *shardedTxPool) ShardDataStore(cacheID string) storage.Cacher {
	cache := txPool.getTxCache(cacheID)
	return cache
}

// getTxCache returns the requested cache
func (txPool *shardedTxPool) getTxCache(cacheID string) *txcache.TxCache {
	shard := txPool.getOrCreateShard(cacheID)
	return shard.Cache
}

func (txPool *shardedTxPool) getOrCreateShard(cacheID string) *txPoolShard {
	txPool.mutex.RLock()
	shard, ok := txPool.backingMap[cacheID]
	txPool.mutex.RUnlock()

	if ok {
		return shard
	}

	shard = txPool.createShard(cacheID)
	return shard
}

func (txPool *shardedTxPool) createShard(cacheID string) *txPoolShard {
	txPool.mutex.Lock()
	defer txPool.mutex.Unlock()

	shard, ok := txPool.backingMap[cacheID]
	if !ok {
		nChunksHint := txPool.cacheConfig.Shards
		evictionConfig := txPool.evictionConfig
		cache := txcache.NewTxCacheWithEviction(nChunksHint, evictionConfig)
		shard = &txPoolShard{
			CacheID: cacheID,
			Cache:   cache,
		}

		txPool.backingMap[cacheID] = shard
	}

	return shard
}

// AddData adds the transaction to the cache
func (txPool *shardedTxPool) AddData(key []byte, value interface{}, cacheID string) {
	valueAsTransaction, ok := value.(data.TransactionHandler)
	if !ok {
		return
	}

	txPool.addTx(key, valueAsTransaction, cacheID)
}

// addTx adds the transaction to the cache
func (txPool *shardedTxPool) addTx(txHash []byte, tx data.TransactionHandler, cacheID string) {
	shard := txPool.getOrCreateShard(cacheID)
	cache := shard.Cache
	_, added := cache.AddTx(txHash, tx)
	if added {
		txPool.onAdded(txHash)
	}
}

func (txPool *shardedTxPool) onAdded(txHash []byte) {
	txPool.mutexAddCallbacks.RLock()
	defer txPool.mutexAddCallbacks.RUnlock()

	for _, handler := range txPool.onAddCallbacks {
		go handler(txHash)
	}
}

// SearchFirstData searches the transaction against all shard data store, retrieving the first found
func (txPool *shardedTxPool) SearchFirstData(key []byte) (interface{}, bool) {
	tx, ok := txPool.searchFirstTx(key)
	return tx, ok
}

// searchFirstTx searches the transaction against all shard data store, retrieving the first found
func (txPool *shardedTxPool) searchFirstTx(txHash []byte) (tx data.TransactionHandler, ok bool) {
	txPool.mutex.RLock()
	defer txPool.mutex.RUnlock()

	var txFromCache data.TransactionHandler
	var hashExists bool
	for _, shard := range txPool.backingMap {
		txFromCache, hashExists = shard.Cache.GetByTxHash(txHash)
		if hashExists {
			return txFromCache, true
		}
	}

	return nil, false
}

// RemoveData removes the transaction from the pool
func (txPool *shardedTxPool) RemoveData(key []byte, cacheID string) {
	txPool.removeTx(key, cacheID)
}

// removeTx removes the transaction from the pool
func (txPool *shardedTxPool) removeTx(txHash []byte, cacheID string) {
	shard := txPool.getOrCreateShard(cacheID)
	_ = shard.Cache.RemoveTxByHash(txHash)
}

// RemoveSetOfDataFromPool removes a bunch of transactions from the pool
func (txPool *shardedTxPool) RemoveSetOfDataFromPool(keys [][]byte, cacheID string) {
	txPool.removeTxBulk(keys, cacheID)
}

// removeTxBulk removes a bunch of transactions from the pool
func (txPool *shardedTxPool) removeTxBulk(txHashes [][]byte, cacheID string) {
	for _, key := range txHashes {
		txPool.removeTx(key, cacheID)
	}
}

// RemoveDataFromAllShards removes the transaction from the pool (it searches in all shards)
func (txPool *shardedTxPool) RemoveDataFromAllShards(key []byte) {
	txPool.removeTxFromAllShards(key)
}

// removeTxFromAllShards removes the transaction from the pool (it searches in all shards)
func (txPool *shardedTxPool) removeTxFromAllShards(txHash []byte) {
	txPool.mutex.RLock()
	defer txPool.mutex.RUnlock()

	for _, shard := range txPool.backingMap {
		cache := shard.Cache
		_ = cache.RemoveTxByHash(txHash)
	}
}

// MergeShardStores merges two shards of the pool
func (txPool *shardedTxPool) MergeShardStores(sourceCacheID, destCacheID string) {
	sourceShard := txPool.getOrCreateShard(sourceCacheID)
	sourceCache := sourceShard.Cache

	sourceCache.ForEachTransaction(func(txHash []byte, tx data.TransactionHandler) {
		txPool.addTx(txHash, tx, destCacheID)
	})

	txPool.mutex.Lock()
	delete(txPool.backingMap, sourceCacheID)
	txPool.mutex.Unlock()
}

// Clear clears everything in the pool
func (txPool *shardedTxPool) Clear() {
	txPool.mutex.Lock()
	txPool.backingMap = make(map[string]*txPoolShard)
	txPool.mutex.Unlock()
}

// ClearShardStore clears a specific cache
func (txPool *shardedTxPool) ClearShardStore(cacheID string) {
	shard := txPool.getOrCreateShard(cacheID)
	shard.Cache.Clear()
}

// CreateShardStore is not implemented for this pool, since shard creations is managed internally
func (txPool *shardedTxPool) CreateShardStore(cacheID string) {
}

// RegisterHandler registers a new handler to be called when a new transaction is added
func (txPool *shardedTxPool) RegisterHandler(handler func(key []byte)) {
	if handler == nil {
		log.Error("attempt to register a nil handler")
		return
	}

	txPool.mutexAddCallbacks.Lock()
	txPool.onAddCallbacks = append(txPool.onAddCallbacks, handler)
	txPool.mutexAddCallbacks.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (txPool *shardedTxPool) IsInterfaceNil() bool {
	return txPool == nil
}
