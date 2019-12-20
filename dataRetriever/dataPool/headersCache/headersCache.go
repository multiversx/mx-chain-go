package headersCache

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go/core"

	"github.com/ElrondNetwork/elrond-go/data"
)

type headersCache struct {
	hdrNonceCache map[uint32]headersMap

	headersByHash headersByHashMap
	hdrsCounter   numHeadersByShard

	numHeadersToRemove int
	maxHeadersPerShard int
}

func newHeadersCache(numHeadersToRemove int, numMaxHeaderPerShard int) *headersCache {
	return &headersCache{
		hdrNonceCache:      make(map[uint32]headersMap),
		hdrsCounter:        make(numHeadersByShard),
		headersByHash:      make(headersByHashMap),
		numHeadersToRemove: numHeadersToRemove,
		maxHeadersPerShard: numMaxHeaderPerShard,
	}
}

func (cache *headersCache) addHeader(headerHash []byte, header data.HeaderHandler) bool {
	headerShardId := header.GetShardID()
	headerNonce := header.GetNonce()

	//check if pool is full and if it is do eviction
	cache.tryToDoEviction(headerShardId)

	// add header info in second map
	alreadyExits := cache.headersByHash.addElement(headerHash, headerInfo{headerNonce, headerShardId})
	if alreadyExits {
		return true
	}

	headersShardMap := cache.getShardMap(headerShardId)
	headersShardMap.appendElement(headerHash, header)

	cache.hdrsCounter.increment(headerShardId)

	return false

}

func (cache *headersCache) tryToDoEviction(hdrShardId uint32) {
	numHeaders := cache.getNumHeaderFromCache(hdrShardId)
	if int(numHeaders) >= cache.maxHeadersPerShard {
		cache.lruEviction(hdrShardId)
	}

	return
}

func (cache *headersCache) lruEviction(shardId uint32) {
	headersShardMap, ok := cache.hdrNonceCache[shardId]
	if !ok {
		return
	}

	nonces := headersShardMap.getNoncesSortedByTimestamp()

	var numHashes int
	maxItemsToRemove := core.MinInt(cache.numHeadersToRemove, len(nonces))
	for i := 0; i < maxItemsToRemove; i++ {
		numHashes += cache.removeHeaderNonceByNonceAndShardId(nonces[i], shardId)

		if numHashes >= maxItemsToRemove {
			break
		}
	}
}

func (cache *headersCache) getShardMap(shardId uint32) headersMap {
	if _, ok := cache.hdrNonceCache[shardId]; !ok {
		cache.hdrNonceCache[shardId] = make(headersMap)
	}

	return cache.hdrNonceCache[shardId]
}

func (cache *headersCache) getNumHeaderFromCache(shardId uint32) int64 {
	return cache.hdrsCounter.getNumHeaderFromCache(shardId)
}

func (cache *headersCache) removeHeaderNonceByNonceAndShardId(hdrNonce uint64, shardId uint32) int {
	headersShardMap, ok := cache.hdrNonceCache[shardId]
	if !ok {
		return 0
	}

	headersWithTimestamp, ok := headersShardMap.getHeadersByNonce(hdrNonce)
	if !ok {
		return 0
	}
	hdrsHashes := headersWithTimestamp.getHashes()

	//remove headers from nonce map
	headersShardMap.removeElement(hdrNonce)
	//remove elements from hashes map
	cache.headersByHash.deleteBulk(hdrsHashes)

	cache.hdrsCounter.decrement(shardId, len(hdrsHashes))

	return len(hdrsHashes)
}

func (cache *headersCache) removeHeaderByHash(hash []byte) {
	info, ok := cache.headersByHash.getElement(hash)
	if !ok {
		return
	}

	//remove header from first map
	cache.removeHeaderNonceCache(info, hash)
	//remove header from second map
	cache.headersByHash.deleteElement(hash)
}

// removeHeaderNonceCache will remove a header from headerWithTimestamp
// when a header is removed by hash we need to remove also header from the map where is stored with nonce
func (cache *headersCache) removeHeaderNonceCache(hdrInfo headerInfo, headerHash []byte) {
	headersShardMap, ok := cache.hdrNonceCache[hdrInfo.headerShardId]
	if !ok {
		return
	}

	headersWithTimestamp, ok := headersShardMap.getHeadersByNonce(hdrInfo.headerNonce)
	if !ok {
		return
	}

	//remove header from header list
	for index, headerD := range headersWithTimestamp.headers {
		if !bytes.Equal(headerD.headerHash, headerHash) {
			continue
		}

		headersWithTimestamp.removeHeader(index)
		cache.hdrsCounter.decrement(hdrInfo.headerShardId, 1)

		if headersWithTimestamp.isEmpty() {
			headersShardMap.removeElement(hdrInfo.headerNonce)
			return
		}

		headersShardMap.addElement(hdrInfo.headerNonce, headersWithTimestamp)
		return
	}
}

func (cache *headersCache) getHeaderByHash(hash []byte) (data.HeaderHandler, error) {
	info, ok := cache.headersByHash.getElement(hash)
	if !ok {
		return nil, ErrHeaderNotFound
	}

	headersList, ok := cache.getHeadersByNonceAndShardId(info.headerNonce, info.headerShardId)
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

func (cache *headersCache) getHeadersByNonceAndShardId(hdrNonce uint64, shardId uint32) ([]headerDetails, bool) {
	headersShardMap, ok := cache.hdrNonceCache[shardId]
	if !ok {
		return nil, false
	}

	headersList, ok := headersShardMap.getHeadersByNonce(hdrNonce)
	if !ok {
		return nil, false
	}

	return headersList.headers, true
}

func (cache *headersCache) getHeadersAndHashesByNonceAndShardId(nonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, bool) {
	headersList, ok := cache.getHeadersByNonceAndShardId(nonce, shardId)
	if !ok || len(headersList) == 0 {
		return nil, nil, false
	}

	headers := make([]data.HeaderHandler, 0, len(headersList))
	hashes := make([][]byte, 0, len(headersList))
	for _, hdrDetails := range headersList {
		headers = append(headers, hdrDetails.header)
		hashes = append(hashes, hdrDetails.headerHash)
	}

	return headers, hashes, true
}

func (cache *headersCache) keys(shardId uint32) []uint64 {
	shardMap := cache.getShardMap(shardId)

	return shardMap.keys()
}

func (cache *headersCache) totalHeaders() int {
	return cache.hdrsCounter.totalHeaders()
}

func (cache *headersCache) clear() {
	cache.hdrNonceCache = make(map[uint32]headersMap)
	cache.hdrsCounter = make(numHeadersByShard)
	cache.headersByHash = make(headersByHashMap)
}
