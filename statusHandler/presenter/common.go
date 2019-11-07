package presenter

import (
	"math/big"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/core"
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

func (psh *PresenterStatusHandler) getBigFloatFromStringMetric(metric string) *big.Float {
	stringValue := psh.getFromCacheAsString(metric)
	bigFloatValue, ok := big.NewFloat(0).SetString(stringValue)
	if !ok {
		return big.NewFloat(0)
	}

	return bigFloatValue
}

func areEqualsWithZero(parameters ...uint64) bool {
	for _, param := range parameters {
		if param == 0 {
			return true
		}
	}

	return false
}

func (psh *PresenterStatusHandler) computeChanceToBeInConsensus() float64 {
	consensusGroupSize := psh.getFromCacheAsUint64(core.MetricConsensusGroupSize)
	numValidators := psh.getFromCacheAsUint64(core.MetricNumValidators)
	areEqualsWithZero := areEqualsWithZero(consensusGroupSize, numValidators)
	if areEqualsWithZero {
		return 0
	}

	return float64(consensusGroupSize) / float64(numValidators)
}

func (psh *PresenterStatusHandler) computeRoundsPerHourAccordingToHitRate() float64 {
	totalBlocks := psh.GetProbableHighestNonce()
	rounds := psh.GetCurrentRound()
	roundDuration := psh.GetRoundTime()
	secondsInAnHour := uint64(3600)
	areEqualsWithZero := areEqualsWithZero(totalBlocks, rounds, roundDuration)
	if areEqualsWithZero {
		return 0
	}

	hitRate := float64(totalBlocks) / float64(rounds)
	roundsPerHour := float64(secondsInAnHour) / float64(roundDuration)
	return hitRate * roundsPerHour
}

func (psh *PresenterStatusHandler) computeRewardsInErd() *big.Float {
	rewardsValue := psh.getBigIntFromStringMetric(core.MetricRewardsValue)
	denominationCoefficient := psh.getBigFloatFromStringMetric(core.MetricDenominationCoefficient)
	if rewardsValue.Cmp(big.NewInt(0)) <= 0 {
		return big.NewFloat(0)
	}

	rewardsInErd := big.NewFloat(0).Mul(big.NewFloat(0).SetInt(rewardsValue), denominationCoefficient)
	return rewardsInErd
}
