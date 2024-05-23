package txcache

import (
	"math"
)

var _ scoreComputer = (*defaultScoreComputer)(nil)

// TODO (continued): The score formula should work even if minGasPrice = 0.
type senderScoreParams struct {
	count uint64
	// Fee score is normalized
	feeScore uint64
	gas      uint64
}

type defaultScoreComputer struct {
	txFeeHelper feeHelper
	ppuDivider  uint64
}

func newDefaultScoreComputer(txFeeHelper feeHelper) *defaultScoreComputer {
	ppuScoreDivider := txFeeHelper.minGasPriceFactor()
	ppuScoreDivider = ppuScoreDivider * ppuScoreDivider * ppuScoreDivider

	return &defaultScoreComputer{
		txFeeHelper: txFeeHelper,
		ppuDivider:  ppuScoreDivider,
	}
}

// computeScore computes the score of the sender, as an integer 0-100
func (computer *defaultScoreComputer) computeScore(scoreParams senderScoreParams) uint32 {
	rawScore := computer.computeRawScore(scoreParams)
	truncatedScore := uint32(rawScore)
	return truncatedScore
}

// TODO (optimization): switch to integer operations (as opposed to float operations).
func (computer *defaultScoreComputer) computeRawScore(params senderScoreParams) float64 {
	allParamsDefined := params.feeScore > 0 && params.gas > 0 && params.count > 0
	if !allParamsDefined {
		return 0
	}

	ppuMin := computer.txFeeHelper.minPricePerUnit()
	normalizedGas := params.gas >> computer.txFeeHelper.gasLimitShift()
	if normalizedGas == 0 {
		normalizedGas = 1
	}
	ppuAvg := params.feeScore / normalizedGas
	// (<< 3)^3 and >> 9 cancel each other; used to preserve a bit more resolution
	ppuRatio := ppuAvg << 3 / ppuMin
	ppuScore := ppuRatio * ppuRatio * ppuRatio >> 9
	ppuScoreAdjusted := float64(ppuScore) / float64(computer.ppuDivider)

	countPow2 := params.count * params.count
	countScore := math.Log(float64(countPow2)+1) + 1

	rawScore := ppuScoreAdjusted / countScore
	// We apply the logistic function,
	// and then subtract 0.5, since we only deal with positive scores,
	// and then we multiply by 2, to have full [0..1] range.
	asymptoticScore := (1/(1+math.Exp(-rawScore)) - 0.5) * 2
	score := asymptoticScore * float64(numberOfScoreChunks)
	return score
}
