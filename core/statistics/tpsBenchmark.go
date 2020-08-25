package statistics

import (
	"math/big"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

var log = logger.GetOrCreate("statistics")

// defaultBlockNumber is used to identify the default value of the value representing the block number fetched from storage.
// it is used to signal that no value was read from storage and the check for not updating total number of processed
// transactions should be skipped.
const defaultBlockNumber = -1

// TpsPersistentData holds the tps benchmark data which is stored between node restarts
type TpsPersistentData struct {
	BlockNumber           uint64
	RoundNumber           uint64
	PeakTPS               float64
	AverageBlockTxCount   *big.Int
	TotalProcessedTxCount *big.Int
	LastBlockTxCount      uint32
}

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
	statusHandler         core.AppStatusHandler
	initialBlockNumber    int64
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

// NewTPSBenchmarkWithInitialData instantiates a new object responsible with calculating statistics for each shard tps
// starting with initial data
func NewTPSBenchmarkWithInitialData(
	appStatusHandler core.AppStatusHandler,
	initialTpsBenchmark *TpsPersistentData,
	nrOfShards uint32,
	roundDuration uint64,
) (*TpsBenchmark, error) {
	if roundDuration == 0 {
		return nil, ErrInvalidRoundDuration
	}
	if initialTpsBenchmark == nil {
		return nil, ErrNilInitialTPSBenchmarks
	}
	if check.IfNil(appStatusHandler) {
		return nil, ErrNilStatusHandler
	}

	shardStats := make(map[uint32]ShardStatistic)
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
		peakTPS:               initialTpsBenchmark.PeakTPS,
		lastBlockTxCount:      initialTpsBenchmark.LastBlockTxCount,
		blockNumber:           initialTpsBenchmark.BlockNumber,
		roundNumber:           initialTpsBenchmark.RoundNumber,
		totalProcessedTxCount: initialTpsBenchmark.TotalProcessedTxCount,
		averageBlockTxCount:   initialTpsBenchmark.AverageBlockTxCount,
		statusHandler:         appStatusHandler,
		missingNonces:         make(map[uint64]struct{}),
		initialBlockNumber:    int64(initialTpsBenchmark.BlockNumber),
	}, nil
}

// NewTPSBenchmark instantiates a new object responsible with calculating statistics for each shard tps.
// nrOfShards represents the total number of shards, roundDuration is the duration for a round in seconds
func NewTPSBenchmark(
	nrOfShards uint32,
	roundDuration uint64,
) (*TpsBenchmark, error) {
	if roundDuration == 0 {
		return nil, ErrInvalidRoundDuration
	}

	shardStats := make(map[uint32]ShardStatistic)
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
		statusHandler:         statusHandler.NewNilStatusHandler(),
		missingNonces:         make(map[uint64]struct{}),
		totalProcessedTxCount: big.NewInt(0),
		averageBlockTxCount:   big.NewInt(0),
		initialBlockNumber:    defaultBlockNumber,
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

	_, isMissing := s.missingNonces[nonce]

	return isMissing
}

func (s *TpsBenchmark) isMetaBlockRelevant(mb *block.MetaBlock) bool {
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
	if mblock == nil || mblock.IsInterfaceNil() {
		return
	}

	mb := mblock.(*block.MetaBlock)
	if mb == nil {
		return
	}

	s.mut.Lock()
	if !s.isMetaBlockRelevant(mb) {
		s.mut.Unlock()
		return
	}

	if mb.Nonce > s.blockNumber {
		for i := s.blockNumber + 1; i < mb.Nonce; i++ {
			s.addMissingNonce(i)
		}
	}
	s.removeMissingNonce(mb.Nonce)
	_ = s.updateStatistics(mb)
	s.mut.Unlock()
}

func (s *TpsBenchmark) addMissingNonce(nonce uint64) {
	s.missingNonces[nonce] = struct{}{}
}

func (s *TpsBenchmark) removeMissingNonce(nonce uint64) {
	delete(s.missingNonces, nonce)
}

