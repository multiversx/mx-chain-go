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

	bbt.processReceivedHeader(shardHeader, shardHeaderHash)
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

	bbt.processReceivedHeader(metaBlock, metaBlockHash)
}

func (bbt *baseBlockTrack) processReceivedHeader(
	header data.HeaderHandler,
	headerHash []byte,
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

	bbt.addHeader(header, headerHash)

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
	_, _, selfNotarizedHeaders, selfNotarizedHeadersHashes := bbt.blockTracker.computeLongestSelfChain()

	if len(selfNotarizedHeaders) > 0 {
		bbt.callSelfNotarizedHeadersHandlers(shardID, selfNotarizedHeaders, selfNotarizedHeadersHashes)
	}
}

func (bbt *baseBlockTrack) doJobOnReceivedCrossNotarizedHeader(shardID uint32) {
	_, _, crossNotarizedHeaders, crossNotarizedHeadersHashes := bbt.computeLongestChainFromLastCrossNotarized(shardID)
	selfNotarizedHeaders, selfNotarizedHeadersHashes := bbt.computeSelfNotarizedHeaders(crossNotarizedHeaders)

	if len(crossNotarizedHeaders) > 0 {
		bbt.callCrossNotarizedHeadersHandlers(shardID, crossNotarizedHeaders, crossNotarizedHeadersHashes)
	}

	if len(selfNotarizedHeaders) > 0 {
		bbt.callSelfNotarizedHeadersHandlers(shardID, selfNotarizedHeaders, selfNotarizedHeadersHashes)
	}
}

func (bbt *baseBlockTrack) computeLongestChainFromLastCrossNotarized(shardID uint32) (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, err := bbt.GetLastCrossNotarizedHeader(shardID)
	if err != nil {
		return nil, nil, nil, nil
	}

	headers, hashes := bbt.ComputeLongestChain(shardID, lastCrossNotarizedHeader)
	return lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, headers, hashes
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
