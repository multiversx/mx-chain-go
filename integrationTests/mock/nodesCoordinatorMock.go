package mock

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type NodesCoordinatorMock struct {
	ComputeValidatorsGroupCalled  func(randomness []byte, round uint64, shardId uint32) ([]sharding.Validator, error)
	GetValidatorsPublicKeysCalled func(randomness []byte, round uint64, shardId uint32) ([]string, error)
}

func (ncm NodesCoordinatorMock) ComputeValidatorsGroup(
	randomness []byte,
	round uint64,
	shardId uint32,
) (validatorsGroup []sharding.Validator, err error) {

	if ncm.ComputeValidatorsGroupCalled != nil {
		return ncm.ComputeValidatorsGroupCalled(randomness, round, shardId)
	}

	list := []sharding.Validator{}

	return list, nil
}

func (ncm NodesCoordinatorMock) GetValidatorsPublicKeys(
	randomness []byte,
	round uint64,
	shardId uint32,
) ([]string, error) {
	if ncm.GetValidatorsPublicKeysCalled != nil {
		return ncm.GetValidatorsPublicKeysCalled(randomness, round, shardId)
	}

	validators, err := ncm.ComputeValidatorsGroup(randomness, round, shardId)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for _, v := range validators {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

func (ncm NodesCoordinatorMock) SetNodesPerShards(map[uint32][]sharding.Validator) error {
	return nil
}

func (ncm NodesCoordinatorMock) GetSelectedPublicKeys(selection []byte, shardId uint32) (publicKeys []string, err error) {
	panic("implement me")
}

func (ncm NodesCoordinatorMock) GetValidatorWithPublicKey(publicKey []byte) (sharding.Validator, uint32, error) {
	panic("implement me")
}
