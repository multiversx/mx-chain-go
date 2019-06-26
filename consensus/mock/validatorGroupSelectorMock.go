package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/consensus"
)

type ValidatorGroupSelectorMock struct {
	ComputeValidatorsGroupCalled          func([]byte) ([]consensus.Validator, error)
	GetSelectedValidatorsPublicKeysCalled func(randomness []byte, bitmap []byte) ([]string, error)
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

func (vgsm ValidatorGroupSelectorMock) GetSelectedValidatorsPublicKeys(randomness []byte, bitmap []byte) ([]string, error) {
	if vgsm.GetSelectedValidatorsPublicKeysCalled != nil {
		return vgsm.GetSelectedValidatorsPublicKeysCalled(randomness, bitmap)
	}

	validators, err := vgsm.ComputeValidatorsGroup(randomness)

	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for i, v := range validators {
		isSelected := (bitmap[i/8] & (1 << (uint16(i) % 8))) != 0
		if !isSelected {
			continue
		}

		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

func (vgsm ValidatorGroupSelectorMock) ConsensusGroupSize() int {
	panic("implement me")
}

func (vgsm ValidatorGroupSelectorMock) LoadNodesPerShards(map[uint32][]consensus.Validator) error {
	return nil
}

func (vgsm ValidatorGroupSelectorMock) SetConsensusGroupSize(int) error {
	panic("implement me")
}

func (vgsm ValidatorGroupSelectorMock) GetSelectedPublicKeys(selection []byte) (publicKeys []string, err error) {
	panic("implement me")
}
