package track

import (
	"bytes"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

func (bbt *baseBlockTrack) receivedShardHeader(shardHeaderHash []byte) {
	shardHeader, err := process.GetShardHeaderFromPool(shardHeaderHash, bbt.shardHeadersPool)
	if err != nil {
		log.Trace("GetShardHeaderFromPool", "error", err.Error())
		return
	}

	log.Debug("received shard header from network in block tracker",
		"shard", shardHeader.GetShardID(),
		"round", shardHeader.GetRound(),
		"nonce", shardHeader.GetNonce(),
		"hash", shardHeaderHash,
	)

	bbt.processReceivedHeader(shardHeader, shardHeaderHash, bbt.getShardHeaderFromPoolWithNonce)
}

func (bbt *baseBlockTrack) receivedMetaBlock(metaBlockHash []byte) {
	metaBlock, err := process.GetMetaHeaderFromPool(metaBlockHash, bbt.metaBlocksPool)
	if err != nil {
		log.Trace("GetMetaHeaderFromPool", "error", err.Error())
		return
	}

	log.Debug("received meta block from network in block tracker",
		"shard", metaBlock.GetShardID(),
		"round", metaBlock.GetRound(),
		"nonce", metaBlock.GetNonce(),
		"hash", metaBlockHash,
	)

	bbt.processReceivedHeader(metaBlock, metaBlockHash, bbt.getMetaHeaderFromPoolWithNonce)
}

func (bbt *baseBlockTrack) processReceivedHeader(
	header data.HeaderHandler,
	headerHash []byte,
	getHeaderFromPoolWithNonce func(nonce uint64, shardID uint32) (data.HeaderHandler, []byte, error),
) {
	isHeaderForSelfShard := header.GetShardID() == bbt.shardCoordinator.SelfId()

	var lastNotarizedHeaderNonce uint64

	if isHeaderForSelfShard {
		lastNotarizedHeaderNonce = bbt.getLastSelfNotarizedHeaderNonce(header.GetShardID())
	} else {
		lastNotarizedHeaderNonce = bbt.getLastCrossNotarizedHeaderNonce(header.GetShardID())
	}

	if bbt.isHeaderOutOfRange(header.GetNonce(), lastNotarizedHeaderNonce) {
		log.Debug("received header is out of range",
			"received nonce", header.GetNonce(),
			"last notarized nonce", lastNotarizedHeaderNonce,
		)
		return
	}

	bbt.addHeaders(header, headerHash, lastNotarizedHeaderNonce, getHeaderFromPoolWithNonce)

	if isHeaderForSelfShard {
		bbt.doJobOnReceivedHeader(header.GetShardID())
	} else {
		bbt.doJobOnReceivedCrossNotarizedHeader(header.GetShardID())
	}
}

func (bbt *baseBlockTrack) isHeaderOutOfRange(receivedNonce uint64, lastNotarizedNonce uint64) bool {
	isHeaderOutOfRange := receivedNonce > lastNotarizedNonce+process.MaxNonceDifferences
	return isHeaderOutOfRange
}

func (bbt *baseBlockTrack) addHeaders(
	header data.HeaderHandler,
	headerHash []byte,
	lastNotarizedHeaderNonce uint64,
	getHeaderFromPoolWithNonce func(nonce uint64, shardID uint32) (data.HeaderHandler, []byte, error),
) {
	nextHeaderNonce := lastNotarizedHeaderNonce + bbt.blockFinality + 1
	headers, headersHashes := bbt.getHeadersIfMissing(
		header.GetShardID(),
		nextHeaderNonce,
		header.GetNonce()-1,
		getHeaderFromPoolWithNonce)

	for i := 0; i < len(headers); i++ {
		bbt.addHeader(headers[i], headersHashes[i])
	}

	bbt.addHeader(header, headerHash)
}

func (bbt *baseBlockTrack) getHeadersIfMissing(
	shardID uint32,
	fromNonce uint64,
	toNonce uint64,
	getHeaderFromPoolWithNonce func(nonce uint64, shardID uint32) (data.HeaderHandler, []byte, error),
) ([]data.HeaderHandler, [][]byte) {

	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	headers := make([]data.HeaderHandler, 0)
	headersHashes := make([][]byte, 0)

	headersForShard, haveHeadersForShard := bbt.headers[shardID]

	for nonce := fromNonce; nonce <= toNonce; nonce++ {
		if haveHeadersForShard && len(headersForShard[nonce]) > 0 {
			continue
		}

		header, headerHash, err := getHeaderFromPoolWithNonce(nonce, shardID)
		if err != nil {
			log.Debug("getHeaderFromPoolWithNonce", "error", err.Error())
			continue
		}

		log.Debug("got missing header",
			"shard", header.GetShardID(),
			"round", header.GetRound(),
			"nonce", header.GetNonce(),
			"hash", headerHash,
		)

		headers = append(headers, header)
		headersHashes = append(headersHashes, headerHash)
	}

	return headers, headersHashes
}

func (bbt *baseBlockTrack) addHeader(header data.HeaderHandler, hash []byte) {
	if check.IfNil(header) {
		return
	}

	shardID := header.GetShardID()
	nonce := header.GetNonce()

	bbt.mutHeaders.Lock()
	defer bbt.mutHeaders.Unlock()

	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		headersForShard = make(map[uint64][]*headerInfo)
		bbt.headers[shardID] = headersForShard
	}

	for _, headerInfo := range headersForShard[nonce] {
		if bytes.Equal(headerInfo.hash, hash) {
			return
		}
	}

	headersForShard[nonce] = append(headersForShard[nonce], &headerInfo{hash: hash, header: header})
}

