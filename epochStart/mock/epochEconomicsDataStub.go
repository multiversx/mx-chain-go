package mock

import "math/big"

// EpochEconomicsData provides a mock implementation of the EpochEconomicsDataProvider interface
type EpochEconomicsData struct {
	RewardsForProtocolSustainabilityVal *big.Int
	RewardsForEcosystemGrowthVal        *big.Int
	RewardsForGrowthDividendVal         *big.Int
}

// SetNumberOfBlocks provides a mock implementation of the SetNumberOfBlocks method
func (eed *EpochEconomicsData) SetNumberOfBlocks(nbBlocks uint64) {
}

// SetNumberOfBlocksPerShard provides a mock implementation of the SetNumberOfBlocksPerShard method
func (eed *EpochEconomicsData) SetNumberOfBlocksPerShard(blocksPerShard map[uint32]uint64) {
}

// SetLeadersFees provides a mock implementation of the SetLeadersFees method
func (eed *EpochEconomicsData) SetLeadersFees(fees *big.Int) {
}

// SetRewardsToBeDistributed provides a mock implementation of the SetRewardsToBeDistributed method
func (eed *EpochEconomicsData) SetRewardsToBeDistributed(rewards *big.Int) {
}

// SetRewardsToBeDistributedForBlocks provides a mock implementation of the SetRewardsToBeDistributedForBlocks method
func (eed *EpochEconomicsData) SetRewardsToBeDistributedForBlocks(rewards *big.Int) {
}

// SetRewardsForProtocolSustainability provides a mock implementation of the SetRewardsForProtocolSustainability method
func (eed *EpochEconomicsData) SetRewardsForProtocolSustainability(rewards *big.Int) {
	eed.RewardsForProtocolSustainabilityVal = rewards
}

// SetRewardsForEcosystemGrowth provides a mock implementation of the SetRewardsForEcosystemGrowth method
func (eed *EpochEconomicsData) SetRewardsForEcosystemGrowth(rewards *big.Int) {
	eed.RewardsForEcosystemGrowthVal = rewards
}

// SetRewardsForGrowthDividend provides a mock implementation of the SetRewardsForGrowthDividend method
func (eed *EpochEconomicsData) SetRewardsForGrowthDividend(rewards *big.Int) {
	eed.RewardsForGrowthDividendVal = rewards
}

// NumberOfBlocks provides a mock implementation of the NumberOfBlocks method
func (eed *EpochEconomicsData) NumberOfBlocks() uint64 {
	return 0
}

// NumberOfBlocksPerShard provides a mock implementation of the NumberOfBlocksPerShard method
func (eed *EpochEconomicsData) NumberOfBlocksPerShard() map[uint32]uint64 {
	return nil
}

// LeaderFees provides a mock implementation of the LeaderFees method
func (eed *EpochEconomicsData) LeaderFees() *big.Int {
	return nil
}

// RewardsToBeDistributed provides a mock implementation of the RewardsToBeDistributed method
func (eed *EpochEconomicsData) RewardsToBeDistributed() *big.Int {
	return nil
}

// RewardsToBeDistributedForBlocks provides a mock implementation of the RewardsToBeDistributedForBlocks method
func (eed *EpochEconomicsData) RewardsToBeDistributedForBlocks() *big.Int {
	return nil
}

// RewardsForProtocolSustainability provides a mock implementation of the RewardsForProtocolSustainability method
func (eed *EpochEconomicsData) RewardsForProtocolSustainability() *big.Int {
	return eed.RewardsForProtocolSustainabilityVal
}

// RewardsForEcosystemGrowth provides a mock implementation of the RewardsForEcosystemGrowth method
func (eed *EpochEconomicsData) RewardsForEcosystemGrowth() *big.Int {
	return eed.RewardsForEcosystemGrowthVal
}

// RewardsForGrowthDividend provides a mock implementation of the RewardsForGrowthDividend method
func (eed *EpochEconomicsData) RewardsForGrowthDividend() *big.Int {
	return eed.RewardsForGrowthDividendVal
}

// IsInterfaceNil provides a mock implementation of the IsInterfaceNil method
func (eed *EpochEconomicsData) IsInterfaceNil() bool {
	return eed == nil
}
