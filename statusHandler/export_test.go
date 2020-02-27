package statusHandler

// StatusMetricsMap will return all metrics in a map
func (sm *statusMetrics) StatusMetricsMap() map[string]interface{} {
	statusMetricsMap := make(map[string]interface{})
	sm.nodeMetrics.Range(func(key, value interface{}) bool {
		statusMetricsMap[key.(string)] = value
		return true
	})

	return statusMetricsMap
}
