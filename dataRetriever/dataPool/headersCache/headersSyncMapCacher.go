package headersCache

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/logger"
	"sync"
)

var ErrHeaderNotFound = errors.New("cannot find header")

var log = logger.GetOrCreate("dataRetriever/headersCache")

type headerDetails struct {
	headerHash []byte
	header     data.HeaderHandler
}

type headersCacher struct {
	headersCacheByShardIdNonce *headersNonceCache
	mutAddedDataHandlers       sync.RWMutex
	addedDataHandlers          []func(shardHeaderHash []byte)
}

func NewHeadersCacher(numMaxHeaderPerShard int, numElementsToRemove int) (*headersCacher, error) {
	if numMaxHeaderPerShard < numElementsToRemove {
		return nil, errors.New("invalid cache parameters")
	}
	return &headersCacher{
		headersCacheByShardIdNonce: NewHeadersNonceCache(numElementsToRemove, numMaxHeaderPerShard),
		mutAddedDataHandlers:       sync.RWMutex{},
		addedDataHandlers:          make([]func(shardHeaderHash []byte), 0),
	}, nil
}

func (hc *headersCacher) Add(headerHash []byte, header data.HeaderHandler) {
	alreadyExits := hc.headersCacheByShardIdNonce.addHeaderInNonceCache(headerHash, header)

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

func (hc *headersCacher) RemoveHeaderByHash(headerHash []byte) {
	hc.headersCacheByShardIdNonce.removeHeaderByHash(headerHash)
}

func (hc *headersCacher) RemoveHeaderByNonceAndShardId(hdrNonce uint64, shardId uint32) {
	_ = hc.headersCacheByShardIdNonce.removeHeaderNonceByNonceAndShardId(hdrNonce, shardId)
}

func (hc *headersCacher) GetHeaderByNonceAndShardId(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
	headersList, ok := hc.headersCacheByShardIdNonce.getHeadersByNonceAndShardId(hdrNonce, shardId)
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

func (hc *headersCacher) GetHeaderByHash(hash []byte) (data.HeaderHandler, error) {
	return hc.headersCacheByShardIdNonce.getHeaderByHash(hash)
}

func (hc *headersCacher) GetNumHeadersFromCacheShard(shardId uint32) int {
	return int(hc.headersCacheByShardIdNonce.getNumHeaderFromCache(shardId))
}

func (hc *headersCacher) Clear() {

}

func (hc *headersCacher) RegisterHandler(handler func(shardHeaderHash []byte)) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a cacher object")
		return
	}

	hc.mutAddedDataHandlers.Lock()
	hc.addedDataHandlers = append(hc.addedDataHandlers, handler)
	hc.mutAddedDataHandlers.Unlock()
}

func (hc *headersCacher) Keys(shardId uint32) []uint64 {
	return hc.headersCacheByShardIdNonce.keys(shardId)
}

func (hc *headersCacher) Len() int {
	return hc.headersCacheByShardIdNonce.totalHeaders()
}

func (hc *headersCacher) MaxSize() int {
	return hc.headersCacheByShardIdNonce.maxHeadersPerShard
}

func (hc *headersCacher) IsInterfaceNil() bool {
	return hc == nil
}
