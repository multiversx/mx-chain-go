package presenter

const metricNotAvailable = "N/A"

func (psh *PresenterStatusHandler) getFromCacheAsUint64(metric string) uint64 {
	psh.mutPresenterMap.RLock()
	defer psh.mutPresenterMap.RUnlock()

	val, ok := psh.presenterMetrics[metric]
	if !ok {
		return 0
	}

	valUint64, ok := val.(uint64)
	if !ok {
		return 0
	}

	return valUint64
}

func (psh *PresenterStatusHandler) getFromCacheAsString(metric string) string {
	psh.mutPresenterMap.RLock()
	defer psh.mutPresenterMap.RUnlock()

	val, ok := psh.presenterMetrics[metric]
	if !ok {
		return metricNotAvailable
	}

	valStr, ok := val.(string)
	if !ok {
		return metricNotAvailable
	}

	return valStr
}
