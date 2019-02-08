package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/mock"
)

type ValidatorGroupSelectorMock struct{}

func (vgsm ValidatorGroupSelectorMock) ComputeValidatorsGroup(randomness []byte) (validatorsGroup []consensus.Validator, err error) {
	list := []consensus.Validator{
		mock.NewValidatorMock(big.NewInt(0), 0, []byte("A")),
		mock.NewValidatorMock(big.NewInt(0), 0, []byte("B")),
		mock.NewValidatorMock(big.NewInt(0), 0, []byte("C")),
		mock.NewValidatorMock(big.NewInt(0), 0, []byte("D")),
		mock.NewValidatorMock(big.NewInt(0), 0, []byte("E")),
		mock.NewValidatorMock(big.NewInt(0), 0, []byte("F")),
		mock.NewValidatorMock(big.NewInt(0), 0, []byte("G")),
		mock.NewValidatorMock(big.NewInt(0), 0, []byte("H")),
		mock.NewValidatorMock(big.NewInt(0), 0, []byte("I")),
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
