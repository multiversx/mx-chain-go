package txcache

import (
	"math"
)

var _ scoreComputer = (*defaultScoreComputer)(nil)
var _ scoreComputer = (*disabledScoreComputer)(nil)

// TODO: the score formula should not be sensitive to the order of magnitude of the minGasPrice.
// TODO (continued): We should not rely on any order of magnitude known a priori.
// TODO (continued): The score formula should work even if minGasPrice = 0.
type senderScoreParams struct {
	count uint64
	// Size is in bytes
	size uint64
	// Fee is in nano ERD
	fee uint64
	gas uint64
}

type defaultScoreComputer struct {
	// Price is in nano ERD
	minGasPrice uint32
}

func newDefaultScoreComputer(minGasPrice uint32) *defaultScoreComputer {
	return &defaultScoreComputer{
		minGasPrice: minGasPrice,
	}
}

// computeScore computes the score of the sender, as an integer 0-100
func (computer *defaultScoreComputer) computeScore(scoreParams senderScoreParams) uint32 {
	rawScore := computer.computeRawScore(scoreParams)
	truncatedScore := uint32(rawScore)
	return truncatedScore
}

// score for a sender is defined as follows:
//
//                           (PPUAvg / PPUMin)^3
// rawScore = ------------------------------------------------
//            [ln(txCount^2 + 1) + 1] * [ln(txSize^2 + 1) + 1]
//
//                              1
// asymptoticScore = [(------------------) - 0.5] * 2
//                     1 + exp(-rawScore)
//
// For asymptoticScore, see (https://en.wikipedia.org/wiki/Logistic_function)
//
// Where:
//  - PPUAvg: average gas points (fee) per processing unit, in nano ERD
//  - PPUMin: minimum gas points (fee) per processing unit (given by economics.toml), in nano ERD
//  - txCount: number of transactions
//  - txSize: size of transactions, in kB (1000 bytes)
//
// TODO (optimization): switch to integer operations (as opposed to float operations).
func (computer *defaultScoreComputer) computeRawScore(params senderScoreParams) float64 {
	allParamsDefined := params.fee > 0 && params.gas > 0 && params.size > 0 && params.count > 0
	if !allParamsDefined {
		return 0
	}

	PPUMin := float64(computer.minGasPrice)
	PPUAvg := float64(params.fee) / float64(params.gas)
	PPUScore := math.Pow(PPUAvg/PPUMin, 3)

	countPow2 := float64(params.count) * float64(params.count)
	countScore := math.Log(countPow2+1) + 1

	// We use size in ~kB
	const bytesInKB = 1000
	size := float64(params.size) / bytesInKB
	sizePow2 := size * size
	sizeScore := math.Log(sizePow2+1) + 1

	rawScore := PPUScore / countScore / sizeScore
	// We apply the logistic function,
	// and then subtract 0.5, since we only deal with positive scores,
	// and then we multiply by 2, to have full [0..1] range.
	asymptoticScore := (1/(1+math.Exp(-rawScore)) - 0.5) * 2
	score := asymptoticScore * float64(numberOfScoreChunks)
	return score
}

type disabledScoreComputer struct {
}

func (computer *disabledScoreComputer) computeScore(scoreParams senderScoreParams) uint32 {
	return 0
}
