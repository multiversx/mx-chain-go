package headersCash

import (
	"bytes"
	"errors"
	"github.com/ElrondNetwork/elrond-go/data"
	"sync"
)

var ErrHeaderNotFound = errors.New("cannot find header")

type headerDetails struct {
	headerHash []byte
	header     data.HeaderHandler
}

type headerInfo struct {
	headerNonce   uint64
	headerShardId uint32
}

type headersCacher struct {
	headersCacheByShardIdNonce *headersNonceCache
	headersByHash              *sync.Map
	maxHeadersPerShard         int
}

func NewHeadersCacher(numMaxHeaderPerShard int, numElementsToRemove int) *headersCacher {
	return &headersCacher{
		headersCacheByShardIdNonce: NewHeadersNonceCache(numElementsToRemove),
		headersByHash:              &sync.Map{},
		maxHeadersPerShard:         numMaxHeaderPerShard,
	}
}

func (hc *headersCacher) Add(headerHash []byte, header data.HeaderHandler) {
	hc.headersCacheByShardIdNonce.addHeaderInNonceCache(headerHash, header)

	headerInfo := headerInfo{
		headerNonce:   header.GetNonce(),
		headerShardId: header.GetShardID(),
	}
	hc.headersByHash.Store(string(headerHash), headerInfo)

	numHeaders := hc.headersCacheByShardIdNonce.getNumHeaderFromCache(header.GetShardID())
	if int(numHeaders) >= hc.maxHeadersPerShard {
		hc.doEviction(header.GetShardID())
	}

}

func (hc *headersCacher) doEviction(shardId uint32) {
	hashes := hc.headersCacheByShardIdNonce.lruEviction(shardId)

	for hash := range hashes {
		hc.headersByHash.Delete(string(hash))
	}
}

func (hc *headersCacher) RemoveHeaderByHash(headerHash []byte) {
	headerInfoI, ok := hc.headersByHash.Load(string(headerHash))
	if !ok {
		return
	}

	headerInfo, ok := headerInfoI.(headerInfo)
	if !ok {
		return
	}

	//remove header from first map
	hc.headersCacheByShardIdNonce.removeHeaderNonceCache(headerInfo, headerHash)
	//remove header from second map
	hc.headersByHash.Delete(string(headerHash))
}

func (hc *headersCacher) RemoveHeaderByNonceAndShardId(hdrNonce uint64, shardId uint32) {
	headersHashes, ok := hc.headersCacheByShardIdNonce.removeHeaderNonceByNonceAndShardId(hdrNonce, shardId)
	if !ok {
		return
	}

	for hash := range headersHashes {
		hc.headersByHash.Delete(string(hash))
	}
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
	infoI, ok := hc.headersByHash.Load(string(hash))
	if !ok {
		return nil, ErrHeaderNotFound
	}

	info, ok := infoI.(headerInfo)
	if !ok {
		return nil, ErrHeaderNotFound
	}

	headersList, ok := hc.headersCacheByShardIdNonce.getHeadersByNonceAndShardId(info.headerNonce, info.headerShardId)
	if !ok {
		return nil, ErrHeaderNotFound
	}

	for _, hdrDetails := range headersList {
		if bytes.Equal(hash, hdrDetails.headerHash) {
			return hdrDetails.header, nil
		}
	}

	return nil, ErrHeaderNotFound
}
