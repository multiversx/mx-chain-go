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
}

func NewHeadersCacher(numMaxHeaderPerShard int, numElementsToRemove int) (*headersCacher, error) {
	if numMaxHeaderPerShard < numElementsToRemove {
		return nil, errors.New("invalid cache parameters")
	}
	return &headersCacher{
		headersCacheByShardIdNonce: NewHeadersNonceCache(numElementsToRemove, numMaxHeaderPerShard),
	}, nil
}

func (hc *headersCacher) Add(headerHash []byte, header data.HeaderHandler) {
	hc.headersCacheByShardIdNonce.addHeaderInNonceCache(headerHash, header)
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
