package metrics

import (
	"fmt"
	"runtime/debug"
	"sort"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const millisecondsInSecond = 1000
const initUint = uint64(0)
const initInt = int64(0)
const initString = ""
const initZeroString = "0"

var log = logger.GetOrCreate("node/metrics")

// InitBaseMetrics will initialize base, default metrics to 0 values
func InitBaseMetrics(appStatusHandler core.AppStatusHandler) error {
	if check.IfNil(appStatusHandler) {
		return ErrNilAppStatusHandler
	}

	appStatusHandler.SetUInt64Value(common.MetricSynchronizedRound, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNonce, initUint)
	appStatusHandler.SetUInt64Value(common.MetricCountConsensus, initUint)
	appStatusHandler.SetUInt64Value(common.MetricCountLeader, initUint)
	appStatusHandler.SetUInt64Value(common.MetricCountAcceptedBlocks, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNumTxInBlock, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNumMiniBlocks, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNumProcessedTxs, initUint)
	appStatusHandler.SetUInt64Value(common.MetricCurrentRoundTimestamp, initUint)
	appStatusHandler.SetUInt64Value(common.MetricHeaderSize, initUint)
	appStatusHandler.SetUInt64Value(common.MetricMiniBlocksSize, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNumShardHeadersFromPool, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNumShardHeadersProcessed, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNumTimesInForkChoice, initUint)
	appStatusHandler.SetUInt64Value(common.MetricHighestFinalBlock, initUint)
	appStatusHandler.SetUInt64Value(common.MetricCountConsensusAcceptedBlocks, initUint)
	appStatusHandler.SetUInt64Value(common.MetricRoundsPassedInCurrentEpoch, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNoncesPassedInCurrentEpoch, initUint)
	appStatusHandler.SetUInt64Value(common.MetricNumConnectedPeers, initUint)
	appStatusHandler.SetUInt64Value(common.MetricEpochForEconomicsData, initUint)
	appStatusHandler.SetUInt64Value(common.MetricAccountsSnapshotNumNodes, initUint)
	appStatusHandler.SetUInt64Value(common.MetricTrieSyncNumProcessedNodes, initUint)
	appStatusHandler.SetUInt64Value(common.MetricTrieSyncNumReceivedBytes, initUint)
	appStatusHandler.SetUInt64Value(common.MetricAccountsSnapshotInProgress, initUint)
	appStatusHandler.SetUInt64Value(common.MetricPeersSnapshotInProgress, initUint)

	appStatusHandler.SetInt64Value(common.MetricLastAccountsSnapshotDurationSec, initInt)
	appStatusHandler.SetInt64Value(common.MetricLastPeersSnapshotDurationSec, initInt)

	appStatusHandler.SetStringValue(common.MetricConsensusState, initString)
	appStatusHandler.SetStringValue(common.MetricConsensusRoundState, initString)
	appStatusHandler.SetStringValue(common.MetricCurrentBlockHash, initString)
	appStatusHandler.SetStringValue(common.MetricNumConnectedPeersClassification, initString)
	appStatusHandler.SetStringValue(common.MetricLatestTagSoftwareVersion, initString)
	appStatusHandler.SetStringValue(common.MetricAreVMQueriesReady, strconv.FormatBool(false))
	appStatusHandler.SetStringValue(common.MetricP2PNumConnectedPeersClassification, initString)
	appStatusHandler.SetStringValue(common.MetricP2PPeerInfo, initString)
	appStatusHandler.SetStringValue(common.MetricP2PIntraShardValidators, initString)
	appStatusHandler.SetStringValue(common.MetricP2PIntraShardObservers, initString)
	appStatusHandler.SetStringValue(common.MetricP2PCrossShardValidators, initString)
	appStatusHandler.SetStringValue(common.MetricP2PCrossShardObservers, initString)
	appStatusHandler.SetStringValue(common.MetricP2PFullHistoryObservers, initString)
	appStatusHandler.SetStringValue(common.MetricP2PUnknownPeers, initString)

	appStatusHandler.SetStringValue(common.MetricInflation, initZeroString)
	appStatusHandler.SetStringValue(common.MetricDevRewardsInEpoch, initZeroString)
	appStatusHandler.SetStringValue(common.MetricTotalFees, initZeroString)

	return nil
}

// InitConfigMetrics will init the "enable epochs" configuration metrics from epoch config
func InitConfigMetrics(
	appStatusHandler core.AppStatusHandler,
	epochConfig config.EpochConfig,
	economicsConfig config.EconomicsConfig,
	genesisNodesConfig sharding.GenesisNodesSetupHandler,
) error {
	if check.IfNil(appStatusHandler) {
		return ErrNilAppStatusHandler
	}

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
	appStatusHandler.SetUInt64Value(common.MetricCorrectLastUnjailedEnableEpoch, uint64(enableEpochs.CorrectLastUnjailedEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricBalanceWaitingListsEnableEpoch, uint64(enableEpochs.BalanceWaitingListsEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricReturnDataToLastTransferEnableEpoch, uint64(enableEpochs.ReturnDataToLastTransferEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricSenderInOutTransferEnableEpoch, uint64(enableEpochs.SenderInOutTransferEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricRelayedTransactionsV2EnableEpoch, uint64(enableEpochs.RelayedTransactionsV2EnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricUnbondTokensV2EnableEpoch, uint64(enableEpochs.UnbondTokensV2EnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricSaveJailedAlwaysEnableEpoch, uint64(enableEpochs.SaveJailedAlwaysEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricValidatorToDelegationEnableEpoch, uint64(enableEpochs.ValidatorToDelegationEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricReDelegateBelowMinCheckEnableEpoch, uint64(enableEpochs.ReDelegateBelowMinCheckEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricIncrementSCRNonceInMultiTransferEnableEpoch, uint64(enableEpochs.IncrementSCRNonceInMultiTransferEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricESDTMultiTransferEnableEpoch, uint64(enableEpochs.ESDTMultiTransferEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricGlobalMintBurnDisableEpoch, uint64(enableEpochs.GlobalMintBurnDisableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricESDTTransferRoleEnableEpoch, uint64(enableEpochs.ESDTTransferRoleEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricBuiltInFunctionOnMetaEnableEpoch, uint64(enableEpochs.BuiltInFunctionOnMetaEnableEpoch))
	appStatusHandler.SetStringValue(common.MetricTotalSupply, economicsConfig.GlobalSettings.GenesisTotalSupply)
	appStatusHandler.SetUInt64Value(common.MetricWaitingListFixEnableEpoch, uint64(enableEpochs.WaitingListFixEnableEpoch))
	appStatusHandler.SetUInt64Value(common.MetricSetGuardianEnableEpoch, uint64(enableEpochs.SetGuardianEnableEpoch))

	for i, nodesChangeConfig := range enableEpochs.MaxNodesChangeEnableEpoch {
		epochEnable := fmt.Sprintf("%s%d%s", common.MetricMaxNodesChangeEnableEpoch, i, common.EpochEnableSuffix)
		appStatusHandler.SetUInt64Value(epochEnable, uint64(nodesChangeConfig.EpochEnable))

		maxNumNodes := fmt.Sprintf("%s%d%s", common.MetricMaxNodesChangeEnableEpoch, i, common.MaxNumNodesSuffix)
		appStatusHandler.SetUInt64Value(maxNumNodes, uint64(nodesChangeConfig.MaxNumNodes))

		nodesToShufflePerShard := fmt.Sprintf("%s%d%s", common.MetricMaxNodesChangeEnableEpoch, i, common.NodesToShufflePerShardSuffix)
		appStatusHandler.SetUInt64Value(nodesToShufflePerShard, uint64(nodesChangeConfig.NodesToShufflePerShard))
	}
	appStatusHandler.SetUInt64Value(common.MetricMaxNodesChangeEnableEpoch+"_count", uint64(len(enableEpochs.MaxNodesChangeEnableEpoch)))

	appStatusHandler.SetStringValue(common.MetricHysteresis, fmt.Sprintf("%f", genesisNodesConfig.GetHysteresis()))
	appStatusHandler.SetStringValue(common.MetricAdaptivity, fmt.Sprintf("%t", genesisNodesConfig.GetAdaptivity()))

	return nil
}

// InitRatingsMetrics will init the ratings configuration metrics
func InitRatingsMetrics(appStatusHandler core.AppStatusHandler, ratingsConfig config.RatingsConfig) error {
	if check.IfNil(appStatusHandler) {
		return ErrNilAppStatusHandler
	}

	appStatusHandler.SetUInt64Value(common.MetricRatingsGeneralStartRating, uint64(ratingsConfig.General.StartRating))
	appStatusHandler.SetUInt64Value(common.MetricRatingsGeneralMaxRating, uint64(ratingsConfig.General.MaxRating))
	appStatusHandler.SetUInt64Value(common.MetricRatingsGeneralMinRating, uint64(ratingsConfig.General.MinRating))
	appStatusHandler.SetStringValue(common.MetricRatingsGeneralSignedBlocksThreshold, fmt.Sprintf("%f", ratingsConfig.General.SignedBlocksThreshold))
	for i, selectionChance := range ratingsConfig.General.SelectionChances {
		maxThresholdStr := fmt.Sprintf("%s%d%s", common.MetricRatingsGeneralSelectionChances, i, common.SelectionChancesMaxThresholdSuffix)
		appStatusHandler.SetUInt64Value(maxThresholdStr, uint64(selectionChance.MaxThreshold))
		chancePercentStr := fmt.Sprintf("%s%d%s", common.MetricRatingsGeneralSelectionChances, i, common.SelectionChancesChancePercentSuffix)
		appStatusHandler.SetUInt64Value(chancePercentStr, uint64(selectionChance.ChancePercent))
	}
	appStatusHandler.SetUInt64Value(common.MetricRatingsGeneralSelectionChances+"_count", uint64(len(ratingsConfig.General.SelectionChances)))

	appStatusHandler.SetUInt64Value(common.MetricRatingsShardChainHoursToMaxRatingFromStartRating, uint64(ratingsConfig.ShardChain.HoursToMaxRatingFromStartRating))
	appStatusHandler.SetStringValue(common.MetricRatingsShardChainProposerValidatorImportance, fmt.Sprintf("%f", ratingsConfig.ShardChain.ProposerValidatorImportance))
	appStatusHandler.SetStringValue(common.MetricRatingsShardChainProposerDecreaseFactor, fmt.Sprintf("%f", ratingsConfig.ShardChain.ProposerDecreaseFactor))
	appStatusHandler.SetStringValue(common.MetricRatingsShardChainValidatorDecreaseFactor, fmt.Sprintf("%f", ratingsConfig.ShardChain.ValidatorDecreaseFactor))
	appStatusHandler.SetStringValue(common.MetricRatingsShardChainConsecutiveMissedBlocksPenalty, fmt.Sprintf("%f", ratingsConfig.ShardChain.ConsecutiveMissedBlocksPenalty))

	appStatusHandler.SetUInt64Value(common.MetricRatingsMetaChainHoursToMaxRatingFromStartRating, uint64(ratingsConfig.MetaChain.HoursToMaxRatingFromStartRating))
	appStatusHandler.SetStringValue(common.MetricRatingsMetaChainProposerValidatorImportance, fmt.Sprintf("%f", ratingsConfig.MetaChain.ProposerValidatorImportance))
	appStatusHandler.SetStringValue(common.MetricRatingsMetaChainProposerDecreaseFactor, fmt.Sprintf("%f", ratingsConfig.MetaChain.ProposerDecreaseFactor))
	appStatusHandler.SetStringValue(common.MetricRatingsMetaChainValidatorDecreaseFactor, fmt.Sprintf("%f", ratingsConfig.MetaChain.ValidatorDecreaseFactor))
	appStatusHandler.SetStringValue(common.MetricRatingsMetaChainConsecutiveMissedBlocksPenalty, fmt.Sprintf("%f", ratingsConfig.MetaChain.ConsecutiveMissedBlocksPenalty))

	appStatusHandler.SetStringValue(common.MetricRatingsPeerHonestyDecayCoefficient, fmt.Sprintf("%f", ratingsConfig.PeerHonesty.DecayCoefficient))
	appStatusHandler.SetUInt64Value(common.MetricRatingsPeerHonestyDecayUpdateIntervalInSeconds, uint64(ratingsConfig.PeerHonesty.DecayUpdateIntervalInSeconds))
	appStatusHandler.SetStringValue(common.MetricRatingsPeerHonestyMaxScore, fmt.Sprintf("%f", ratingsConfig.PeerHonesty.MaxScore))
	appStatusHandler.SetStringValue(common.MetricRatingsPeerHonestyMinScore, fmt.Sprintf("%f", ratingsConfig.PeerHonesty.MinScore))
	appStatusHandler.SetStringValue(common.MetricRatingsPeerHonestyBadPeerThreshold, fmt.Sprintf("%f", ratingsConfig.PeerHonesty.BadPeerThreshold))
	appStatusHandler.SetStringValue(common.MetricRatingsPeerHonestyUnitValue, fmt.Sprintf("%f", ratingsConfig.PeerHonesty.UnitValue))

	return nil
}

// InitMetrics will init metrics for status handler
func InitMetrics(
	appStatusHandler core.AppStatusHandler,
	pubkeyStr string,
	nodeType core.NodeType,
	shardCoordinator sharding.Coordinator,
	nodesConfig sharding.GenesisNodesSetupHandler,
	version string,
	economicsConfig *config.EconomicsConfig,
	roundsPerEpoch int64,
	minTransactionVersion uint32,
) error {
	if check.IfNil(appStatusHandler) {
		return ErrNilAppStatusHandler
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

	leaderPercentage := float64(0)
	rewardsConfigs := make([]config.EpochRewardSettings, len(economicsConfig.RewardsSettings.RewardsConfigByEpoch))
	_ = copy(rewardsConfigs, economicsConfig.RewardsSettings.RewardsConfigByEpoch)

	sort.Slice(rewardsConfigs, func(i, j int) bool {
		return rewardsConfigs[i].EpochEnable < rewardsConfigs[j].EpochEnable
	})

	if len(rewardsConfigs) > 0 {
		leaderPercentage = rewardsConfigs[0].LeaderPercentage
	}

	appStatusHandler.SetStringValue(common.MetricPublicKeyBlockSign, pubkeyStr)
	appStatusHandler.SetUInt64Value(common.MetricShardId, shardId)
	appStatusHandler.SetUInt64Value(common.MetricNumShardsWithoutMetachain, numOfShards)
	appStatusHandler.SetStringValue(common.MetricNodeType, string(nodeType))
	appStatusHandler.SetUInt64Value(common.MetricRoundTime, roundDuration/millisecondsInSecond)
	appStatusHandler.SetStringValue(common.MetricAppVersion, version)
	appStatusHandler.SetUInt64Value(common.MetricRoundsPerEpoch, uint64(roundsPerEpoch))
	appStatusHandler.SetStringValue(common.MetricCrossCheckBlockHeight, "0")
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		key := fmt.Sprintf("%s_%d", common.MetricCrossCheckBlockHeight, i)
		appStatusHandler.SetUInt64Value(key, 0)
	}
	appStatusHandler.SetUInt64Value(common.MetricCrossCheckBlockHeightMeta, 0)
	appStatusHandler.SetUInt64Value(common.MetricIsSyncing, isSyncing)
	appStatusHandler.SetStringValue(common.MetricLeaderPercentage, fmt.Sprintf("%f", leaderPercentage))
	appStatusHandler.SetUInt64Value(common.MetricDenomination, uint64(economicsConfig.GlobalSettings.Denomination))

	appStatusHandler.SetUInt64Value(common.MetricShardConsensusGroupSize, uint64(nodesConfig.GetShardConsensusGroupSize()))
	appStatusHandler.SetUInt64Value(common.MetricMetaConsensusGroupSize, uint64(nodesConfig.GetMetaConsensusGroupSize()))
	appStatusHandler.SetUInt64Value(common.MetricNumNodesPerShard, uint64(nodesConfig.MinNumberOfShardNodes()))
	appStatusHandler.SetUInt64Value(common.MetricNumMetachainNodes, uint64(nodesConfig.MinNumberOfMetaNodes()))
	appStatusHandler.SetUInt64Value(common.MetricStartTime, uint64(nodesConfig.GetStartTime()))
	appStatusHandler.SetUInt64Value(common.MetricRoundDuration, nodesConfig.GetRoundDuration())
	appStatusHandler.SetUInt64Value(common.MetricMinTransactionVersion, uint64(minTransactionVersion))

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

	return nil
}

// SaveUint64Metric will save an uint64 metric in status handler
func SaveUint64Metric(ash core.AppStatusHandler, key string, value uint64) {
	if check.IfNil(ash) {
		log.Error("programming error: nil AppStatusHandler in SaveUint64Metric", "stack", string(debug.Stack()))
		return
	}

	ash.SetUInt64Value(key, value)
}

// SaveStringMetric will save a string metric in status handler
func SaveStringMetric(ash core.AppStatusHandler, key, value string) {
	if check.IfNil(ash) {
		log.Error("programming error: nil AppStatusHandler in SaveStringMetric", "stack", string(debug.Stack()))
		return
	}

	ash.SetStringValue(key, value)
}
