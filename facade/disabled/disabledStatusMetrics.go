package disabled

const responseKey = "error"
const responseValue = "node is starting"

// disabledStatusMetricsHandler represents a disabled implementation of the StatusMetricsHandler interface
type disabledStatusMetricsHandler struct {
}

// NewDisabledStatusMetricsHandler returns a new instance of disabledStatusMetricsHandler
func NewDisabledStatusMetricsHandler() *disabledStatusMetricsHandler {
	return &disabledStatusMetricsHandler{}
}

// StatusMetricsMapWithoutP2P returns a default response map
func (d *disabledStatusMetricsHandler) StatusMetricsMapWithoutP2P() map[string]interface{} {
	return getReturnMap()
}

// StatusP2pMetricsMap returns a default response map
func (d *disabledStatusMetricsHandler) StatusP2pMetricsMap() map[string]interface{} {
	return getReturnMap()
}

// StatusMetricsWithoutP2PPrometheusString returns the message that signals that the node is starting
func (d *disabledStatusMetricsHandler) StatusMetricsWithoutP2PPrometheusString() string {
	return responseValue
}

// EconomicsMetrics returns a default response map
func (d *disabledStatusMetricsHandler) EconomicsMetrics() map[string]interface{} {
	return getReturnMap()
}

// ConfigMetrics returns a default response map
func (d *disabledStatusMetricsHandler) ConfigMetrics() map[string]interface{} {
	return getReturnMap()
}

// NetworkMetrics returns a default response map
func (d *disabledStatusMetricsHandler) NetworkMetrics() map[string]interface{} {
	return getReturnMap()
}

func getReturnMap() map[string]interface{} {
	return map[string]interface{}{
		responseKey: responseValue,
	}
}

// IsInterfaceNil returns true if there is nu value under the interface
func (d *disabledStatusMetricsHandler) IsInterfaceNil() bool {
	return d == nil
}
