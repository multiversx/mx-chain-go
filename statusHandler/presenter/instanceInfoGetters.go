package presenter

import (
	"math/big"
	"strings"

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

// CheckSoftwareVersion will check if node is the latest version and will return latest stable version
func (psh *PresenterStatusHandler) CheckSoftwareVersion() (bool, string) {
	latestStableVersion := psh.getFromCacheAsString(core.MetricLatestTagSoftwareVersion)
	appVersion := psh.getFromCacheAsString(core.MetricAppVersion)

	if strings.Contains(appVersion, latestStableVersion) || latestStableVersion == "" {
		return false, latestStableVersion
	}

	return true, latestStableVersion
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
// numSignedBlocks * denomination * Rewards
func (psh *PresenterStatusHandler) GetTotalRewardsValue() (string, string) {
	numSignedBlocks := psh.getFromCacheAsUint64(core.MetricCountConsensusAcceptedBlocks)
	rewardsInErd := psh.computeRewardsInErd()

	totalRewardsFloat := big.NewFloat(float64(numSignedBlocks))
	totalRewardsFloat.Mul(totalRewardsFloat, rewardsInErd)

	totalRewards := new(big.Int)
	totalRewardsFloat.Int(totalRewards)

	difRewards := big.NewInt(0).Sub(totalRewards, psh.totalRewardsOld)
	psh.totalRewardsOld.Set(totalRewards)
	totalRewards.Sub(totalRewards, difRewards)

	return totalRewards.Text(10), difRewards.Text(10)
}

// CalculateRewardsPerHour will return an approximation of how many ERDs a validator will earn per hour
// Rewards estimation per hour will be equals with :
// chanceToBeInConsensus * roundsPerHour * hitRate * denominationCoefficient * Rewards
func (psh *PresenterStatusHandler) CalculateRewardsPerHour() string {
	chanceToBeInConsensus := psh.computeChanceToBeInConsensus()
	roundsPerHourAccordingToHitRate := psh.computeRoundsPerHourAccordingToHitRate()
	rewardsInErd := psh.computeRewardsInErd()
	if chanceToBeInConsensus == 0 || roundsPerHourAccordingToHitRate == 0 || rewardsInErd.Cmp(big.NewFloat(0)) <= 0 {
		return "0"
	}

	rewardsPerHourCoefficient := chanceToBeInConsensus * roundsPerHourAccordingToHitRate
	totalRewardsPerHourFloat := big.NewFloat(rewardsPerHourCoefficient)
	totalRewardsPerHourFloat.Mul(totalRewardsPerHourFloat, rewardsInErd)

	totalRewardsPerHour := new(big.Int)
	totalRewardsPerHourFloat.Int(totalRewardsPerHour)

	return totalRewardsPerHour.Text(10)
}
