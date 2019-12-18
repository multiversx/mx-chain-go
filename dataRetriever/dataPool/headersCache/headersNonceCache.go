package headersCache

import (
	"bytes"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
)

type headerListDetails struct {
	headerList []headerDetails
	timestamp  time.Time
}

type headerInfo struct {
	headerNonce   uint64
	headerShardId uint32
}

type headersNonceCache struct {
	hdrNonceCache map[uint32]*headersMap

	headersByHash *headersHashMap
	hdrsCounter   *headersCounter

	mutHeadersMap      sync.RWMutex
	numHeadersToRemove int

	maxHeadersPerShard int
	canDoEviction      map[uint32]chan struct{}
}

func newHeadersNonceCache(numHeadersToRemove int, numMaxHeaderPerShard int) *headersNonceCache {
	return &headersNonceCache{
		hdrNonceCache:      make(map[uint32]*headersMap),
		hdrsCounter:        newHeadersCounter(),
		headersByHash:      newHeadersHashMap(),
		mutHeadersMap:      sync.RWMutex{},
		numHeadersToRemove: numHeadersToRemove,
		canDoEviction:      make(map[uint32]chan struct{}),
		maxHeadersPerShard: numMaxHeaderPerShard,
	}
}

func (hnc *headersNonceCache) addHeaderInNonceCache(headerHash []byte, header data.HeaderHandler) bool {
	headerShardId := header.GetShardID()
	headerNonce := header.GetNonce()

	// add header info in second map
	headerInfo := headerInfo{
		headerNonce:   headerNonce,
		headerShardId: headerShardId,
	}
	alreadyExits := hnc.headersByHash.addElement(headerHash, headerInfo)
	if alreadyExits {
		return true
	}

	headerShardPool := hnc.getShardMap(headerShardId)

	hnc.mutHeadersMap.Lock()
	headerListD := headerShardPool.getElement(headerNonce)

	headerDetails := headerDetails{
		headerHash: headerHash,
		header:     header,
	}
	headerListD.headerList = append(headerListD.headerList, headerDetails)
	headerShardPool.addElement(headerNonce, headerListD)
	hnc.mutHeadersMap.Unlock()

	hnc.hdrsCounter.increment(headerShardId)

	hnc.tryToDoEviction(headerShardId)

	return false

}

func (hnc *headersNonceCache) getShardMap(shardId uint32) *headersMap {
	hnc.mutHeadersMap.Lock()
	defer hnc.mutHeadersMap.Unlock()

	if _, ok := hnc.hdrNonceCache[shardId]; !ok {
		hnc.hdrNonceCache[shardId] = newHeadersMap()
		hnc.canDoEviction[shardId] = make(chan struct{}, 1)
	}

	return hnc.hdrNonceCache[shardId]
}

func (hnc *headersNonceCache) removeHeaderNonceCache(hdrInfo headerInfo, headerHash []byte) {
	_, ok := hnc.hdrNonceCache[hdrInfo.headerShardId]
	if !ok {
		return
	}

	hdrListD, ok := hnc.hdrNonceCache[hdrInfo.headerShardId].getHeadersDetailsListFromSMap(hdrInfo.headerNonce)
	if !ok {
		return
	}

	//remove header from header list
	for index, headerD := range hdrListD.headerList {
		if !bytes.Equal(headerD.headerHash, headerHash) {
			continue
		}

		hdrListD.headerList = append(hdrListD.headerList[:index], hdrListD.headerList[index+1:]...)
		hnc.hdrsCounter.decrement(hdrInfo.headerShardId)

		if len(hdrListD.headerList) == 0 {
			hnc.hdrNonceCache[hdrInfo.headerShardId].removeElement(hdrInfo.headerNonce)
			return
		}

		hnc.hdrNonceCache[hdrInfo.headerShardId].addElement(hdrInfo.headerNonce, hdrListD)
		return
	}
}

