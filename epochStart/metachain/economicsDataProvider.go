package metachain

import (
	"math/big"
	"sync"
)

type epochEconomicsStatistics struct {
	numberOfBlocks                  uint64
	numberOfBlocksPerShard          map[uint32]uint64
	leaderFees                      *big.Int
	rewardsToBeDistributed          *big.Int
	rewardsToBeDistributedForBlocks *big.Int // without leader fees, protocol sustainability and developer fees
	mutEconomicsStatistics          sync.RWMutex
}

// NewEpochEconomicsStatistics creates the end of epoch economics statistics
func NewEpochEconomicsStatistics() *epochEconomicsStatistics {
	return &epochEconomicsStatistics{
		numberOfBlocksPerShard:          make(map[uint32]uint64),
		rewardsToBeDistributed:          big.NewInt(0),
		rewardsToBeDistributedForBlocks: big.NewInt(0),
	}
}

// SetNumberOfBlocks sets the number of blocks produced in the epoch
func (es *epochEconomicsStatistics) SetNumberOfBlocks(nbBlocks uint64) {
	es.mutEconomicsStatistics.Lock()
	defer es.mutEconomicsStatistics.Unlock()
	es.numberOfBlocks = nbBlocks
}

// SetNumberOfBlocksPerShard sets the number of blocks per each shard produced in the epoch
func (es *epochEconomicsStatistics) SetNumberOfBlocksPerShard(blocksPerShard map[uint32]uint64) {
	es.mutEconomicsStatistics.Lock()
	defer es.mutEconomicsStatistics.Unlock()
	es.numberOfBlocks = 0

	es.numberOfBlocksPerShard = make(map[uint32]uint64)
	for k, v := range blocksPerShard {
		es.numberOfBlocksPerShard[k] = v
		es.numberOfBlocks += v
	}
}

// SetLeadersFees sets the leader fees for the previous epoch
func (es *epochEconomicsStatistics) SetLeadersFees(fees *big.Int) {
	es.mutEconomicsStatistics.Lock()
	defer es.mutEconomicsStatistics.Unlock()

	es.leaderFees = fees
}

// SetRewardsToBeDistributed sets the rewards to be distributed at the end of the epoch (includes the rewards per block,
//the block producers fees, protocol sustainability rewards and developer fees)
func (es *epochEconomicsStatistics) SetRewardsToBeDistributed(rewards *big.Int) {
	es.mutEconomicsStatistics.Lock()
	defer es.mutEconomicsStatistics.Unlock()

	es.rewardsToBeDistributed = big.NewInt(0).Set(rewards)
}

// SetRewardsToBeDistributedForBlocks sets the rewards to be distributed at the ed of the epoch for produced blocks
func (es *epochEconomicsStatistics) SetRewardsToBeDistributedForBlocks(rewards *big.Int) {
	es.mutEconomicsStatistics.Lock()
	defer es.mutEconomicsStatistics.Unlock()

	es.rewardsToBeDistributedForBlocks = big.NewInt(0).Set(rewards)
}

// NumberOfBlocks returns the number of blocks produced in the epoch
func (es *epochEconomicsStatistics) NumberOfBlocks() uint64 {
	es.mutEconomicsStatistics.RLock()
	defer es.mutEconomicsStatistics.RUnlock()

	return es.numberOfBlocks
}

// NumberOfBlocksPerShard returns the number of blocks produced in each shard in the epoch
func (es *epochEconomicsStatistics) NumberOfBlocksPerShard() map[uint32]uint64 {
	es.mutEconomicsStatistics.RLock()
	defer es.mutEconomicsStatistics.RUnlock()

	retMap := make(map[uint32]uint64)
	for k, v := range es.numberOfBlocksPerShard {
		retMap[k] = v
	}

	return retMap
}

// LeaderFees returns the leader fees given in the previous epoch
func (es *epochEconomicsStatistics) LeaderFees() *big.Int {
	es.mutEconomicsStatistics.RLock()
	defer es.mutEconomicsStatistics.RUnlock()

	return big.NewInt(0).Set(es.leaderFees)
}

// RewardsToBeDistributed returns the rewards to be distributed at the end of epoch (includes rewards for produced
//blocks, protocol sustainability rewards, block producer fees and developer fees)
func (es *epochEconomicsStatistics) RewardsToBeDistributed() *big.Int {
	es.mutEconomicsStatistics.RLock()
	defer es.mutEconomicsStatistics.RUnlock()

	return big.NewInt(0).Set(es.rewardsToBeDistributed)
}

// RewardsToBeDistributedForBlocks returns the rewards to be distributed at the end of epoch for produced blocks
func (es *epochEconomicsStatistics) RewardsToBeDistributedForBlocks() *big.Int {
	es.mutEconomicsStatistics.RLock()
	defer es.mutEconomicsStatistics.RUnlock()

	return big.NewInt(0).Set(es.rewardsToBeDistributedForBlocks)
}

// IsInterfaceNil returns nil if the underlying object is nil
func (es *epochEconomicsStatistics) IsInterfaceNil() bool {
	return es == nil
}
