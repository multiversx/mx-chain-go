package mock

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NodesCoordinatorStub -
type NodesCoordinatorStub struct {
	ComputeValidatorsGroupCalled        func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]sharding.Validator, error)
	GetValidatorsPublicKeysCalled       func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	GetValidatorsRewardsAddressesCalled func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	GetValidatorWithPublicKeyCalled     func(publicKey []byte) (validator sharding.Validator, shardId uint32, err error)
	GetAllValidatorsPublicKeysCalled    func() (map[uint32][][]byte, error)
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
func (ncm *NodesCoordinatorStub) GetValidatorsIndexes(_ []string, _ uint32) ([]uint64, error) {
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
func (ncm *NodesCoordinatorStub) ConsensusGroupSize(uint32) int {
	return 1
}

// GetConsensusValidatorsPublicKeys -
func (ncm *NodesCoordinatorStub) GetConsensusValidatorsPublicKeys(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) ([]string, error) {
	if ncm.GetValidatorsPublicKeysCalled != nil {
		return ncm.GetValidatorsPublicKeysCalled(randomness, round, shardId, epoch)
	}

	return nil, nil
}

// GetConsensusValidatorsRewardsAddresses -
func (ncm *NodesCoordinatorStub) GetConsensusValidatorsRewardsAddresses(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) ([]string, error) {
	if ncm.GetValidatorsPublicKeysCalled != nil {
		return ncm.GetValidatorsRewardsAddressesCalled(randomness, round, shardId, epoch)
	}

	return nil, nil
}

// SetNodesPerShards -
func (ncm *NodesCoordinatorStub) SetNodesPerShards(_ map[uint32][]sharding.Validator, _ map[uint32][]sharding.Validator, _ uint32) error {
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
func (ncm *NodesCoordinatorStub) GetValidatorWithPublicKey(address []byte, _ uint32) (sharding.Validator, uint32, error) {
	if ncm.GetValidatorWithPublicKeyCalled != nil {
		return ncm.GetValidatorWithPublicKeyCalled(address)
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
