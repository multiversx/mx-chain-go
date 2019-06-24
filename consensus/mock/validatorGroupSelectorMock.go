package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/consensus"
)

type ValidatorGroupSelectorMock struct {
	ComputeValidatorsGroupCalled func([]byte) ([]consensus.Validator, error)
}

func (vgsm ValidatorGroupSelectorMock) ComputeValidatorsGroup(randomness []byte) (validatorsGroup []consensus.Validator, err error) {
	if vgsm.ComputeValidatorsGroupCalled != nil {
		return vgsm.ComputeValidatorsGroupCalled(randomness)
	}

	list := []consensus.Validator{
		NewValidatorMock(big.NewInt(0), 0, []byte("A")),
		NewValidatorMock(big.NewInt(0), 0, []byte("B")),
		NewValidatorMock(big.NewInt(0), 0, []byte("C")),
		NewValidatorMock(big.NewInt(0), 0, []byte("D")),
		NewValidatorMock(big.NewInt(0), 0, []byte("E")),
		NewValidatorMock(big.NewInt(0), 0, []byte("F")),
		NewValidatorMock(big.NewInt(0), 0, []byte("G")),
		NewValidatorMock(big.NewInt(0), 0, []byte("H")),
		NewValidatorMock(big.NewInt(0), 0, []byte("I")),
	}

	return list, nil
}

func (vgsm ValidatorGroupSelectorMock) ConsensusGroupSize() int {
	panic("implement me")
}

func (vgsm ValidatorGroupSelectorMock) LoadEligibleList(eligibleList []consensus.Validator) error {
	return nil
}

func (vgsm ValidatorGroupSelectorMock) SetConsensusGroupSize(int) error {
	panic("implement me")
}

func (vgsm ValidatorGroupSelectorMock) GetSelectedPublicKeys(selection []byte) (publicKeys []string, err error) {
	panic("implement me")
}
