package presenter

import (
	"math/big"
	"strconv"
)

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

func (psh *PresenterStatusHandler) getFloatFromStringMetric(metric string) float64 {
	stringValue := psh.getFromCacheAsString(metric)
	floatMetric, err := strconv.ParseFloat(stringValue, 64)
	if err != nil {
		return 0.0
	}

	return floatMetric
}

func (psh *PresenterStatusHandler) getBigIntFromStringMetric(metric string) *big.Int {
	stringValue := psh.getFromCacheAsString(metric)
	bigIntValue, ok := big.NewInt(0).SetString(stringValue, 10)
	if !ok {
		return big.NewInt(0)
	}

	return bigIntValue
}