func (hnc *headersNonceCache) removeHeaderNonceByNonceAndShardId(hdrNonce uint64, shardId uint32) int {
	_, ok := hnc.hdrNonceCache[shardId]
	if !ok {
		return 0
	}

	hdrListD, ok := hnc.hdrNonceCache[shardId].getHeadersDetailsListFromSMap(hdrNonce)
	if !ok {
		return 0
	}

	hdrsHashes := make([][]byte, 0)
	for _, hdrDetails := range hdrListD.headerList {
		hdrsHashes = append(hdrsHashes, hdrDetails.headerHash)
		hnc.hdrsCounter.decrement(shardId)
	}

	hnc.hdrNonceCache[shardId].removeElement(hdrNonce)

	//remove elements from hashes map
	for _, hash := range hdrsHashes {
		hnc.headersByHash.deleteElement(hash)
	}

	return len(hdrsHashes)
}

func (hnc *headersNonceCache) getHeadersByNonceAndShardId(hdrNonce uint64, shardId uint32) ([]headerDetails, bool) {
	hnc.mutHeadersMap.RLock()
	defer hnc.mutHeadersMap.RUnlock()

	_, ok := hnc.hdrNonceCache[shardId]
	if !ok {
		return nil, false
	}

	headersList, ok := hnc.hdrNonceCache[shardId].getHeadersDetailsListFromSMap(hdrNonce)
	if !ok {
		return nil, false
	}

	return headersList.headerList, true
}

func (hnc *headersNonceCache) getNumHeaderFromCache(shardId uint32) int64 {
	return hnc.hdrsCounter.getNumHeaderFromCache(shardId)
}

func (hnc *headersNonceCache) lruEviction(shardId uint32) {
	nonces := hnc.hdrNonceCache[shardId].getNoncesTimestampSorted()

	var numHashes int
	for i := 0; i < hnc.numHeadersToRemove && i < len(nonces); i++ {
		numHashes += hnc.removeHeaderNonceByNonceAndShardId(nonces[i], shardId)

		if numHashes >= hnc.numHeadersToRemove {
			break
		}
	}
}

func (hnc *headersNonceCache) removeHeaderByHash(hash []byte) {
	hnc.mutHeadersMap.Lock()
	defer hnc.mutHeadersMap.Unlock()

	info, ok := hnc.headersByHash.getElement(hash)
	if !ok {
		return
	}

	//remove header from first map
	hnc.removeHeaderNonceCache(info, hash)
	//remove header from second map
	hnc.headersByHash.deleteElement(hash)
}

func (hnc *headersNonceCache) getHeaderByHash(hash []byte) (data.HeaderHandler, error) {
	info, ok := hnc.headersByHash.getElement(hash)
	if !ok {
		return nil, ErrHeaderNotFound
	}

	headersList, ok := hnc.getHeadersByNonceAndShardId(info.headerNonce, info.headerShardId)
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

func (hnc *headersNonceCache) keys(shardId uint32) []uint64 {
	shardMap := hnc.getShardMap(shardId)

	return shardMap.keys()
}

func (hnc *headersNonceCache) tryToDoEviction(hdrShardId uint32) {
	hnc.mutHeadersMap.Lock()
	c := hnc.canDoEviction[hdrShardId]
	hnc.mutHeadersMap.Unlock()

	c <- struct{}{}
	numHeaders := hnc.getNumHeaderFromCache(hdrShardId)
	if int(numHeaders) > hnc.maxHeadersPerShard {
		hnc.lruEviction(hdrShardId)
	}

	<-c

	return
}

func (hnc *headersNonceCache) totalHeaders() int {
	return hnc.hdrsCounter.totalHeaders()
}

func (hnc *headersNonceCache) clear() {
	hnc.mutHeadersMap.Lock()
	defer hnc.mutHeadersMap.Unlock()

	hnc.hdrNonceCache = make(map[uint32]*headersMap)
	hnc.hdrsCounter = newHeadersCounter()
	hnc.headersByHash = newHeadersHashMap()
}
