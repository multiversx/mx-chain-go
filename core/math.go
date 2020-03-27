package core

import "math/big"

func Pow(a *big.Float, e uint64) *big.Float {
	if e == 0 {
		return big.NewFloat(1)
	}
	result := big.NewFloat(0).Copy(a)
	for i := uint64(0); i < e-1; i++ {
		result = result.Mul(result, a)
	}
	return result
}
