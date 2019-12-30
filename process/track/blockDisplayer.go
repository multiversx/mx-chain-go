package track

import (
	"sort"
)

func (bbt *baseBlockTrack) displayHeadersForShard(shardID uint32) {
	bbt.displayTrackedHeadersForShard(shardID)
	bbt.displayCrossNotarizedHeadersForShard(shardID)
	bbt.displaySelfNotarizedHeadersForShard(shardID)
}

func (bbt *baseBlockTrack) displayTrackedHeadersForShard(shardID uint32) {
	headers, hashes := bbt.sortHeadersForShardFromNonce(shardID, 0)
	shouldNotDisplay := len(headers) == 0 ||
		len(headers) == 1 && headers[0].GetNonce() == 0
	if shouldNotDisplay {
		return
	}

	log.Trace("tracked headers", "shard", shardID)
	for index, header := range headers {
		log.Trace("tracked header info",
			"round", header.GetRound(),
			"nonce", header.GetNonce(),
			"hash", hashes[index])
	}
}

func (bbt *baseBlockTrack) displayCrossNotarizedHeadersForShard(shardID uint32) {
	bbt.mutCrossNotarizedHeaders.RLock()
	defer bbt.mutCrossNotarizedHeaders.RUnlock()

	crossNotarizedHeadersForShard, ok := bbt.crossNotarizedHeaders[shardID]
	if ok {
		if len(crossNotarizedHeadersForShard) > 1 {
			sort.Slice(crossNotarizedHeadersForShard, func(i, j int) bool {
				return crossNotarizedHeadersForShard[i].header.GetNonce() < crossNotarizedHeadersForShard[j].header.GetNonce()
			})
		}

		shouldNotDisplay := len(crossNotarizedHeadersForShard) == 0 ||
			len(crossNotarizedHeadersForShard) == 1 && crossNotarizedHeadersForShard[0].header.GetNonce() == 0
		if shouldNotDisplay {
			return
		}

		log.Trace("cross notarized headers", "shard", shardID)
		for _, headerInfo := range crossNotarizedHeadersForShard {
			log.Trace("cross notarized header info",
				"round", headerInfo.header.GetRound(),
				"nonce", headerInfo.header.GetNonce(),
				"hash", headerInfo.hash)
		}
	}
}

func (bbt *baseBlockTrack) displaySelfNotarizedHeadersForShard(shardID uint32) {
	bbt.mutSelfNotarizedHeaders.RLock()
	defer bbt.mutSelfNotarizedHeaders.RUnlock()

	selfNotarizedHeadersForShard, ok := bbt.selfNotarizedHeaders[shardID]
	if ok {
		if len(selfNotarizedHeadersForShard) > 1 {
			sort.Slice(selfNotarizedHeadersForShard, func(i, j int) bool {
				return selfNotarizedHeadersForShard[i].header.GetNonce() < selfNotarizedHeadersForShard[j].header.GetNonce()
			})
		}

		shouldNotDisplay := len(selfNotarizedHeadersForShard) == 0 ||
			len(selfNotarizedHeadersForShard) == 1 && selfNotarizedHeadersForShard[0].header.GetNonce() == 0
		if shouldNotDisplay {
			return
		}

		log.Trace("self notarized headers", "shard", shardID)
		for _, headerInfo := range selfNotarizedHeadersForShard {
			log.Trace("self notarized header info",
				"round", headerInfo.header.GetRound(),
				"nonce", headerInfo.header.GetNonce(),
				"hash", headerInfo.hash)
		}
	}
}
