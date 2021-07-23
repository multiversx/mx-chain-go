package presenter

import (
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/common"
)

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

func (psh *PresenterStatusHandler) getBigIntFromStringMetric(metric string) *big.Int {
	stringValue := psh.getFromCacheAsString(metric)
	bigIntValue, ok := big.NewInt(0).SetString(stringValue, 10)
	if !ok {
		return big.NewInt(0)
	}

	return bigIntValue
}

func areEqualWithZero(parameters ...uint64) bool {
	for _, param := range parameters {
		if param == 0 {
			return true
		}
	}

	return false
}

func (psh *PresenterStatusHandler) computeChanceToBeInConsensus() float64 {
	consensusGroupSize := psh.getFromCacheAsUint64(common.MetricConsensusGroupSize)
	numValidators := psh.getFromCacheAsUint64(common.MetricNumValidators)
	isChanceZero := areEqualWithZero(consensusGroupSize, numValidators)
	if isChanceZero {
		return 0
	}

	return float64(consensusGroupSize) / float64(numValidators)
}

func (psh *PresenterStatusHandler) computeRoundsPerHourAccordingToHitRate() float64 {
	totalBlocks := psh.GetProbableHighestNonce()
	rounds := psh.GetCurrentRound()
	roundDuration := psh.GetRoundTime()
	secondsInAnHour := uint64(3600)
	isRoundsPerHourZero := areEqualWithZero(totalBlocks, rounds, roundDuration)
	if isRoundsPerHourZero {
		return 0
	}

	hitRate := float64(totalBlocks) / float64(rounds)
	roundsPerHour := float64(secondsInAnHour) / float64(roundDuration)
	return hitRate * roundsPerHour
}

func (psh *PresenterStatusHandler) computeRewardsInErd() *big.Float {
	rewardsValue := psh.getBigIntFromStringMetric(common.MetricRewardsValue)
	denomination := psh.getFromCacheAsUint64(common.MetricDenomination)
	denominationCoefficientFloat := 1.0
	if denomination > 0 {
		denominationCoefficientFloat /= math.Pow10(int(denomination))
	}

	denominationCoefficient := big.NewFloat(denominationCoefficientFloat)

	if rewardsValue.Cmp(big.NewInt(0)) <= 0 {
		return big.NewFloat(0)
	}

	rewardsInErd := big.NewFloat(0).Mul(big.NewFloat(0).SetInt(rewardsValue), denominationCoefficient)
	return rewardsInErd
}
