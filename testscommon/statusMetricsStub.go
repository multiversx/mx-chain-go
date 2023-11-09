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
	BootstrapMetricsCalled                        func() (map[string]interface{}, error)
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
	if sms.ConfigMetricsCalled != nil {
		return sms.ConfigMetricsCalled()
	}
	return baseReturnValues()
}

// NetworkMetrics -
func (sms *StatusMetricsStub) NetworkMetrics() (map[string]interface{}, error) {
	if sms.NetworkMetricsCalled != nil {
		return sms.NetworkMetricsCalled()
	}
	return baseReturnValues()
}

// EconomicsMetrics -
func (sms *StatusMetricsStub) EconomicsMetrics() (map[string]interface{}, error) {
	if sms.EconomicsMetricsCalled != nil {
		return sms.EconomicsMetricsCalled()
	}
	return baseReturnValues()
}

// StatusMetricsMapWithoutP2P -
func (sms *StatusMetricsStub) StatusMetricsMapWithoutP2P() (map[string]interface{}, error) {
	if sms.StatusMetricsMapWithoutP2PCalled != nil {
		return sms.StatusMetricsMapWithoutP2PCalled()
	}
	return baseReturnValues()
}

// StatusP2pMetricsMap -
func (sms *StatusMetricsStub) StatusP2pMetricsMap() (map[string]interface{}, error) {
	if sms.StatusP2pMetricsMapCalled != nil {
		return sms.StatusP2pMetricsMapCalled()
	}
	return baseReturnValues()
}

// EnableEpochsMetrics -
func (sms *StatusMetricsStub) EnableEpochsMetrics() (map[string]interface{}, error) {
	if sms.EnableEpochsMetricsCalled != nil {
		return sms.EnableEpochsMetricsCalled()
	}
	return baseReturnValues()
}

// RatingsMetrics -
func (sms *StatusMetricsStub) RatingsMetrics() (map[string]interface{}, error) {
	if sms.RatingsMetricsCalled != nil {
		return sms.RatingsMetricsCalled()
	}
	return baseReturnValues()
}

// BootstrapMetrics -
func (sms *StatusMetricsStub) BootstrapMetrics() (map[string]interface{}, error) {
	if sms.BootstrapMetricsCalled != nil {
		return sms.BootstrapMetricsCalled()
	}
	return baseReturnValues()
}

func baseReturnValues() (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

// IsInterfaceNil -
func (sms *StatusMetricsStub) IsInterfaceNil() bool {
	return sms == nil
}
