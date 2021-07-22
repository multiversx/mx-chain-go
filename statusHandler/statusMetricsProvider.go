package statusHandler

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/common"
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
	shardID := sm.loadUint64Metric(common.MetricShardId)
	metrics := sm.StatusMetricsMapWithoutP2P()
	stringBuilder := strings.Builder{}
	for key, value := range metrics {
		_, isUint64 := value.(uint64)
		_, isInt64 := value.(int64)
		isNumericValue := isUint64 || isInt64
		if isNumericValue {
			stringBuilder.WriteString(fmt.Sprintf("%s{%s=\"%d\"} %v\n", key, common.MetricShardId, shardID, value))
		}
	}

	return stringBuilder.String()
}

// EconomicsMetrics returns the economics related metrics
func (sm *statusMetrics) EconomicsMetrics() map[string]interface{} {
	economicsMetrics := make(map[string]interface{})

	economicsMetrics[common.MetricTotalSupply] = sm.loadStringMetric(common.MetricTotalSupply)
	economicsMetrics[common.MetricTotalFees] = sm.loadStringMetric(common.MetricTotalFees)
	economicsMetrics[common.MetricDevRewardsInEpoch] = sm.loadStringMetric(common.MetricDevRewardsInEpoch)
	economicsMetrics[common.MetricInflation] = sm.loadStringMetric(common.MetricInflation)
	economicsMetrics[common.MetricEpochForEconomicsData] = sm.loadUint64Metric(common.MetricEpochForEconomicsData)

	return economicsMetrics
}

// ConfigMetrics will return metrics related to current configuration
func (sm *statusMetrics) ConfigMetrics() map[string]interface{} {
	configMetrics := make(map[string]interface{})

	configMetrics[common.MetricNumShardsWithoutMetachain] = sm.loadUint64Metric(common.MetricNumShardsWithoutMetachain)
	configMetrics[common.MetricNumNodesPerShard] = sm.loadUint64Metric(common.MetricNumNodesPerShard)
	configMetrics[common.MetricNumMetachainNodes] = sm.loadUint64Metric(common.MetricNumMetachainNodes)
	configMetrics[common.MetricShardConsensusGroupSize] = sm.loadUint64Metric(common.MetricShardConsensusGroupSize)
	configMetrics[common.MetricMetaConsensusGroupSize] = sm.loadUint64Metric(common.MetricMetaConsensusGroupSize)
	configMetrics[common.MetricMinGasPrice] = sm.loadUint64Metric(common.MetricMinGasPrice)
	configMetrics[common.MetricMinGasLimit] = sm.loadUint64Metric(common.MetricMinGasLimit)
	configMetrics[common.MetricRewardsTopUpGradientPoint] = sm.loadStringMetric(common.MetricRewardsTopUpGradientPoint)
	configMetrics[common.MetricGasPerDataByte] = sm.loadUint64Metric(common.MetricGasPerDataByte)
	configMetrics[common.MetricChainId] = sm.loadStringMetric(common.MetricChainId)
	configMetrics[common.MetricRoundDuration] = sm.loadUint64Metric(common.MetricRoundDuration)
	configMetrics[common.MetricStartTime] = sm.loadUint64Metric(common.MetricStartTime)
	configMetrics[common.MetricLatestTagSoftwareVersion] = sm.loadStringMetric(common.MetricLatestTagSoftwareVersion)
	configMetrics[common.MetricDenomination] = sm.loadUint64Metric(common.MetricDenomination)
	configMetrics[common.MetricMinTransactionVersion] = sm.loadUint64Metric(common.MetricMinTransactionVersion)
	configMetrics[common.MetricTopUpFactor] = sm.loadStringMetric(common.MetricTopUpFactor)
	configMetrics[common.MetricGasPriceModifier] = sm.loadStringMetric(common.MetricGasPriceModifier)
	configMetrics[common.MetricRoundsPerEpoch] = sm.loadUint64Metric(common.MetricRoundsPerEpoch)

	return configMetrics
}

