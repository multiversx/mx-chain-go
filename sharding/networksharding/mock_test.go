package networksharding_test

import (
	state "github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NodesCoordinatorStub can not be moved inside mock package as it generates cyclic imports.
//TODO refactor mock package & sharding package & remove this file. Put tests in sharding_test package
type nodesCoordinatorStub struct {
	GetValidatorWithPublicKeyCalled func(publicKey []byte) (validator sharding.Validator, shardId uint32, err error)
}

// GetChance -
func (ncm *nodesCoordinatorStub) GetChance(uint32) uint32 {
	return 1
}

// GetAllLeavingValidatorsPublicKeys -
func (ncs *nodesCoordinatorStub) GetAllLeavingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// ComputeAdditionalLeaving -
func (ncs *nodesCoordinatorStub) ComputeAdditionalLeaving(_ []*state.ShardValidatorInfo) (map[uint32][]sharding.Validator, error) {
	panic("implement me")
}

// GetAllEligibleValidatorsPublicKeys -
func (ncs *nodesCoordinatorStub) GetAllEligibleValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	panic("implement me")
}

// GetAllWaitingValidatorsPublicKeys -
func (ncs *nodesCoordinatorStub) GetAllWaitingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	panic("implement me")
}

// GetNumTotalEligible -
func (ncs *nodesCoordinatorStub) GetNumTotalEligible() uint64 {
	panic("implement me")
}

// GetValidatorsIndexes -
func (ncs *nodesCoordinatorStub) GetValidatorsIndexes(_ []string, _ uint32) ([]uint64, error) {
	panic("implement me")
}

// GetConsensusValidatorsPublicKeys -
func (ncs *nodesCoordinatorStub) GetConsensusValidatorsPublicKeys(_ []byte, _ uint64, _ uint32, _ uint32) ([]string, error) {
	panic("implement me")
}

// GetConsensusValidatorsRewardsAddresses -
func (ncs *nodesCoordinatorStub) GetConsensusValidatorsRewardsAddresses(_ []byte, _ uint64, _ uint32, _ uint32) ([]string, error) {
	panic("implement me")
}

// GetOwnPublicKey -
func (ncs *nodesCoordinatorStub) GetOwnPublicKey() []byte {
	panic("implement me")
}

// SetNodesPerShards -
func (ncs *nodesCoordinatorStub) SetNodesPerShards(_ map[uint32][]sharding.Validator, _ map[uint32][]sharding.Validator, _ []sharding.Validator, _ uint32) error {
	panic("implement me")
}

// ComputeConsensusGroup -
func (ncs *nodesCoordinatorStub) ComputeConsensusGroup(_ []byte, _ uint64, _ uint32, _ uint32) (validatorsGroup []sharding.Validator, err error) {
	panic("implement me")
}

// LoadState -
func (ncs *nodesCoordinatorStub) LoadState(_ []byte) error {
	panic("implement me")
}

// GetSavedStateKey -
func (ncs *nodesCoordinatorStub) GetSavedStateKey() []byte {
	panic("implement me")
}

// ShardIdForEpoch -
func (ncs *nodesCoordinatorStub) ShardIdForEpoch(_ uint32) (uint32, error) {
	panic("implement me")
}

// ShuffleOutForEpoch verifies if the shards changed in the new epoch and calls the shuffleOutHandler
func (ncm *nodesCoordinatorStub) ShuffleOutForEpoch(_ uint32) {
	panic("not implemented")
}

// GetConsensusWhitelistedNodes -
func (ncs *nodesCoordinatorStub) GetConsensusWhitelistedNodes(_ uint32) (map[string]struct{}, error) {
	panic("implement me")
}

// ConsensusGroupSize -
func (ncs *nodesCoordinatorStub) ConsensusGroupSize(uint32) int {
	panic("implement me")
}

// GetValidatorWithPublicKey -
func (ncs *nodesCoordinatorStub) GetValidatorWithPublicKey(publicKey []byte) (sharding.Validator, uint32, error) {
	if ncs.GetValidatorWithPublicKeyCalled != nil {
		return ncs.GetValidatorWithPublicKeyCalled(publicKey)
	}

	return nil, 0, sharding.ErrValidatorNotFound
}

// ValidatorsWeights -
func (ncs *nodesCoordinatorStub) ValidatorsWeights(validators []sharding.Validator) ([]uint32, error) {
	weights := make([]uint32, len(validators))
	for i := range validators {
		weights[i] = 1
	}

	return weights, nil
}

// IsInterfaceNil -
func (ncs *nodesCoordinatorStub) IsInterfaceNil() bool {
	return ncs == nil
}
