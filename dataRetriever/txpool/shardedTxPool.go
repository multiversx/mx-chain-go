package txpool

import (
	"strconv"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

var log = logger.GetOrCreate("txpool")

// shardedTxPool holds transaction caches organised by destination shard
type shardedTxPool struct {
	mutex                sync.RWMutex
	backingMap           map[string]*txPoolShard
	mutexAddCallbacks    sync.RWMutex
	onAddCallbacks       []func(key []byte)
	cacheConfigPrototype txcache.CacheConfig
	selfShardID          uint32
	numberOfShards       uint32
}

type txPoolShard struct {
	CacheID string
	Cache   *txcache.TxCache
}

// NewShardedTxPool creates a new sharded tx pool
// Implements "dataRetriever.TxPool"
func NewShardedTxPool(args ArgShardedTxPool) (dataRetriever.ShardedDataCacherNotifier, error) {
	log.Trace("NewShardedTxPool", "args", args)

	err := args.verify()
	if err != nil {
		return nil, err
	}

	const oneTrilion = 1000000 * 1000000
	numCaches := 2*args.NumberOfShards - 1

	cacheConfigPrototype := txcache.CacheConfig{
		NumChunksHint:              args.Config.Shards,
		EvictionEnabled:            true,
		NumBytesThreshold:          args.Config.SizeInBytes / numCaches,
		CountThreshold:             args.Config.Size / numCaches,
		NumSendersToEvictInOneStep: dataRetriever.TxPoolNumSendersToEvictInOneStep,
		LargeNumOfTxsForASender:    dataRetriever.TxPoolLargeNumOfTxsForASender,
		NumTxsToEvictFromASender:   dataRetriever.TxPoolNumTxsToEvictFromASender,
		MinGasPriceMicroErd:        uint32(args.MinGasPrice / oneTrilion),
	}

	shardedTxPoolObject := &shardedTxPool{
		mutex:                sync.RWMutex{},
		backingMap:           make(map[string]*txPoolShard),
		mutexAddCallbacks:    sync.RWMutex{},
		onAddCallbacks:       make([]func(key []byte), 0),
		cacheConfigPrototype: cacheConfigPrototype,
		selfShardID:          args.SelfShardID,
		numberOfShards:       args.NumberOfShards,
	}

	return shardedTxPoolObject, nil
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
	cacheID = txPool.routeToCache(cacheID)

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
		// Here we clone the config structure
		cacheConfig := txPool.cacheConfigPrototype
		cacheConfig.Name = cacheID

		if txPool.isCacheForSelfShard(cacheID) {
			cacheConfig.CountThreshold *= txPool.numberOfShards
			cacheConfig.NumBytesThreshold *= txPool.numberOfShards
		}

		cache := txcache.NewTxCache(cacheConfig)
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

	sourceShardID, destinationShardID := parseCacheID(cacheID)

	wrapper := &txcache.WrappedTransaction{
		Tx:              valueAsTransaction,
		TxHash:          key,
		SenderShardID:   sourceShardID,
		ReceiverShardID: destinationShardID,
	}

	txPool.addTx(wrapper, cacheID)
}

// addTx adds the transaction to the cache
func (txPool *shardedTxPool) addTx(tx *txcache.WrappedTransaction, cacheID string) {
	shard := txPool.getOrCreateShard(cacheID)
	cache := shard.Cache
	_, added := cache.AddTx(tx)
	if added {
		txPool.onAdded(tx.TxHash)
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

	var txFromCache *txcache.WrappedTransaction
	var hashExists bool
	for _, shard := range txPool.backingMap {
		txFromCache, hashExists = shard.Cache.GetByTxHash(txHash)
		if hashExists {
			return txFromCache.Tx, true
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
	sourceCacheID = txPool.routeToCache(sourceCacheID)
	destCacheID = txPool.routeToCache(destCacheID)

	sourceShard := txPool.getOrCreateShard(sourceCacheID)
	sourceCache := sourceShard.Cache

	sourceCache.ForEachTransaction(func(txHash []byte, tx *txcache.WrappedTransaction) {
		txPool.addTx(tx, destCacheID)
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

func (txPool *shardedTxPool) routeToCache(cacheID string) string {
	sourceShardID, _ := parseCacheID(cacheID)

	if sourceShardID == txPool.selfShardID {
		return strconv.Itoa(int(sourceShardID))
	}

	return cacheID
}

func (txPool *shardedTxPool) isCacheForSelfShard(cacheID string) bool {
	shardID, err := strconv.ParseUint(cacheID, 10, 64)
	if err != nil {
		return false
	}

	return uint32(shardID) == txPool.selfShardID
}

func parseCacheID(cacheID string) (sourceShardID uint32, destinationShardID uint32) {
	parts := strings.Split(cacheID, "_")
	sourceShardID = 0
	destinationShardID = 0

	switch len(parts) {
	case 1:
		shardID := parseCacheIDPart(parts[0])
		sourceShardID = shardID
		destinationShardID = shardID
	case 2:
		sourceShardID = parseCacheIDPart(parts[0])
		destinationShardID = parseCacheIDPart(parts[1])
	default:
		log.Error("parseCacheID", "cacheID", cacheID)
	}

	return
}

func parseCacheIDPart(cacheIDPart string) uint32 {
	part, err := strconv.ParseUint(cacheIDPart, 10, 64)
	if err != nil {
		log.Error("parseCacheIDPart", "cacheIDPart", cacheIDPart)
		return 0
	}

	return uint32(part)
}
