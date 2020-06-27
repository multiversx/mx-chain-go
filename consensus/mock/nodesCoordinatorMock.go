package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const defaultSelectionChances = 1

// NodesCoordinatorMock -
type NodesCoordinatorMock struct {
	ComputeValidatorsGroupCalled        func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]sharding.Validator, error)
	GetValidatorsPublicKeysCalled       func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	GetValidatorsRewardsAddressesCalled func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
}

// GetChance -
func (ncm *NodesCoordinatorMock) GetChance(uint32) uint32 {
	return 1
}

// GetAllLeavingValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetAllLeavingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// ComputeConsensusGroup -
func (ncm *NodesCoordinatorMock) ComputeConsensusGroup(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) (validatorsGroup []sharding.Validator, err error) {

	if ncm.ComputeValidatorsGroupCalled != nil {
		return ncm.ComputeValidatorsGroupCalled(randomness, round, shardId, epoch)
	}

	list := []sharding.Validator{
		NewValidator([]byte("A"), 1, defaultSelectionChances),
		NewValidator([]byte("B"), 1, defaultSelectionChances),
		NewValidator([]byte("C"), 1, defaultSelectionChances),
		NewValidator([]byte("D"), 1, defaultSelectionChances),
		NewValidator([]byte("E"), 1, defaultSelectionChances),
		NewValidator([]byte("F"), 1, defaultSelectionChances),
		NewValidator([]byte("G"), 1, defaultSelectionChances),
		NewValidator([]byte("H"), 1, defaultSelectionChances),
		NewValidator([]byte("I"), 1, defaultSelectionChances),
	}

	return list, nil
}

// GetNumTotalEligible -
func (ncm *NodesCoordinatorMock) GetNumTotalEligible() uint64 {
	return 1
}

// ConsensusGroupSize -
func (ncm *NodesCoordinatorMock) ConsensusGroupSize(uint32) int {
	return 1
}

// GetAllEligibleValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetAllEligibleValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// GetAllWaitingValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetAllWaitingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// GetValidatorsIndexes -
func (ncm *NodesCoordinatorMock) GetValidatorsIndexes(_ []string, _ uint32) ([]uint64, error) {
	return nil, nil
}

// GetConsensusValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetConsensusValidatorsPublicKeys(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error) {
	if ncm.GetValidatorsPublicKeysCalled != nil {
		return ncm.GetValidatorsPublicKeysCalled(randomness, round, shardId, epoch)
	}

	validators, err := ncm.ComputeConsensusGroup(randomness, round, shardId, epoch)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for _, v := range validators {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

// ValidatorsWeights -
func (ncm *NodesCoordinatorMock) ValidatorsWeights(validators []sharding.Validator) ([]uint32, error) {
	weights := make([]uint32, len(validators))
	for i := range validators {
		weights[i] = 1
	}

	return weights, nil
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
	panic("not implemented")
}

// GetConsensusWhitelistedNodes return the whitelisted nodes allowed to send consensus messages, for each of the shards
func (ncm *NodesCoordinatorMock) GetConsensusWhitelistedNodes(
	_ uint32,
) (map[string]struct{}, error) {
	panic("not implemented")
}

// ComputeAdditionalLeaving -
func (ncm *NodesCoordinatorMock) ComputeAdditionalLeaving([]*state.ShardValidatorInfo) (map[uint32][]sharding.Validator, error) {
	return make(map[uint32][]sharding.Validator), nil
}

// GetValidatorWithPublicKey -
func (ncm *NodesCoordinatorMock) GetValidatorWithPublicKey(_ []byte) (sharding.Validator, uint32, error) {
	panic("implement me")
}

// GetOwnPublicKey -
func (ncm *NodesCoordinatorMock) GetOwnPublicKey() []byte {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ncm *NodesCoordinatorMock) IsInterfaceNil() bool {
	return ncm == nil
}
