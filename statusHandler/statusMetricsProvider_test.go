package statusHandler_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusMetricsProvider(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	assert.NotNil(t, sm)
	assert.False(t, sm.IsInterfaceNil())
}

func TestStatusMetricsProvider_IncrementCallNonExistingKey(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key1 := "test-key1"
	sm.Increment(key1)

	retMap := sm.StatusMetricsMap()

	assert.Nil(t, retMap[key1])
}

func TestStatusMetricsProvider_IncrementNonUint64ValueShouldNotWork(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key1 := "test-key2"
	value1 := "value2"
	// set a key which is initialized with a string and the Increment method won't affect the key
	sm.SetStringValue(key1, value1)
	sm.Increment(key1)

	retMap := sm.StatusMetricsMap()
	assert.Equal(t, value1, retMap[key1])
}

func TestStatusMetricsProvider_IncrementShouldWork(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key1 := "test-key3"
	sm.SetUInt64Value(key1, 0)
	sm.Increment(key1)

	retMap := sm.StatusMetricsMap()

	assert.Equal(t, uint64(1), retMap[key1])
}

func TestStatusMetricsProvider_Decrement(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key := "test-key4"
	sm.SetUInt64Value(key, 2)
	sm.Decrement(key)

	retMap := sm.StatusMetricsMap()

	assert.Equal(t, uint64(1), retMap[key])
}

func TestStatusMetricsProvider_SetInt64Value(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key := "test-key5"
	value := int64(5)
	sm.SetInt64Value(key, value)
	sm.Decrement(key)

	retMap := sm.StatusMetricsMap()

	assert.Equal(t, value, retMap[key])
}

func TestStatusMetricsProvider_SetStringValue(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key := "test-key6"
	value := "value"
	sm.SetStringValue(key, value)
	sm.Decrement(key)

	retMap := sm.StatusMetricsMap()

	assert.Equal(t, value, retMap[key])
}

func TestStatusMetricsProvider_AddUint64Value(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key := "test-key6"
	value := uint64(100)
	sm.SetUInt64Value(key, value)
	sm.AddUint64(key, value)

	retMap := sm.StatusMetricsMap()
	assert.Equal(t, value+value, retMap[key])
}

func TestStatusMetrics_StatusMetricsWithoutP2PPrometheusStringShouldPutDefaultShardIDLabel(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key1, value1 := "test-key7", uint64(100)
	key2, value2 := "test-key8", "value8"
	sm.SetUInt64Value(key1, value1)
	sm.SetStringValue(key2, value2)

	strRes, _ := sm.StatusMetricsWithoutP2PPrometheusString()

	expectedMetricOutput := fmt.Sprintf("%s{%s=\"%d\"} %v", key1, common.MetricShardId, 0, value1)
	assert.True(t, strings.Contains(strRes, expectedMetricOutput))
}

func TestStatusMetrics_StatusMetricsWithoutP2PPrometheusStringShouldPutCorrectShardIDLabel(t *testing.T) {
	t.Parallel()

	shardID := uint32(37)
	sm := statusHandler.NewStatusMetrics()
	key1, value1 := "test-key7", uint64(100)
	key2, value2 := "test-key8", "value8"
	key3, value3 := common.MetricShardId, shardID
	sm.SetUInt64Value(key1, value1)
	sm.SetStringValue(key2, value2)
	sm.SetUInt64Value(key3, uint64(value3))

	strRes, _ := sm.StatusMetricsWithoutP2PPrometheusString()

	expectedMetricOutput := fmt.Sprintf("%s{%s=\"%d\"} %v", key1, common.MetricShardId, shardID, value1)
	assert.True(t, strings.Contains(strRes, expectedMetricOutput))
}

func TestStatusMetrics_StatusMetricsWithoutP2PPrometheusStringShouldComputeRoundsAndNoncesPassedInEpoch(t *testing.T) {
	t.Parallel()

	shardID := uint32(2)
	sm := statusHandler.NewStatusMetrics()
	sm.SetUInt64Value(common.MetricRoundsPassedInCurrentEpoch, 0)
	sm.SetUInt64Value(common.MetricNoncesPassedInCurrentEpoch, 0)
	sm.SetUInt64Value(common.MetricShardId, uint64(shardID))
	sm.SetUInt64Value(common.MetricRoundAtEpochStart, 100)
	sm.SetUInt64Value(common.MetricCurrentRound, 137)
	sm.SetUInt64Value(common.MetricNonceAtEpochStart, 100)
	sm.SetUInt64Value(common.MetricNonce, 138)

	strRes, _ := sm.StatusMetricsWithoutP2PPrometheusString()

	assert.Contains(t, strRes, `erd_rounds_passed_in_current_epoch{erd_shard_id="2"} 37`)
	assert.Contains(t, strRes, `erd_nonces_passed_in_current_epoch{erd_shard_id="2"} 38`)
}

