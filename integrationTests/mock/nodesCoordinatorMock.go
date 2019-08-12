package mock

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type NodesCoordinatorMock struct {
	ComputeValidatorsGroupCalled  func([]byte) ([]sharding.Validator, error)
	GetValidatorsPublicKeysCalled func(randomness []byte) ([]string, error)
}

func (ncm NodesCoordinatorMock) ComputeValidatorsGroup(
	randomness []byte,
) (validatorsGroup []sharding.Validator, err error) {

	if ncm.ComputeValidatorsGroupCalled != nil {
		return ncm.ComputeValidatorsGroupCalled(randomness)
	}

	list := []sharding.Validator{}

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

func (ncm NodesCoordinatorMock) SetNodesPerShards(map[uint32][]sharding.Validator) error {
	return nil
}

func (ncm NodesCoordinatorMock) SetConsensusGroupSize(int) error {
	panic("implement me")
}

func (ncm NodesCoordinatorMock) GetSelectedPublicKeys(selection []byte) (publicKeys []string, err error) {
	panic("implement me")
}

func (ncm NodesCoordinatorMock) GetValidatorWithPublicKey(publicKey []byte) (sharding.Validator, uint32, error) {
	panic("implement me")
}
