package core

import (
	"math/big"
	"strconv"
	"strings"
)

// MaxInt32 returns the maximum of two given numbers
func MaxInt32(a int32, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// MinInt returns the minimum of two given numbers
func MinInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt returns the maximum of two given numbers
func MaxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt32 returns the minimum of two given numbers
func MinInt32(a int32, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

// MaxUint32 returns the maximum of two given numbers
func MaxUint32(a uint32, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

// MinUint32 returns the minimum of two given numbers
func MinUint32(a uint32, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// MaxUint64 returns the maximum of two given numbers
func MaxUint64(a uint64, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// MinUint64 returns the minimum of two given numbers
func MinUint64(a uint64, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// MaxInt64 returns the maximum of two given numbers
func MaxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// MinInt64 returns the minimum of two given numbers
func MinInt64(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// MaxFloat64 returns the maximum of two given numbers
func MaxFloat64(a float64, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// GetApproximatePercentageOfValue returns the approximate percentage of value
// the approximation comes from floating point operations, which in case of large numbers
// has some loss in accuracy and can cause the result to be slightly lower or higher than the actual value
func GetApproximatePercentageOfValue(value *big.Int, percentage float64) *big.Int {
	x := new(big.Float).SetInt(value)
	y := big.NewFloat(percentage)

	z := new(big.Float).Mul(x, y)

	op := big.NewInt(0)
	result, _ := z.Int(op)

	return result
}

// GetIntTrimmedPercentageOfValue returns the exact percentage of value, that fits into the integer (with loss of division remainder)
func GetIntTrimmedPercentageOfValue(value *big.Int, percentage float64) *big.Int {
	x := big.NewInt(0).Set(value)
	percentageString := strconv.FormatFloat(percentage, 'f', -1, 64)
	exp, fra := splitExponentFraction(percentageString)
	concatExpFra := exp + fra
	concatBigInt, _ := big.NewInt(0).SetString(concatExpFra, 10)
	intMultiplier, _ := big.NewInt(0).SetString("1"+strings.Repeat("0", len(fra)), 10)
	x.Mul(x, concatBigInt)
	x.Div(x, intMultiplier)
	return x
}

// IsInRangeExclusive returns true if the provided value is in the given range, including the provided min and max values
func IsInRangeInclusive(value, min, max *big.Int) bool {
	return value.Cmp(min) >= 0 && value.Cmp(max) <= 0
}

// IsInRangeInclusiveFloat64 returns true if the provided value is in the given range including the provided min and max values
func IsInRangeInclusiveFloat64(value, min, max float64) bool {
	return value >= min && value <= max
}

func splitExponentFraction(val string) (string, string) {
	split := strings.Split(val, ".")
	if len(split) == 2 {
		return split[0], split[1]
	}
	return val, ""
}

// SafeMul returns multiplication results of two uint64 values into a big int
func SafeMul(a uint64, b uint64) *big.Int {
	return big.NewInt(0).Mul(big.NewInt(0).SetUint64(a), big.NewInt(0).SetUint64(b))
}

// SafeSubUint64 performs subtraction on uint64 and returns an error if it overflows
func SafeSubUint64(a, b uint64) (uint64, error) {
	if a < b {
		return 0, ErrSubtractionOverflow
	}
	return a - b, nil
}

// SafeAddUint64 performs addition on uint64 and returns an error if the addition overflows
func SafeAddUint64(a, b uint64) (uint64, error) {
	s := a + b
	if s >= a && s >= b {
		return s, nil
	}
	return 0, ErrAdditionOverflow
}
