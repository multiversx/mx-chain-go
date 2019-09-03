package accumulator

import (
	"math/big"
)

// Accumulator is an interface for a cryptographic accumulator that can accumulate data and
// verify if some data was added to the accumulator
type Accumulator interface {
	Accumulate(...[]byte) []*big.Int
	Verify(*big.Int, []byte) bool
	IsInterfaceNil() bool
}
