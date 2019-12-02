package presenter

import (
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core/constants"
)

const precisionRewards = 2

// GetAppVersion will return application version
func (psh *PresenterStatusHandler) GetAppVersion() string {
	return psh.getFromCacheAsString(constants.MetricAppVersion)
}

// GetNodeType will return type of node
func (psh *PresenterStatusHandler) GetNodeType() string {
	return psh.getFromCacheAsString(constants.MetricNodeType)
}

// GetPublicKeyTxSign will return node public key for sign transaction
func (psh *PresenterStatusHandler) GetPublicKeyTxSign() string {
	return psh.getFromCacheAsString(constants.MetricPublicKeyTxSign)
}

// GetPublicKeyBlockSign will return node public key for sign blocks
func (psh *PresenterStatusHandler) GetPublicKeyBlockSign() string {
	return psh.getFromCacheAsString(constants.MetricPublicKeyBlockSign)
}

// GetShardId will return shard Id of node
func (psh *PresenterStatusHandler) GetShardId() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricShardId)
}

// GetCountConsensus will return count of how many times node was in consensus group
func (psh *PresenterStatusHandler) GetCountConsensus() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricCountConsensus)
}

// GetCountConsensusAcceptedBlocks will return a count if how many times the node was in consensus group and
// a block was produced
func (psh *PresenterStatusHandler) GetCountConsensusAcceptedBlocks() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricCountConsensusAcceptedBlocks)
}

// GetCountLeader will return count of how many time node was leader in consensus group
func (psh *PresenterStatusHandler) GetCountLeader() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricCountLeader)
}

// GetCountAcceptedBlocks will return count of how many accepted blocks was proposed by the node
func (psh *PresenterStatusHandler) GetCountAcceptedBlocks() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricCountAcceptedBlocks)
}

// CheckSoftwareVersion will check if node is the latest version and will return latest stable version
func (psh *PresenterStatusHandler) CheckSoftwareVersion() (bool, string) {
	latestStableVersion := psh.getFromCacheAsString(constants.MetricLatestTagSoftwareVersion)
	appVersion := psh.getFromCacheAsString(constants.MetricAppVersion)

	if strings.Contains(appVersion, latestStableVersion) || latestStableVersion == "" {
		return false, latestStableVersion
	}

	return true, latestStableVersion
}

// GetNodeName will return node's display name
func (psh *PresenterStatusHandler) GetNodeName() string {
	nodeName := psh.getFromCacheAsString(constants.MetricNodeDisplayName)
	if nodeName == "" {
		nodeName = "noname"
	}

	return nodeName
}

// GetTotalRewardsValue will return total value of rewards and how the rewards were increased on every second
// Rewards estimation will be equal with :
// numSignedBlocks * denomination * Rewards
func (psh *PresenterStatusHandler) GetTotalRewardsValue() (string, string) {
	numSignedBlocks := psh.getFromCacheAsUint64(constants.MetricCountConsensusAcceptedBlocks)
	rewardsInErd := psh.computeRewardsInErd()

	totalRewardsFloat := big.NewFloat(float64(numSignedBlocks))
	totalRewardsFloat.Mul(totalRewardsFloat, rewardsInErd)
	difRewards := big.NewFloat(0).Sub(totalRewardsFloat, psh.totalRewardsOld)

	defer func() {
		psh.totalRewardsOld = totalRewardsFloat
	}()

	return psh.totalRewardsOld.Text('f', precisionRewards), difRewards.Text('f', precisionRewards)
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
	return totalRewardsPerHourFloat.Text('f', precisionRewards)
}

// GetZeros will return a string with a specific number of zeros
func (psh *PresenterStatusHandler) GetZeros() string {
	retValue := "." + strings.Repeat("0", precisionRewards)
	if retValue == "." {
		return ""
	}

	return retValue
}
