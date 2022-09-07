package testscommon

// StatusMetricsStub -
type StatusMetricsStub struct {
	StatusMetricsMapWithoutP2PCalled              func() (map[string]interface{}, error)
	StatusP2pMetricsMapCalled                     func() (map[string]interface{}, error)
	ConfigMetricsCalled                           func() (map[string]interface{}, error)
	NetworkMetricsCalled                          func() (map[string]interface{}, error)
	EconomicsMetricsCalled                        func() (map[string]interface{}, error)
	EnableEpochsMetricsCalled                     func() (map[string]interface{}, error)
	RatingsMetricsCalled                          func() (map[string]interface{}, error)
	StatusMetricsWithoutP2PPrometheusStringCalled func() (string, error)
}

// StatusMetricsWithoutP2PPrometheusString -
func (sms *StatusMetricsStub) StatusMetricsWithoutP2PPrometheusString() (string, error) {
	if sms.StatusMetricsWithoutP2PPrometheusStringCalled != nil {
		return sms.StatusMetricsWithoutP2PPrometheusStringCalled()
	}

	return "metric 10", nil
}

// ConfigMetrics -
func (sms *StatusMetricsStub) ConfigMetrics() (map[string]interface{}, error) {
	return sms.ConfigMetricsCalled()
}

// NetworkMetrics -
func (sms *StatusMetricsStub) NetworkMetrics() (map[string]interface{}, error) {
	return sms.NetworkMetricsCalled()
}

// EconomicsMetrics -
func (sms *StatusMetricsStub) EconomicsMetrics() (map[string]interface{}, error) {
	return sms.EconomicsMetricsCalled()
}

// StatusMetricsMapWithoutP2P -
func (sms *StatusMetricsStub) StatusMetricsMapWithoutP2P() (map[string]interface{}, error) {
	return sms.StatusMetricsMapWithoutP2PCalled()
}

// StatusP2pMetricsMap -
func (sms *StatusMetricsStub) StatusP2pMetricsMap() (map[string]interface{}, error) {
	return sms.StatusP2pMetricsMapCalled()
}

// EnableEpochsMetrics -
func (sms *StatusMetricsStub) EnableEpochsMetrics() (map[string]interface{}, error) {
	return sms.EnableEpochsMetricsCalled()
}

// RatingsConfig -
func (sms *StatusMetricsStub) RatingsMetrics() (map[string]interface{}, error) {
	return sms.RatingsMetricsCalled()
}

// IsInterfaceNil -
func (sms *StatusMetricsStub) IsInterfaceNil() bool {
	return sms == nil
}
