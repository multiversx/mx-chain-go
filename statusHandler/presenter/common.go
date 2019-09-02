package presenter

func (psh *PresenterStatusHandler) getFromCacheAsUint64(metric string) uint64 {
	val, ok := psh.presenterMetrics.Load(metric)
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
	val, ok := psh.presenterMetrics.Load(metric)
	if !ok {
		return "[invalid key]"
	}

	valStr, ok := val.(string)
	if !ok {
		return "[not a string]"
	}

	return valStr
}
