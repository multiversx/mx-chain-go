package statusHandler

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/common"
)

// statusMetrics will handle displaying at /node/details all metrics already collected for other status handlers
type statusMetrics struct {
	uint64Metrics       map[string]uint64
	mutUint64Operations sync.RWMutex

	stringMetrics       map[string]string
	mutStringOperations sync.RWMutex

	int64Metrics       map[string]int64
	mutInt64Operations sync.RWMutex
}

// NewStatusMetrics will return an instance of the struct
func NewStatusMetrics() *statusMetrics {
	return &statusMetrics{
		uint64Metrics: make(map[string]uint64),
		stringMetrics: make(map[string]string),
		int64Metrics:  make(map[string]int64),
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *statusMetrics) IsInterfaceNil() bool {
	return sm == nil
}

// Increment method increment a metric
func (sm *statusMetrics) Increment(key string) {
	sm.mutUint64Operations.Lock()
	defer sm.mutUint64Operations.Unlock()

	value, ok := sm.uint64Metrics[key]
	if !ok {
		return
	}

	value++
	sm.uint64Metrics[key] = value
}

// AddUint64 method increase a metric with a specific value
func (sm *statusMetrics) AddUint64(key string, val uint64) {
	sm.mutUint64Operations.Lock()
	defer sm.mutUint64Operations.Unlock()

	value, ok := sm.uint64Metrics[key]
	if !ok {
		return
	}

	value += val
	sm.uint64Metrics[key] = value
}

// Decrement method - decrement a metric
func (sm *statusMetrics) Decrement(key string) {
	sm.mutUint64Operations.Lock()
	defer sm.mutUint64Operations.Unlock()

	value, ok := sm.uint64Metrics[key]
	if !ok {
		return
	}

	if value == 0 {
		return
	}

	value--
	sm.uint64Metrics[key] = value
}

// SetInt64Value method - sets an int64 value for a key
func (sm *statusMetrics) SetInt64Value(key string, value int64) {
	sm.mutInt64Operations.Lock()
	defer sm.mutInt64Operations.Unlock()

	sm.int64Metrics[key] = value
}

// SetUInt64Value method - sets an uint64 value for a key
func (sm *statusMetrics) SetUInt64Value(key string, value uint64) {
	sm.mutUint64Operations.Lock()
	defer sm.mutUint64Operations.Unlock()

	sm.uint64Metrics[key] = value
}

// SetStringValue method - sets a string value for a key
func (sm *statusMetrics) SetStringValue(key string, value string) {
	sm.mutStringOperations.Lock()
	defer sm.mutStringOperations.Unlock()

	sm.stringMetrics[key] = value
}

// Close method - won't do anything
func (sm *statusMetrics) Close() {
}

// StatusMetricsMapWithoutP2P will return the non-p2p metrics in a map
func (sm *statusMetrics) StatusMetricsMapWithoutP2P() map[string]interface{} {
	return sm.getMetricsWithKeyFilterMutexProtected(func(input string) bool {
		return !strings.Contains(input, "_p2p_")
	})
}

// StatusP2pMetricsMap will return the p2p metrics in a map
func (sm *statusMetrics) StatusP2pMetricsMap() map[string]interface{} {
	return sm.getMetricsWithKeyFilterMutexProtected(func(input string) bool {
		return strings.Contains(input, "_p2p_")
	})
}

func (sm *statusMetrics) getMetricsWithKeyFilterMutexProtected(filterFunc func(input string) bool) map[string]interface{} {
	statusMetricsMap := make(map[string]interface{})

	sm.mutUint64Operations.RLock()
	for key, value := range sm.uint64Metrics {
		if !filterFunc(key) {
			continue
		}

		statusMetricsMap[key] = value
	}
	sm.mutUint64Operations.RUnlock()

	sm.mutStringOperations.RLock()
	for key, value := range sm.stringMetrics {
		if !filterFunc(key) {
			continue
		}

		statusMetricsMap[key] = value
	}
	sm.mutStringOperations.RUnlock()

	sm.mutInt64Operations.RLock()
	for key, value := range sm.int64Metrics {
		if !filterFunc(key) {
			continue
		}

		statusMetricsMap[key] = value
	}
	sm.mutInt64Operations.RUnlock()

	return statusMetricsMap
}

// StatusMetricsWithoutP2PPrometheusString returns the metrics in a string format which respects prometheus style
func (sm *statusMetrics) StatusMetricsWithoutP2PPrometheusString() string {
	metrics := sm.StatusMetricsMapWithoutP2P()

	sm.mutUint64Operations.RLock()
	shardID := sm.uint64Metrics[common.MetricShardId]
	sm.mutUint64Operations.RUnlock()

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

	sm.mutStringOperations.RLock()
	economicsMetrics[common.MetricTotalSupply] = sm.stringMetrics[common.MetricTotalSupply]
	economicsMetrics[common.MetricTotalFees] = sm.stringMetrics[common.MetricTotalFees]
	economicsMetrics[common.MetricDevRewardsInEpoch] = sm.stringMetrics[common.MetricDevRewardsInEpoch]
	economicsMetrics[common.MetricInflation] = sm.stringMetrics[common.MetricInflation]
	sm.mutStringOperations.RUnlock()

	sm.mutUint64Operations.RLock()
	economicsMetrics[common.MetricEpochForEconomicsData] = sm.uint64Metrics[common.MetricEpochForEconomicsData]
	sm.mutUint64Operations.RUnlock()

	return economicsMetrics
}

// ConfigMetrics will return metrics related to current configuration
func (sm *statusMetrics) ConfigMetrics() map[string]interface{} {
	configMetrics := make(map[string]interface{})

	sm.mutUint64Operations.RLock()
	configMetrics[common.MetricNumShardsWithoutMetachain] = sm.uint64Metrics[common.MetricNumShardsWithoutMetachain]
	configMetrics[common.MetricNumNodesPerShard] = sm.uint64Metrics[common.MetricNumNodesPerShard]
	configMetrics[common.MetricNumMetachainNodes] = sm.uint64Metrics[common.MetricNumMetachainNodes]
	configMetrics[common.MetricShardConsensusGroupSize] = sm.uint64Metrics[common.MetricShardConsensusGroupSize]
	configMetrics[common.MetricMetaConsensusGroupSize] = sm.uint64Metrics[common.MetricMetaConsensusGroupSize]
	configMetrics[common.MetricMinGasPrice] = sm.uint64Metrics[common.MetricMinGasPrice]
	configMetrics[common.MetricMinGasLimit] = sm.uint64Metrics[common.MetricMinGasLimit]
	configMetrics[common.MetricMaxGasPerTransaction] = sm.uint64Metrics[common.MetricMaxGasPerTransaction]
	configMetrics[common.MetricRoundDuration] = sm.uint64Metrics[common.MetricRoundDuration]
	configMetrics[common.MetricStartTime] = sm.uint64Metrics[common.MetricStartTime]
	configMetrics[common.MetricDenomination] = sm.uint64Metrics[common.MetricDenomination]
	configMetrics[common.MetricMinTransactionVersion] = sm.uint64Metrics[common.MetricMinTransactionVersion]
	configMetrics[common.MetricRoundsPerEpoch] = sm.uint64Metrics[common.MetricRoundsPerEpoch]
	configMetrics[common.MetricGasPerDataByte] = sm.uint64Metrics[common.MetricGasPerDataByte]
	sm.mutUint64Operations.RUnlock()

	sm.mutStringOperations.RLock()
	configMetrics[common.MetricRewardsTopUpGradientPoint] = sm.stringMetrics[common.MetricRewardsTopUpGradientPoint]
	configMetrics[common.MetricChainId] = sm.stringMetrics[common.MetricChainId]
	configMetrics[common.MetricLatestTagSoftwareVersion] = sm.stringMetrics[common.MetricLatestTagSoftwareVersion]
	configMetrics[common.MetricTopUpFactor] = sm.stringMetrics[common.MetricTopUpFactor]
	configMetrics[common.MetricGasPriceModifier] = sm.stringMetrics[common.MetricGasPriceModifier]
	sm.mutStringOperations.RUnlock()

	return configMetrics
}

// EnableEpochsMetrics will return metrics related to activation epochs
func (sm *statusMetrics) EnableEpochsMetrics() map[string]interface{} {
	enableEpochsMetrics := make(map[string]interface{})

	sm.mutUint64Operations.RLock()
	enableEpochsMetrics[common.MetricScDeployEnableEpoch] = sm.uint64Metrics[common.MetricScDeployEnableEpoch]
	enableEpochsMetrics[common.MetricBuiltInFunctionsEnableEpoch] = sm.uint64Metrics[common.MetricBuiltInFunctionsEnableEpoch]
	enableEpochsMetrics[common.MetricRelayedTransactionsEnableEpoch] = sm.uint64Metrics[common.MetricRelayedTransactionsEnableEpoch]
	enableEpochsMetrics[common.MetricPenalizedTooMuchGasEnableEpoch] = sm.uint64Metrics[common.MetricPenalizedTooMuchGasEnableEpoch]
	enableEpochsMetrics[common.MetricSwitchJailWaitingEnableEpoch] = sm.uint64Metrics[common.MetricSwitchJailWaitingEnableEpoch]
	enableEpochsMetrics[common.MetricSwitchHysteresisForMinNodesEnableEpoch] = sm.uint64Metrics[common.MetricSwitchHysteresisForMinNodesEnableEpoch]
	enableEpochsMetrics[common.MetricBelowSignedThresholdEnableEpoch] = sm.uint64Metrics[common.MetricBelowSignedThresholdEnableEpoch]
	enableEpochsMetrics[common.MetricTransactionSignedWithTxHashEnableEpoch] = sm.uint64Metrics[common.MetricTransactionSignedWithTxHashEnableEpoch]
	enableEpochsMetrics[common.MetricMetaProtectionEnableEpoch] = sm.uint64Metrics[common.MetricMetaProtectionEnableEpoch]
	enableEpochsMetrics[common.MetricAheadOfTimeGasUsageEnableEpoch] = sm.uint64Metrics[common.MetricAheadOfTimeGasUsageEnableEpoch]
	enableEpochsMetrics[common.MetricGasPriceModifierEnableEpoch] = sm.uint64Metrics[common.MetricGasPriceModifierEnableEpoch]
	enableEpochsMetrics[common.MetricRepairCallbackEnableEpoch] = sm.uint64Metrics[common.MetricRepairCallbackEnableEpoch]
	enableEpochsMetrics[common.MetricBlockGasAndFreeRecheckEnableEpoch] = sm.uint64Metrics[common.MetricBlockGasAndFreeRecheckEnableEpoch]
	enableEpochsMetrics[common.MetricStakingV2EnableEpoch] = sm.uint64Metrics[common.MetricStakingV2EnableEpoch]
	enableEpochsMetrics[common.MetricStakeEnableEpoch] = sm.uint64Metrics[common.MetricStakeEnableEpoch]
	enableEpochsMetrics[common.MetricDoubleKeyProtectionEnableEpoch] = sm.uint64Metrics[common.MetricDoubleKeyProtectionEnableEpoch]
	enableEpochsMetrics[common.MetricEsdtEnableEpoch] = sm.uint64Metrics[common.MetricEsdtEnableEpoch]
	enableEpochsMetrics[common.MetricGovernanceEnableEpoch] = sm.uint64Metrics[common.MetricGovernanceEnableEpoch]
	enableEpochsMetrics[common.MetricDelegationManagerEnableEpoch] = sm.uint64Metrics[common.MetricDelegationManagerEnableEpoch]
	enableEpochsMetrics[common.MetricDelegationSmartContractEnableEpoch] = sm.uint64Metrics[common.MetricDelegationSmartContractEnableEpoch]
	enableEpochsMetrics[common.MetricIncrementSCRNonceInMultiTransferEnableEpoch] = sm.uint64Metrics[common.MetricIncrementSCRNonceInMultiTransferEnableEpoch]
	enableEpochsMetrics[common.MetricHeartbeatDisableEpoch] = sm.uint64Metrics[common.MetricHeartbeatDisableEpoch]
	sm.mutUint64Operations.RUnlock()

	return enableEpochsMetrics
}

// NetworkMetrics will return metrics related to current configuration
func (sm *statusMetrics) NetworkMetrics() map[string]interface{} {
	sm.mutUint64Operations.RLock()
	defer sm.mutUint64Operations.RUnlock()

	networkMetrics := make(map[string]interface{})

	currentRound := sm.uint64Metrics[common.MetricCurrentRound]
	roundNumberAtEpochStart := sm.uint64Metrics[common.MetricRoundAtEpochStart]

	currentNonce := sm.uint64Metrics[common.MetricNonce]
	nonceAtEpochStart := sm.uint64Metrics[common.MetricNonceAtEpochStart]
	networkMetrics[common.MetricNonce] = currentNonce
	networkMetrics[common.MetricHighestFinalBlock] = sm.uint64Metrics[common.MetricHighestFinalBlock]
	networkMetrics[common.MetricCurrentRound] = currentRound
	networkMetrics[common.MetricRoundAtEpochStart] = roundNumberAtEpochStart
	networkMetrics[common.MetricNonceAtEpochStart] = nonceAtEpochStart
	networkMetrics[common.MetricEpochNumber] = sm.uint64Metrics[common.MetricEpochNumber]
	networkMetrics[common.MetricRoundsPerEpoch] = sm.uint64Metrics[common.MetricRoundsPerEpoch]
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
