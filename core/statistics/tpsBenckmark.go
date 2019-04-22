package statistics

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// TpsBenchmark will calculate statistics for the network activity
type TpsBenchmark struct {
	nrOfShards            uint32
	activeNodes           uint32
	roundTime             uint32
	blockNumber           uint64
	peakTPS               float32
	averageBlockTxCount   float32
	lastBlockTxCount      uint32
	totalProcessedTxCount uint32
	shardStatisticsMut    sync.RWMutex
	shardStatistics       map[uint32]*shardStatistics
	headerCacher          storage.Cacher
}

type shardStatistics struct {
	shardID               uint32
	roundTime             uint32
	averageTPS            float32
	peakTPS				  float32
	lastBlockTxCount      uint32
	averageBlockTxCount   uint32
	currentBlockNonce     uint64
	missingNonces         map[uint64]struct{}
	missingNoncesLock     sync.RWMutex
	totalProcessedTxCount uint32
}

// NewTPSBenchmark instantiates a new object responsible with calculating statistics for each shard tps
func NewTPSBenchmark(nrOfShards uint32, headerCacher storage.Cacher) (*TpsBenchmark, error) {
	shardStats := make(map[uint32]*shardStatistics, 0)
	for i := uint32(0); i < nrOfShards; i++ {
		shardStats[i] = &shardStatistics{
			roundTime: 4,
			missingNonces: make(map[uint64]struct{}, 0),
		}
	}
	return &TpsBenchmark{
		nrOfShards: nrOfShards,
		roundTime: 4,
		shardStatistics: shardStats,
		headerCacher: headerCacher,
	}, nil
}

// ActiveNodes returns the number of active nodes
func (s *TpsBenchmark) ActiveNodes() uint32 {
	return s.activeNodes
}

// RoundTime returns the round duration in seconds
func (s *TpsBenchmark) RoundTime() uint32 {
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
func (s *TpsBenchmark) LiveTPS() float32 {
	return float32(s.lastBlockTxCount / s.roundTime)
}

// PeakTPS returns tps for the last block
func (s *TpsBenchmark) PeakTPS() float32 {
	return s.peakTPS
}

// NrOfShards returns the number of shards
func (s *TpsBenchmark) NrOfShards() uint32 {
	return s.nrOfShards
}

// ShardStatistic returns the current statistical state for a given shard
func (s *TpsBenchmark) ShardStatistic(shardID uint32) *shardStatistics {
	ss, ok := s.shardStatistics[shardID]
	if !ok {
		return nil
	}
	return ss
}

// UpdateShardStatistics receives a shard id, block nounce and a transaction count and updates all fields accordingly
func (s *TpsBenchmark) 	UpdateShardStatistics(currentNonce uint64) {

	for i := 0; i < len(s.shardStatistics); i++ {
		//currentStats := s.shardStatistics[uint32(i)]

		// if there's a gap between current round and last nonce that was updated add that gap as missing nonces
		if uint64(currentNonce) > s.shardStatistics[uint32(i)].currentBlockNonce + 1 {
			for j := s.shardStatistics[uint32(i)].currentBlockNonce + 1; j < uint64(currentNonce); j++ {
				s.shardStatistics[uint32(i)].addMissingNonce(uint64(j))
			}
		}

		// Update missing nonces
		missingNonces := s.shardStatistics[uint32(i)].missingNonces

		for nonce := range missingNonces {
			missingHeader := s.getHeader(uint32(i), nonce)
			if missingHeader == nil {
				continue
			}

			s.shardStatistics[uint32(i)].removeMissingNonce(nonce)
			_ = s.updateStatistics(missingHeader)
		}

		currentShardHeader := s.getHeader(uint32(i), currentNonce)
		if currentShardHeader == nil {
			continue
		}
		s.shardStatistics[uint32(i)].removeMissingNonce(currentNonce)
		_ = s.updateStatistics(currentShardHeader)
	}
}

func (ss *shardStatistics) addMissingNonce(nonce uint64) {
	ss.missingNonces[nonce] = struct{}{}

}

func (ss *shardStatistics) removeMissingNonce(nonce uint64) {
	delete(ss.missingNonces, nonce)
}

func (s *TpsBenchmark) getHeader(shardID uint32, nonce uint64) *block.Header {
	headerKey := fmt.Sprintf("%d_%d", shardID, nonce)
	header, ok := s.headerCacher.Peek([]byte(headerKey))
	if !ok {
		return nil
	}

	currentShardHeader, ok := header.(*block.Header)
	if !ok {
		return nil
	}

	return currentShardHeader
}

func (s *TpsBenchmark) updateStatistics(header *block.Header) error {
	shardStat, ok := s.shardStatistics[header.ShardId]

	if !ok {
		return core.ErrInvalidShardId
	}

	if shardStat.currentBlockNonce >= header.Nonce {
		return core.ErrInvalidNonce
	}

	s.totalProcessedTxCount += header.TxCount
	s.lastBlockTxCount = header.TxCount
	s.averageBlockTxCount = float32(uint64(s.totalProcessedTxCount) / header.Nonce)
	s.blockNumber = header.Nonce

	currentTPS := float32(header.TxCount / s.roundTime)
	if currentTPS > s.peakTPS {
		s.peakTPS = currentTPS
	}

	shardPeakTPS := shardStat.peakTPS
	if currentTPS > shardStat.peakTPS {
		shardPeakTPS = currentTPS
	}

	updatedShardStats := &shardStatistics{
		shardID: header.ShardId,
		roundTime: s.roundTime,
		currentBlockNonce: header.Nonce,
		totalProcessedTxCount: shardStat.totalProcessedTxCount + header.TxCount,
		averageTPS: float32(uint64(shardStat.totalProcessedTxCount + header.TxCount) / (header.Nonce * uint64(s.roundTime))),
		peakTPS: shardPeakTPS,
		lastBlockTxCount: header.TxCount,
		missingNonces: shardStat.missingNonces,
	}

	s.shardStatistics[header.ShardId] = updatedShardStats

	return nil
}

// ShardID returns the shard id of the current statistic object
func (ss *shardStatistics) ShardID() uint32 {
	return ss.shardID
}

// AverageTPS returns an average tps for all processed blocks in a shard
func (ss *shardStatistics) AverageTPS() float32 {
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
func (ss *shardStatistics) LiveTPS() float32 {
	return float32(ss.lastBlockTxCount / ss.roundTime)
}

// PeakTPS returns peak tps for for all the blocks of the current shard
func (ss *shardStatistics) PeakTPS() float32 {
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