func (bbt *baseBlockTrack) doJobOnReceivedHeader(shardID uint32) {
	selfNotarizedHeaders := bbt.computeLongestChainFromLastSelfNotarized(shardID)
	bbt.addSelfNotarizedHeaders(shardID, selfNotarizedHeaders)

	//log.Trace("display on received header", "shard", shardID)
	//bbt.displayHeadersForShard(shardID)
}

func (bbt *baseBlockTrack) computeLongestChainFromLastSelfNotarized(shardID uint32) []data.HeaderHandler {
	lastSelfNotarizedHeader, err := bbt.getLastSelfNotarizedHeader(shardID)
	if err != nil {
		return nil
	}

	return bbt.ComputeLongestChain(shardID, lastSelfNotarizedHeader)
}

func (bbt *baseBlockTrack) doJobOnReceivedCrossNotarizedHeader(shardID uint32) {
	crossNotarizedHeaders := bbt.computeLongestChainFromLastCrossNotarized(shardID)
	selfNotarizedHeaders, selfNotarizedHeadersHashes := bbt.computeSelfNotarizedHeaders(crossNotarizedHeaders)

	bbt.addCrossNotarizedHeaders(shardID, crossNotarizedHeaders)
	bbt.addSelfNotarizedHeaders(shardID, selfNotarizedHeaders)

	bbt.callSelfNotarizedHeadersHandlers(selfNotarizedHeaders, selfNotarizedHeadersHashes)

	log.Trace("display on received cross notarized header", "shard", shardID)
	bbt.displayHeadersForShard(shardID)
}

func (bbt *baseBlockTrack) computeLongestChainFromLastCrossNotarized(shardID uint32) []data.HeaderHandler {
	lastCrossNotarizedHeader, err := bbt.getLastCrossNotarizedHeader(shardID)
	if err != nil {
		return nil
	}

	return bbt.ComputeLongestChain(shardID, lastCrossNotarizedHeader)
}

func (bbt *baseBlockTrack) computeSelfNotarizedHeaders(headers []data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	selfNotarizedHeadersInfo := make([]*headerInfo, 0)

	for _, header := range headers {
		selfHeadersInfo := bbt.blockTracker.getSelfHeaders(header)
		if len(selfHeadersInfo) > 0 {
			selfNotarizedHeadersInfo = append(selfNotarizedHeadersInfo, selfHeadersInfo...)
		}
	}

	if len(selfNotarizedHeadersInfo) > 1 {
		sort.Slice(selfNotarizedHeadersInfo, func(i, j int) bool {
			return selfNotarizedHeadersInfo[i].header.GetNonce() < selfNotarizedHeadersInfo[j].header.GetNonce()
		})
	}

	selfNotarizedHeaders := make([]data.HeaderHandler, 0)
	selfNotarizedHeadersHashes := make([][]byte, 0)

	for _, selfNotarizedHeaderInfo := range selfNotarizedHeadersInfo {
		selfNotarizedHeaders = append(selfNotarizedHeaders, selfNotarizedHeaderInfo.header)
		selfNotarizedHeadersHashes = append(selfNotarizedHeadersHashes, selfNotarizedHeaderInfo.hash)
	}

	return selfNotarizedHeaders, selfNotarizedHeadersHashes
}

func (bbt *baseBlockTrack) getShardHeaderFromPoolWithNonce(
	nonce uint64,
	shardID uint32,
) (data.HeaderHandler, []byte, error) {
	return process.GetShardHeaderFromPoolWithNonce(nonce, shardID, bbt.shardHeadersPool, bbt.headersNoncesPool)
}

func (bbt *baseBlockTrack) getMetaHeaderFromPoolWithNonce(
	nonce uint64,
	_ uint32,
) (data.HeaderHandler, []byte, error) {
	return process.GetMetaHeaderFromPoolWithNonce(nonce, bbt.metaBlocksPool, bbt.headersNoncesPool)
}
