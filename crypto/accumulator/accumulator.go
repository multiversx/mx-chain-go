package accumulator

import (
	"math/big"
)

type Accumulator interface {
	HashToPrime([]byte) *big.Int
	Accumulate(...[]byte) []*big.Int
	Verify([]byte, *big.Int) bool
}
