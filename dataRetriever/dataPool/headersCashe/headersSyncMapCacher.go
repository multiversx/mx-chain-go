package headersCashe

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go/data"
)

var ErrHeaderNotFound = errors.New("cannot find header")

type headerDetails struct {
	headerHash []byte
	header     data.HeaderHandler
}

type headersCacher struct {
	headersCacheByShardIdNonce *headersNonceCache
	maxHeadersPerShard         int
	canDoEviction              chan struct{}
}

func NewHeadersCacher(numMaxHeaderPerShard int, numElementsToRemove int) (*headersCacher, error) {
	if numMaxHeaderPerShard < numElementsToRemove {
		return nil, errors.New("invalid cache parameters")
	}
	return &headersCacher{
		headersCacheByShardIdNonce: NewHeadersNonceCache(numElementsToRemove),
		maxHeadersPerShard:         numMaxHeaderPerShard,
		canDoEviction:              make(chan struct{}, 1),
	}, nil
}

func (hc *headersCacher) Add(headerHash []byte, header data.HeaderHandler) {
	hc.headersCacheByShardIdNonce.addHeaderInNonceCache(headerHash, header)

	hc.tryToDoEviction(header.GetShardID())
}

func (hc *headersCacher) tryToDoEviction(hdrShardId uint32) {
	hc.canDoEviction <- struct{}{}
	numHeaders := hc.headersCacheByShardIdNonce.getNumHeaderFromCache(hdrShardId)
	if int(numHeaders) > hc.maxHeadersPerShard {
		hc.doEviction(hdrShardId)
	}

	<-hc.canDoEviction

	return
}

func (hc *headersCacher) doEviction(shardId uint32) {
	hc.headersCacheByShardIdNonce.lruEviction(shardId)
}

func (hc *headersCacher) RemoveHeaderByHash(headerHash []byte) {
	hc.headersCacheByShardIdNonce.removeHeaderByHash(headerHash)
}

func (hc *headersCacher) RemoveHeaderByNonceAndShardId(hdrNonce uint64, shardId uint32) {
	_ = hc.headersCacheByShardIdNonce.removeHeaderNonceByNonceAndShardId(hdrNonce, shardId)
}

func (hc *headersCacher) GetHeaderByNonceAndShardId(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, error) {
	headersList, ok := hc.headersCacheByShardIdNonce.getHeadersByNonceAndShardId(hdrNonce, shardId)
	if !ok {
		return nil, ErrHeaderNotFound
	}

	headers := make([]data.HeaderHandler, 0)
	for _, hdrDetails := range headersList {
		headers = append(headers, hdrDetails.header)
	}

	return headers, nil
}

func (hc *headersCacher) GetHeaderByHash(hash []byte) (data.HeaderHandler, error) {
	return hc.headersCacheByShardIdNonce.getHeaderByHash(hash)
}

func (hc *headersCacher) GetNumHeadersFromCacheShard(shardId uint32) int {
	return int(hc.headersCacheByShardIdNonce.getNumHeaderFromCache(shardId))
}
