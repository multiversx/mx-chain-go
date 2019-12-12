package track

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

	headers := bbt.sortHeadersForShardFromNonce(shardID, 0)
	for _, header := range headers {
		log.Trace("traked header info",
			"round", header.GetRound(),
			"nonce", header.GetNonce())
	}
}

func (bbt *baseBlockTrack) displayCrossNotarizedHeadersForShard(shardID uint32) {
	log.Trace("cross notarized headers", "shard", shardID)

	bbt.mutCrossNotarizedHeaders.RLock()

	crossNotarizedHeadersForShard, ok := bbt.crossNotarizedHeaders[shardID]
	if ok {
		for _, header := range crossNotarizedHeadersForShard {
			log.Trace("cross notarized header info",
				"round", header.GetRound(),
				"nonce", header.GetNonce())
		}
	}

	bbt.mutCrossNotarizedHeaders.RUnlock()
}

func (bbt *baseBlockTrack) displaySelfNotarizedHeadersForShard(shardID uint32) {
	log.Debug("self notarized headers", "shard", shardID)

	bbt.mutSelfNotarizedHeaders.RLock()

	selfNotarizedHeadersForShard, ok := bbt.selfNotarizedHeaders[shardID]
	if ok {
		for _, header := range selfNotarizedHeadersForShard {
			log.Debug("self notarized header info",
				"round", header.GetRound(),
				"nonce", header.GetNonce())
		}
	}

	bbt.mutSelfNotarizedHeaders.RUnlock()
}
