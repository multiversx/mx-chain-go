package testscommon

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// TpsBenchmarkMock will calculate statistics for the network activity
type TpsBenchmarkMock struct {
	mut                   sync.RWMutex
	nrOfShards            uint32
	activeNodes           uint32
	roundTime             uint64
	blockNumber           uint64
	roundNumber           uint64
	peakTPS               float64
	averageBlockTxCount   *big.Int
	lastBlockTxCount      uint32
	totalProcessedTxCount *big.Int
	shardStatistics       map[uint32]statistics.ShardStatistic
}

// ActiveNodes returns the number of active nodes
func (s *TpsBenchmarkMock) ActiveNodes() uint32 {
	return s.activeNodes
}

// RoundTime returns the round duration in seconds
func (s *TpsBenchmarkMock) RoundTime() uint64 {
	return s.roundTime
}

// BlockNumber returns the last processed block number
func (s *TpsBenchmarkMock) BlockNumber() uint64 {
	return s.blockNumber
}

// RoundNumber returns the round index for this benchmark object
func (s *TpsBenchmarkMock) RoundNumber() uint64 {
	return s.roundNumber
}

// AverageBlockTxCount returns an average of the tx/block
func (s *TpsBenchmarkMock) AverageBlockTxCount() *big.Int {
	return s.averageBlockTxCount
}

// LastBlockTxCount returns the number of transactions processed in the last block
func (s *TpsBenchmarkMock) LastBlockTxCount() uint32 {
	return s.lastBlockTxCount
}

// TotalProcessedTxCount returns the total number of processed transactions
func (s *TpsBenchmarkMock) TotalProcessedTxCount() *big.Int {
	return s.totalProcessedTxCount
}

// LiveTPS returns tps for the last block
func (s *TpsBenchmarkMock) LiveTPS() float64 {
	return float64(uint64(s.lastBlockTxCount) / s.roundTime)
}

// PeakTPS returns tps for the last block
func (s *TpsBenchmarkMock) PeakTPS() float64 {
	return s.peakTPS
}

// NrOfShards returns the number of shards
func (s *TpsBenchmarkMock) NrOfShards() uint32 {
	return s.nrOfShards
}

// ShardStatistics returns the current statistical state for a given shard
func (s *TpsBenchmarkMock) ShardStatistics() map[uint32]statistics.ShardStatistic {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.shardStatistics
}

// ShardStatistic returns the current statistical state for a given shard
func (s *TpsBenchmarkMock) ShardStatistic(shardID uint32) statistics.ShardStatistic {
	s.mut.RLock()
	defer s.mut.RUnlock()

	ss, ok := s.shardStatistics[shardID]
	if !ok {
		return nil
	}
	return ss
}

// Update receives a metablock and updates all fields accordingly for each shard available in the meta block
func (s *TpsBenchmarkMock) Update(mb data.HeaderHandler) {
	if mb == nil {
		return
	}

	s.blockNumber = mb.GetNonce()
	s.roundNumber = mb.GetRound()
	s.lastBlockTxCount = mb.GetTxCount()
	s.roundTime = 6

	currentTPS := float64(uint64(mb.GetTxCount()) / s.roundTime)
	if currentTPS > s.peakTPS {
		s.peakTPS = currentTPS
	}
}

// UpdateWithShardStats -
func (s *TpsBenchmarkMock) UpdateWithShardStats(mb *block.MetaBlock) {
	s.blockNumber = mb.Nonce
	s.roundNumber = mb.Round
	s.lastBlockTxCount = mb.TxCount
	s.roundTime = 6

	currentTPS := float64(uint64(mb.TxCount) / s.roundTime)
	if currentTPS > s.peakTPS {
		s.peakTPS = currentTPS
	}

	s.shardStatistics = make(map[uint32]statistics.ShardStatistic)
	for _, shardInfo := range mb.ShardInfo {
		updatedShardStats := &ShardStatisticsMock{}
		updatedShardStats.roundTime = 6
		s.shardStatistics[shardInfo.ShardID] = updatedShardStats
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *TpsBenchmarkMock) IsInterfaceNil() bool {
	return s == nil
}
