package headersCache

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("dataRetriever/headersCache")

var _ dataRetriever.HeadersPool = (*headersPool)(nil)

type headersPool struct {
	cache                *headersCache
	mutAddedDataHandlers sync.RWMutex
	mutHeadersPool       sync.RWMutex
	addedDataHandlers    []func(headerHandler data.HeaderHandler, headerHash []byte)
}

// NewHeadersPool will create a new items cacher
func NewHeadersPool(hdrsPoolConfig config.HeadersPoolConfig) (*headersPool, error) {
	err := checkHeadersPoolConfig(hdrsPoolConfig)
	if err != nil {
		return nil, err
	}

	headersCacheObject := newHeadersCache(hdrsPoolConfig.MaxHeadersPerShard, hdrsPoolConfig.NumElementsToRemoveOnEviction)

	return &headersPool{
		cache:                headersCacheObject,
		mutAddedDataHandlers: sync.RWMutex{},
		mutHeadersPool:       sync.RWMutex{},
		addedDataHandlers:    make([]func(headerHandler data.HeaderHandler, headerHash []byte), 0),
	}, nil
}

func checkHeadersPoolConfig(hdrsPoolConfig config.HeadersPoolConfig) error {
	maxHdrsPerShard := hdrsPoolConfig.MaxHeadersPerShard
	numElementsToRemove := hdrsPoolConfig.NumElementsToRemoveOnEviction

	if maxHdrsPerShard <= 0 {
		return fmt.Errorf("%w, maxHdrsPerShard should be greater than 0", ErrInvalidHeadersCacheParameter)
	}
	if numElementsToRemove <= 0 {
		return fmt.Errorf("%w, numElementsToRemove should be greater than 0", ErrInvalidHeadersCacheParameter)
	}

	if maxHdrsPerShard < numElementsToRemove {
		return fmt.Errorf("%w, maxHdrsPerShard should be greater than numElementsToRemove", ErrInvalidHeadersCacheParameter)
	}

	return nil
}

// AddHeader is used to add a header in pool
func (pool *headersPool) AddHeader(headerHash []byte, header data.HeaderHandler) {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()

	added := pool.cache.addHeader(headerHash, header)

	if added {
		pool.callAddedDataHandlers(header, headerHash)
	}
}

// AddHeaderInShard adds header in pool as specified shard id
func (pool *headersPool) AddHeaderInShard(headerHash []byte, header data.HeaderHandler, shardID uint32) {
	if check.IfNil(header) || len(headerHash) == 0 {
		return
	}

	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()

	added := pool.cache.addHeaderByShardID(headerHash, header, shardID)
	if added {
		pool.callAddedDataHandlers(header, headerHash)
	}
}

func (pool *headersPool) callAddedDataHandlers(headerHandler data.HeaderHandler, headerHash []byte) {
	pool.mutAddedDataHandlers.RLock()
	for _, handler := range pool.addedDataHandlers {
		go handler(headerHandler, headerHash)
	}
	pool.mutAddedDataHandlers.RUnlock()
}

// RemoveHeaderByHash will remove a header with a specific hash from pool
func (pool *headersPool) RemoveHeaderByHash(headerHash []byte) {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()
	pool.cache.removeHeaderByHash(headerHash)
}

// RemoveHeaderByNonceAndShardId will remove a header with a nonce and shard id from pool
func (pool *headersPool) RemoveHeaderByNonceAndShardId(hdrNonce uint64, shardId uint32) {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()
	_ = pool.cache.removeHeaderByNonceAndShardId(hdrNonce, shardId)
}

// GetHeadersByNonceAndShardId will return a list of items from pool
func (pool *headersPool) GetHeadersByNonceAndShardId(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()

	headers, hashes, ok := pool.cache.getHeadersAndHashesByNonceAndShardId(hdrNonce, shardId)
	if !ok {
		return nil, nil, ErrHeaderNotFound
	}

	return headers, hashes, nil
}

// GetHeaderByHash will return a header handler from pool with a specific hash
func (pool *headersPool) GetHeaderByHash(hash []byte) (data.HeaderHandler, error) {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()

	return pool.cache.getHeaderByHash(hash)
}

// GetNumHeaders will return how many header are in pool for a specific shard
func (pool *headersPool) GetNumHeaders(shardId uint32) int {
	pool.mutHeadersPool.RLock()
	defer pool.mutHeadersPool.RUnlock()

	return int(pool.cache.getNumHeaders(shardId))
}

// Clear will clear items pool
func (pool *headersPool) Clear() {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()

	pool.cache.clear()
}

// RegisterHandler registers a new handler to be called when a new data is added
func (pool *headersPool) RegisterHandler(handler func(headerHandler data.HeaderHandler, headerHash []byte)) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a cacher object")
		return
	}

	pool.mutAddedDataHandlers.Lock()
	pool.addedDataHandlers = append(pool.addedDataHandlers, handler)
	pool.mutAddedDataHandlers.Unlock()
}

// Nonces will return a slice of all items nonce that are in pool
func (pool *headersPool) Nonces(shardId uint32) []uint64 {
	pool.mutHeadersPool.RLock()
	defer pool.mutHeadersPool.RUnlock()

	return pool.cache.keys(shardId)
}

// Len will return how many items are in pool
func (pool *headersPool) Len() int {
	pool.mutHeadersPool.RLock()
	defer pool.mutHeadersPool.RUnlock()

	return pool.cache.totalHeaders()
}

// MaxSize will return how many header can be added in a pool ( per shard)
func (pool *headersPool) MaxSize() int {
	pool.mutHeadersPool.RLock()
	defer pool.mutHeadersPool.RUnlock()

	return pool.cache.maxHeadersPerShard
}

// IsInterfaceNil returns true if there is no value under the interface
func (pool *headersPool) IsInterfaceNil() bool {
	return pool == nil
}
