package block

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

func getMetricsFromMetaHeader(
	header *block.MetaBlock,
	marshalizer marshal.Marshalizer,
	appStatusHandler core.AppStatusHandler,
	headersCountInPool int,
) {
	numMiniBlocksMetaBlock := uint64(0)
	headerSize := uint64(0)

	for _, shardInfo := range header.ShardInfo {
		numMiniBlocksMetaBlock += uint64(len(shardInfo.ShardMiniBlockHeaders))
	}

	marshalizedHeader, err := marshalizer.Marshal(header)
	if err == nil {
		headerSize = uint64(len(marshalizedHeader))
	}

	appStatusHandler.SetUInt64Value(core.MetricHeaderSize, headerSize)
	appStatusHandler.SetUInt64Value(core.MetricNumTxInBlock, uint64(header.TxCount))
	appStatusHandler.SetUInt64Value(core.MetricNumMiniBlocks, numMiniBlocksMetaBlock)

	numShardHeadersFromPool := 0
	shardMBHeaderCounterMutex.Lock()
	appStatusHandler.SetUInt64Value(core.MetricNumShardHeadersProcessed, uint64(shardMBHeadersTotalProcessed))
	numShardHeadersFromPool = headersCountInPool
	appStatusHandler.SetUInt64Value(core.MetricNumShardHeadersFromPool, uint64(numShardHeadersFromPool))
	shardMBHeaderCounterMutex.Unlock()
}

func getMetricsFromBlockBody(
	body block.Body,
	marshalizer marshal.Marshalizer,
	appStatusHandler core.AppStatusHandler,
) {
	mbLen := len(body)
	miniblocksSize := uint64(0)
	totalTxCount := 0
	for i := 0; i < mbLen; i++ {
		totalTxCount += len(body[i].TxHashes)

		marshalizedBlock, err := marshalizer.Marshal(body[i])
		if err == nil {
			miniblocksSize += uint64(len(marshalizedBlock))
		}
	}
	appStatusHandler.SetUInt64Value(core.MetricNumTxInBlock, uint64(totalTxCount))
	appStatusHandler.SetUInt64Value(core.MetricNumMiniBlocks, uint64(mbLen))
	appStatusHandler.SetUInt64Value(core.MetricMiniBlocksSize, miniblocksSize)
}

func getMetricsFromHeader(
	header *block.Header,
	numTxWithDst uint64,
	totalTx int,
	marshalizer marshal.Marshalizer,
	appStatusHandler core.AppStatusHandler,
) {
	headerSize := uint64(0)
	marshalizedHeader, err := marshalizer.Marshal(header)
	if err == nil {
		headerSize = uint64(len(marshalizedHeader))
	}

	appStatusHandler.SetUInt64Value(core.MetricHeaderSize, headerSize)
	appStatusHandler.SetUInt64Value(core.MetricTxPoolLoad, numTxWithDst)
	appStatusHandler.SetUInt64Value(core.MetricNumProcessedTxs, uint64(totalTx))
}
