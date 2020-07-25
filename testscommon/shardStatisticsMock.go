package testscommon

import "math/big"

// ShardStatisticsMock will hold the tps statistics for each shard
type ShardStatisticsMock struct {
	shardID               uint32
	roundTime             uint64
	averageTPS            *big.Int
	peakTPS               float64
	lastBlockTxCount      uint32
	averageBlockTxCount   uint32
	currentBlockNonce     uint64
	totalProcessedTxCount *big.Int
}

// ShardID returns the shard id of the current statistic object
func (ss *ShardStatisticsMock) ShardID() uint32 {
	return ss.shardID
}

// AverageTPS returns an average tps for all processed blocks in a shard
func (ss *ShardStatisticsMock) AverageTPS() *big.Int {
	return ss.averageTPS
}

// AverageBlockTxCount returns an average transaction count for
func (ss *ShardStatisticsMock) AverageBlockTxCount() uint32 {
	return ss.averageBlockTxCount
}

// CurrentBlockNonce returns the block nounce of the last processed block in a shard
func (ss *ShardStatisticsMock) CurrentBlockNonce() uint64 {
	return ss.currentBlockNonce
}

// LiveTPS returns tps for the last block
func (ss *ShardStatisticsMock) LiveTPS() float64 {
	return float64(uint64(ss.lastBlockTxCount) / ss.roundTime)
}

// PeakTPS returns peak tps for for all the blocks of the current shard
func (ss *ShardStatisticsMock) PeakTPS() float64 {
	return ss.peakTPS
}

// LastBlockTxCount returns the number of transactions included in the last block
func (ss *ShardStatisticsMock) LastBlockTxCount() uint32 {
	return ss.lastBlockTxCount
}

// TotalProcessedTxCount returns the total number of processed transactions for this shard
func (ss *ShardStatisticsMock) TotalProcessedTxCount() *big.Int {
	return ss.totalProcessedTxCount
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *ShardStatisticsMock) IsInterfaceNil() bool {
	return ss == nil
}