func TestStatusMetrics_NetworkConfig(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()

	sm.SetUInt64Value(common.MetricNumShardsWithoutMetachain, 1)
	sm.SetUInt64Value(common.MetricNumNodesPerShard, 100)
	sm.SetUInt64Value(common.MetricNumMetachainNodes, 50)
	sm.SetUInt64Value(common.MetricShardConsensusGroupSize, 20)
	sm.SetUInt64Value(common.MetricMetaConsensusGroupSize, 25)
	sm.SetUInt64Value(common.MetricMinGasPrice, 1000)
	sm.SetUInt64Value(common.MetricMinGasLimit, 50000)
	sm.SetUInt64Value(common.MetricExtraGasLimitGuardedTx, 50000)
	sm.SetUInt64Value(common.MetricExtraGasLimitRelayedTx, 50000)
	sm.SetStringValue(common.MetricRewardsTopUpGradientPoint, "12345")
	sm.SetUInt64Value(common.MetricGasPerDataByte, 1500)
	sm.SetStringValue(common.MetricChainId, "local-id")
	sm.SetUInt64Value(common.MetricMaxGasPerTransaction, 15000)
	sm.SetUInt64Value(common.MetricRoundDuration, 5000)
	sm.SetUInt64Value(common.MetricStartTime, 9999)
	sm.SetStringValue(common.MetricLatestTagSoftwareVersion, "version1.0")
	sm.SetUInt64Value(common.MetricDenomination, 18)
	sm.SetUInt64Value(common.MetricMinTransactionVersion, 2)
	sm.SetStringValue(common.MetricTopUpFactor, fmt.Sprintf("%g", 12.134))
	sm.SetStringValue(common.MetricGasPriceModifier, fmt.Sprintf("%g", 0.5))
	sm.SetUInt64Value(common.MetricRoundsPerEpoch, uint64(144))
	sm.SetStringValue(common.MetricAdaptivity, fmt.Sprintf("%t", true))
	sm.SetStringValue(common.MetricHysteresis, fmt.Sprintf("%f", 0.0))

	expectedConfig := map[string]interface{}{
		"erd_chain_id":                      "local-id",
		"erd_denomination":                  uint64(18),
		"erd_gas_per_data_byte":             uint64(1500),
		"erd_latest_tag_software_version":   "version1.0",
		"erd_meta_consensus_group_size":     uint64(25),
		"erd_min_gas_limit":                 uint64(50000),
		"erd_extra_gas_limit_guarded_tx":    uint64(50000),
		"erd_extra_gas_limit_relayed_tx":    uint64(50000),
		"erd_min_gas_price":                 uint64(1000),
		"erd_min_transaction_version":       uint64(2),
		"erd_num_metachain_nodes":           uint64(50),
		"erd_num_nodes_in_shard":            uint64(100),
		"erd_num_shards_without_meta":       uint64(1),
		"erd_rewards_top_up_gradient_point": "12345",
		"erd_round_duration":                uint64(5000),
		"erd_shard_consensus_group_size":    uint64(20),
		"erd_start_time":                    uint64(9999),
		"erd_top_up_factor":                 "12.134",
		"erd_gas_price_modifier":            "0.5",
		"erd_rounds_per_epoch":              uint64(144),
		"erd_max_gas_per_transaction":       uint64(15000),
		"erd_adaptivity":                    "true",
		"erd_hysteresis":                    "0.000000",
	}

	configMetrics, _ := sm.ConfigMetrics()
	assert.Equal(t, expectedConfig, configMetrics)
}

func TestStatusMetrics_NetworkMetrics(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()

	sm.SetUInt64Value(common.MetricCurrentRound, 200)
	sm.SetUInt64Value(common.MetricRoundAtEpochStart, 100)
	sm.SetUInt64Value(common.MetricNonce, 180)
	sm.SetUInt64Value(common.MetricBlockTimestamp, 18000)
	sm.SetUInt64Value(common.MetricBlockTimestampMs, 18000000)
	sm.SetUInt64Value(common.MetricHighestFinalBlock, 181)
	sm.SetUInt64Value(common.MetricNonceAtEpochStart, 95)
	sm.SetUInt64Value(common.MetricEpochNumber, 1)
	sm.SetUInt64Value(common.MetricRoundsPerEpoch, 50)

	expectedConfig := map[string]interface{}{
		"erd_current_round":                  uint64(200),
		"erd_round_at_epoch_start":           uint64(100),
		"erd_nonce":                          uint64(180),
		"erd_block_timestamp":                uint64(18000),
		"erd_block_timestamp_ms":             uint64(18000000),
		"erd_highest_final_nonce":            uint64(181),
		"erd_nonce_at_epoch_start":           uint64(95),
		"erd_epoch_number":                   uint64(1),
		"erd_rounds_per_epoch":               uint64(50),
		"erd_rounds_passed_in_current_epoch": uint64(100),
		"erd_nonces_passed_in_current_epoch": uint64(85),
	}

	t.Run("no cross check value", func(t *testing.T) {
		configMetrics, _ := sm.NetworkMetrics()
		assert.Equal(t, expectedConfig, configMetrics)
	})
	t.Run("with cross check value", func(t *testing.T) {
		crossCheckValue := "0: 9169897, 1: 9166353, 2: 9170524, "
		sm.SetStringValue(common.MetricCrossCheckBlockHeight, crossCheckValue)

		configMetrics, _ := sm.NetworkMetrics()
		expectedConfig[common.MetricCrossCheckBlockHeight] = crossCheckValue
		assert.Equal(t, expectedConfig, configMetrics)
	})
}

