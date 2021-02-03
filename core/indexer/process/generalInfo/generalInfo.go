package generalInfo

import (
	"math/big"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
)

const (
	metachainTpsDocID   = "meta"
	shardTpsDocIDPrefix = "shard"
)

var log = logger.GetOrCreate("indexer/process/generalInfo")

type infoProcessor struct {
}

// NewGeneralInfoProcessor will create a new instance of general info processor
func NewGeneralInfoProcessor() *infoProcessor {
	return &infoProcessor{}
}

// PrepareGeneralInfo will prepare and general information about chain
func (gip *infoProcessor) PrepareGeneralInfo(tpsBenchmark statistics.TPSBenchmark) (*types.TPS, []*types.TPS) {
	generalInfo := &types.TPS{
		LiveTPS:               tpsBenchmark.LiveTPS(),
		PeakTPS:               tpsBenchmark.PeakTPS(),
		NrOfShards:            tpsBenchmark.NrOfShards(),
		BlockNumber:           tpsBenchmark.BlockNumber(),
		RoundNumber:           tpsBenchmark.RoundNumber(),
		RoundTime:             tpsBenchmark.RoundTime(),
		AverageBlockTxCount:   tpsBenchmark.AverageBlockTxCount(),
		LastBlockTxCount:      tpsBenchmark.LastBlockTxCount(),
		TotalProcessedTxCount: tpsBenchmark.TotalProcessedTxCount(),
	}

	shardsInfo := make([]*types.TPS, 0)
	for _, shardInfo := range tpsBenchmark.ShardStatistics() {
		bigTxCount := big.NewInt(int64(shardInfo.AverageBlockTxCount()))
		shardTPS := &types.TPS{
			ShardID:               shardInfo.ShardID(),
			LiveTPS:               shardInfo.LiveTPS(),
			PeakTPS:               shardInfo.PeakTPS(),
			AverageTPS:            shardInfo.AverageTPS(),
			AverageBlockTxCount:   bigTxCount,
			CurrentBlockNonce:     shardInfo.CurrentBlockNonce(),
			LastBlockTxCount:      shardInfo.LastBlockTxCount(),
			TotalProcessedTxCount: shardInfo.TotalProcessedTxCount(),
		}

		shardsInfo = append(shardsInfo, shardTPS)
	}

	return generalInfo, shardsInfo
}
