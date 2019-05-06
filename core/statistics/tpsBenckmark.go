package statistics

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// TpsBenchmark will calculate statistics for the network activity
type TpsBenchmark struct {
	nrOfShards            uint32
	activeNodes           uint32
	roundTime             uint64
	blockNumber           uint64
	peakTPS               float64
	averageBlockTxCount   float32
	lastBlockTxCount      uint32
	totalProcessedTxCount uint32
	shardStatisticsMut    sync.RWMutex
	shardStatistics       map[uint32]*shardStatistics
	missingNonces         map[uint64]struct{}
	missingNoncesLock     sync.RWMutex
}

type shardStatistics struct {
	shardID               uint32
	roundTime             uint64
	averageTPS            float64
	peakTPS				  float64
	lastBlockTxCount      uint32
	averageBlockTxCount   uint32
	currentBlockNonce     uint64
	totalProcessedTxCount uint32
}

// NewTPSBenchmark instantiates a new object responsible with calculating statistics for each shard tps
func NewTPSBenchmark(nrOfShards uint32, roundDuration uint64) (*TpsBenchmark, error) {
	if roundDuration == 0 {
		return nil, core.ErrInvalidRoundDuration
	}

	shardStats := make(map[uint32]*shardStatistics, 0)
	for i := uint32(0); i < nrOfShards; i++ {
		shardStats[i] = &shardStatistics{
			roundTime: roundDuration,
		}
	}
	return &TpsBenchmark{
		nrOfShards: nrOfShards,
		roundTime: roundDuration,
		shardStatistics: shardStats,
		missingNonces: make(map[uint64]struct{}, 0),
	}, nil
}

// ActiveNodes returns the number of active nodes
func (s *TpsBenchmark) ActiveNodes() uint32 {
	return s.activeNodes
}

// RoundTime returns the round duration in seconds
func (s *TpsBenchmark) RoundTime() uint64 {
	return s.roundTime
}

// BlockNumber returns the last processed block number
func (s *TpsBenchmark) BlockNumber() uint64 {
	return s.blockNumber
}

// AverageBlockTxCount returns an average of the tx/block
func (s *TpsBenchmark) AverageBlockTxCount() float32 {
	return s.averageBlockTxCount
}

// LastBlockTxCount returns the number of transactions processed in the last block
func (s *TpsBenchmark) LastBlockTxCount() uint32 {
	return s.lastBlockTxCount
}

// TotalProcessedTxCount returns the total number of processed transactions
func (s *TpsBenchmark) TotalProcessedTxCount() uint32 {
	return s.totalProcessedTxCount
}

// LiveTPS returns tps for the last block
func (s *TpsBenchmark) LiveTPS() float64 {
	return float64(uint64(s.lastBlockTxCount) / s.roundTime)
}

// PeakTPS returns tps for the last block
func (s *TpsBenchmark) PeakTPS() float64 {
	return s.peakTPS
}

// NrOfShards returns the number of shards
func (s *TpsBenchmark) NrOfShards() uint32 {
	return s.nrOfShards
}

// ShardStatistics returns the current statistical state for a given shard
func (s *TpsBenchmark) ShardStatistics() map[uint32]*shardStatistics {
	return s.shardStatistics
}

// ShardStatistic returns the current statistical state for a given shard
func (s *TpsBenchmark) ShardStatistic(shardID uint32) *shardStatistics {
	ss, ok := s.shardStatistics[shardID]
	if !ok {
		return nil
	}
	return ss
}

func (s *TpsBenchmark) isMissingNonce(nonce uint64) bool {
	if nonce >= s.blockNumber {
		return false
	}

	s.missingNoncesLock.RLock()
	_, isMissing := s.missingNonces[nonce]
	s.missingNoncesLock.RUnlock()

	return isMissing
}