func TestStatusMetrics_StatusMetricsMapWithoutP2P(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()

	sm.SetUInt64Value(common.MetricCurrentRound, 100)
	sm.SetUInt64Value(common.MetricRoundAtEpochStart, 200)
	sm.SetUInt64Value(common.MetricNonce, 300)
	sm.SetUInt64Value(common.MetricBlockTimestamp, 30000)
	sm.SetUInt64Value(common.MetricBlockTimestampMs, 30000000)
	sm.SetStringValue(common.MetricAppVersion, "400")
	sm.SetUInt64Value(common.MetricRoundsPassedInCurrentEpoch, 95)
	sm.SetUInt64Value(common.MetricNoncesPassedInCurrentEpoch, 1)
	sm.SetUInt64Value(common.MetricTrieSyncNumReceivedBytes, 100)
	sm.SetUInt64Value(common.MetricTrieSyncNumProcessedNodes, 101)

	res, _ := sm.StatusMetricsMapWithoutP2P()

	require.Equal(t, uint64(100), res[common.MetricCurrentRound])
	require.Equal(t, uint64(200), res[common.MetricRoundAtEpochStart])
	require.Equal(t, uint64(300), res[common.MetricNonce])
	require.Equal(t, uint64(30000), res[common.MetricBlockTimestamp])
	require.Equal(t, "400", res[common.MetricAppVersion])
	require.NotContains(t, res, common.MetricRoundsPassedInCurrentEpoch)
	require.NotContains(t, res, common.MetricNoncesPassedInCurrentEpoch)
	require.NotContains(t, res, common.MetricTrieSyncNumReceivedBytes)
	require.NotContains(t, res, common.MetricTrieSyncNumProcessedNodes)
}

