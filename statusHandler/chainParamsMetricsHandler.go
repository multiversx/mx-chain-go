package statusHandler

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
)

type chainParamsMetricsHandler struct {
	appStatusHandler core.AppStatusHandler
}

// NewChainParamsMetricsHandler will create a new chain parameters metrics handler component
func NewChainParamsMetricsHandler(appStatusHandler core.AppStatusHandler) (*chainParamsMetricsHandler, error) {
	if check.IfNil(appStatusHandler) {
		return nil, ErrNilAppStatusHandler
	}

	return &chainParamsMetricsHandler{
		appStatusHandler: appStatusHandler,
	}, nil
}

// ChainParametersChanged will be called when new chain parameters are confirmed on the network
func (cpm *chainParamsMetricsHandler) ChainParametersChanged(chainParameters config.ChainParametersByEpochConfig) {
	cpm.appStatusHandler.SetUInt64Value(common.MetricRoundsPerEpoch, uint64(chainParameters.RoundsPerEpoch))
	cpm.appStatusHandler.SetUInt64Value(common.MetricShardConsensusGroupSize, uint64(chainParameters.ShardConsensusGroupSize))
	cpm.appStatusHandler.SetUInt64Value(common.MetricMetaConsensusGroupSize, uint64(chainParameters.MetachainConsensusGroupSize))
	cpm.appStatusHandler.SetUInt64Value(common.MetricNumNodesPerShard, uint64(chainParameters.ShardMinNumNodes))
	cpm.appStatusHandler.SetUInt64Value(common.MetricNumMetachainNodes, uint64(chainParameters.MetachainMinNumNodes))
}

// IsInterfaceNil returns true if there is no value under the interface
func (cpm *chainParamsMetricsHandler) IsInterfaceNil() bool {
	return cpm == nil
}
