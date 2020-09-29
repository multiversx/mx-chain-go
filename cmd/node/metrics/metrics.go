package metrics

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const millisecondsInSecond = 1000

// StatusHandlersUtils provides some functionality for statusHandlers
type StatusHandlersUtils interface {
	StatusHandler() core.AppStatusHandler
	SignalStartViews()
	SignalLogRewrite()
	IsInterfaceNil() bool
}

// InitMetrics will init metrics for status handler
func InitMetrics(
	statusHandlerUtils StatusHandlersUtils,
	pubkeyStr string,
	nodeType core.NodeType,
	shardCoordinator sharding.Coordinator,
	nodesConfig sharding.GenesisNodesSetupHandler,
	version string,
	economicsConfig *config.EconomicsConfig,
	roundsPerEpoch int64,
	minTransactionVersion uint32,
) error {
	if check.IfNil(statusHandlerUtils) {
		return fmt.Errorf("nil StatusHandlerUtils when initializing metrics")
	}
	if check.IfNil(shardCoordinator) {
		return fmt.Errorf("nil shard coordinator when initializing metrics")
	}
	if nodesConfig == nil {
		return fmt.Errorf("nil nodes config when initializing metrics")
	}
	if economicsConfig == nil {
		return fmt.Errorf("nil economics config when initializing metrics")
	}

	shardId := uint64(shardCoordinator.SelfId())
	numOfShards := uint64(shardCoordinator.NumberOfShards())
	roundDuration := nodesConfig.GetRoundDuration()
	isSyncing := uint64(1)
	initUint := uint64(0)
	initString := ""
	appStatusHandler := statusHandlerUtils.StatusHandler()

	appStatusHandler.SetUInt64Value(core.MetricSynchronizedRound, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNonce, initUint)
	appStatusHandler.SetStringValue(core.MetricPublicKeyBlockSign, pubkeyStr)
	appStatusHandler.SetUInt64Value(core.MetricShardId, shardId)
	appStatusHandler.SetUInt64Value(core.MetricNumShardsWithoutMetacahin, numOfShards)
	appStatusHandler.SetStringValue(core.MetricNodeType, string(nodeType))
	appStatusHandler.SetUInt64Value(core.MetricRoundTime, roundDuration/millisecondsInSecond)
	appStatusHandler.SetStringValue(core.MetricAppVersion, version)
	appStatusHandler.SetUInt64Value(core.MetricRoundsPerEpoch, uint64(roundsPerEpoch))
	appStatusHandler.SetUInt64Value(core.MetricCountConsensus, initUint)
	appStatusHandler.SetUInt64Value(core.MetricCountLeader, initUint)
	appStatusHandler.SetUInt64Value(core.MetricCountAcceptedBlocks, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumTxInBlock, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumMiniBlocks, initUint)
	appStatusHandler.SetStringValue(core.MetricConsensusState, initString)
	appStatusHandler.SetStringValue(core.MetricConsensusRoundState, initString)
	appStatusHandler.SetStringValue(core.MetricCrossCheckBlockHeight, "0")
	appStatusHandler.SetUInt64Value(core.MetricIsSyncing, isSyncing)
	appStatusHandler.SetStringValue(core.MetricCurrentBlockHash, initString)
	appStatusHandler.SetUInt64Value(core.MetricNumProcessedTxs, initUint)
	appStatusHandler.SetUInt64Value(core.MetricCurrentRoundTimestamp, initUint)
	appStatusHandler.SetUInt64Value(core.MetricHeaderSize, initUint)
	appStatusHandler.SetUInt64Value(core.MetricMiniBlocksSize, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumShardHeadersFromPool, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumShardHeadersProcessed, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumTimesInForkChoice, initUint)
	appStatusHandler.SetUInt64Value(core.MetricHighestFinalBlock, initUint)
	appStatusHandler.SetUInt64Value(core.MetricCountConsensusAcceptedBlocks, initUint)
	appStatusHandler.SetUInt64Value(core.MetricRoundAtEpochStart, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNonceAtEpochStart, initUint)
	appStatusHandler.SetUInt64Value(core.MetricRoundsPassedInCurrentEpoch, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNoncesPassedInCurrentEpoch, initUint)
	appStatusHandler.SetStringValue(core.MetricLeaderPercentage, fmt.Sprintf("%f", economicsConfig.RewardsSettings.LeaderPercentage))
	appStatusHandler.SetUInt64Value(core.MetricDenomination, uint64(economicsConfig.GlobalSettings.Denomination))
	appStatusHandler.SetUInt64Value(core.MetricNumConnectedPeers, initUint)
	appStatusHandler.SetStringValue(core.MetricNumConnectedPeersClassification, initString)
	appStatusHandler.SetStringValue(core.MetricLatestTagSoftwareVersion, initString)

	appStatusHandler.SetStringValue(core.MetricP2PNumConnectedPeersClassification, initString)
	appStatusHandler.SetStringValue(core.MetricP2PPeerInfo, initString)
	appStatusHandler.SetStringValue(core.MetricP2PIntraShardValidators, initString)
	appStatusHandler.SetStringValue(core.MetricP2PIntraShardObservers, initString)
	appStatusHandler.SetStringValue(core.MetricP2PCrossShardValidators, initString)
	appStatusHandler.SetStringValue(core.MetricP2PCrossShardObservers, initString)
	appStatusHandler.SetStringValue(core.MetricP2PUnknownPeers, initString)
	appStatusHandler.SetUInt64Value(core.MetricShardConsensusGroupSize, uint64(nodesConfig.GetShardConsensusGroupSize()))
	appStatusHandler.SetUInt64Value(core.MetricMetaConsensusGroupSize, uint64(nodesConfig.GetMetaConsensusGroupSize()))
	appStatusHandler.SetUInt64Value(core.MetricNumNodesPerShard, uint64(nodesConfig.MinNumberOfShardNodes()))
	appStatusHandler.SetUInt64Value(core.MetricNumMetachainNodes, uint64(nodesConfig.MinNumberOfMetaNodes()))
	appStatusHandler.SetUInt64Value(core.MetricStartTime, uint64(nodesConfig.GetStartTime()))
	appStatusHandler.SetUInt64Value(core.MetricRoundDuration, nodesConfig.GetRoundDuration())
	appStatusHandler.SetUInt64Value(core.MetricMinTransactionVersion, uint64(minTransactionVersion))

	var consensusGroupSize uint32
	switch {
	case shardCoordinator.SelfId() < shardCoordinator.NumberOfShards():
		consensusGroupSize = nodesConfig.GetShardConsensusGroupSize()
	case shardCoordinator.SelfId() == core.MetachainShardId:
		consensusGroupSize = nodesConfig.GetMetaConsensusGroupSize()
	default:
		consensusGroupSize = 0
	}

	validatorsNodes, _ := nodesConfig.InitialNodesInfo()
	numValidators := len(validatorsNodes[shardCoordinator.SelfId()])

	appStatusHandler.SetUInt64Value(core.MetricNumValidators, uint64(numValidators))
	appStatusHandler.SetUInt64Value(core.MetricConsensusGroupSize, uint64(consensusGroupSize))

	statusHandlerUtils.SignalLogRewrite()
	statusHandlerUtils.SignalStartViews()

	return nil
}

// SaveUint64Metric will save a uint64 metric in status handler
func SaveUint64Metric(ash core.AppStatusHandler, key string, value uint64) {
	ash.SetUInt64Value(key, value)
}

// SaveStringMetric will save a string metric in status handler
func SaveStringMetric(ash core.AppStatusHandler, key, value string) {
	ash.SetStringValue(key, value)
}