func TestStatusMetrics_EnableEpochMetrics(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()

	sm.SetUInt64Value(common.MetricScDeployEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricBuiltInFunctionsEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricRelayedTransactionsEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricPenalizedTooMuchGasEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricSwitchJailWaitingEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricSwitchHysteresisForMinNodesEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricBelowSignedThresholdEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricTransactionSignedWithTxHashEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricMetaProtectionEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricAheadOfTimeGasUsageEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricGasPriceModifierEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricRepairCallbackEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricBlockGasAndFreeRecheckEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricStakingV2EnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricStakeEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricDoubleKeyProtectionEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricEsdtEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricGovernanceEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricDelegationManagerEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricDelegationSmartContractEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricCorrectLastUnjailedEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricBalanceWaitingListsEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricReturnDataToLastTransferEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricSenderInOutTransferEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricRelayedTransactionsV2EnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricUnbondTokensV2EnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricSaveJailedAlwaysEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricValidatorToDelegationEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricReDelegateBelowMinCheckEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricIncrementSCRNonceInMultiTransferEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricScheduledMiniBlocksEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricESDTMultiTransferEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricGlobalMintBurnDisableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricESDTTransferRoleEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricComputeRewardCheckpointEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricSCRSizeInvariantCheckEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricBackwardCompSaveKeyValueEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricESDTNFTCreateOnMultiShardEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricMetaESDTSetEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricAddTokensToDelegationEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricMultiESDTTransferFixOnCallBackOnEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricOptimizeGasUsedInCrossMiniBlocksEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricCorrectFirstQueuedEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricCorrectJailedNotUnstakedEmptyQueueEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricFixOOGReturnCodeEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricRemoveNonUpdatedStorageEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricDeleteDelegatorAfterClaimRewardsEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricOptimizeNFTStoreEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricCreateNFTThroughExecByCallerEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricStopDecreasingValidatorRatingWhenStuckEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricFrontRunningProtectionEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricIsPayableBySCEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricStorageAPICostOptimizationEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricTransformToMultiShardCreateEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricESDTRegisterAndSetAllRolesEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricDoNotReturnOldBlockInBlockchainHookEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricAddFailedRelayedTxToInvalidMBsDisableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricSCRSizeInvariantOnBuiltInResultEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricCheckCorrectTokenIDForTransferRoleEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricDisableExecByCallerEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricFailExecutionOnEveryAPIErrorEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricManagedCryptoAPIsEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricRefactorContextEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricCheckFunctionArgumentEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricCheckExecuteOnReadOnlyEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricMiniBlockPartialExecutionEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricESDTMetadataContinuousCleanupEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricFixAsyncCallBackArgsListEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricFixOldTokenLiquidityEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricRuntimeMemStoreLimitEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricRuntimeCodeSizeFixEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricSetSenderInEeiOutputTransferEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricRefactorPeersMiniBlocksEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricSCProcessorV2EnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricMaxBlockchainHookCountersEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricWipeSingleNFTLiquidityDecreaseEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricAlwaysSaveTokenMetaDataEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricCleanUpInformativeSCRsEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricSetGuardianEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricSetScToScLogEventEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricRelayedNonceFixEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricDeterministicSortOnValidatorsInfoEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricKeepExecOrderOnCreatedSCRsEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricMultiClaimOnDelegationEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricChangeUsernameEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricAutoBalanceDataTriesEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricMigrateDataTrieEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricConsistentTokensValuesLengthCheckEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricFixDelegationChangeOwnerOnAccountEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricDynamicGasCostForDataTrieStorageLoadEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricNFTStopCreateEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricChangeOwnerAddressCrossShardThroughSCEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricFixGasRemainingForSaveKeyValueBuiltinFunctionEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricCurrentRandomnessOnSortingEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricStakeLimitsEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricStakingV4Step1EnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricStakingV4Step2EnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricStakingV4Step3EnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricCleanupAuctionOnLowWaitingListEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricAlwaysMergeContextsInEEIEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricDynamicESDTEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricEGLDInMultiTransferEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricCryptoOpcodesV2EnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricMultiESDTNFTTransferAndExecuteByUserEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricFixRelayedMoveBalanceToNonPayableSCEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricRelayedTransactionsV3EnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricRelayedTransactionsV3FixESDTTransferEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricCheckBuiltInCallOnTransferValueAndFailEnableRound, uint64(4))
	sm.SetUInt64Value(common.MetricMaskVMInternalDependenciesErrorsEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricFixBackTransferOPCODEEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricValidationOnGobDecodeEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricBarnardOpcodesEnableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricAutomaticActivationOfNodesDisableEpoch, uint64(4))
	sm.SetUInt64Value(common.MetricFixGetBalanceEnableEpoch, uint64(4))

	maxNodesChangeConfig := []map[string]uint64{
		{
			"EpochEnable":            0,
			"MaxNumNodes":            1,
			"NodesToShufflePerShard": 2,
		},
		{
			"EpochEnable":            3,
			"MaxNumNodes":            4,
			"NodesToShufflePerShard": 5,
		},
	}
	for i, nodesChangeConfig := range maxNodesChangeConfig {
		epochEnable := fmt.Sprintf("%s%d%s", common.MetricMaxNodesChangeEnableEpoch, i, common.EpochEnableSuffix)
		sm.SetUInt64Value(epochEnable, nodesChangeConfig["EpochEnable"])

		maxNumNodes := fmt.Sprintf("%s%d%s", common.MetricMaxNodesChangeEnableEpoch, i, common.MaxNumNodesSuffix)
		sm.SetUInt64Value(maxNumNodes, nodesChangeConfig["MaxNumNodes"])

		nodesToShufflePerShard := fmt.Sprintf("%s%d%s", common.MetricMaxNodesChangeEnableEpoch, i, common.NodesToShufflePerShardSuffix)
		sm.SetUInt64Value(nodesToShufflePerShard, nodesChangeConfig["NodesToShufflePerShard"])
	}
	sm.SetUInt64Value(common.MetricMaxNodesChangeEnableEpoch+"_count", uint64(len(maxNodesChangeConfig)))

	expectedMetrics := map[string]interface{}{
		common.MetricScDeployEnableEpoch:                                      uint64(4),
		common.MetricBuiltInFunctionsEnableEpoch:                              uint64(4),
		common.MetricRelayedTransactionsEnableEpoch:                           uint64(4),
		common.MetricPenalizedTooMuchGasEnableEpoch:                           uint64(4),
		common.MetricSwitchJailWaitingEnableEpoch:                             uint64(4),
		common.MetricSwitchHysteresisForMinNodesEnableEpoch:                   uint64(4),
		common.MetricBelowSignedThresholdEnableEpoch:                          uint64(4),
		common.MetricTransactionSignedWithTxHashEnableEpoch:                   uint64(4),
		common.MetricMetaProtectionEnableEpoch:                                uint64(4),
		common.MetricAheadOfTimeGasUsageEnableEpoch:                           uint64(4),
		common.MetricGasPriceModifierEnableEpoch:                              uint64(4),
		common.MetricRepairCallbackEnableEpoch:                                uint64(4),
		common.MetricBlockGasAndFreeRecheckEnableEpoch:                        uint64(4),
		common.MetricStakingV2EnableEpoch:                                     uint64(4),
		common.MetricStakeEnableEpoch:                                         uint64(4),
		common.MetricDoubleKeyProtectionEnableEpoch:                           uint64(4),
		common.MetricEsdtEnableEpoch:                                          uint64(4),
		common.MetricGovernanceEnableEpoch:                                    uint64(4),
		common.MetricDelegationManagerEnableEpoch:                             uint64(4),
		common.MetricDelegationSmartContractEnableEpoch:                       uint64(4),
		common.MetricCorrectLastUnjailedEnableEpoch:                           uint64(4),
		common.MetricBalanceWaitingListsEnableEpoch:                           uint64(4),
		common.MetricReturnDataToLastTransferEnableEpoch:                      uint64(4),
		common.MetricSenderInOutTransferEnableEpoch:                           uint64(4),
		common.MetricRelayedTransactionsV2EnableEpoch:                         uint64(4),
		common.MetricUnbondTokensV2EnableEpoch:                                uint64(4),
		common.MetricSaveJailedAlwaysEnableEpoch:                              uint64(4),
		common.MetricValidatorToDelegationEnableEpoch:                         uint64(4),
		common.MetricReDelegateBelowMinCheckEnableEpoch:                       uint64(4),
		common.MetricIncrementSCRNonceInMultiTransferEnableEpoch:              uint64(4),
		common.MetricScheduledMiniBlocksEnableEpoch:                           uint64(4),
		common.MetricESDTMultiTransferEnableEpoch:                             uint64(4),
		common.MetricGlobalMintBurnDisableEpoch:                               uint64(4),
		common.MetricESDTTransferRoleEnableEpoch:                              uint64(4),
		common.MetricComputeRewardCheckpointEnableEpoch:                       uint64(4),
		common.MetricSCRSizeInvariantCheckEnableEpoch:                         uint64(4),
		common.MetricBackwardCompSaveKeyValueEnableEpoch:                      uint64(4),
		common.MetricESDTNFTCreateOnMultiShardEnableEpoch:                     uint64(4),
		common.MetricMetaESDTSetEnableEpoch:                                   uint64(4),
		common.MetricAddTokensToDelegationEnableEpoch:                         uint64(4),
		common.MetricMultiESDTTransferFixOnCallBackOnEnableEpoch:              uint64(4),
		common.MetricOptimizeGasUsedInCrossMiniBlocksEnableEpoch:              uint64(4),
		common.MetricCorrectFirstQueuedEpoch:                                  uint64(4),
		common.MetricCorrectJailedNotUnstakedEmptyQueueEpoch:                  uint64(4),
		common.MetricFixOOGReturnCodeEnableEpoch:                              uint64(4),
		common.MetricRemoveNonUpdatedStorageEnableEpoch:                       uint64(4),
		common.MetricDeleteDelegatorAfterClaimRewardsEnableEpoch:              uint64(4),
		common.MetricOptimizeNFTStoreEnableEpoch:                              uint64(4),
		common.MetricCreateNFTThroughExecByCallerEnableEpoch:                  uint64(4),
		common.MetricStopDecreasingValidatorRatingWhenStuckEnableEpoch:        uint64(4),
		common.MetricFrontRunningProtectionEnableEpoch:                        uint64(4),
		common.MetricIsPayableBySCEnableEpoch:                                 uint64(4),
		common.MetricStorageAPICostOptimizationEnableEpoch:                    uint64(4),
		common.MetricTransformToMultiShardCreateEnableEpoch:                   uint64(4),
		common.MetricESDTRegisterAndSetAllRolesEnableEpoch:                    uint64(4),
		common.MetricDoNotReturnOldBlockInBlockchainHookEnableEpoch:           uint64(4),
		common.MetricAddFailedRelayedTxToInvalidMBsDisableEpoch:               uint64(4),
		common.MetricSCRSizeInvariantOnBuiltInResultEnableEpoch:               uint64(4),
		common.MetricCheckCorrectTokenIDForTransferRoleEnableEpoch:            uint64(4),
		common.MetricDisableExecByCallerEnableEpoch:                           uint64(4),
		common.MetricFailExecutionOnEveryAPIErrorEnableEpoch:                  uint64(4),
		common.MetricManagedCryptoAPIsEnableEpoch:                             uint64(4),
		common.MetricRefactorContextEnableEpoch:                               uint64(4),
		common.MetricCheckFunctionArgumentEnableEpoch:                         uint64(4),
		common.MetricCheckExecuteOnReadOnlyEnableEpoch:                        uint64(4),
		common.MetricMiniBlockPartialExecutionEnableEpoch:                     uint64(4),
		common.MetricESDTMetadataContinuousCleanupEnableEpoch:                 uint64(4),
		common.MetricFixAsyncCallBackArgsListEnableEpoch:                      uint64(4),
		common.MetricFixOldTokenLiquidityEnableEpoch:                          uint64(4),
		common.MetricRuntimeMemStoreLimitEnableEpoch:                          uint64(4),
		common.MetricRuntimeCodeSizeFixEnableEpoch:                            uint64(4),
		common.MetricSetSenderInEeiOutputTransferEnableEpoch:                  uint64(4),
		common.MetricRefactorPeersMiniBlocksEnableEpoch:                       uint64(4),
		common.MetricSCProcessorV2EnableEpoch:                                 uint64(4),
		common.MetricMaxBlockchainHookCountersEnableEpoch:                     uint64(4),
		common.MetricWipeSingleNFTLiquidityDecreaseEnableEpoch:                uint64(4),
		common.MetricAlwaysSaveTokenMetaDataEnableEpoch:                       uint64(4),
		common.MetricCleanUpInformativeSCRsEnableEpoch:                        uint64(4),
		common.MetricSetGuardianEnableEpoch:                                   uint64(4),
		common.MetricSetScToScLogEventEnableEpoch:                             uint64(4),
		common.MetricRelayedNonceFixEnableEpoch:                               uint64(4),
		common.MetricDeterministicSortOnValidatorsInfoEnableEpoch:             uint64(4),
		common.MetricKeepExecOrderOnCreatedSCRsEnableEpoch:                    uint64(4),
		common.MetricMultiClaimOnDelegationEnableEpoch:                        uint64(4),
		common.MetricChangeUsernameEnableEpoch:                                uint64(4),
		common.MetricAutoBalanceDataTriesEnableEpoch:                          uint64(4),
		common.MetricMigrateDataTrieEnableEpoch:                               uint64(4),
		common.MetricConsistentTokensValuesLengthCheckEnableEpoch:             uint64(4),
		common.MetricFixDelegationChangeOwnerOnAccountEnableEpoch:             uint64(4),
		common.MetricDynamicGasCostForDataTrieStorageLoadEnableEpoch:          uint64(4),
		common.MetricNFTStopCreateEnableEpoch:                                 uint64(4),
		common.MetricChangeOwnerAddressCrossShardThroughSCEnableEpoch:         uint64(4),
		common.MetricFixGasRemainingForSaveKeyValueBuiltinFunctionEnableEpoch: uint64(4),
		common.MetricCurrentRandomnessOnSortingEnableEpoch:                    uint64(4),
		common.MetricStakeLimitsEnableEpoch:                                   uint64(4),
		common.MetricStakingV4Step1EnableEpoch:                                uint64(4),
		common.MetricStakingV4Step2EnableEpoch:                                uint64(4),
		common.MetricStakingV4Step3EnableEpoch:                                uint64(4),
		common.MetricCleanupAuctionOnLowWaitingListEnableEpoch:                uint64(4),
		common.MetricAlwaysMergeContextsInEEIEnableEpoch:                      uint64(4),
		common.MetricDynamicESDTEnableEpoch:                                   uint64(4),
		common.MetricEGLDInMultiTransferEnableEpoch:                           uint64(4),
		common.MetricCryptoOpcodesV2EnableEpoch:                               uint64(4),
		common.MetricMultiESDTNFTTransferAndExecuteByUserEnableEpoch:          uint64(4),
		common.MetricFixRelayedMoveBalanceToNonPayableSCEnableEpoch:           uint64(4),
		common.MetricRelayedTransactionsV3EnableEpoch:                         uint64(4),
		common.MetricRelayedTransactionsV3FixESDTTransferEnableEpoch:          uint64(4),
		common.MetricCheckBuiltInCallOnTransferValueAndFailEnableRound:        uint64(4),
		common.MetricMaskVMInternalDependenciesErrorsEnableEpoch:              uint64(4),
		common.MetricFixBackTransferOPCODEEnableEpoch:                         uint64(4),
		common.MetricValidationOnGobDecodeEnableEpoch:                         uint64(4),
		common.MetricBarnardOpcodesEnableEpoch:                                uint64(4),
		common.MetricAutomaticActivationOfNodesDisableEpoch:                   uint64(4),
		common.MetricFixGetBalanceEnableEpoch:                                 uint64(4),
		common.MetricMaxNodesChangeEnableEpoch: []map[string]interface{}{
			{
				common.MetricEpochEnable:            uint64(0),
				common.MetricMaxNumNodes:            uint64(1),
				common.MetricNodesToShufflePerShard: uint64(2),
			},
			{
				common.MetricEpochEnable:            uint64(3),
				common.MetricMaxNumNodes:            uint64(4),
				common.MetricNodesToShufflePerShard: uint64(5),
			},
		},
	}

	epochsMetrics, _ := sm.EnableEpochsMetrics()
	assert.Equal(t, expectedMetrics, epochsMetrics)
}

