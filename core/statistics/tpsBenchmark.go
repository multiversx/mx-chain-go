package statistics

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// TpsBenchmark will calculate statistics for the network activity
type TpsBenchmark struct {
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
	shardStatistics       map[uint32]ShardStatistic
	missingNonces         map[uint64]struct{}
	missingNoncesLock     sync.RWMutex
}

// ShardStatistics will hold the tps statistics for each shard
type ShardStatistics struct {
	shardID               uint32
	roundTime             uint64
	averageTPS            *big.Int
	peakTPS               float64
	lastBlockTxCount      uint32
	averageBlockTxCount   uint32
	currentBlockNonce     uint64
	totalProcessedTxCount *big.Int
}

// NewTPSBenchmark instantiates a new object responsible with calculating statistics for each shard tps.
// nrOfShards represents the total number of shards, roundDuration is the duration for a round in seconds
func NewTPSBenchmark(nrOfShards uint32, roundDuration uint64) (*TpsBenchmark, error) {
	if roundDuration == 0 {
		return nil, ErrInvalidRoundDuration
	}

	shardStats := make(map[uint32]ShardStatistic, 0)
	for i := uint32(0); i < nrOfShards; i++ {
		shardStats[i] = &ShardStatistics{
			roundTime:             roundDuration,
			totalProcessedTxCount: big.NewInt(0),
		}
	}
	return &TpsBenchmark{
		nrOfShards:            nrOfShards,
		roundTime:             roundDuration,
		shardStatistics:       shardStats,
		missingNonces:         make(map[uint64]struct{}, 0),
		totalProcessedTxCount: big.NewInt(0),
		averageBlockTxCount:   big.NewInt(0),
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

// RoundNumber returns the round index for this benchmark object
func (s *TpsBenchmark) RoundNumber() uint64 {
	return s.roundNumber
}

// AverageBlockTxCount returns an average of the tx/block
func (s *TpsBenchmark) AverageBlockTxCount() *big.Int {
	return s.averageBlockTxCount
}

// LastBlockTxCount returns the number of transactions processed in the last block
func (s *TpsBenchmark) LastBlockTxCount() uint32 {
	return s.lastBlockTxCount
}

// TotalProcessedTxCount returns the total number of processed transactions
func (s *TpsBenchmark) TotalProcessedTxCount() *big.Int {
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
func (s *TpsBenchmark) ShardStatistics() map[uint32]ShardStatistic {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.shardStatistics
}

// ShardStatistic returns the current statistical state for a given shard
func (s *TpsBenchmark) ShardStatistic(shardID uint32) ShardStatistic {
	s.mut.RLock()
	defer s.mut.RUnlock()

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
	s.mut.RLock()
	defer s.mut.RUnlock()
	if mb == nil {
		return false
	}
	if mb.Nonce <= s.blockNumber && !s.isMissingNonce(mb.Nonce) {
		return false
	}
	if len(mb.ShardInfo) < 1 {
		return false
	}

	return true
}

// Update receives a metablock and updates all fields accordingly for each shard available in the meta block
func (s *TpsBenchmark) Update(mblock data.HeaderHandler) {
	if mblock == nil {
		return
	}

	mb := mblock.(*block.MetaBlock)
	if mb == nil {
		return
	}

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

func (s *TpsBenchmark) setAverageTxCountForRound(round uint64) {
	bigNonce := big.NewInt(0)
	bigNonce.SetUint64(round)
	s.averageBlockTxCount.Quo(s.totalProcessedTxCount, bigNonce)
}

func (s *TpsBenchmark) updateStatistics(header *block.MetaBlock) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.blockNumber = header.Nonce
	s.roundNumber = uint64(header.Round)
	s.lastBlockTxCount = header.TxCount
	s.totalProcessedTxCount.Add(s.totalProcessedTxCount, big.NewInt(int64(header.TxCount)))
	s.averageBlockTxCount.Quo(s.totalProcessedTxCount, big.NewInt(int64(header.Nonce)))

	currentTPS := float64(uint64(header.TxCount) / s.roundTime)
	if currentTPS > s.peakTPS {
		s.peakTPS = currentTPS
	}

	for _, shardInfo := range header.ShardInfo {
		shardStat, ok := s.shardStatistics[shardInfo.ShardId]
		if !ok {
			return ErrInvalidShardId
		}

		shardPeakTPS := shardStat.PeakTPS()
		currentShardTPS := float64(uint64(shardInfo.TxCount) / s.roundTime)
		if currentShardTPS > shardStat.PeakTPS() {
			shardPeakTPS = currentShardTPS
		}

		bigTxCount := big.NewInt(int64(shardInfo.TxCount))
		newTotalProcessedTxCount := big.NewInt(0).Add(shardStat.TotalProcessedTxCount(), bigTxCount)
		roundsPassed := big.NewInt(int64(header.Round))
		newAverageTPS := big.NewInt(0).Quo(newTotalProcessedTxCount, roundsPassed)

		updatedShardStats := &ShardStatistics{
			shardID:               shardInfo.ShardId,
			roundTime:             s.roundTime,
			currentBlockNonce:     header.Nonce,
			totalProcessedTxCount: newTotalProcessedTxCount,

			averageTPS:       newAverageTPS,
			peakTPS:          shardPeakTPS,
			lastBlockTxCount: header.TxCount,
		}

		s.shardStatistics[shardInfo.ShardId] = updatedShardStats
	}

	return nil
}

// ShardID returns the shard id of the current statistic object
func (ss *ShardStatistics) ShardID() uint32 {
	return ss.shardID
}

// AverageTPS returns an average tps for all processed blocks in a shard
func (ss *ShardStatistics) AverageTPS() *big.Int {
	return ss.averageTPS
}

// AverageBlockTxCount returns an average transaction count for
func (ss *ShardStatistics) AverageBlockTxCount() uint32 {
	return ss.averageBlockTxCount
}

// CurrentBlockNonce returns the block nounce of the last processed block in a shard
func (ss *ShardStatistics) CurrentBlockNonce() uint64 {
	return ss.currentBlockNonce
}

// LiveTPS returns tps for the last block
func (ss *ShardStatistics) LiveTPS() float64 {
	return float64(uint64(ss.lastBlockTxCount) / ss.roundTime)
}

// PeakTPS returns peak tps for for all the blocks of the current shard
func (ss *ShardStatistics) PeakTPS() float64 {
	return ss.peakTPS
}

// LastBlockTxCount returns the number of transactions included in the last block
func (ss *ShardStatistics) LastBlockTxCount() uint32 {
	return ss.lastBlockTxCount
}

// TotalProcessedTxCount returns the total number of processed transactions for this shard
func (ss *ShardStatistics) TotalProcessedTxCount() *big.Int {
	return ss.totalProcessedTxCount
}
