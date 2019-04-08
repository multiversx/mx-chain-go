package node

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// statisticsManager will calculate statistics for the network activity
type statisticsManager struct {
	activeNodes           uint32
	roundTime             uint32
	blockNumber           uint32
	averageBlockTxCount   float32
	lastBlockTxCount      uint32
	totalProcessedTxCount uint32
	shardStatisticsMut    sync.RWMutex
	shardStatistics       map[uint32]*shardStatistics
}

type shardStatistics struct {
	shardID             uint32
	averageTPS          float32
	lastBlockTxCount    uint32
	averageBlockTxCount uint32
	currentBlockNonce   uint32
}

func NewStatisticsManager(nrOfShards uint32) (*statisticsManager, error) {
	shardStats := make(map[uint32]*shardStatistics, 0)
	for i := uint32(0); i < nrOfShards; i++ {
		shardStats[i] = &shardStatistics{}
	}
	return &statisticsManager{
		shardStatistics: shardStats,
	}, nil
}

func (s *statisticsManager) ActiveNodes() uint32 {
	return s.activeNodes
}

func (s *statisticsManager) RoundTime() uint32 {
	return s.roundTime
}

func (s *statisticsManager) BlockNumber() uint32 {
	return s.blockNumber
}

func (s *statisticsManager) AverageBlockTxCount() float32 {
	return s.averageBlockTxCount
}

func (s *statisticsManager) LastBlockTxCount() uint32 {
	return s.lastBlockTxCount
}

func (s *statisticsManager) TotalProcessedTxCount() uint32 {
	return s.totalProcessedTxCount
}

func (s *statisticsManager) UpdateShardStatistics(shardID, blockNonce, txCount uint32) error {
	s.shardStatisticsMut.RLock()
	shardStat, ok := s.shardStatistics[shardID]
	s.shardStatisticsMut.RUnlock()
	if !ok {
		return process.ErrInvalidShardId
	}



	return nil
}

func (ss *shardStatistics) ShardID() uint32 {
	return ss.shardID
}

func (ss *shardStatistics) AverageTPS() float32 {
	return ss.averageTPS
}

func (ss *shardStatistics) AverageBlockTxCount() uint32 {
	return ss.averageBlockTxCount
}

func (ss *shardStatistics) CurrentBlockNonce() uint32 {
	return ss.currentBlockNonce
}
