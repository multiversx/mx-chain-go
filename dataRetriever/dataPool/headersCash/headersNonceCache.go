package headersCash

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go/data"
	"sort"
	"sync"
	"time"
)

type headerListDetails struct {
	headerList []headerDetails
	timestamp  time.Time
}

type headersNonceCache struct {
	hdrNonceCache      map[uint32]*sync.Map
	numHeaderPerShard  *sync.Map
	numHeadersToRemove int
}

func NewHeadersNonceCache(numHeadersToRemove int) *headersNonceCache {
	return &headersNonceCache{
		hdrNonceCache:      make(map[uint32]*sync.Map),
		numHeaderPerShard:  &sync.Map{},
		numHeadersToRemove: numHeadersToRemove,
	}
}

func (hnc *headersNonceCache) addHeaderInNonceCache(headerHash []byte, header data.HeaderHandler) {
	headerShardId := header.GetShardID()
	headerNonce := header.GetNonce()

	headerShardPool, ok := hnc.hdrNonceCache[headerShardId]
	if !ok {
		hnc.hdrNonceCache[headerShardId] = &sync.Map{}

		headerList := make([]headerDetails, 0)

		headerDetails := headerDetails{
			headerHash: headerHash,
			header:     header,
		}
		headerList = append(headerList, headerDetails)

		headerListD := headerListDetails{
			headerList: headerList,
			timestamp:  time.Now(),
		}

		hnc.hdrNonceCache[headerShardId].Store(headerNonce, headerListD)
		hnc.incrementNumHeaders(headerShardId)
		return
	}

	headerListI, ok := headerShardPool.Load(headerNonce)
	if !ok {
		headerList := make([]headerDetails, 0)

		headerDetails := headerDetails{
			headerHash: headerHash,
			header:     header,
		}
		headerList = append(headerList, headerDetails)
		headerListD := headerListDetails{
			headerList: headerList,
			timestamp:  time.Now(),
		}

		hnc.hdrNonceCache[headerShardId].Store(headerNonce, headerListD)

		hnc.incrementNumHeaders(headerShardId)
		return
	}

	headerListD, ok := headerListI.(headerListDetails)
	if !ok {
		headerList := make([]headerDetails, 0)

		headerDetails := headerDetails{
			headerHash: headerHash,
			header:     header,
		}
		headerList = append(headerList, headerDetails)
		headerListD := headerListDetails{
			headerList: headerList,
			timestamp:  time.Now(),
		}

		hnc.hdrNonceCache[headerShardId].Store(headerNonce, headerListD)

		hnc.incrementNumHeaders(headerShardId)
		return
	}

	headerDetails := headerDetails{
		headerHash: headerHash,
		header:     header,
	}
	headerListD.headerList = append(headerListD.headerList, headerDetails)
	hnc.hdrNonceCache[headerShardId].Store(headerNonce, headerListD)

	hnc.incrementNumHeaders(headerShardId)
	return
}

func (hnc *headersNonceCache) removeHeaderNonceCache(hdrInfo headerInfo, headerHash []byte) {
	hdrListD, ok := hnc.getHeadersDetailsListFromSMap(hdrInfo.headerNonce, hdrInfo.headerShardId)
	if !ok {
		return
	}

	//remove header from header list
	for index, headerD := range hdrListD.headerList {
		if !bytes.Equal(headerD.headerHash, headerHash) {
			continue
		}

		hdrListD.headerList = append(hdrListD.headerList[:index], hdrListD.headerList[index+1:]...)
		hnc.decrementNumHeaders(hdrInfo.headerShardId)

		if len(hdrListD.headerList) == 0 {
			hnc.hdrNonceCache[hdrInfo.headerShardId].Delete(hdrInfo.headerNonce)
			return
		}

		hnc.hdrNonceCache[hdrInfo.headerShardId].Store(hdrInfo.headerNonce, hdrListD)
		return
	}
}

func (hnc *headersNonceCache) removeHeaderNonceByNonceAndShardId(hdrNonce uint64, shardId uint32) ([][]byte, bool) {
	hdrListD, ok := hnc.getHeadersDetailsListFromSMap(hdrNonce, shardId)
	if !ok {
		return nil, false
	}

	hdrsHash := make([][]byte, 0)
	for _, hdrDetails := range hdrListD.headerList {
		hdrsHash = append(hdrsHash, hdrDetails.headerHash)
		hnc.decrementNumHeaders(shardId)
	}

	hnc.hdrNonceCache[shardId].Delete(hdrNonce)

	return hdrsHash, true
}

