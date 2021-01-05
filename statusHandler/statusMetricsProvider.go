package statusHandler

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
)

// statusMetrics will handle displaying at /node/details all metrics already collected for other status handlers
type statusMetrics struct {
	nodeMetrics *sync.Map
}

// NewStatusMetrics will return an instance of the struct
func NewStatusMetrics() *statusMetrics {
	return &statusMetrics{
		nodeMetrics: &sync.Map{},
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *statusMetrics) IsInterfaceNil() bool {
	return sm == nil
}

// Increment method increment a metric
func (sm *statusMetrics) Increment(key string) {
	keyValueI, ok := sm.nodeMetrics.Load(key)
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}

	keyValue++
	sm.nodeMetrics.Store(key, keyValue)
}

// AddUint64 method increase a metric with a specific value
func (sm *statusMetrics) AddUint64(key string, val uint64) {
	keyValueI, ok := sm.nodeMetrics.Load(key)
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}

	keyValue += val
	sm.nodeMetrics.Store(key, keyValue)
}

// Decrement method - decrement a metric
func (sm *statusMetrics) Decrement(key string) {
	keyValueI, ok := sm.nodeMetrics.Load(key)
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}
	if keyValue == 0 {
		return
	}

	keyValue--
	sm.nodeMetrics.Store(key, keyValue)
}

// SetInt64Value method - sets an int64 value for a key
func (sm *statusMetrics) SetInt64Value(key string, value int64) {
	sm.nodeMetrics.Store(key, value)
}

// SetUInt64Value method - sets an uint64 value for a key
func (sm *statusMetrics) SetUInt64Value(key string, value uint64) {
	sm.nodeMetrics.Store(key, value)
}

// SetStringValue method - sets a string value for a key
func (sm *statusMetrics) SetStringValue(key string, value string) {
	sm.nodeMetrics.Store(key, value)
}

// Close method - won't do anything
func (sm *statusMetrics) Close() {
}

// StatusMetricsMapWithoutP2P will return the non-p2p metrics in a map
func (sm *statusMetrics) StatusMetricsMapWithoutP2P() map[string]interface{} {
	statusMetricsMap := make(map[string]interface{})
	sm.nodeMetrics.Range(func(key, value interface{}) bool {
		keyString := key.(string)
		if strings.Contains(keyString, "_p2p_") {
			return true
		}

		statusMetricsMap[key.(string)] = value
		return true
	})

	return statusMetricsMap
}

// StatusP2pMetricsMap will return the p2p metrics in a map
func (sm *statusMetrics) StatusP2pMetricsMap() map[string]interface{} {
	statusMetricsMap := make(map[string]interface{})
	sm.nodeMetrics.Range(func(key, value interface{}) bool {
		keyString := key.(string)
		if !strings.Contains(keyString, "_p2p_") {
			return true
		}

		statusMetricsMap[key.(string)] = value
		return true
	})

	return statusMetricsMap
}

// StatusMetricsWithoutP2PPrometheusString returns the metrics in a string format which respects prometheus style
func (sm *statusMetrics) StatusMetricsWithoutP2PPrometheusString() string {
	shardID := sm.loadUint64Metric(core.MetricShardId)
	metrics := sm.StatusMetricsMapWithoutP2P()
	stringBuilder := strings.Builder{}
	for key, value := range metrics {
		_, isUint64 := value.(uint64)
		_, isInt64 := value.(int64)
		isNumericValue := isUint64 || isInt64
		if isNumericValue {
			stringBuilder.WriteString(fmt.Sprintf("%s{%s=\"%d\"} %v\n", key, core.MetricShardId, shardID, value))
		}
	}

	return stringBuilder.String()
}

// EconomicsMetrics returns the economics related metrics
func (sm *statusMetrics) EconomicsMetrics() map[string]interface{} {
	economicsMetrics := make(map[string]interface{})

	economicsMetrics[core.MetricTotalSupply] = sm.loadStringMetric(core.MetricTotalSupply)
	economicsMetrics[core.MetricTotalFees] = sm.loadStringMetric(core.MetricTotalFees)
	economicsMetrics[core.MetricDevRewards] = sm.loadStringMetric(core.MetricDevRewards)
	economicsMetrics[core.MetricInflation] = sm.loadStringMetric(core.MetricInflation)
	economicsMetrics[core.MetricEpochForEconomicsData] = sm.loadUint64Metric(core.MetricEpochForEconomicsData)

	return economicsMetrics
}

