package statusHandler_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
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

	expectedMetricOutput := fmt.Sprintf("%s{%s=\"%d\"} %v", key1, core.MetricShardId, 0, value1)
	assert.True(t, strings.Contains(strRes, expectedMetricOutput))
}

func TestStatusMetrics_StatusMetricsWithoutP2PPrometheusStringShouldPutCorrectShardIDLabel(t *testing.T) {
	t.Parallel()

	shardID := uint32(37)
	sm := statusHandler.NewStatusMetrics()
	key1, value1 := "test-key7", uint64(100)
	key2, value2 := "test-key8", "value8"
	key3, value3 := core.MetricShardId, shardID
	sm.SetUInt64Value(key1, value1)
	sm.SetStringValue(key2, value2)
	sm.SetUInt64Value(key3, uint64(value3))

	strRes := sm.StatusMetricsWithoutP2PPrometheusString()

	expectedMetricOutput := fmt.Sprintf("%s{%s=\"%d\"} %v", key1, core.MetricShardId, shardID, value1)
	assert.True(t, strings.Contains(strRes, expectedMetricOutput))
}

func TestStatusMetrics_NetworkConfig(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()

	sm.SetUInt64Value(core.MetricNumShardsWithoutMetacahin, 1)
	sm.SetUInt64Value(core.MetricNumNodesPerShard, 100)
	sm.SetUInt64Value(core.MetricNumMetachainNodes, 50)
	sm.SetUInt64Value(core.MetricShardConsensusGroupSize, 20)
	sm.SetUInt64Value(core.MetricMetaConsensusGroupSize, 25)
	sm.SetUInt64Value(core.MetricMinGasPrice, 1000)
	sm.SetUInt64Value(core.MetricMinGasLimit, 50000)
	sm.SetStringValue(core.MetricRewardsTopUpGradientPoint, "12345")
	sm.SetUInt64Value(core.MetricGasPerDataByte, 1500)
	sm.SetStringValue(core.MetricChainId, "local-id")
	sm.SetUInt64Value(core.MetricRoundDuration, 5000)
	sm.SetUInt64Value(core.MetricStartTime, 9999)
	sm.SetStringValue(core.MetricLatestTagSoftwareVersion, "version1.0")
	sm.SetUInt64Value(core.MetricDenomination, 18)
	sm.SetUInt64Value(core.MetricMinTransactionVersion, 2)
	sm.SetStringValue(core.MetricTopUpFactor, fmt.Sprintf("%g", 12.134))
	sm.SetStringValue(core.MetricGasPriceModifier, fmt.Sprintf("%g", 0.5))

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
	}

	configMetrics := sm.ConfigMetrics()
	assert.Equal(t, expectedConfig, configMetrics)
}
