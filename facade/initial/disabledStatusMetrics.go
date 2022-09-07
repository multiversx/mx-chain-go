package initial

// disabledStatusMetricsHandler represents a disabled implementation of the StatusMetricsHandler interface
type disabledStatusMetricsHandler struct {
}

// NewDisabledStatusMetricsHandler returns a new instance of disabledStatusMetricsHandler
func NewDisabledStatusMetricsHandler() *disabledStatusMetricsHandler {
	return &disabledStatusMetricsHandler{}
}

// StatusMetricsMapWithoutP2P returns an empty map and the error which specifies that the node is starting
func (d *disabledStatusMetricsHandler) StatusMetricsMapWithoutP2P() (map[string]interface{}, error) {
	return getReturnValues()
}

// StatusP2pMetricsMap returns an empty map and the error which specifies that the node is starting
func (d *disabledStatusMetricsHandler) StatusP2pMetricsMap() (map[string]interface{}, error) {
	return getReturnValues()
}

// StatusMetricsWithoutP2PPrometheusString returns an empty string and the error which specifies that the node is starting
func (d *disabledStatusMetricsHandler) StatusMetricsWithoutP2PPrometheusString() (string, error) {
	return "", errNodeStarting
}

// EconomicsMetrics returns an empty map and the error which specifies that the node is starting
func (d *disabledStatusMetricsHandler) EconomicsMetrics() (map[string]interface{}, error) {
	return getReturnValues()
}

// ConfigMetrics returns an empty map and the error which specifies that the node is starting
func (d *disabledStatusMetricsHandler) ConfigMetrics() (map[string]interface{}, error) {
	return getReturnValues()
}

//EnableEpochsMetrics returns an empty map and the error which specifies that the node is starting
func (d *disabledStatusMetricsHandler) EnableEpochsMetrics() (map[string]interface{}, error) {
	return getReturnValues()
}

// NetworkMetrics returns an empty map and the error which specifies that the node is starting
func (d *disabledStatusMetricsHandler) NetworkMetrics() (map[string]interface{}, error) {
	return getReturnValues()
}

// RatingsMetrics returns an empty map and the error which specifies that the node is starting
func (d *disabledStatusMetricsHandler) RatingsMetrics() (map[string]interface{}, error) {
	return getReturnValues()
}

func getReturnValues() (map[string]interface{}, error) {
	return map[string]interface{}{}, errNodeStarting
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledStatusMetricsHandler) IsInterfaceNil() bool {
	return d == nil
}
