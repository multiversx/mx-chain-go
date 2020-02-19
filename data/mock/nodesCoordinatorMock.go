package mock

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NodesCoordinatorMock defines the behaviour of a struct able to do validator group selection
type NodesCoordinatorMock struct {
	Validators                          map[uint32][]sharding.Validator
	ShardConsensusSize                  uint32
	MetaConsensusSize                   uint32
	ShardId                             uint32
	NbShards                            uint32
	GetOwnPublicKeyCalled               func() []byte
	GetSelectedPublicKeysCalled         func(selection []byte, shardId uint32) (publicKeys []string, err error)
	GetValidatorsPublicKeysCalled       func(randomness []byte, round uint64, shardId uint32) ([]string, error)
	GetValidatorsRewardsAddressesCalled func(randomness []byte, round uint64, shardId uint32) ([]string, error)
	LoadNodesPerShardsCalled            func(nodes map[uint32][]sharding.Validator) error
	ComputeValidatorsGroupCalled        func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error)
	GetValidatorWithPublicKeyCalled     func(publicKey []byte) (validator sharding.Validator, shardId uint32, err error)
}

// GetNumTotalEligible -
func (ncm *NodesCoordinatorMock) GetNumTotalEligible() uint64 {
	return 1
}

// GetValidatorsIndexes -
func (ncm *NodesCoordinatorMock) GetValidatorsIndexes(_ []string) []uint64 {
	return nil
}

// GetAllValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetAllValidatorsPublicKeys() map[uint32][][]byte {
	return nil
}

// GetSelectedPublicKeys -
func (ncm *NodesCoordinatorMock) GetSelectedPublicKeys(selection []byte, shardId uint32) (publicKeys []string, err error) {
	if ncm.GetSelectedPublicKeysCalled != nil {
		return ncm.GetSelectedPublicKeysCalled(selection, shardId)
	}

	if len(ncm.Validators) == 0 {
		return nil, sharding.ErrNilInputNodesMap
	}

	pubKeys := make([]string, 0)

	for _, v := range ncm.Validators[shardId] {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

// GetValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetValidatorsPublicKeys(
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

	valGrStr := make([]string, 0)

	for _, v := range validators {
		valGrStr = append(valGrStr, string(v.PubKey()))
	}

	return valGrStr, nil
}

// GetValidatorsRewardsAddresses -
func (ncm *NodesCoordinatorMock) GetValidatorsRewardsAddresses(
	randomness []byte,
	round uint64,
	shardId uint32,
) ([]string, error) {
	if ncm.GetValidatorsPublicKeysCalled != nil {
		return ncm.GetValidatorsRewardsAddressesCalled(randomness, round, shardId)
	}

	validators, err := ncm.ComputeValidatorsGroup(randomness, round, shardId)
	if err != nil {
		return nil, err
	}

	addresses := make([]string, 0)
	for _, v := range validators {
		addresses = append(addresses, string(v.Address()))
	}

	return addresses, nil
}

// SetNodesPerShards -
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

// ComputeValidatorsGroup -
func (ncm *NodesCoordinatorMock) ComputeValidatorsGroup(
	randomess []byte,
	round uint64,
	shardId uint32,
) ([]sharding.Validator, error) {
	var consensusSize uint32

	if ncm.ComputeValidatorsGroupCalled != nil {
		return ncm.ComputeValidatorsGroupCalled(randomess, round, shardId)
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
		validatorsGroup = append(validatorsGroup, ncm.Validators[shardId][i])
	}

	return validatorsGroup, nil
}

// ConsensusGroupSize -
func (ncm *NodesCoordinatorMock) ConsensusGroupSize(uint32) int {
	return 1
}

// GetValidatorWithPublicKey -
func (ncm *NodesCoordinatorMock) GetValidatorWithPublicKey(publicKey []byte) (sharding.Validator, uint32, error) {
	if ncm.GetValidatorWithPublicKeyCalled != nil {
		return ncm.GetValidatorWithPublicKeyCalled(publicKey)
	}

	if publicKey == nil {
		return nil, 0, sharding.ErrNilPubKey
	}

	for shardId, shardEligible := range ncm.Validators {
		for i := 0; i < len(shardEligible); i++ {
			if bytes.Equal(publicKey, shardEligible[i].PubKey()) {
				return shardEligible[i], shardId, nil
			}
		}
	}

	return nil, 0, sharding.ErrValidatorNotFound
}

// GetOwnPublicKey -
func (ncm *NodesCoordinatorMock) GetOwnPublicKey() []byte {
	if ncm.GetOwnPublicKeyCalled != nil {
		return ncm.GetOwnPublicKeyCalled()
	}

	return []byte("key")
}

// IsInterfaceNil -
func (ncm *NodesCoordinatorMock) IsInterfaceNil() bool {
	if ncm == nil {
		return true
	}
	return false
}