// ConfigMetrics will return metrics related to current configuration
func (sm *statusMetrics) ConfigMetrics() map[string]interface{} {
	configMetrics := make(map[string]interface{})

	configMetrics[core.MetricNumShardsWithoutMetacahin] = sm.loadUint64Metric(core.MetricNumShardsWithoutMetacahin)
	configMetrics[core.MetricNumNodesPerShard] = sm.loadUint64Metric(core.MetricNumNodesPerShard)
	configMetrics[core.MetricNumMetachainNodes] = sm.loadUint64Metric(core.MetricNumMetachainNodes)
	configMetrics[core.MetricShardConsensusGroupSize] = sm.loadUint64Metric(core.MetricShardConsensusGroupSize)
	configMetrics[core.MetricMetaConsensusGroupSize] = sm.loadUint64Metric(core.MetricMetaConsensusGroupSize)
	configMetrics[core.MetricMinGasPrice] = sm.loadUint64Metric(core.MetricMinGasPrice)
	configMetrics[core.MetricMinGasLimit] = sm.loadUint64Metric(core.MetricMinGasLimit)
	configMetrics[core.MetricRewardsTopUpGradientPoint] = sm.loadStringMetric(core.MetricRewardsTopUpGradientPoint)
	configMetrics[core.MetricGasPerDataByte] = sm.loadUint64Metric(core.MetricGasPerDataByte)
	configMetrics[core.MetricChainId] = sm.loadStringMetric(core.MetricChainId)
	configMetrics[core.MetricRoundDuration] = sm.loadUint64Metric(core.MetricRoundDuration)
	configMetrics[core.MetricStartTime] = sm.loadUint64Metric(core.MetricStartTime)
	configMetrics[core.MetricLatestTagSoftwareVersion] = sm.loadStringMetric(core.MetricLatestTagSoftwareVersion)
	configMetrics[core.MetricDenomination] = sm.loadUint64Metric(core.MetricDenomination)
	configMetrics[core.MetricMinTransactionVersion] = sm.loadUint64Metric(core.MetricMinTransactionVersion)

	return configMetrics
}

// NetworkMetrics will return metrics related to current configuration
func (sm *statusMetrics) NetworkMetrics() map[string]interface{} {
	networkMetrics := make(map[string]interface{})

	currentRound := sm.loadUint64Metric(core.MetricCurrentRound)
	roundNumberAtEpochStart := sm.loadUint64Metric(core.MetricRoundAtEpochStart)

	currentNonce := sm.loadUint64Metric(core.MetricNonce)
	nonceAtEpochStart := sm.loadUint64Metric(core.MetricNonceAtEpochStart)

	networkMetrics[core.MetricNonce] = currentNonce
	networkMetrics[core.MetricHighestFinalBlock] = sm.loadUint64Metric(core.MetricHighestFinalBlock)
	networkMetrics[core.MetricCurrentRound] = currentRound
	networkMetrics[core.MetricRoundAtEpochStart] = roundNumberAtEpochStart
	networkMetrics[core.MetricNonceAtEpochStart] = nonceAtEpochStart
	networkMetrics[core.MetricEpochNumber] = sm.loadUint64Metric(core.MetricEpochNumber)
	networkMetrics[core.MetricRoundsPerEpoch] = sm.loadUint64Metric(core.MetricRoundsPerEpoch)
	roundsPassedInEpoch := uint64(0)
	if currentRound >= roundNumberAtEpochStart {
		roundsPassedInEpoch = currentRound - roundNumberAtEpochStart
	}
	networkMetrics[core.MetricRoundsPassedInCurrentEpoch] = roundsPassedInEpoch

	noncesPassedInEpoch := uint64(0)
	if currentNonce >= nonceAtEpochStart {
		noncesPassedInEpoch = currentNonce - nonceAtEpochStart
	}
	networkMetrics[core.MetricNoncesPassedInCurrentEpoch] = noncesPassedInEpoch

	return networkMetrics
}

func (sm *statusMetrics) loadUint64Metric(metric string) uint64 {
	metricObj, ok := sm.nodeMetrics.Load(metric)
	if !ok {
		return 0
	}
	metricAsUint64, ok := metricObj.(uint64)
	if !ok {
		return 0
	}

	return metricAsUint64
}

func (sm *statusMetrics) loadStringMetric(metric string) string {
	metricObj, ok := sm.nodeMetrics.Load(metric)
	if !ok {
		return ""
	}
	metricAsString, ok := metricObj.(string)
	if !ok {
		return ""
	}

	return metricAsString
}
