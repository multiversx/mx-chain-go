package accumulator

import (
	"math/big"
)

type Accumulator interface {
	Accumulate(...[]byte) []*big.Int
	Verify(*big.Int, []byte) bool
}
