package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/sharding"
)

type NodesCoordinatorMock struct {
	ComputeValidatorsGroupCalled  func([]byte) ([]sharding.Validator, error)
	GetValidatorsPublicKeysCalled func(randomness []byte) ([]string, error)
}

func (ncm NodesCoordinatorMock) ComputeValidatorsGroup(randomness []byte) (validatorsGroup []sharding.Validator, err error) {
	if ncm.ComputeValidatorsGroupCalled != nil {
		return ncm.ComputeValidatorsGroupCalled(randomness)
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

func (ncm NodesCoordinatorMock) GetValidatorsPublicKeys(randomness []byte) ([]string, error) {
	if ncm.GetValidatorsPublicKeysCalled != nil {
		return ncm.GetValidatorsPublicKeysCalled(randomness)
	}

	validators, err := ncm.ComputeValidatorsGroup(randomness)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for _, v := range validators {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

func (ncm NodesCoordinatorMock) ConsensusGroupSize() int {
	panic("implement me")
}

func (ncm NodesCoordinatorMock) LoadNodesPerShards(map[uint32][]sharding.Validator) error {
	return nil
}

func (ncm NodesCoordinatorMock) SetConsensusGroupSize(int) error {
	panic("implement me")
}

func (ncm NodesCoordinatorMock) GetSelectedPublicKeys(selection []byte) (publicKeys []string, err error) {
	panic("implement me")
}
