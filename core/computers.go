package core

import "math/big"

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

// GetPercentageOfValue returns the percentage part of the value
func GetPercentageOfValue(value *big.Int, percentage float64) *big.Int {
	x := new(big.Float).SetInt(value)
	y := big.NewFloat(percentage)

	z := new(big.Float).Mul(x, y)

	op := big.NewInt(0)
	result, _ := z.Int(op)

	return result
}
