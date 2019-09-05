package presenter

const invalidKey = "[invalid key]"
const invalidType = "[not a string]"

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
		return invalidKey
	}

	valStr, ok := val.(string)
	if !ok {
		return invalidType
	}

	return valStr
}
