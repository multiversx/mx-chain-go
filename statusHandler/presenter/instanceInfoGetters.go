package presenter

import (
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
)

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

// GetCountConsensusAcceptedBlocks will return a count if how many times the node was in consensus group and
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

// GetNodeName will return node's display name
func (psh *PresenterStatusHandler) GetNodeName() string {
	nodeName := psh.getFromCacheAsString(core.MetricNodeDisplayName)
	if nodeName == "" {
		nodeName = "noname"
	}

	return nodeName
}

// GetTotalRewardsValue will return total value of rewards and how the rewards were increased on every second
// Rewards estimation will be equal with :
// (numSignedBlocks - numProposedBlocks) * communityPercentage * Rewards +
// numProposedBlocks * leaderPercentage * Rewards
func (psh *PresenterStatusHandler) GetTotalRewardsValue() (string, string) {
	rewardsValue := psh.getBigIntFromStringMetric(core.MetricRewardsValue)
	communityPercentage := psh.getFloatFromStringMetric(core.MetricCommunityPercentage)
	leaderPercentage := psh.getFloatFromStringMetric(core.MetricLeaderPercentage)
	numSignedBlocks := psh.getFromCacheAsUint64(core.MetricCountConsensusAcceptedBlocks)
	numProposedBlocks := psh.getFromCacheAsUint64(core.MetricCountAcceptedBlocks)

	rewardsCommunityCoefficient := math.Abs(float64(int(numSignedBlocks-numProposedBlocks))) * communityPercentage
	consensusSignedBlocksRewards := big.NewFloat(rewardsCommunityCoefficient)
	consensusSignedBlocksRewards.Mul(consensusSignedBlocksRewards, big.NewFloat(0).SetInt(rewardsValue))

	rewardsLeaderCoefficient := float64(numProposedBlocks) * leaderPercentage
	leaderProposedBlocksRewards := big.NewFloat(rewardsLeaderCoefficient)
	leaderProposedBlocksRewards.Mul(leaderProposedBlocksRewards, big.NewFloat(0).SetInt(rewardsValue))

	totalRewardsBigFloat := big.NewFloat(0).Add(consensusSignedBlocksRewards, leaderProposedBlocksRewards)
	totalRewards := new(big.Int)
	totalRewardsBigFloat.Int(totalRewards)

	difRewards := big.NewInt(0).Sub(totalRewards, psh.totalRewardsOld)
	psh.totalRewardsOld.Set(totalRewards)
	totalRewards.Sub(totalRewards, difRewards)

	return totalRewards.Text(10), difRewards.Text(10)
}

// CalculateRewardsPerHour will return an approximation of how many ERDs a validator will earn per hour
// Rewards estimation per hour will be equals with :
// (changeToBeInConsensus - changeToBeLeader) * roundsPerHour * hitRate * communityPercentage * Rewards +
// changeToBeLeader * roundsPerHour * hitRate * leaderPercentage * Rewards
func (psh *PresenterStatusHandler) CalculateRewardsPerHour() string {
	consensusGroupSize := psh.getFromCacheAsUint64(core.MetricConsensusGroupSize)
	numValidators := psh.getFromCacheAsUint64(core.MetricNumValidators)
	totalBlocks := psh.GetProbableHighestNonce()
	rounds := psh.GetCurrentRound()
	roundDuration := psh.GetRoundTime()
	secondsInAHour := uint64(3600)
	rewardsValue := psh.getBigIntFromStringMetric(core.MetricRewardsValue)
	communityPercentage := psh.getFloatFromStringMetric(core.MetricCommunityPercentage)
	leaderPercentage := psh.getFloatFromStringMetric(core.MetricLeaderPercentage)

	areEqualsWithZero := areEqualsWithZero(consensusGroupSize, numValidators, totalBlocks, rounds, roundDuration)
	if areEqualsWithZero || rewardsValue.Cmp(big.NewInt(0)) <= 0 {
		return "0"
	}

	chanceToBeInConsensus := float64(consensusGroupSize) / float64(numValidators)
	chanceToBeLeader := 1 / float64(numValidators)
	hitRate := float64(totalBlocks) / float64(rounds)
	roundsPerHour := float64(secondsInAHour) / float64(roundDuration)

	consensusCoefficient := (chanceToBeInConsensus - chanceToBeLeader) * hitRate * roundsPerHour * communityPercentage
	consensusRewardsPerHour := big.NewFloat(consensusCoefficient)
	consensusRewardsPerHour.Mul(consensusRewardsPerHour, big.NewFloat(0).SetInt(rewardsValue))

	leaderCoefficient := chanceToBeLeader * hitRate * roundsPerHour * leaderPercentage
	leaderRewardsPerHour := big.NewFloat(leaderCoefficient)
	leaderRewardsPerHour.Mul(leaderRewardsPerHour, big.NewFloat(0).SetInt(rewardsValue))

	totalRewardsPerHourFloat := big.NewFloat(0).Add(consensusRewardsPerHour, leaderRewardsPerHour)
	totalRewardsPerHour := new(big.Int)
	totalRewardsPerHourFloat.Int(totalRewardsPerHour)

	return totalRewardsPerHour.Text(10)
}
