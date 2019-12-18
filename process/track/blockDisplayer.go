package track

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

func (bbt *baseBlockTrack) displayHeaders() {
	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	for shardID := range bbt.headers {
		bbt.displayHeadersForShard(shardID)
	}
}
func (bbt *baseBlockTrack) displayHeadersForShard(shardID uint32) {
	bbt.displayTrackedHeadersForShard(shardID)
	bbt.displayCrossNotarizedHeadersForShard(shardID)
	bbt.displaySelfNotarizedHeadersForShard(shardID)
}

func (bbt *baseBlockTrack) displayTrackedHeadersForShard(shardID uint32) {
	log.Trace("tracked headers", "shard", shardID)

	headers, hashes := bbt.sortHeadersForShardFromNonce(shardID, 0)
	for index, header := range headers {
		log.Trace("tracked header info",
			"round", header.GetRound(),
			"nonce", header.GetNonce(),
			"hash", hashes[index])
	}
}

func (bbt *baseBlockTrack) displayCrossNotarizedHeadersForShard(shardID uint32) {
	log.Trace("cross notarized headers", "shard", shardID)

	bbt.mutCrossNotarizedHeaders.RLock()

	crossNotarizedHeadersForShard, ok := bbt.crossNotarizedHeaders[shardID]
	if ok {
		process.SortHeadersByNonce(crossNotarizedHeadersForShard)
		for _, header := range crossNotarizedHeadersForShard {
			log.Trace("cross notarized header info",
				"round", header.GetRound(),
				"nonce", header.GetNonce())
		}
	}

	bbt.mutCrossNotarizedHeaders.RUnlock()
}

func (bbt *baseBlockTrack) displaySelfNotarizedHeadersForShard(shardID uint32) {
	log.Trace("self notarized headers", "shard", shardID)

	bbt.mutSelfNotarizedHeaders.RLock()

	selfNotarizedHeadersForShard, ok := bbt.selfNotarizedHeaders[shardID]
	if ok {
		process.SortHeadersByNonce(selfNotarizedHeadersForShard)
		for _, header := range selfNotarizedHeadersForShard {
			log.Trace("self notarized header info",
				"round", header.GetRound(),
				"nonce", header.GetNonce())
		}
	}

	bbt.mutSelfNotarizedHeaders.RUnlock()
}
