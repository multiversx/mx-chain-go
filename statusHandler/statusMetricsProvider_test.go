package statusHandler_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
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

	strRes := sm.StatusMetricsWithoutP2PPrometheusString()

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

	strRes := sm.StatusMetricsWithoutP2PPrometheusString()

	expectedMetricOutput := fmt.Sprintf("%s{%s=\"%d\"} %v", key1, common.MetricShardId, shardID, value1)
	assert.True(t, strings.Contains(strRes, expectedMetricOutput))
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
	sm.SetStringValue(common.MetricRewardsTopUpGradientPoint, "12345")
	sm.SetUInt64Value(common.MetricGasPerDataByte, 1500)
	sm.SetStringValue(common.MetricChainId, "local-id")
	sm.SetUInt64Value(common.MetricRoundDuration, 5000)
	sm.SetUInt64Value(common.MetricStartTime, 9999)
	sm.SetStringValue(common.MetricLatestTagSoftwareVersion, "version1.0")
	sm.SetUInt64Value(common.MetricDenomination, 18)
	sm.SetUInt64Value(common.MetricMinTransactionVersion, 2)
	sm.SetStringValue(common.MetricTopUpFactor, fmt.Sprintf("%g", 12.134))
	sm.SetStringValue(common.MetricGasPriceModifier, fmt.Sprintf("%g", 0.5))
	sm.SetUInt64Value(common.MetricRoundsPerEpoch, uint64(144))

	expectedConfig := map[string]interface{}{
		"erd_chain_id":                      "local-id",
		"erd_denomination":                  uint64(18),
		"erd_gas_per_data_byte":             uint64(1500),
		"erd_latest_tag_software_version":   "version1.0",
		"erd_meta_consensus_group_size":     uint64(25),
		"erd_min_gas_limit":                 uint64(50000),
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
	}

	configMetrics := sm.ConfigMetrics()
	assert.Equal(t, expectedConfig, configMetrics)
}

func TestStatusMetrics_EnableEpochMetrics(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()

	sm.SetUInt64Value(common.MetricScDeployEnableEpoch, 4)
	sm.SetUInt64Value(common.MetricBuiltInFunctionsEnableEpoch, 2)
	sm.SetUInt64Value(common.MetricRelayedTransactionsEnableEpoch, 4)
	sm.SetUInt64Value(common.MetricPenalizedTooMuchGasEnableEpoch, 2)
	sm.SetUInt64Value(common.MetricSwitchJailWaitingEnableEpoch, 2)
	sm.SetUInt64Value(common.MetricSwitchHysteresisForMinNodesEnableEpoch, 4)
	sm.SetUInt64Value(common.MetricBelowSignedThresholdEnableEpoch, 2)
	sm.SetUInt64Value(common.MetricTransactionSignedWithTxHashEnableEpoch, 4)
	sm.SetUInt64Value(common.MetricMetaProtectionEnableEpoch, 6)
	sm.SetUInt64Value(common.MetricAheadOfTimeGasUsageEnableEpoch, 2)
	sm.SetUInt64Value(common.MetricGasPriceModifierEnableEpoch, 2)
	sm.SetUInt64Value(common.MetricRepairCallbackEnableEpoch, 2)
	sm.SetUInt64Value(common.MetricBlockGasAndFreeRecheckEnableEpoch, 2)
	sm.SetUInt64Value(common.MetricStakingV2EnableEpoch, 2)
	sm.SetUInt64Value(common.MetricStakeEnableEpoch, 2)
	sm.SetUInt64Value(common.MetricDoubleKeyProtectionEnableEpoch, 2)
	sm.SetUInt64Value(common.MetricEsdtEnableEpoch, 4)
	sm.SetUInt64Value(common.MetricGovernanceEnableEpoch, 3)
	sm.SetUInt64Value(common.MetricDelegationManagerEnableEpoch, 1)
	sm.SetUInt64Value(common.MetricDelegationSmartContractEnableEpoch, 2)
	sm.SetUInt64Value(common.MetricIncrementSCRNonceInMultiTransferEnableEpoch, 3)

	expectedMetrics := map[string]interface{}{
		common.MetricScDeployEnableEpoch:                    uint64(4),
		common.MetricBuiltInFunctionsEnableEpoch:            uint64(2),
		common.MetricRelayedTransactionsEnableEpoch:         uint64(4),
		common.MetricPenalizedTooMuchGasEnableEpoch:         uint64(2),
		common.MetricSwitchJailWaitingEnableEpoch:           uint64(2),
		common.MetricSwitchHysteresisForMinNodesEnableEpoch: uint64(4),
		common.MetricBelowSignedThresholdEnableEpoch:        uint64(2),
		common.MetricTransactionSignedWithTxHashEnableEpoch: uint64(4),
		common.MetricMetaProtectionEnableEpoch:              uint64(6),
		common.MetricAheadOfTimeGasUsageEnableEpoch:         uint64(2),
		common.MetricGasPriceModifierEnableEpoch:            uint64(2),
		common.MetricRepairCallbackEnableEpoch:              uint64(2),
		common.MetricBlockGasAndFreeRecheckEnableEpoch:      uint64(2),
		common.MetricStakingV2EnableEpoch:                   uint64(2),
		common.MetricStakeEnableEpoch:                       uint64(2),
		common.MetricDoubleKeyProtectionEnableEpoch:         uint64(2),
		common.MetricEsdtEnableEpoch:                        uint64(4),
		common.MetricGovernanceEnableEpoch:                  uint64(3),
		common.MetricDelegationManagerEnableEpoch:           uint64(1),
		common.MetricDelegationSmartContractEnableEpoch:     uint64(2),

		common.MetricIncrementSCRNonceInMultiTransferEnableEpoch: uint64(3),
	}

	epochsMetrics := sm.EnableEpochsMetrics()
	assert.Equal(t, expectedMetrics, epochsMetrics)
}
