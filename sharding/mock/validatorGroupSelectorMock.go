package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/sharding"
)

type ValidatorGroupSelectorMock struct {
	ComputeValidatorsGroupCalled  func([]byte) ([]sharding.Validator, error)
	GetValidatorsPublicKeysCalled func(randomness []byte) ([]string, error)
}

func (vgsm ValidatorGroupSelectorMock) ComputeValidatorsGroup(randomness []byte) (validatorsGroup []sharding.Validator, err error) {
	if vgsm.ComputeValidatorsGroupCalled != nil {
		return vgsm.ComputeValidatorsGroupCalled(randomness)
	}

	list := []sharding.Validator{
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

func (vgsm ValidatorGroupSelectorMock) GetValidatorsPublicKeys(randomness []byte) ([]string, error) {
	if vgsm.GetValidatorsPublicKeysCalled != nil {
		return vgsm.GetValidatorsPublicKeysCalled(randomness)
	}

	validators, err := vgsm.ComputeValidatorsGroup(randomness)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for _, v := range validators {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

func (vgsm ValidatorGroupSelectorMock) ConsensusGroupSize() int {
	panic("implement me")
}

func (vgsm ValidatorGroupSelectorMock) LoadNodesPerShards(map[uint32][]sharding.Validator) error {
	return nil
}

func (vgsm ValidatorGroupSelectorMock) SetConsensusGroupSize(int) error {
	panic("implement me")
}

func (vgsm ValidatorGroupSelectorMock) GetSelectedPublicKeys(selection []byte) (publicKeys []string, err error) {
	panic("implement me")
}