func TestStatusMetrics_RatingsConfig(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()

	sm.SetUInt64Value(common.MetricRatingsGeneralStartRating, uint64(5001))
	sm.SetUInt64Value(common.MetricRatingsGeneralMaxRating, uint64(10000))
	sm.SetUInt64Value(common.MetricRatingsGeneralMinRating, uint64(1))
	sm.SetStringValue(common.MetricRatingsGeneralSignedBlocksThreshold, "0.01")

	selectionChances := []map[string]uint64{
		{
			"MaxThreshold":  0,
			"ChancePercent": 5,
		},
		{
			"MaxThreshold":  1000,
			"ChancePercent": 10,
		},
	}

	for i, selectionChance := range selectionChances {
		maxThresholdStr := fmt.Sprintf("%s%d%s", common.MetricRatingsGeneralSelectionChances, i, common.SelectionChancesMaxThresholdSuffix)
		sm.SetUInt64Value(maxThresholdStr, selectionChance["MaxThreshold"])
		chancePercentStr := fmt.Sprintf("%s%d%s", common.MetricRatingsGeneralSelectionChances, i, common.SelectionChancesChancePercentSuffix)
		sm.SetUInt64Value(chancePercentStr, selectionChance["ChancePercent"])
	}
	sm.SetUInt64Value(common.MetricRatingsGeneralSelectionChances+"_count", uint64(len(selectionChances)))

	sm.SetUInt64Value(common.MetricRatingsShardChainHoursToMaxRatingFromStartRating, uint64(72))
	sm.SetStringValue(common.MetricRatingsShardChainProposerValidatorImportance, "1.0")
	sm.SetStringValue(common.MetricRatingsShardChainProposerDecreaseFactor, "-4.0")
	sm.SetStringValue(common.MetricRatingsShardChainValidatorDecreaseFactor, "-4.0")
	sm.SetStringValue(common.MetricRatingsShardChainConsecutiveMissedBlocksPenalty, "1.50")

	sm.SetUInt64Value(common.MetricRatingsMetaChainHoursToMaxRatingFromStartRating, uint64(55))
	sm.SetStringValue(common.MetricRatingsMetaChainProposerValidatorImportance, "1.0")
	sm.SetStringValue(common.MetricRatingsMetaChainProposerDecreaseFactor, "-4.0")
	sm.SetStringValue(common.MetricRatingsMetaChainValidatorDecreaseFactor, "-4.0")
	sm.SetStringValue(common.MetricRatingsMetaChainConsecutiveMissedBlocksPenalty, "1.50")

	sm.SetStringValue(common.MetricRatingsPeerHonestyDecayCoefficient, "0.97")
	sm.SetUInt64Value(common.MetricRatingsPeerHonestyDecayUpdateIntervalInSeconds, uint64(10))
	sm.SetStringValue(common.MetricRatingsPeerHonestyMaxScore, "100.0")
	sm.SetStringValue(common.MetricRatingsPeerHonestyMinScore, "-100.0")
	sm.SetStringValue(common.MetricRatingsPeerHonestyBadPeerThreshold, "-80.0")
	sm.SetStringValue(common.MetricRatingsPeerHonestyUnitValue, "1.0")

	expectedConfig := map[string]interface{}{
		common.MetricRatingsGeneralStartRating:           uint64(5001),
		common.MetricRatingsGeneralMaxRating:             uint64(10000),
		common.MetricRatingsGeneralMinRating:             uint64(1),
		common.MetricRatingsGeneralSignedBlocksThreshold: "0.01",

		common.MetricRatingsGeneralSelectionChances: []map[string]uint64{
			{
				common.MetricSelectionChancesMaxThreshold:  uint64(0),
				common.MetricSelectionChancesChancePercent: uint64(5),
			},
			{
				common.MetricSelectionChancesMaxThreshold:  uint64(1000),
				common.MetricSelectionChancesChancePercent: uint64(10),
			},
		},

		common.MetricRatingsShardChainHoursToMaxRatingFromStartRating: uint64(72),
		common.MetricRatingsShardChainProposerValidatorImportance:     "1.0",
		common.MetricRatingsShardChainProposerDecreaseFactor:          "-4.0",
		common.MetricRatingsShardChainValidatorDecreaseFactor:         "-4.0",
		common.MetricRatingsShardChainConsecutiveMissedBlocksPenalty:  "1.50",

		common.MetricRatingsMetaChainHoursToMaxRatingFromStartRating: uint64(55),
		common.MetricRatingsMetaChainProposerValidatorImportance:     "1.0",
		common.MetricRatingsMetaChainProposerDecreaseFactor:          "-4.0",
		common.MetricRatingsMetaChainValidatorDecreaseFactor:         "-4.0",
		common.MetricRatingsMetaChainConsecutiveMissedBlocksPenalty:  "1.50",

		common.MetricRatingsPeerHonestyDecayCoefficient:             "0.97",
		common.MetricRatingsPeerHonestyDecayUpdateIntervalInSeconds: uint64(10),
		common.MetricRatingsPeerHonestyMaxScore:                     "100.0",
		common.MetricRatingsPeerHonestyMinScore:                     "-100.0",
		common.MetricRatingsPeerHonestyBadPeerThreshold:             "-80.0",
		common.MetricRatingsPeerHonestyUnitValue:                    "1.0",
	}

	configMetrics, err := sm.RatingsMetrics()
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, configMetrics)
}

