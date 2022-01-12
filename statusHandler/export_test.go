package statusHandler

// StatusMetricsMap will return all metrics in a map
func (sm *statusMetrics) StatusMetricsMap() map[string]interface{} {
	return sm.getMetricsWithKeyFilterMutexProtected(func(_ string) bool {
		return true
	})
}