func (s *TpsBenchmark) updateStatistics(header *block.MetaBlock) error {
	s.blockNumber = header.Nonce
	s.roundNumber = header.Round

	totalTxsWithoutPeerTxs := getNumOfTxsWithoutPeerTxs(header)

	s.lastBlockTxCount = uint32(totalTxsWithoutPeerTxs)
	shouldUpdateTotalNumAndPeak := s.shouldUpdateFields(header)
	if shouldUpdateTotalNumAndPeak {
		s.totalProcessedTxCount.Add(s.totalProcessedTxCount, big.NewInt(0).SetUint64(totalTxsWithoutPeerTxs))
		s.statusHandler.AddUint64(core.MetricNumProcessedTxs, totalTxsWithoutPeerTxs)
	}
	s.averageBlockTxCount.Quo(s.totalProcessedTxCount, big.NewInt(int64(header.Nonce)))

	currentTPS := float64(totalTxsWithoutPeerTxs / s.roundTime)
	if currentTPS > s.peakTPS && shouldUpdateTotalNumAndPeak {
		s.peakTPS = currentTPS
	}

	s.statusHandler.SetUInt64Value(core.MetricNonceForTPS, header.Nonce)
	s.statusHandler.SetUInt64Value(core.MetricLastBlockTxCount, totalTxsWithoutPeerTxs)
	s.statusHandler.SetUInt64Value(core.MetricPeakTPS, uint64(s.peakTPS))
	s.statusHandler.SetStringValue(core.MetricAverageBlockTxCount, s.averageBlockTxCount.String())

	for _, shardInfo := range header.ShardInfo {
		shardStat, ok := s.shardStatistics[shardInfo.ShardID]
		if !ok {
			return ErrInvalidShardId
		}

		totalTxsFromShardBlockWithoutPeerTxs := getNumTxsFromMiniblocksWithoutPeerTxs(shardInfo.ShardMiniBlockHeaders)

		shardPeakTPS := shardStat.PeakTPS()
		currentShardTPS := float64(totalTxsFromShardBlockWithoutPeerTxs / s.roundTime)
		if currentShardTPS > shardStat.PeakTPS() {
			shardPeakTPS = currentShardTPS
		}

		bigTxCount := big.NewInt(0).SetUint64(totalTxsFromShardBlockWithoutPeerTxs)
		newTotalProcessedTxCount := big.NewInt(0).Add(shardStat.TotalProcessedTxCount(), bigTxCount)
		roundsPassed := big.NewInt(int64(header.Round))
		newAverageTPS := big.NewInt(0).Quo(newTotalProcessedTxCount, roundsPassed)

		updatedShardStats := &ShardStatistics{
			shardID:               shardInfo.ShardID,
			roundTime:             s.roundTime,
			currentBlockNonce:     header.Nonce,
			totalProcessedTxCount: newTotalProcessedTxCount,

			averageTPS:       newAverageTPS,
			peakTPS:          shardPeakTPS,
			lastBlockTxCount: uint32(totalTxsFromShardBlockWithoutPeerTxs),
		}

		log.Debug("TpsBenchmark.updateStatistics",
			"shard", updatedShardStats.shardID,
			"block", updatedShardStats.currentBlockNonce,
			"avgTPS", updatedShardStats.averageTPS,
			"peakTPS", updatedShardStats.peakTPS,
			"lastBlockTxCount", updatedShardStats.lastBlockTxCount,
			"avgBlockTxCount", updatedShardStats.averageBlockTxCount,
			"totalProcessedTxCount", updatedShardStats.totalProcessedTxCount,
		)

		s.shardStatistics[shardInfo.ShardID] = updatedShardStats
	}

	return nil
}

func getNumOfTxsWithoutPeerTxs(metaBlock *block.MetaBlock) uint64 {
	// get number of transactions from metablock miniblocks
	totalTxs := getNumTxsFromMiniblocksWithoutPeerTxs(metaBlock.MiniBlockHeaders)

	// get number of transactions from shard blocks that are included in metablock
	for idx := 0; idx < len(metaBlock.ShardInfo); idx++ {
		totalTxs += getNumTxsFromMiniblocksWithoutPeerTxs(metaBlock.ShardInfo[idx].ShardMiniBlockHeaders)
	}

	return totalTxs
}

func getNumTxsFromMiniblocksWithoutPeerTxs(miniblocks []block.MiniBlockHeader) uint64 {
	totalTxs := uint64(0)
	for idx := 0; idx < len(miniblocks); idx++ {
		if miniblocks[idx].Type == block.PeerBlock {
			continue
		}

		totalTxs += uint64(miniblocks[idx].TxCount)
	}

	return totalTxs
}

func (s *TpsBenchmark) shouldUpdateFields(metaBlock *block.MetaBlock) bool {
	if s.initialBlockNumber == defaultBlockNumber {
		return true
	}

	return uint64(s.initialBlockNumber) < metaBlock.Nonce
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *TpsBenchmark) IsInterfaceNil() bool {
	return s == nil
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

// IsInterfaceNil returns true if there is no value under the interface
func (ss *ShardStatistics) IsInterfaceNil() bool {
	return ss == nil
}
