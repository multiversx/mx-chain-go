package headersCache

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("dataRetriever/headersCache")

type headersPool struct {
	cache                *headersCache
	mutAddedDataHandlers sync.RWMutex
	mutHeadersPool       sync.RWMutex
	addedDataHandlers    []func(shardHeaderHash []byte)
}

func NewHeadersCacher(numMaxHeaderPerShard int, numElementsToRemove int) (*headersPool, error) {
	if numMaxHeaderPerShard < numElementsToRemove {
		return nil, ErrInvalidHeadersCacheParameter
	}

	headersCache := newHeadersCache(numElementsToRemove, numMaxHeaderPerShard)

	return &headersPool{
		cache:                headersCache,
		mutAddedDataHandlers: sync.RWMutex{},
		mutHeadersPool:       sync.RWMutex{},
		addedDataHandlers:    make([]func(shardHeaderHash []byte), 0),
	}, nil
}

// Add is used to add a header in pool
func (pool *headersPool) Add(headerHash []byte, header data.HeaderHandler) {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()

	alreadyExits := pool.cache.addHeader(headerHash, header)

	if !alreadyExits {
		pool.callAddedDataHandlers(headerHash)
	}
}

func (pool *headersPool) callAddedDataHandlers(key []byte) {
	pool.mutAddedDataHandlers.RLock()
	for _, handler := range pool.addedDataHandlers {
		go handler(key)
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

	_ = pool.cache.removeHeaderNonceByNonceAndShardId(hdrNonce, shardId)
}

// GetHeaderByNonceAndShardId will return a list of headers from pool
func (pool *headersPool) GetHeaderByNonceAndShardId(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()

	headersList, ok := pool.cache.getHeadersByNonceAndShardId(hdrNonce, shardId)
	if !ok {
		return nil, nil, ErrHeaderNotFound
	}

	headers := make([]data.HeaderHandler, 0)
	hashes := make([][]byte, 0)
	for _, hdrDetails := range headersList {
		headers = append(headers, hdrDetails.header)
		hashes = append(hashes, hdrDetails.headerHash)
	}

	if len(headers) == 0 {
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

// GetNumHeadersFromCacheShard will return how many header are in pool for a specific shard
func (pool *headersPool) GetNumHeadersFromCacheShard(shardId uint32) int {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()

	return int(pool.cache.getNumHeaderFromCache(shardId))
}

// Clear will clear headers pool
func (pool *headersPool) Clear() {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()

	pool.cache.clear()
}

// RegisterHandler registers a new handler to be called when a new data is added
func (pool *headersPool) RegisterHandler(handler func(shardHeaderHash []byte)) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a cacher object")
		return
	}

	pool.mutAddedDataHandlers.Lock()
	pool.addedDataHandlers = append(pool.addedDataHandlers, handler)
	pool.mutAddedDataHandlers.Unlock()
}

// Keys will return a slice of all headers nonce that are in pool
func (pool *headersPool) Keys(shardId uint32) []uint64 {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()

	return pool.cache.keys(shardId)
}

// Len will return how many headers are in pool
func (pool *headersPool) Len() int {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()

	return pool.cache.totalHeaders()
}

// MaxSize will return how many header can be added in a pool ( per shard)
func (pool *headersPool) MaxSize() int {
	pool.mutHeadersPool.Lock()
	defer pool.mutHeadersPool.Unlock()

	return pool.cache.maxHeadersPerShard
}

// IsInterfaceNil returns true if there is no value under the interface
func (pool *headersPool) IsInterfaceNil() bool {
	return pool == nil
}