func (hnc *headersNonceCache) getHeadersByNonceAndShardId(hdrNonce uint64, shardId uint32) ([]headerDetails, bool) {
	headersList, ok := hnc.getHeadersDetailsListFromSMap(hdrNonce, shardId)
	if !ok {
		return nil, false
	}

	return headersList.headerList, true
}

func (hnc *headersNonceCache) getHeadersDetailsListFromSMap(hdrNonce uint64, shardId uint32) (headerListDetails, bool) {
	headersShardPool, ok := hnc.hdrNonceCache[shardId]
	if !ok {
		return headerListDetails{}, false
	}

	headersListDI, ok := headersShardPool.Load(hdrNonce)
	if !ok {
		return headerListDetails{}, false
	}

	headerListD, ok := headersListDI.(headerListDetails)
	if !ok {
		return headerListDetails{}, false
	}

	//update timestamp
	headerListD.timestamp = time.Now()
	hnc.hdrNonceCache[shardId].Store(hdrNonce, headerListD)

	return headerListD, true
}

func (hnc *headersNonceCache) incrementNumHeaders(shardId uint32) {
	headersPerShardI, ok := hnc.numHeaderPerShard.Load(shardId)
	if !ok {
		hnc.numHeaderPerShard.Store(shardId, int64(1))

		return
	}

	headersPerShard, ok := headersPerShardI.(int64)
	if !ok {
		return
	}

	headersPerShard++
	hnc.numHeaderPerShard.Store(shardId, headersPerShard)
}

func (hnc *headersNonceCache) decrementNumHeaders(shardId uint32) {
	headersPerShardI, ok := hnc.numHeaderPerShard.Load(shardId)
	if !ok {
		return
	}

	headersPerShard, ok := headersPerShardI.(int64)
	if !ok {
		return
	}

	headersPerShard--
	hnc.numHeaderPerShard.Store(shardId, headersPerShard)
}

func (hnc *headersNonceCache) getNumHeaderFromCache(shardId uint32) int64 {
	numShardHeadersI, ok := hnc.numHeaderPerShard.Load(shardId)
	if !ok {
		return 0
	}

	numShardHeaders, ok := numShardHeadersI.(int64)
	if !ok {
		return 0
	}

	return numShardHeaders
}

func (hnc *headersNonceCache) lruEviction(shardId uint32) [][]byte {
	nonces := hnc.getNoncesTimestampSorted(shardId)

	hashesSlice := make([][]byte, 0)
	for i := 0; i < hnc.numHeadersToRemove && i < len(nonces); i++ {
		hashes, ok := hnc.removeHeaderNonceByNonceAndShardId(nonces[i], shardId)
		if !ok {
			continue
		}

		hashesSlice = append(hashesSlice, hashes...)
	}

	return hashesSlice
}

func (hnc *headersNonceCache) getNoncesTimestampSorted(shardId uint32) []uint64 {
	headersShardPool, ok := hnc.hdrNonceCache[shardId]
	if !ok {
		return nil
	}

	type nonceTimestamp struct {
		nonce     uint64
		timestamp time.Time
	}

	noncesTimestampsSlice := make([]nonceTimestamp, 0)
	headersShardPool.Range(
		func(key, value interface{}) bool {
			keyUint, ok := key.(uint64)
			if !ok {
				return false
			}
			headersListD, ok := value.(headerListDetails)
			if !ok {
				return false
			}

			noncesTimestampsSlice = append(noncesTimestampsSlice, nonceTimestamp{nonce: keyUint, timestamp: headersListD.timestamp})
			return true
		},
	)

	sort.Slice(noncesTimestampsSlice, func(i, j int) bool {
		return noncesTimestampsSlice[j].timestamp.After(noncesTimestampsSlice[i].timestamp)
	})

	nonceSlice := make([]uint64, 0)
	for _, d := range noncesTimestampsSlice {
		nonceSlice = append(nonceSlice, d.nonce)
	}

	return nonceSlice
}
