package shardingMocks

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
)

// NodesCoordinatorMock defines the behaviour of a struct able to do validator group selection
type NodesCoordinatorMock struct {
	Validators                                  map[uint32][]nodesCoordinator.Validator
	ShardConsensusSize                          uint32
	MetaConsensusSize                           uint32
	ShardId                                     uint32
	NbShards                                    uint32
	GetSelectedPublicKeysCalled                 func(selection []byte, shardId uint32, epoch uint32) (publicKeys []string, err error)
	GetValidatorsPublicKeysCalled               func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	GetValidatorsRewardsAddressesCalled         func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	SetNodesPerShardsCalled                     func(nodes map[uint32][]nodesCoordinator.Validator, epoch uint32) error
	ComputeValidatorsGroupCalled                func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error)
	GetValidatorWithPublicKeyCalled             func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error)
	GetAllEligibleValidatorsPublicKeysCalled    func(epoch uint32) (map[uint32][][]byte, error)
	GetAllWaitingValidatorsPublicKeysCalled     func() (map[uint32][][]byte, error)
	ConsensusGroupSizeCalled                    func(uint32) int
	GetValidatorsIndexesCalled                  func(publicKeys []string, epoch uint32) ([]uint64, error)
	GetAllShuffledOutValidatorsPublicKeysCalled func(epoch uint32) (map[uint32][][]byte, error)
	GetNumTotalEligibleCalled                   func() uint64
}

// NewNodesCoordinatorMock -
func NewNodesCoordinatorMock() *NodesCoordinatorMock {
	nbShards := uint32(1)
	nodesPerShard := 2
	validatorsMap := make(map[uint32][]nodesCoordinator.Validator)

	for sh := uint32(0); sh < nbShards; sh++ {
		validatorsList := make([]nodesCoordinator.Validator, nodesPerShard)
		for v := 0; v < nodesPerShard; v++ {
			validatorsList[v], _ = nodesCoordinator.NewValidator(
				[]byte(fmt.Sprintf("pubKey%d%d", sh, v)),
				1,
				uint32(v),
			)
		}
		validatorsMap[sh] = validatorsList
	}

	validatorsList := make([]nodesCoordinator.Validator, nodesPerShard)
	for v := 0; v < nodesPerShard; v++ {
		validatorsList[v], _ = nodesCoordinator.NewValidator(
			[]byte(fmt.Sprintf("pubKey%d%d", core.MetachainShardId, v)),
			1,
			uint32(v),
		)
	}

	validatorsMap[core.MetachainShardId] = validatorsList

	return &NodesCoordinatorMock{
		ShardConsensusSize: 1,
		MetaConsensusSize:  1,
		ShardId:            0,
		NbShards:           nbShards,
		Validators:         validatorsMap,
	}
}

// GetChance -
func (ncm *NodesCoordinatorMock) GetChance(uint32) uint32 {
	return 1
}

// GetAllLeavingValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetAllLeavingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// GetNumTotalEligible -
func (ncm *NodesCoordinatorMock) GetNumTotalEligible() uint64 {
	if ncm.GetNumTotalEligibleCalled != nil {
		return ncm.GetNumTotalEligibleCalled()
	}
	return 1
}

// GetAllEligibleValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	if ncm.GetAllEligibleValidatorsPublicKeysCalled != nil {
		return ncm.GetAllEligibleValidatorsPublicKeysCalled(epoch)
	}
	return nil, nil
}

// GetAllWaitingValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetAllWaitingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	if ncm.GetAllWaitingValidatorsPublicKeysCalled != nil {
		return ncm.GetAllWaitingValidatorsPublicKeysCalled()
	}
	return nil, nil
}

// GetAllShuffledOutValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetAllShuffledOutValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	if ncm.GetAllShuffledOutValidatorsPublicKeysCalled != nil {
		return ncm.GetAllShuffledOutValidatorsPublicKeysCalled(epoch)
	}
	return nil, nil
}

// GetValidatorsIndexes -
func (ncm *NodesCoordinatorMock) GetValidatorsIndexes(publicKeys []string, epoch uint32) ([]uint64, error) {
	if ncm.GetValidatorsIndexesCalled != nil {
		return ncm.GetValidatorsIndexesCalled(publicKeys, epoch)
	}

	return nil, nil
}

