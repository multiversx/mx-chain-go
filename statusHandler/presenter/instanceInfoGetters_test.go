package presenter

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestPresenterStatusHandler_GetAppVersion(t *testing.T) {
	t.Parallel()

	appVersion := "version001"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricAppVersion, appVersion)
	result := presenterStatusHandler.GetAppVersion()

	assert.Equal(t, appVersion, result)
}

func TestPresenterStatusHandler_GetNodeType(t *testing.T) {
	t.Parallel()

	nodeType := "validator"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricNodeType, nodeType)
	result := presenterStatusHandler.GetNodeType()

	assert.Equal(t, nodeType, result)
}

func TestPresenterStatusHandler_GetPublicKeyTxSign(t *testing.T) {
	t.Parallel()

	publicKey := "publicKeyTxSign"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricPublicKeyTxSign, publicKey)
	result := presenterStatusHandler.GetPublicKeyTxSign()

	assert.Equal(t, publicKey, result)
}

func TestPresenterStatusHandler_GetPublicKeyBlockSign(t *testing.T) {
	t.Parallel()

	publicKeyBlock := "publicKeyBlockSign"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricPublicKeyBlockSign, publicKeyBlock)
	result := presenterStatusHandler.GetPublicKeyBlockSign()

	assert.Equal(t, publicKeyBlock, result)
}

func TestPresenterStatusHandler_GetShardId(t *testing.T) {
	t.Parallel()

	shardId := uint64(1)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricShardId, shardId)
	result := presenterStatusHandler.GetShardId()

	assert.Equal(t, shardId, result)
}

func TestPresenterStatusHandler_GetCountConsensus(t *testing.T) {
	t.Parallel()

	countConsensus := uint64(100)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricCountConsensus, countConsensus)
	result := presenterStatusHandler.GetCountConsensus()

	assert.Equal(t, countConsensus, result)
}

func TestPresenterStatusHandler_GetCountLeader(t *testing.T) {
	t.Parallel()

	countLeader := uint64(100)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricCountLeader, countLeader)
	result := presenterStatusHandler.GetCountLeader()

	assert.Equal(t, countLeader, result)
}

func TestPresenterStatusHandler_GetCountAcceptedBlocks(t *testing.T) {
	t.Parallel()

	countAcceptedBlocks := uint64(100)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricCountAcceptedBlocks, countAcceptedBlocks)
	result := presenterStatusHandler.GetCountAcceptedBlocks()

	assert.Equal(t, countAcceptedBlocks, result)
}

func TestPresenterStatusHandler_GetCountConsensusAcceptedBlocks(t *testing.T) {
	t.Parallel()

	countConsensusAcceptedBlocks := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricCountConsensusAcceptedBlocks, countConsensusAcceptedBlocks)
	result := presenterStatusHandler.GetCountConsensusAcceptedBlocks()

	assert.Equal(t, countConsensusAcceptedBlocks, result)

}

func TestPresenterStatusHandler_GetNodeNameShouldReturnDefaultName(t *testing.T) {
	t.Parallel()

	nodeName := ""
	expectedName := "noname"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricNodeDisplayName, nodeName)
	result := presenterStatusHandler.GetNodeName()

	assert.Equal(t, expectedName, result)
}

func TestPresenterStatusHandler_GetNodeName(t *testing.T) {
	t.Parallel()

	nodeName := "node"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricNodeDisplayName, nodeName)
	result := presenterStatusHandler.GetNodeName()

	assert.Equal(t, nodeName, result)
}

func TestPresenterStatusHandler_CalculateRewardsTotal(t *testing.T) {
	t.Parallel()

	rewardsValue := "1000"
	expectedDifValue := "5"
	numSignedBlocks := uint64(50)

	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricRewardsValue, rewardsValue)
	presenterStatusHandler.SetUInt64Value(core.MetricCountConsensusAcceptedBlocks, numSignedBlocks)
	totalRewards, diff := presenterStatusHandler.GetTotalRewardsValue()

	assert.Equal(t, "0", totalRewards)
	assert.Equal(t, expectedDifValue, diff)
}

func TestPresenterStatusHandler_CalculateRewardsTotalRewards(t *testing.T) {
	t.Parallel()

	rewardsValue := "1000"
	numSignedBlocks := uint64(50000)
	expectedDiffValue := "4000"

	presenterStatusHandler := NewPresenterStatusHandler()
	totalRewardsOld, _ := big.NewInt(0).SetString(rewardsValue, 10)
	presenterStatusHandler.totalRewardsOld = big.NewInt(0).Set(totalRewardsOld)
	presenterStatusHandler.SetStringValue(core.MetricRewardsValue, rewardsValue)
	presenterStatusHandler.SetUInt64Value(core.MetricCountConsensusAcceptedBlocks, numSignedBlocks)
	totalRewards, diff := presenterStatusHandler.GetTotalRewardsValue()

	assert.Equal(t, totalRewardsOld.Text(10), totalRewards)
	assert.Equal(t, expectedDiffValue, diff)
}

func TestPresenterStatusHandler_CalculateRewardsPerHourReturnZero(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := NewPresenterStatusHandler()
	result := presenterStatusHandler.CalculateRewardsPerHour()

	assert.Equal(t, "0", result)
}

func TestPresenterStatusHandler_CalculateRewardsPerHourShouldWork(t *testing.T) {
	t.Parallel()

	consensusGroupSize := uint64(50)
	numValidators := uint64(100)
	totalBlocks := uint64(1000)
	totalRounds := uint64(1000)
	roundTime := uint64(6)
	rewardsValue := "10000"
	expectedValue := "300"

	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricConsensusGroupSize, consensusGroupSize)
	presenterStatusHandler.SetUInt64Value(core.MetricNumValidators, numValidators)
	presenterStatusHandler.SetUInt64Value(core.MetricProbableHighestNonce, totalBlocks)
	presenterStatusHandler.SetStringValue(core.MetricRewardsValue, rewardsValue)
	presenterStatusHandler.SetUInt64Value(core.MetricCurrentRound, totalRounds)
	presenterStatusHandler.SetUInt64Value(core.MetricRoundTime, roundTime)

	result := presenterStatusHandler.CalculateRewardsPerHour()
	assert.Equal(t, expectedValue, result)
}
