package headersCache

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("dataRetriever/headersCache")

type headerDetails struct {
	headerHash []byte
	header     data.HeaderHandler
}

type headersCacher struct {
	hdrsCache            *headersNonceCache
	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func(shardHeaderHash []byte)
}

func NewHeadersCacher(numMaxHeaderPerShard int, numElementsToRemove int) (*headersCacher, error) {
	if numMaxHeaderPerShard < numElementsToRemove {
		return nil, ErrInvalidHeadersCacheParameter
	}

	headersCache := newHeadersNonceCache(numElementsToRemove, numMaxHeaderPerShard)

	return &headersCacher{
		hdrsCache:            headersCache,
		mutAddedDataHandlers: sync.RWMutex{},
		addedDataHandlers:    make([]func(shardHeaderHash []byte), 0),
	}, nil
}

// Add is used to add a header in pool
func (hc *headersCacher) Add(headerHash []byte, header data.HeaderHandler) {
	alreadyExits := hc.hdrsCache.addHeaderInNonceCache(headerHash, header)

	if !alreadyExits {
		hc.callAddedDataHandlers(headerHash)
	}
}

func (hc *headersCacher) callAddedDataHandlers(key []byte) {
	hc.mutAddedDataHandlers.RLock()
	for _, handler := range hc.addedDataHandlers {
		go handler(key)
	}
	hc.mutAddedDataHandlers.RUnlock()
}

// RemoveHeaderByHash will remove a header with a specific hash from pool
func (hc *headersCacher) RemoveHeaderByHash(headerHash []byte) {
	hc.hdrsCache.removeHeaderByHash(headerHash)
}

// RemoveHeaderByNonceAndShardId will remove a header with a nonce and shard id from pool
func (hc *headersCacher) RemoveHeaderByNonceAndShardId(hdrNonce uint64, shardId uint32) {
	_ = hc.hdrsCache.removeHeaderNonceByNonceAndShardId(hdrNonce, shardId)
}

// GetHeaderByNonceAndShardId will return a list of headers from pool
func (hc *headersCacher) GetHeaderByNonceAndShardId(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
	headersList, ok := hc.hdrsCache.getHeadersByNonceAndShardId(hdrNonce, shardId)
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
func (hc *headersCacher) GetHeaderByHash(hash []byte) (data.HeaderHandler, error) {
	return hc.hdrsCache.getHeaderByHash(hash)
}

// GetNumHeadersFromCacheShard will return how many header are in pool for a specific shard
func (hc *headersCacher) GetNumHeadersFromCacheShard(shardId uint32) int {
	return int(hc.hdrsCache.getNumHeaderFromCache(shardId))
}

// Clear will clear headers pool
func (hc *headersCacher) Clear() {
	hc.hdrsCache.clear()
}

// RegisterHandler registers a new handler to be called when a new data is added
func (hc *headersCacher) RegisterHandler(handler func(shardHeaderHash []byte)) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a cacher object")
		return
	}

	hc.mutAddedDataHandlers.Lock()
	hc.addedDataHandlers = append(hc.addedDataHandlers, handler)
	hc.mutAddedDataHandlers.Unlock()
}

// Keys will return a slice of all headers nonce that are in pool
func (hc *headersCacher) Keys(shardId uint32) []uint64 {
	return hc.hdrsCache.keys(shardId)
}

// Len will return how many headers are in pool
func (hc *headersCacher) Len() int {
	return hc.hdrsCache.totalHeaders()
}

// MaxSize will return how many header can be added in a pool ( per shard)
func (hc *headersCacher) MaxSize() int {
	return hc.hdrsCache.maxHeadersPerShard
}

// IsInterfaceNil returns true if there is no value under the interface
func (hc *headersCacher) IsInterfaceNil() bool {
	return hc == nil
}
