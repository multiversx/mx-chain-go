package mock

// StatusMetricsStub -
type StatusMetricsStub struct {
	StatusMetricsMapWithoutP2PCalled              func() map[string]interface{}
	StatusP2pMetricsMapCalled                     func() map[string]interface{}
	ConfigMetricsCalled                           func() map[string]interface{}
	EnableEpochMetricsCalled                      func() map[string]interface{}
	NetworkMetricsCalled                          func() map[string]interface{}
	EconomicsMetricsCalled                        func() map[string]interface{}
	StatusMetricsWithoutP2PPrometheusStringCalled func() string
}

// StatusMetricsWithoutP2PPrometheusString -
func (sms *StatusMetricsStub) StatusMetricsWithoutP2PPrometheusString() string {
	if sms.StatusMetricsWithoutP2PPrometheusStringCalled != nil {
		return sms.StatusMetricsWithoutP2PPrometheusStringCalled()
	}

	return "metric 10"
}

// ConfigMetrics -
func (sms *StatusMetricsStub) ConfigMetrics() map[string]interface{} {
	return sms.ConfigMetricsCalled()
}

// EnableEpochsMetrics
func (sms *StatusMetricsStub) EnableEpochsMetrics() map[string]interface{} {
	return sms.EnableEpochMetricsCalled()
}

// NetworkMetrics -
func (sms *StatusMetricsStub) NetworkMetrics() map[string]interface{} {
	return sms.NetworkMetricsCalled()
}

// EconomicsMetrics -
func (sms *StatusMetricsStub) EconomicsMetrics() map[string]interface{} {
	return sms.EconomicsMetricsCalled()
}

// StatusMetricsMapWithoutP2P -
func (sms *StatusMetricsStub) StatusMetricsMapWithoutP2P() map[string]interface{} {
	return sms.StatusMetricsMapWithoutP2PCalled()
}

// StatusP2pMetricsMap -
func (sms *StatusMetricsStub) StatusP2pMetricsMap() map[string]interface{} {
	return sms.StatusP2pMetricsMapCalled()
}

// IsInterfaceNil -
func (sms *StatusMetricsStub) IsInterfaceNil() bool {
	return sms == nil
}
