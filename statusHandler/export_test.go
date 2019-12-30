package statusHandler

// StatusMetricsMap will return all metrics in a map
func (nd *statusMetrics) StatusMetricsMap() map[string]interface{} {
	statusMetricsMap := make(map[string]interface{})
	nd.nodeMetrics.Range(func(key, value interface{}) bool {
		statusMetricsMap[key.(string)] = value
		return true
	})

	return statusMetricsMap
}