func TestStatusMetrics_BootstrapMetrics(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()

	sm.SetUInt64Value(common.MetricTrieSyncNumReceivedBytes, uint64(5001))
	sm.SetUInt64Value(common.MetricTrieSyncNumProcessedNodes, uint64(10000))
	sm.SetUInt64Value(common.MetricShardId, uint64(2))
	sm.SetStringValue(common.MetricGatewayMetricsEndpoint, "http://localhost:8080")

	expectedMetrics := map[string]interface{}{
		common.MetricTrieSyncNumReceivedBytes:  uint64(5001),
		common.MetricTrieSyncNumProcessedNodes: uint64(10000),
		common.MetricShardId:                   uint64(2),
		common.MetricGatewayMetricsEndpoint:    "http://localhost:8080",
	}

	bootstrapMetrics, err := sm.BootstrapMetrics()
	assert.NoError(t, err)
	assert.Equal(t, expectedMetrics, bootstrapMetrics)
}

func TestStatusMetrics_IncrementConcurrentOperations(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()

	testKey := "test key"
	sm.SetUInt64Value(testKey, 0)

	numIterations := 1000
	wg := sync.WaitGroup{}
	wg.Add(numIterations)

	for i := 0; i < numIterations; i++ {
		go func() {
			sm.Increment(testKey)
			wg.Done()
		}()
	}
	wg.Wait()

	val := sm.StatusMetricsMap()[testKey]
	require.Equal(t, uint64(numIterations), val.(uint64))
}

