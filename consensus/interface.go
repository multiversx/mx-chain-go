package consensus

import (
	"math/big"
)

// Validator defines what a consensus validator implementation should do.
type Validator interface {
	Stake() big.Int
	Rating() int32
	PubKey() []byte
}

// ValidatorGroupSelector defines the behaviour of a struct able to do validator group selection
type ValidatorGroupSelector interface {
	LoadEligibleList(eligibleList []Validator) error
	ComputeValidatorsGroup(randomness []byte) (validatorsGroup []Validator, err error)
	ConsensusGroupSize() int
	SetConsensusGroupSize(int) error
}
