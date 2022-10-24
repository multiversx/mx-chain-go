package statusHandler

// StatusMetricsHandler is the interface that defines what a node details handler/provider should do
type StatusMetricsHandler interface {
	StatusMetricsMapWithoutP2P() (map[string]interface{}, error)
	StatusP2pMetricsMap() (map[string]interface{}, error)
	StatusMetricsWithoutP2PPrometheusString() (string, error)
	EconomicsMetrics() (map[string]interface{}, error)
	ConfigMetrics() (map[string]interface{}, error)
	EnableEpochsMetrics() (map[string]interface{}, error)
	NetworkMetrics() (map[string]interface{}, error)
	RatingsMetrics() (map[string]interface{}, error)
	IsInterfaceNil() bool
}
