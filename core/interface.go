package core

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

// Core interface will abstract all the subpackage functionalities and will
//  provide access to it's members where needed
type Core interface {
	Indexer() Indexer
	TPSBenchmark() TPSBenchmark
}

// Indexer is an interface for saving node specific data to other storages.
// This could be an elasticsearch intex, a MySql database or any other external services.
type Indexer interface {
	SaveBlock(body block.Body, header *block.Header, txPool map[string]*transaction.Transaction)
	SaveMetaBlock(metaBlock *block.MetaBlock, headerPool map[string]*block.Header)
	UpdateTPS(tpsBenchmark TPSBenchmark)
}

// TPSBenchmark is an interface used to calculate statistics for the network activity
type TPSBenchmark interface {
	Update(mb *block.MetaBlock)
	ActiveNodes() uint32
	RoundTime() uint64
	BlockNumber() uint64
	RoundNumber() uint64
	AverageBlockTxCount() *big.Int
	LastBlockTxCount() uint32
	TotalProcessedTxCount() *big.Int
	LiveTPS() float64
	PeakTPS() float64
	NrOfShards() uint32
	ShardStatistics() map[uint32]ShardStatistic
	ShardStatistic(shardID uint32) ShardStatistic
}

// ShardStatistic is an interface used to calculate statistics for the network activity of a specific shard
type ShardStatistic interface {
	ShardID() uint32
	AverageTPS() *big.Int
	AverageBlockTxCount() uint32
	CurrentBlockNonce() uint64
	LiveTPS() float64
	PeakTPS() float64
	LastBlockTxCount() uint32
	TotalProcessedTxCount() *big.Int
}
