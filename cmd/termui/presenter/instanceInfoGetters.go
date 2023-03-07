package presenter

import (
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-go/common"
)

// GetAppVersion will return application version
func (psh *PresenterStatusHandler) GetAppVersion() string {
	return psh.getFromCacheAsString(common.MetricAppVersion)
}

// GetNodeType will return type of node
func (psh *PresenterStatusHandler) GetNodeType() string {
	return psh.getFromCacheAsString(common.MetricNodeType)
}

// GetPeerType will return type of peer (eligible, waiting, and so on)
func (psh *PresenterStatusHandler) GetPeerType() string {
	return psh.getFromCacheAsString(common.MetricPeerType)
}

// GetPeerSubType will return subtype of peer (regular, full history)
func (psh *PresenterStatusHandler) GetPeerSubType() string {
	return psh.getFromCacheAsString(common.MetricPeerSubType)
}

// GetPublicKeyBlockSign will return node public key for sign blocks
func (psh *PresenterStatusHandler) GetPublicKeyBlockSign() string {
	return psh.getFromCacheAsString(common.MetricPublicKeyBlockSign)
}

// GetRedundancyLevel will return the redundancy level of the node
func (psh *PresenterStatusHandler) GetRedundancyLevel() int64 {
	// redundancy level is sent as string as JSON unmarshal doesn't treat well the casting from interface{} to int64
	redundancyLevelStr := psh.getFromCacheAsString(common.MetricRedundancyLevel)
	i64Val, err := strconv.ParseInt(redundancyLevelStr, 10, 64)
	if err != nil {
		return 0
	}

	return i64Val
}

// GetRedundancyIsMainActive will return the info about redundancy main machine
func (psh *PresenterStatusHandler) GetRedundancyIsMainActive() string {
	return psh.getFromCacheAsString(common.MetricRedundancyIsMainActive)
}

// GetShardId will return shard ID of node
func (psh *PresenterStatusHandler) GetShardId() uint64 {
	return psh.getFromCacheAsUint64(common.MetricShardId)
}

// GetChainID returns the chain ID
func (psh *PresenterStatusHandler) GetChainID() string {
	return psh.getFromCacheAsString(common.MetricChainId)
}

// GetCountConsensus will return count of how many times node was in consensus group
func (psh *PresenterStatusHandler) GetCountConsensus() uint64 {
	return psh.getFromCacheAsUint64(common.MetricCountConsensus)
}

// GetCountConsensusAcceptedBlocks will return a count if how many times the node was in consensus group and
// a block was produced
func (psh *PresenterStatusHandler) GetCountConsensusAcceptedBlocks() uint64 {
	return psh.getFromCacheAsUint64(common.MetricCountConsensusAcceptedBlocks)
}

// GetCountLeader will return count of how many times node was leader in consensus group
func (psh *PresenterStatusHandler) GetCountLeader() uint64 {
	return psh.getFromCacheAsUint64(common.MetricCountLeader)
}

// GetCountAcceptedBlocks will return count of how many accepted blocks was proposed by the node
func (psh *PresenterStatusHandler) GetCountAcceptedBlocks() uint64 {
	return psh.getFromCacheAsUint64(common.MetricCountAcceptedBlocks)
}

// CheckSoftwareVersion will check if node is the latest version and will return the latest stable version
func (psh *PresenterStatusHandler) CheckSoftwareVersion() (bool, string) {
	latestStableVersion := psh.getFromCacheAsString(common.MetricLatestTagSoftwareVersion)
	appVersion := psh.getFromCacheAsString(common.MetricAppVersion)

	if strings.Contains(appVersion, latestStableVersion) || latestStableVersion == "" {
		return false, latestStableVersion
	}

	return true, latestStableVersion
}

// GetNodeName will return node's display name
func (psh *PresenterStatusHandler) GetNodeName() string {
	nodeName := psh.getFromCacheAsString(common.MetricNodeDisplayName)
	if nodeName == "" {
		nodeName = "noname"
	}

	return nodeName
}
