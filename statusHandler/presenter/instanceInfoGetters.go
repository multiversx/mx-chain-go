package presenter

import "github.com/ElrondNetwork/elrond-go/core"

// GetAppVersion will return application version
func (psh *PresenterStatusHandler) GetAppVersion() string {
	return psh.getFromCacheAsString(core.MetricAppVersion)
}

// GetNodeType will return type of node
func (psh *PresenterStatusHandler) GetNodeType() string {
	return psh.getFromCacheAsString(core.MetricNodeType)
}

// GetPublicKeyTxSign will return node public key for sign transaction
func (psh *PresenterStatusHandler) GetPublicKeyTxSign() string {
	return psh.getFromCacheAsString(core.MetricPublicKeyTxSign)
}

// GetPublicKeyBlockSign will return node public key for sign blocks
func (psh *PresenterStatusHandler) GetPublicKeyBlockSign() string {
	return psh.getFromCacheAsString(core.MetricPublicKeyBlockSign)
}

// GetShardId will return shard Id of node
func (psh *PresenterStatusHandler) GetShardId() uint64 {
	return psh.getFromCacheAsUint64(core.MetricShardId)
}

// GetCountConsensus will return count of how many times node was in consensus group
func (psh *PresenterStatusHandler) GetCountConsensus() uint64 {
	return psh.getFromCacheAsUint64(core.MetricCountConsensus)
}

// GetCountConsensusAcceptedBlocks will return count of how many times node was in concensus group and
// a block was produced
func (psh *PresenterStatusHandler) GetCountConsensusAcceptedBlocks() uint64 {
	return psh.getFromCacheAsUint64(core.MetricCountConsensusAcceptedBlocks)
}

// GetCountLeader will return count of how many time node was leader in consensus group
func (psh *PresenterStatusHandler) GetCountLeader() uint64 {
	return psh.getFromCacheAsUint64(core.MetricCountLeader)
}

// GetCountAcceptedBlocks will return count of how many accepted blocks was proposed by the node
func (psh *PresenterStatusHandler) GetCountAcceptedBlocks() uint64 {
	return psh.getFromCacheAsUint64(core.MetricCountAcceptedBlocks)
}

// GetNodeName will return node name
func (psh *PresenterStatusHandler) GetNodeName() string {
	nodeName := psh.getFromCacheAsString(core.MetricNodeDisplayName)
	if nodeName == "" {
		nodeName = "noname"
	}

	return nodeName
}

// GetTotalRewardsValue will return total value of rewards and how rewards was increased on every second
func (psh *PresenterStatusHandler) GetTotalRewardsValue() (uint64, uint64) {
	rewardsValue := psh.getFromCacheAsUint64(core.MetricRewardsValue)
	numSignedBlocks := psh.getFromCacheAsUint64(core.MetricCountConsensusAcceptedBlocks)
	totalRewards := numSignedBlocks * rewardsValue
	difRewards := totalRewards - psh.totalRewardsOld
	psh.totalRewardsOld = totalRewards
	totalRewards = totalRewards - difRewards

	return totalRewards, difRewards
}

// CalculateRewardsPerHour will return an approximation of how many elronds will earn a validator per hour
func (psh *PresenterStatusHandler) CalculateRewardsPerHour() uint64 {
	consensusGroupSize := psh.getFromCacheAsUint64(core.MetricConsensusGroupSize)
	numValidators := psh.getFromCacheAsUint64(core.MetricNumValidators)
	totalBlocks := psh.GetProbableHighestNonce()
	rounds := psh.GetCurrentRound()
	rewardsValue := psh.getFromCacheAsUint64(core.MetricRewardsValue)
	roundDuration := psh.GetRoundTime()
	secondsInAHour := uint64(3600)

	if consensusGroupSize == 0 || numValidators == 0 || totalBlocks == 0 ||
		rounds == 0 || rewardsValue == 0 || roundDuration == 0 {
		return 0
	}

	chanceToBeInConsensus := float64(consensusGroupSize) / float64(numValidators)
	hitRate := float64(totalBlocks) / float64(rounds)
	roundsPerHour := float64(secondsInAHour) / float64(roundDuration)

	erdPerHour := uint64(chanceToBeInConsensus * hitRate * roundsPerHour * float64(rewardsValue))
	return erdPerHour
}
