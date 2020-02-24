package mock

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NodesCoordinatorMock -
type NodesCoordinatorMock struct {
	ComputeValidatorsGroupCalled        func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]sharding.Validator, error)
	GetValidatorsPublicKeysCalled       func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
	GetValidatorsRewardsAddressesCalled func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error)
}

// GetWaitingPublicKeysPerShard -
func (ncm *NodesCoordinatorMock) GetWaitingPublicKeysPerShard(epoch uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// UpdatePeersListAndIndex -
func (ncm *NodesCoordinatorMock) UpdatePeersListAndIndex() error {
	return nil
}

// GetEligiblePublicKeysPerShard -
func (ncm *NodesCoordinatorMock) GetEligiblePublicKeysPerShard(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// GetValidatorsIndexes -
func (ncm *NodesCoordinatorMock) GetValidatorsIndexes(_ []string, _ uint32) ([]uint64, error) {
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

	var list []sharding.Validator

	return list, nil
}

// ConsensusGroupSize -
func (ncm *NodesCoordinatorMock) ConsensusGroupSize(uint32) int {
	return 1
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

	pubKeys := make([]string, 0)

	for _, v := range validators {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
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
func (ncm *NodesCoordinatorMock) SetNodesPerShards(_ map[uint32][]sharding.Validator, _ map[uint32][]sharding.Validator, _ uint32, _ bool) error {
	return nil
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

// GetSelectedPublicKeys -
func (ncm *NodesCoordinatorMock) GetSelectedPublicKeys(_ []byte, _ uint32, _ uint32) ([]string, error) {
	panic("implement me")
}

// GetValidatorWithPublicKey -
func (ncm *NodesCoordinatorMock) GetValidatorWithPublicKey(_ []byte, _ uint32) (sharding.Validator, uint32, error) {
	panic("implement me")
}

// GetOwnPublicKey -
func (ncm *NodesCoordinatorMock) GetOwnPublicKey() []byte {
	return []byte("key")
}

// GetNodesPerShard -
func (ncm *NodesCoordinatorMock) GetNodesPerShard(epoch uint32) (map[uint32][]sharding.Validator, error) {
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ncm *NodesCoordinatorMock) IsInterfaceNil() bool {
	return ncm == nil
}
