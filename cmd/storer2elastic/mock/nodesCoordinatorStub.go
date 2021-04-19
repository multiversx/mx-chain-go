package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	state "github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NodesCoordinatorStub -
type NodesCoordinatorStub struct {
	ComputeValidatorsGroupCalled           func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]sharding.Validator, error)
	GetValidatorsPublicKeysCalled          func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	GetValidatorsRewardsAddressesCalled    func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	GetValidatorWithPublicKeyCalled        func(publicKey []byte) (validator sharding.Validator, shardId uint32, err error)
	GetConsensusValidatorsPublicKeysCalled func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	GetValidatorsIndexesCalled             func(publicKeys []string, epoch uint32) ([]uint64, error)
	GetAllValidatorsPublicKeysCalled       func() (map[uint32][][]byte, error)
	ConsensusGroupSizeCalled               func(shardID uint32) int
}

// EpochStartPrepare -
func (ncm *NodesCoordinatorStub) EpochStartPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
	panic("implement me")
}

// GetChance -
func (ncm *NodesCoordinatorStub) GetChance(uint32) uint32 {
	return 1
}

// ValidatorsWeights -
func (ncm *NodesCoordinatorStub) ValidatorsWeights(_ []sharding.Validator) ([]uint32, error) {
	return nil, nil
}

// GetAllLeavingValidatorsPublicKeys -
func (ncm *NodesCoordinatorStub) GetAllLeavingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// SetConfig -
func (ncm *NodesCoordinatorStub) SetConfig(_ *sharding.NodesCoordinatorRegistry) error {
	return nil
}

// ComputeAdditionalLeaving -
func (ncm *NodesCoordinatorStub) ComputeAdditionalLeaving(_ []*state.ShardValidatorInfo) (map[uint32][]sharding.Validator, error) {
	return nil, nil
}

// GetAllEligibleValidatorsPublicKeys -
func (ncm *NodesCoordinatorStub) GetAllEligibleValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// GetAllWaitingValidatorsPublicKeys -
func (ncm *NodesCoordinatorStub) GetAllWaitingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// GetNumTotalEligible -
func (ncm *NodesCoordinatorStub) GetNumTotalEligible() uint64 {
	return 1
}

// GetAllValidatorsPublicKeys -
func (ncm *NodesCoordinatorStub) GetAllValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	if ncm.GetAllValidatorsPublicKeysCalled != nil {
		return ncm.GetAllValidatorsPublicKeysCalled()
	}

	return nil, nil
}

// GetValidatorsIndexes -
func (ncm *NodesCoordinatorStub) GetValidatorsIndexes(p []string, e uint32) ([]uint64, error) {
	if ncm.GetValidatorsIndexesCalled != nil {
		return ncm.GetValidatorsIndexesCalled(p, e)
	}

	return nil, nil
}

// ComputeConsensusGroup -
func (ncm *NodesCoordinatorStub) ComputeConsensusGroup(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) (validatorsGroup []sharding.Validator, err error) {

	if ncm.ComputeValidatorsGroupCalled != nil {
		return ncm.ComputeValidatorsGroupCalled(randomness, round, shardId, epoch)
	}

	var list []sharding.Validator

	return list, nil
}

// ConsensusGroupSize -
func (ncm *NodesCoordinatorStub) ConsensusGroupSize(shardID uint32) int {
	if ncm.ConsensusGroupSizeCalled != nil {
		return ncm.ConsensusGroupSizeCalled(shardID)
	}
	return 1
}

// GetConsensusValidatorsPublicKeys -
func (ncm *NodesCoordinatorStub) GetConsensusValidatorsPublicKeys(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) ([]string, error) {
	if ncm.GetConsensusValidatorsPublicKeysCalled != nil {
		return ncm.GetConsensusValidatorsPublicKeysCalled(randomness, round, shardId, epoch)
	}

	return nil, nil
}

// SetNodesPerShards -
func (ncm *NodesCoordinatorStub) SetNodesPerShards(_ map[uint32][]sharding.Validator, _ map[uint32][]sharding.Validator, _ []sharding.Validator, _ uint32) error {
	return nil
}

// LoadState -
func (ncm *NodesCoordinatorStub) LoadState(_ []byte) error {
	return nil
}

// GetSavedStateKey -
func (ncm *NodesCoordinatorStub) GetSavedStateKey() []byte {
	return []byte("key")
}

// ShardIdForEpoch returns the nodesCoordinator configured ShardId for specified epoch if epoch configuration exists,
// otherwise error
func (ncm *NodesCoordinatorStub) ShardIdForEpoch(_ uint32) (uint32, error) {
	panic("not implemented")
}

// ShuffleOutForEpoch verifies if the shards changed in the new epoch and calls the shuffleOutHandler
func (ncm *NodesCoordinatorStub) ShuffleOutForEpoch(_ uint32) {
	panic("not implemented")
}

// GetConsensusWhitelistedNodes return the whitelisted nodes allowed to send consensus messages, for each of the shards
func (ncm *NodesCoordinatorStub) GetConsensusWhitelistedNodes(
	_ uint32,
) (map[string]struct{}, error) {
	panic("not implemented")
}

// GetSelectedPublicKeys -
func (ncm *NodesCoordinatorStub) GetSelectedPublicKeys(_ []byte, _ uint32, _ uint32) ([]string, error) {
	panic("implement me")
}

// GetValidatorWithPublicKey -
func (ncm *NodesCoordinatorStub) GetValidatorWithPublicKey(publicKey []byte) (sharding.Validator, uint32, error) {
	if ncm.GetValidatorWithPublicKeyCalled != nil {
		return ncm.GetValidatorWithPublicKeyCalled(publicKey)
	}
	return nil, 0, nil
}

// GetOwnPublicKey -
func (ncm *NodesCoordinatorStub) GetOwnPublicKey() []byte {
	return []byte("key")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ncm *NodesCoordinatorStub) IsInterfaceNil() bool {
	return ncm == nil
}