// GetSelectedPublicKeys -
func (ncm *NodesCoordinatorMock) GetSelectedPublicKeys(selection []byte, shardId uint32, epoch uint32) (publicKeys []string, err error) {
	if ncm.GetSelectedPublicKeysCalled != nil {
		return ncm.GetSelectedPublicKeysCalled(selection, shardId, epoch)
	}

	if len(ncm.Validators) == 0 {
		return nil, nodesCoordinator.ErrNilInputNodesMap
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

// SetNodesPerShards -
func (ncm *NodesCoordinatorMock) SetNodesPerShards(
	eligible map[uint32][]nodesCoordinator.Validator,
	_ map[uint32][]nodesCoordinator.Validator,
	epoch uint32,
) error {
	if ncm.SetNodesPerShardsCalled != nil {
		return ncm.SetNodesPerShardsCalled(eligible, epoch)
	}

	if eligible == nil {
		return nodesCoordinator.ErrNilInputNodesMap
	}

	ncm.Validators = eligible

	return nil
}

// ComputeAdditionalLeaving -
func (ncm *NodesCoordinatorMock) ComputeAdditionalLeaving(_ []*state.ShardValidatorInfo) (map[uint32][]nodesCoordinator.Validator, error) {
	return make(map[uint32][]nodesCoordinator.Validator), nil
}

// ComputeConsensusGroup -
func (ncm *NodesCoordinatorMock) ComputeConsensusGroup(
	randomess []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) ([]nodesCoordinator.Validator, error) {
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
		return nil, nodesCoordinator.ErrNilRandomness
	}

	validatorsGroup := make([]nodesCoordinator.Validator, 0)

	for i := uint32(0); i < consensusSize; i++ {
		validatorsGroup = append(validatorsGroup, ncm.Validators[shardId][i])
	}

	return validatorsGroup, nil
}

// ConsensusGroupSize -
func (ncm *NodesCoordinatorMock) ConsensusGroupSize(shardId uint32) int {
	if ncm.ConsensusGroupSizeCalled != nil {
		return ncm.ConsensusGroupSizeCalled(shardId)
	}
	return 1
}

// GetValidatorWithPublicKey -
func (ncm *NodesCoordinatorMock) GetValidatorWithPublicKey(publicKey []byte) (nodesCoordinator.Validator, uint32, error) {
	if ncm.GetValidatorWithPublicKeyCalled != nil {
		return ncm.GetValidatorWithPublicKeyCalled(publicKey)
	}

	if publicKey == nil {
		return nil, 0, nodesCoordinator.ErrNilPubKey
	}

	for shardId, shardEligible := range ncm.Validators {
		for i := 0; i < len(shardEligible); i++ {
			if bytes.Equal(publicKey, shardEligible[i].PubKey()) {
				return shardEligible[i], shardId, nil
			}
		}
	}

	return nil, 0, nodesCoordinator.ErrValidatorNotFound
}

// GetOwnPublicKey -
func (ncm *NodesCoordinatorMock) GetOwnPublicKey() []byte {
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
	return 0, nil
}

// ShuffleOutForEpoch verifies if the shards changed in the new epoch and calls the shuffleOutHandler
func (ncm *NodesCoordinatorMock) ShuffleOutForEpoch(_ uint32) {
}

// GetConsensusWhitelistedNodes return the whitelisted nodes allowed to send consensus messages, for each of the shards
func (ncm *NodesCoordinatorMock) GetConsensusWhitelistedNodes(
	_ uint32,
) (map[string]struct{}, error) {
	return make(map[string]struct{}), nil
}

// ValidatorsWeights -
func (ncm *NodesCoordinatorMock) ValidatorsWeights(validators []nodesCoordinator.Validator) ([]uint32, error) {
	weights := make([]uint32, len(validators))
	for i := range validators {
		weights[i] = 1
	}

	return weights, nil
}

// GetWaitingEpochsLeftForPublicKey always returns 0
func (ncm *NodesCoordinatorMock) GetWaitingEpochsLeftForPublicKey(_ []byte) (uint32, error) {
	return 0, nil
}

// IsInterfaceNil -
func (ncm *NodesCoordinatorMock) IsInterfaceNil() bool {
	return ncm == nil
}