func TestStatusMetrics_ConcurrentIncrementAndDecrement(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()

	initialValue := uint64(5000)

	testKey := "test key"
	sm.SetUInt64Value(testKey, initialValue)

	numIterations := 1001
	wg := sync.WaitGroup{}
	wg.Add(numIterations)

	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			if idx%2 == 0 {
				sm.Increment(testKey)
			} else {
				sm.Decrement(testKey)
			}

			wg.Done()
		}(i)
	}
	wg.Wait()

	val := sm.StatusMetricsMap()[testKey]
	// we started with a value of initialValue (5000). after X + 1 increments and X decrements, the final
	// value should be the original value, plus one (5001)

	require.Equal(t, initialValue+1, val.(uint64))
}

func TestStatusMetrics_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()

	startTime := time.Now()

	defer func() {
		r := recover()
		require.Nil(t, r)
	}()
	numIterations := 1000
	wg := sync.WaitGroup{}
	wg.Add(numIterations)

	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			switch idx % 14 {
			case 0:
				sm.AddUint64("test", uint64(idx))
			case 1:
				sm.SetUInt64Value("test", uint64(idx))
			case 2:
				sm.Increment("test")
			case 3:
				sm.Decrement("test")
			case 4:
				sm.SetInt64Value("test i64", int64(idx))
			case 5:
				sm.SetStringValue("test str", "test val")
			case 6:
				_, _ = sm.NetworkMetrics()
			case 7:
				_, _ = sm.ConfigMetrics()
			case 8:
				_, _ = sm.EconomicsMetrics()
			case 9:
				_ = sm.StatusMetricsMap()
			case 10:
				_, _ = sm.StatusMetricsMapWithoutP2P()
			case 11:
				_, _ = sm.StatusMetricsWithoutP2PPrometheusString()
			case 12:
				_, _ = sm.StatusP2pMetricsMap()
			case 13:
				_, _ = sm.BootstrapMetrics()
			}
			wg.Done()
		}(i)
	}

	wg.Wait()

	elapsedTime := time.Since(startTime)
	require.True(t, elapsedTime < 10*time.Second, "if the test isn't finished within 10 seconds, there might be a deadlock somewhere")
}
