package mock

import "github.com/ElrondNetwork/elrond-go/core"

// StatusMetricsStub -
type StatusMetricsStub struct {
	StatusMetricsMapWithoutP2PCalled func() map[string]interface{}
	StatusP2pMetricsMapCalled        func() map[string]interface{}
	EpochMetricsCalled               func() map[string]interface{}
	ConfigMetricsCalled              func() map[string]interface{}
	NetworkMetricsCalled             func() map[string]interface{}
}

// ConfigMetrics -
func (sms *StatusMetricsStub) ConfigMetrics() map[string]interface{} {
	return sms.ConfigMetricsCalled()
}

// NetworkMetrics -
func (sms *StatusMetricsStub) NetworkMetrics() map[string]interface{} {
	return sms.NetworkMetricsCalled()
}

// StatusMetricsMapWithoutP2P -
func (sms *StatusMetricsStub) StatusMetricsMapWithoutP2P() map[string]interface{} {
	return sms.StatusMetricsMapWithoutP2PCalled()
}

// StatusP2pMetricsMap -
func (sms *StatusMetricsStub) StatusP2pMetricsMap() map[string]interface{} {
	return sms.StatusP2pMetricsMapCalled()
}

// EpochMetrics -
func (sms *StatusMetricsStub) EpochMetrics() map[string]interface{} {
	if sms.EpochMetricsCalled != nil {
		return sms.EpochMetricsCalled()
	}

	return map[string]interface{}{
		core.MetricEpochNumber: 37,
	}
}

// IsInterfaceNil -
func (sms *StatusMetricsStub) IsInterfaceNil() bool {
	return sms == nil
}
