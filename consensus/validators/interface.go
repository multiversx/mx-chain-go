package validators

import (
	"math/big"
)

// Validator defines what a consensus validator implementation should do.
type Validator interface {
	Stake() *big.Int
	Rating() int32
	PubKey() []byte
}
