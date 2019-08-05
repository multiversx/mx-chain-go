package mock

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NodesCoordinator defines the behaviour of a struct able to do validator group selection
type NodesCoordinatorMock struct {
	Validators                    map[uint32][]sharding.Validator
	ShardConsensusSize            uint32
	MetaConsensusSize             uint32
	ShardId                       uint32
	NbShards                      uint32
	GetSelectedPublicKeysCalled   func(selection []byte) (publicKeys []string, err error)
	GetValidatorsPublicKeysCalled func(randomness []byte) ([]string, error)
	LoadNodesPerShardsCalled      func(nodes map[uint32][]sharding.Validator) error
	ComputeValidatorsGroupCalled  func(randomness []byte) (validatorsGroup []sharding.Validator, err error)
}

func NewNodesCoordinatorMock() *NodesCoordinatorMock {
	return &NodesCoordinatorMock{
		ShardConsensusSize: 1,
		MetaConsensusSize:  1,
		ShardId:            0,
		NbShards:           1,
		Validators:         make(map[uint32][]sharding.Validator),
	}
}

func (ncm *NodesCoordinatorMock) GetSelectedPublicKeys(selection []byte) (publicKeys []string, err error) {
	if ncm.GetSelectedPublicKeysCalled != nil {
		return ncm.GetSelectedPublicKeysCalled(selection)
	}

	if len(ncm.Validators) == 0 {
		return nil, sharding.ErrNilInputNodesMap
	}

	pubKeys := make([]string, 0)

	for _, v := range ncm.Validators[ncm.ShardId] {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

func (ncm *NodesCoordinatorMock) GetValidatorsPublicKeys(randomness []byte) ([]string, error) {
	if ncm.GetValidatorsPublicKeysCalled != nil {
		return ncm.GetValidatorsPublicKeysCalled(randomness)
	}

	validators, err := ncm.ComputeValidatorsGroup(randomness)
	if err != nil {
		return nil, err
	}

	valGrStr := make([]string, 0)

	for _, v := range validators {
		valGrStr = append(valGrStr, string(v.PubKey()))
	}

	return valGrStr, nil
}

func (ncm *NodesCoordinatorMock) SetNodesPerShards(nodes map[uint32][]sharding.Validator) error {
	if ncm.LoadNodesPerShardsCalled != nil {
		return ncm.LoadNodesPerShardsCalled(nodes)
	}

	if nodes == nil {
		return sharding.ErrNilInputNodesMap
	}

	ncm.Validators = nodes

	return nil
}

func (ncm *NodesCoordinatorMock) ComputeValidatorsGroup(randomess []byte) ([]sharding.Validator, error) {
	var consensusSize uint32

	if ncm.ComputeValidatorsGroupCalled != nil {
		return ncm.ComputeValidatorsGroupCalled(randomess)
	}

	if ncm.ShardId == sharding.MetachainShardId {
		consensusSize = ncm.MetaConsensusSize
	} else {
		consensusSize = ncm.ShardConsensusSize
	}

	if randomess == nil {
		return nil, sharding.ErrNilRandomness
	}

	validatorsGroup := make([]sharding.Validator, 0)

	for i := uint32(0); i < consensusSize; i++ {
		validatorsGroup = append(validatorsGroup, ncm.Validators[ncm.ShardId][i])
	}

	return validatorsGroup, nil
}
