package block

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
)

type baseBlockNotarizer struct {
	blockTracker process.BlockTracker
}

func (bbn *baseBlockNotarizer) getLastCrossNotarizedHeadersForShard(shardID uint32) *bootstrapStorage.BootstrapHeaderInfo {
	lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, err := bbn.blockTracker.GetLastCrossNotarizedHeader(shardID)
	if err != nil {
		log.Warn("getLastCrossNotarizedHeadersForShard",
			"shard", shardID,
			"error", err.Error())
		return nil
	}

	if lastCrossNotarizedHeader.GetNonce() == 0 {
		return nil
	}

	headerInfo := &bootstrapStorage.BootstrapHeaderInfo{
		ShardId: lastCrossNotarizedHeader.GetShardID(),
		Nonce:   lastCrossNotarizedHeader.GetNonce(),
		Hash:    lastCrossNotarizedHeaderHash,
	}

	return headerInfo
}
