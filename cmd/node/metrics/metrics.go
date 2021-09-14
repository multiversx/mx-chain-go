package metrics

import (
	"fmt"
	"sort"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
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
	epochConfig *config.EpochConfig,
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
	if epochConfig == nil {
		return fmt.Errorf("nil epoch config when initializing metrics")
	}

	shardId := uint64(shardCoordinator.SelfId())
	numOfShards := uint64(shardCoordinator.NumberOfShards())
	roundDuration := nodesConfig.GetRoundDuration()
	isSyncing := uint64(1)
	initUint := uint64(0)
	initString := ""
	initZeroString := "0"
	appStatusHandler := statusHandlerUtils.StatusHandler()

	leaderPercentage := float64(0)
	rewardsConfigs := make([]config.EpochRewardSettings, len(economicsConfig.RewardsSettings.RewardsConfigByEpoch))
	_ = copy(rewardsConfigs, economicsConfig.RewardsSettings.RewardsConfigByEpoch)

	sort.Slice(rewardsConfigs, func(i, j int) bool {
		return rewardsConfigs[i].EpochEnable < rewardsConfigs[j].EpochEnable
	})

	if len(rewardsConfigs) > 0 {
		leaderPercentage = rewardsConfigs[0].LeaderPercentage
	}

	appStatusHandler.SetUInt64Value(common.MetricSynchronizedRound, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNonce, initUint)
	appStatusHandler.SetStringValue(common.MetricPublicKeyBlockSign, pubkeyStr)
	appStatusHandler.SetUInt64Value(common.MetricShardId, shardId)
	appStatusHandler.SetUInt64Value(common.MetricNumShardsWithoutMetachain, numOfShards)
	appStatusHandler.SetStringValue(common.MetricNodeType, string(nodeType))
	appStatusHandler.SetUInt64Value(common.MetricRoundTime, roundDuration/millisecondsInSecond)
	appStatusHandler.SetStringValue(common.MetricAppVersion, version)
	appStatusHandler.SetUInt64Value(common.MetricRoundsPerEpoch, uint64(roundsPerEpoch))
	appStatusHandler.SetUInt64Value(common.MetricCountConsensus, initUint)
	appStatusHandler.SetUInt64Value(common.MetricCountLeader, initUint)
	appStatusHandler.SetUInt64Value(common.MetricCountAcceptedBlocks, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNumTxInBlock, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNumMiniBlocks, initUint)
	appStatusHandler.SetStringValue(common.MetricConsensusState, initString)
	appStatusHandler.SetStringValue(common.MetricConsensusRoundState, initString)
	appStatusHandler.SetStringValue(common.MetricCrossCheckBlockHeight, "0")
	appStatusHandler.SetUInt64Value(common.MetricIsSyncing, isSyncing)
	appStatusHandler.SetStringValue(common.MetricCurrentBlockHash, initString)
	appStatusHandler.SetUInt64Value(common.MetricNumProcessedTxs, initUint)
	appStatusHandler.SetUInt64Value(common.MetricCurrentRoundTimestamp, initUint)
	appStatusHandler.SetUInt64Value(common.MetricHeaderSize, initUint)
	appStatusHandler.SetUInt64Value(common.MetricMiniBlocksSize, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNumShardHeadersFromPool, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNumShardHeadersProcessed, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNumTimesInForkChoice, initUint)
	appStatusHandler.SetUInt64Value(common.MetricHighestFinalBlock, initUint)
	appStatusHandler.SetUInt64Value(common.MetricCountConsensusAcceptedBlocks, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNonceAtEpochStart, initUint)
	appStatusHandler.SetUInt64Value(common.MetricRoundsPassedInCurrentEpoch, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNoncesPassedInCurrentEpoch, initUint)
	// TODO: add all other rewards parameters
	appStatusHandler.SetStringValue(common.MetricLeaderPercentage, fmt.Sprintf("%f", leaderPercentage))
	appStatusHandler.SetUInt64Value(common.MetricDenomination, uint64(economicsConfig.GlobalSettings.Denomination))
	appStatusHandler.SetUInt64Value(common.MetricNumConnectedPeers, initUint)
	appStatusHandler.SetStringValue(common.MetricNumConnectedPeersClassification, initString)
	appStatusHandler.SetStringValue(common.MetricLatestTagSoftwareVersion, initString)

	appStatusHandler.SetStringValue(common.MetricP2PNumConnectedPeersClassification, initString)
	appStatusHandler.SetStringValue(common.MetricP2PPeerInfo, initString)
	appStatusHandler.SetStringValue(common.MetricP2PIntraShardValidators, initString)
	appStatusHandler.SetStringValue(common.MetricP2PIntraShardObservers, initString)
	appStatusHandler.SetStringValue(common.MetricP2PCrossShardValidators, initString)
	appStatusHandler.SetStringValue(common.MetricP2PCrossShardObservers, initString)
	appStatusHandler.SetStringValue(common.MetricP2PFullHistoryObservers, initString)
	appStatusHandler.SetStringValue(common.MetricP2PUnknownPeers, initString)
	appStatusHandler.SetUInt64Value(common.MetricShardConsensusGroupSize, uint64(nodesConfig.GetShardConsensusGroupSize()))
	appStatusHandler.SetUInt64Value(common.MetricMetaConsensusGroupSize, uint64(nodesConfig.GetMetaConsensusGroupSize()))
	appStatusHandler.SetUInt64Value(common.MetricNumNodesPerShard, uint64(nodesConfig.MinNumberOfShardNodes()))
	appStatusHandler.SetUInt64Value(common.MetricNumMetachainNodes, uint64(nodesConfig.MinNumberOfMetaNodes()))
	appStatusHandler.SetUInt64Value(common.MetricStartTime, uint64(nodesConfig.GetStartTime()))
	appStatusHandler.SetUInt64Value(common.MetricRoundDuration, nodesConfig.GetRoundDuration())
	appStatusHandler.SetUInt64Value(common.MetricMinTransactionVersion, uint64(minTransactionVersion))
	appStatusHandler.SetStringValue(common.MetricTotalSupply, economicsConfig.GlobalSettings.GenesisTotalSupply)
	appStatusHandler.SetStringValue(common.MetricInflation, initZeroString)
	appStatusHandler.SetStringValue(common.MetricDevRewardsInEpoch, initZeroString)
	appStatusHandler.SetStringValue(common.MetricTotalFees, initZeroString)
	appStatusHandler.SetUInt64Value(common.MetricEpochForEconomicsData, initUint)

	enableEpochs := epochConfig.EnableEpochs
	appStatusHandler.SetUInt64Value(common.MetricScDeployEnableEpoch, uint64(enableEpochs.SCDeployEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricBuiltInFunctionsEnableEpoch, uint64(enableEpochs.BuiltInFunctionsEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricRelayedTransactionsEnableEpoch, uint64(enableEpochs.RelayedTransactionsEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricPenalizedTooMuchGasEnableEpoch, uint64(enableEpochs.PenalizedTooMuchGasEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricSwitchJailWaitingEnableEpoch, uint64(enableEpochs.SwitchJailWaitingEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricSwitchHysteresisForMinNodesEnableEpoch, uint64(enableEpochs.SwitchHysteresisForMinNodesEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricBelowSignedThresholdEnableEpoch, uint64(enableEpochs.BelowSignedThresholdEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricTransactionSignedWithTxHashEnableEpoch, uint64(enableEpochs.TransactionSignedWithTxHashEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricMetaProtectionEnableEpoch, uint64(enableEpochs.MetaProtectionEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricAheadOfTimeGasUsageEnableEpoch, uint64(enableEpochs.AheadOfTimeGasUsageEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricGasPriceModifierEnableEpoch, uint64(enableEpochs.GasPriceModifierEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricRepairCallbackEnableEpoch, uint64(enableEpochs.RepairCallbackEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricBlockGasAndFreeRecheckEnableEpoch, uint64(enableEpochs.BlockGasAndFeesReCheckEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricStakingV2EnableEpoch, uint64(enableEpochs.StakingV2EnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricStakeEnableEpoch, uint64(enableEpochs.StakeEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricDoubleKeyProtectionEnableEpoch, uint64(enableEpochs.DoubleKeyProtectionEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricEsdtEnableEpoch, uint64(enableEpochs.ESDTEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricGovernanceEnableEpoch, uint64(enableEpochs.GovernanceEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricDelegationManagerEnableEpoch, uint64(enableEpochs.DelegationManagerEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricDelegationSmartContractEnableEpoch, uint64(enableEpochs.DelegationSmartContractEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricIncrementSCRNonceInMultiTransferEnableEpoch, uint64(enableEpochs.IncrementSCRNonceInMultiTransferEnableEpoch))

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

	appStatusHandler.SetUInt64Value(common.MetricNumValidators, uint64(numValidators))
	appStatusHandler.SetUInt64Value(common.MetricConsensusGroupSize, uint64(consensusGroupSize))

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