func (s *TpsBenchmark) isMetaBlockRelevant(mb *block.MetaBlock) bool {
	if mb == nil {
		return false
	}
	if mb.Nonce < s.blockNumber && !s.isMissingNonce(mb.Nonce) {
		return false
	}
	if len(mb.ShardInfo) < 1 {
		return false
	}

	return true
}

// Update receives a metablock and updates all fields accordingly for each shard available in the meta block
func (s *TpsBenchmark) Update(mb *block.MetaBlock) {
	if !s.isMetaBlockRelevant(mb) {
		return
	}

	if mb.Nonce > s.blockNumber {
		for i := s.blockNumber + 1; i < mb.Nonce; i++ {
			s.addMissingNonce(i)
		}
	}
	s.removeMissingNonce(mb.Nonce)
	_ = s.updateStatistics(mb)
}

func (s *TpsBenchmark) addMissingNonce(nonce uint64) {
	s.missingNoncesLock.Lock()
	s.missingNonces[nonce] = struct{}{}
	s.missingNoncesLock.Unlock()

}

func (s *TpsBenchmark) removeMissingNonce(nonce uint64) {
	s.missingNoncesLock.Lock()
	delete(s.missingNonces, nonce)
	s.missingNoncesLock.Unlock()
}

func (s *TpsBenchmark) updateStatistics(header *block.MetaBlock) error {
	for _, shardInfo := range header.ShardInfo {
		shardStat, ok := s.shardStatistics[shardInfo.ShardId]
		if !ok {
			return core.ErrInvalidShardId
		}

		s.totalProcessedTxCount += header.TxCount
		s.lastBlockTxCount = header.TxCount
		s.averageBlockTxCount = float32(uint64(s.totalProcessedTxCount) / header.Nonce)
		s.blockNumber = header.Nonce

		currentTPS := float64(uint64(header.TxCount) / s.roundTime)
		if currentTPS > s.peakTPS {
			s.peakTPS = currentTPS
		}

		shardPeakTPS := shardStat.peakTPS
		if currentTPS > shardStat.peakTPS {
			shardPeakTPS = currentTPS
		}

		updatedShardStats := &shardStatistics{
			shardID: shardInfo.ShardId,
			roundTime: s.roundTime,
			currentBlockNonce: header.Nonce,
			totalProcessedTxCount: shardStat.totalProcessedTxCount + shardInfo.TxCount,
			averageTPS: float64(uint64(shardStat.totalProcessedTxCount + shardInfo.TxCount) / (header.Nonce * s.roundTime)),
			peakTPS: shardPeakTPS,
			lastBlockTxCount: header.TxCount,
		}

		s.shardStatistics[shardInfo.ShardId] = updatedShardStats
	}

	return nil
}

// ShardID returns the shard id of the current statistic object
func (ss *shardStatistics) ShardID() uint32 {
	return ss.shardID
}

// AverageTPS returns an average tps for all processed blocks in a shard
func (ss *shardStatistics) AverageTPS() float64 {
	return ss.averageTPS
}

// AverageBlockTxCount returns an average transaction count for
func (ss *shardStatistics) AverageBlockTxCount() uint32 {
	return ss.averageBlockTxCount
}

// CurrentBlockNonce returns the block nounce of the last processed block in a shard
func (ss *shardStatistics) CurrentBlockNonce() uint64 {
	return ss.currentBlockNonce
}

// LiveTPS returns tps for the last block
func (ss *shardStatistics) LiveTPS() float64 {
	return float64(uint64(ss.lastBlockTxCount) / ss.roundTime)
}

// PeakTPS returns peak tps for for all the blocks of the current shard
func (ss *shardStatistics) PeakTPS() float64 {
	return ss.peakTPS
}

// LastBlockTxCount returns the number of transactions included in the last block
func (ss *shardStatistics) LastBlockTxCount() uint32 {
	return ss.lastBlockTxCount
}

// TotalProcessedTxCount returns the total number of processed transactions for this shard
func (ss *shardStatistics) TotalProcessedTxCount() uint32 {
	return ss.totalProcessedTxCount
}