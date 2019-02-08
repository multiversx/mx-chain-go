package groupSelectors

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/validators"
)

// ValidatorGroupSelector defines the behaviour of a struct able to do validator group selection
type ValidatorGroupSelector interface {
	LoadEligibleList(eligibleList []validators.Validator) error
	ComputeValidatorsGroup(randomness []byte) (validatorsGroup []validators.Validator, err error)
	ConsensusGroupSize() int
	SetConsensusGroupSize(int) error
}
