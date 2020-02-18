package mock

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
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
	GetSelectedPublicKeysCalled         func(selection []byte, shardId uint32, epoch uint32) (publicKeys []string, err error)
	GetValidatorsPublicKeysCalled       func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	GetValidatorsRewardsAddressesCalled func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	SetNodesPerShardsCalled             func(nodes map[uint32][]sharding.Validator, waiting map[uint32][]sharding.Validator, epoch uint32) error
	ComputeValidatorsGroupCalled        func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []sharding.Validator, err error)
	GetValidatorWithPublicKeyCalled     func(publicKey []byte, epoch uint32) (validator sharding.Validator, shardId uint32, err error)
}

// NewNodesCoordinatorMock -
func NewNodesCoordinatorMock() *NodesCoordinatorMock {
	nbShards := uint32(1)
	nodesPerShard := 2
	validatorsMap := make(map[uint32][]sharding.Validator)

	shards := make([]uint32, nbShards+1)
	for i := uint32(0); i < nbShards; i++ {
		shards[i] = i
	}
	shards[nbShards] = core.MetachainShardId

	for _, sh := range shards {
		validatorsList := make([]sharding.Validator, nodesPerShard)
		for v := 0; v < nodesPerShard; v++ {
			validatorsList[v], _ = sharding.NewValidator(
				[]byte(fmt.Sprintf("pubKey%d%d", sh, v)),
				[]byte(fmt.Sprintf("address%d%d", sh, v)),
			)
		}
		validatorsMap[sh] = validatorsList
	}

	return &NodesCoordinatorMock{
		ShardConsensusSize: 1,
		MetaConsensusSize:  1,
		ShardId:            0,
		NbShards:           nbShards,
		Validators:         validatorsMap,
	}
}

// GetValidatorsIndexes -
func (ncm *NodesCoordinatorMock) GetValidatorsIndexes(_ []string, _ uint32) ([]uint64, error) {
	return nil, nil
}

// GetAllValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetAllEligibleValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// GetSelectedPublicKeys -
func (ncm *NodesCoordinatorMock) GetSelectedPublicKeys(selection []byte, shardId uint32, epoch uint32) (publicKeys []string, err error) {
	if ncm.GetSelectedPublicKeysCalled != nil {
		return ncm.GetSelectedPublicKeysCalled(selection, shardId, epoch)
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

// GetConsensusValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetConsensusValidatorsPublicKeys(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) ([]string, error) {
	if ncm.GetValidatorsPublicKeysCalled != nil {
		return ncm.GetValidatorsPublicKeysCalled(randomness, round, shardId, epoch)
	}

	validators, err := ncm.ComputeConsensusGroup(randomness, round, shardId, epoch)
	if err != nil {
		return nil, err
	}

	valGrStr := make([]string, 0)

	for _, v := range validators {
		valGrStr = append(valGrStr, string(v.PubKey()))
	}

	return valGrStr, nil
}

// GetConsensusValidatorsRewardsAddresses -
func (ncm *NodesCoordinatorMock) GetConsensusValidatorsRewardsAddresses(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) ([]string, error) {
	if ncm.GetValidatorsPublicKeysCalled != nil {
		return ncm.GetValidatorsRewardsAddressesCalled(randomness, round, shardId, epoch)
	}

	validators, err := ncm.ComputeConsensusGroup(randomness, round, shardId, epoch)
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
func (ncm *NodesCoordinatorMock) SetNodesPerShards(
	eligible map[uint32][]sharding.Validator,
	waiting map[uint32][]sharding.Validator,
	epoch uint32,
) error {
	if ncm.SetNodesPerShardsCalled != nil {
		return ncm.SetNodesPerShardsCalled(eligible, waiting, epoch)
	}

	if eligible == nil {
		return sharding.ErrNilInputNodesMap
	}

	ncm.Validators = eligible

	return nil
}

// ComputeConsensusGroup -
func (ncm *NodesCoordinatorMock) ComputeConsensusGroup(
	randomess []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) ([]sharding.Validator, error) {
	var consensusSize uint32

	if ncm.ComputeValidatorsGroupCalled != nil {
		return ncm.ComputeValidatorsGroupCalled(randomess, round, shardId, epoch)
	}

	if ncm.ShardId == core.MetachainShardId {
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
func (ncm *NodesCoordinatorMock) GetValidatorWithPublicKey(publicKey []byte, epoch uint32) (sharding.Validator, uint32, error) {
	if ncm.GetValidatorWithPublicKeyCalled != nil {
		return ncm.GetValidatorWithPublicKeyCalled(publicKey, epoch)
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

// LoadState -
func (ncm *NodesCoordinatorMock) LoadState(_ []byte) error {
	return nil
}

// GetSavedStateKey -
func (ncm *NodesCoordinatorMock) GetSavedStateKey() []byte {
	return []byte("key")
}

// ShardIdForEpoch returns the nodesCoordinator configured ShardId for specified epoch if epoch configuration exists,
// otherwise error
func (ncm *NodesCoordinatorMock) ShardIdForEpoch(_ uint32) (uint32, error) {
	panic("not implemented")
}

// GetConsensusWhitelistedNodes return the whitelisted nodes allowed to send consensus messages, for each of the shards
func (ncm *NodesCoordinatorMock) GetConsensusWhitelistedNodes(
	_ uint32,
) (map[string]struct{}, error) {
	panic("not implemented")
}

// IsInterfaceNil -
func (ncm *NodesCoordinatorMock) IsInterfaceNil() bool {
	return ncm == nil
}
