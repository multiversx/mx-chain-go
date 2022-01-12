package statusHandler

// StatusMetricsMap will return all metrics in a map
func (sm *statusMetrics) StatusMetricsMap() map[string]interface{} {
	sm.mutOperations.RLock()
	defer sm.mutOperations.RUnlock()

	statusMetricsMap := make(map[string]interface{})
	for key, value := range sm.nodeMetrics {
		statusMetricsMap[key] = value
	}

	return statusMetricsMap
}