// EnableEpochsMetrics will return metrics related to activation epochs
func (sm *statusMetrics) EnableEpochsMetrics() map[string]interface{} {
	enableEpochsMetrics := make(map[string]interface{})

	enableEpochsMetrics[common.MetricScDeployEnableEpoch] = sm.loadUint64Metric(common.MetricScDeployEnableEpoch)
	enableEpochsMetrics[common.MetricBuiltInFunctionsEnableEpoch] = sm.loadUint64Metric(common.MetricBuiltInFunctionsEnableEpoch)
	enableEpochsMetrics[common.MetricRelayedTransactionsEnableEpoch] = sm.loadUint64Metric(common.MetricRelayedTransactionsEnableEpoch)
	enableEpochsMetrics[common.MetricPenalizedTooMuchGasEnableEpoch] = sm.loadUint64Metric(common.MetricPenalizedTooMuchGasEnableEpoch)
	enableEpochsMetrics[common.MetricSwitchJailWaitingEnableEpoch] = sm.loadUint64Metric(common.MetricSwitchJailWaitingEnableEpoch)
	enableEpochsMetrics[common.MetricSwitchHysteresisForMinNodesEnableEpoch] = sm.loadUint64Metric(common.MetricSwitchHysteresisForMinNodesEnableEpoch)
	enableEpochsMetrics[common.MetricBelowSignedThresholdEnableEpoch] = sm.loadUint64Metric(common.MetricBelowSignedThresholdEnableEpoch)
	enableEpochsMetrics[common.MetricTransactionSignedWithTxHashEnableEpoch] = sm.loadUint64Metric(common.MetricTransactionSignedWithTxHashEnableEpoch)
	enableEpochsMetrics[common.MetricMetaProtectionEnableEpoch] = sm.loadUint64Metric(common.MetricMetaProtectionEnableEpoch)
	enableEpochsMetrics[common.MetricAheadOfTimeGasUsageEnableEpoch] = sm.loadUint64Metric(common.MetricAheadOfTimeGasUsageEnableEpoch)
	enableEpochsMetrics[common.MetricGasPriceModifierEnableEpoch] = sm.loadUint64Metric(common.MetricGasPriceModifierEnableEpoch)
	enableEpochsMetrics[common.MetricRepairCallbackEnableEpoch] = sm.loadUint64Metric(common.MetricRepairCallbackEnableEpoch)
	enableEpochsMetrics[common.MetricBlockGasAndFreeRecheckEnableEpoch] = sm.loadUint64Metric(common.MetricBlockGasAndFreeRecheckEnableEpoch)
	enableEpochsMetrics[common.MetricStakingV2EnableEpoch] = sm.loadUint64Metric(common.MetricStakingV2EnableEpoch)
	enableEpochsMetrics[common.MetricStakeEnableEpoch] = sm.loadUint64Metric(common.MetricStakeEnableEpoch)
	enableEpochsMetrics[common.MetricDoubleKeyProtectionEnableEpoch] = sm.loadUint64Metric(common.MetricDoubleKeyProtectionEnableEpoch)
	enableEpochsMetrics[common.MetricEsdtEnableEpoch] = sm.loadUint64Metric(common.MetricEsdtEnableEpoch)
	enableEpochsMetrics[common.MetricGovernanceEnableEpoch] = sm.loadUint64Metric(common.MetricGovernanceEnableEpoch)
	enableEpochsMetrics[common.MetricDelegationManagerEnableEpoch] = sm.loadUint64Metric(common.MetricDelegationManagerEnableEpoch)
	enableEpochsMetrics[common.MetricDelegationSmartContractEnableEpoch] = sm.loadUint64Metric(common.MetricDelegationSmartContractEnableEpoch)
	enableEpochsMetrics[common.MetricIncrementSCRNonceInMultiTransferEnableEpoch] = sm.loadUint64Metric(common.MetricIncrementSCRNonceInMultiTransferEnableEpoch)

	return enableEpochsMetrics
}

// NetworkMetrics will return metrics related to current configuration
func (sm *statusMetrics) NetworkMetrics() map[string]interface{} {
	networkMetrics := make(map[string]interface{})

	currentRound := sm.loadUint64Metric(common.MetricCurrentRound)
	roundNumberAtEpochStart := sm.loadUint64Metric(common.MetricRoundAtEpochStart)

	currentNonce := sm.loadUint64Metric(common.MetricNonce)
	nonceAtEpochStart := sm.loadUint64Metric(common.MetricNonceAtEpochStart)

	networkMetrics[common.MetricNonce] = currentNonce
	networkMetrics[common.MetricHighestFinalBlock] = sm.loadUint64Metric(common.MetricHighestFinalBlock)
	networkMetrics[common.MetricCurrentRound] = currentRound
	networkMetrics[common.MetricRoundAtEpochStart] = roundNumberAtEpochStart
	networkMetrics[common.MetricNonceAtEpochStart] = nonceAtEpochStart
	networkMetrics[common.MetricEpochNumber] = sm.loadUint64Metric(common.MetricEpochNumber)
	networkMetrics[common.MetricRoundsPerEpoch] = sm.loadUint64Metric(common.MetricRoundsPerEpoch)
	roundsPassedInEpoch := uint64(0)
	if currentRound >= roundNumberAtEpochStart {
		roundsPassedInEpoch = currentRound - roundNumberAtEpochStart
	}
	networkMetrics[common.MetricRoundsPassedInCurrentEpoch] = roundsPassedInEpoch

	noncesPassedInEpoch := uint64(0)
	if currentNonce >= nonceAtEpochStart {
		noncesPassedInEpoch = currentNonce - nonceAtEpochStart
	}
	networkMetrics[common.MetricNoncesPassedInCurrentEpoch] = noncesPassedInEpoch

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
