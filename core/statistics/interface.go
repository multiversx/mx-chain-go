package statistics

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// TPSBenchmark is an interface used to calculate statistics for the network activity
type TPSBenchmark interface {
	Update(mb data.HeaderHandler)
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
	IsInterfaceNil() bool
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
	IsInterfaceNil() bool
}

// SoftwareVersionChecker holds the actions needed to be handled by a components which will check the software version
type SoftwareVersionChecker interface {
	StartCheckSoftwareVersion()
	IsInterfaceNil() bool
	Close() error
}

// ResourceMonitorHandler defines the resource monitor supported actions
type ResourceMonitorHandler interface {
	GenerateStatistics() []interface{}
	SaveStatistics()
	StartMonitoring()
	Close() error
	IsInterfaceNil() bool
}
