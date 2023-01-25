package initial

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/node/external"
)

type initialStatusMetricsProvider struct {
	realStatusMetricsProvider external.StatusMetricsHandler
}

// NewInitialStatusMetricsProvider returns a new instance of initial status metrics provider which should be used before the node is started
func NewInitialStatusMetricsProvider(statusMetricsProvider external.StatusMetricsHandler) (*initialStatusMetricsProvider, error) {
	if check.IfNil(statusMetricsProvider) {
		return nil, facade.ErrNilStatusMetrics
	}

	return &initialStatusMetricsProvider{
		realStatusMetricsProvider: statusMetricsProvider,
	}, nil
}

// BootstrapMetrics returns the metrics available during bootstrap
func (provider *initialStatusMetricsProvider) BootstrapMetrics() (map[string]interface{}, error) {
	return provider.realStatusMetricsProvider.BootstrapMetrics()
}

// StatusMetricsMapWithoutP2P returns an empty map and the error which specifies that the node is starting
func (provider *initialStatusMetricsProvider) StatusMetricsMapWithoutP2P() (map[string]interface{}, error) {
	return getEmptyReturnValues()
}

// StatusP2pMetricsMap returns an empty map and the error which specifies that the node is starting
func (provider *initialStatusMetricsProvider) StatusP2pMetricsMap() (map[string]interface{}, error) {
	return getEmptyReturnValues()
}

// StatusMetricsWithoutP2PPrometheusString returns an empty string and the error which specifies that the node is starting
func (provider *initialStatusMetricsProvider) StatusMetricsWithoutP2PPrometheusString() (string, error) {
	return "", errNodeStarting
}

// EconomicsMetrics returns an empty map and the error which specifies that the node is starting
func (provider *initialStatusMetricsProvider) EconomicsMetrics() (map[string]interface{}, error) {
	return getEmptyReturnValues()
}

// ConfigMetrics returns an empty map and the error which specifies that the node is starting
func (provider *initialStatusMetricsProvider) ConfigMetrics() (map[string]interface{}, error) {
	return getEmptyReturnValues()
}

// EnableEpochsMetrics returns an empty map and the error which specifies that the node is starting
func (provider *initialStatusMetricsProvider) EnableEpochsMetrics() (map[string]interface{}, error) {
	return getEmptyReturnValues()
}

// NetworkMetrics returns an empty map and the error which specifies that the node is starting
func (provider *initialStatusMetricsProvider) NetworkMetrics() (map[string]interface{}, error) {
	return getEmptyReturnValues()
}

// RatingsMetrics returns an empty map and the error which specifies that the node is starting
func (provider *initialStatusMetricsProvider) RatingsMetrics() (map[string]interface{}, error) {
	return getEmptyReturnValues()
}

// getEmptyReturnValues returns an empty map and the error which specifies that the node is starting
func getEmptyReturnValues() (map[string]interface{}, error) {
	return map[string]interface{}{}, errNodeStarting
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *initialStatusMetricsProvider) IsInterfaceNil() bool {
	return provider == nil
}
