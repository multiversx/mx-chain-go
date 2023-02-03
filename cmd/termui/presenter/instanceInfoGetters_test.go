package presenter

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPresenterStatusHandler_GetAppVersion(t *testing.T) {
	t.Parallel()

	appVersion := "version001"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(common.MetricAppVersion, appVersion)
	result := presenterStatusHandler.GetAppVersion()

	assert.Equal(t, appVersion, result)
}

func TestPresenterStatusHandler_GetNodeType(t *testing.T) {
	t.Parallel()

	nodeType := "validator"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(common.MetricNodeType, nodeType)
	result := presenterStatusHandler.GetNodeType()

	assert.Equal(t, nodeType, result)
}

func TestPresenterStatusHandler_GetPublicKeyBlockSign(t *testing.T) {
	t.Parallel()

	publicKeyBlock := "publicKeyBlockSign"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(common.MetricPublicKeyBlockSign, publicKeyBlock)
	result := presenterStatusHandler.GetPublicKeyBlockSign()

	assert.Equal(t, publicKeyBlock, result)
}

func TestPresenterStatusHandler_GetShardId(t *testing.T) {
	t.Parallel()

	shardId := uint64(1)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(common.MetricShardId, shardId)
	result := presenterStatusHandler.GetShardId()

	assert.Equal(t, shardId, result)
}

func TestPresenterStatusHandler_GetCountConsensus(t *testing.T) {
	t.Parallel()

	countConsensus := uint64(100)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(common.MetricCountConsensus, countConsensus)
	result := presenterStatusHandler.GetCountConsensus()

	assert.Equal(t, countConsensus, result)
}

func TestPresenterStatusHandler_GetCountLeader(t *testing.T) {
	t.Parallel()

	countLeader := uint64(100)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(common.MetricCountLeader, countLeader)
	result := presenterStatusHandler.GetCountLeader()

	assert.Equal(t, countLeader, result)
}

func TestPresenterStatusHandler_GetCountAcceptedBlocks(t *testing.T) {
	t.Parallel()

	countAcceptedBlocks := uint64(100)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(common.MetricCountAcceptedBlocks, countAcceptedBlocks)
	result := presenterStatusHandler.GetCountAcceptedBlocks()

	assert.Equal(t, countAcceptedBlocks, result)
}

func TestPresenterStatusHandler_CheckSoftwareVersionNeedUpdate(t *testing.T) {
	t.Parallel()

	appVersion := "v20/go123/adsds"
	softwareVersion := "v21"

	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(common.MetricAppVersion, appVersion)
	presenterStatusHandler.SetStringValue(common.MetricLatestTagSoftwareVersion, softwareVersion)
	needUpdate, latestSoftwareVersion := presenterStatusHandler.CheckSoftwareVersion()

	assert.Equal(t, true, needUpdate)
	assert.Equal(t, softwareVersion, latestSoftwareVersion)
}

func TestPresenterStatusHandler_CheckSoftwareVersion(t *testing.T) {
	t.Parallel()

	appVersion := "v21/go123/adsds"
	softwareVersion := "v21"

	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(common.MetricAppVersion, appVersion)
	presenterStatusHandler.SetStringValue(common.MetricLatestTagSoftwareVersion, softwareVersion)
	needUpdate, latestSoftwareVersion := presenterStatusHandler.CheckSoftwareVersion()

	assert.Equal(t, false, needUpdate)
	assert.Equal(t, softwareVersion, latestSoftwareVersion)
}

func TestPresenterStatusHandler_GetCountConsensusAcceptedBlocks(t *testing.T) {
	t.Parallel()

	countConsensusAcceptedBlocks := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(common.MetricCountConsensusAcceptedBlocks, countConsensusAcceptedBlocks)
	result := presenterStatusHandler.GetCountConsensusAcceptedBlocks()

	assert.Equal(t, countConsensusAcceptedBlocks, result)

}

func TestPresenterStatusHandler_GetNodeNameShouldReturnDefaultName(t *testing.T) {
	t.Parallel()

	nodeName := ""
	expectedName := "noname"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(common.MetricNodeDisplayName, nodeName)
	result := presenterStatusHandler.GetNodeName()

	assert.Equal(t, expectedName, result)
}

func TestPresenterStatusHandler_GetNodeName(t *testing.T) {
	t.Parallel()

	nodeName := "node"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(common.MetricNodeDisplayName, nodeName)
	result := presenterStatusHandler.GetNodeName()

	assert.Equal(t, nodeName, result)
}

func TestPresenterStatusHandler_GetRedundancyLevel(t *testing.T) {
	t.Parallel()

	testRedundancyParsing(t, "-1", -1)
	testRedundancyParsing(t, "0", 0)
	testRedundancyParsing(t, "invalid", 0)
	testRedundancyParsing(t, "1", 1)
}

func testRedundancyParsing(t *testing.T, input string, desiredOutput int64) {
	psh := NewPresenterStatusHandler()
	psh.SetStringValue(common.MetricRedundancyLevel, input)
	redLev := psh.GetRedundancyLevel()
	require.Equal(t, desiredOutput, redLev)
}
