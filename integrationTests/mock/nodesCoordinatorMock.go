package mock

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NodesCoordinatorMock -
type NodesCoordinatorMock struct {
	ComputeValidatorsGroupCalled        func(randomness []byte, round uint64, shardId uint32) ([]sharding.Validator, error)
	GetValidatorsPublicKeysCalled       func(randomness []byte, round uint64, shardId uint32) ([]string, error)
	GetValidatorsRewardsAddressesCalled func(randomness []byte, round uint64, shardId uint32) ([]string, error)
	GetValidatorWithPublicKeyCalled     func(publicKey []byte) (validator sharding.Validator, shardId uint32, err error)
	GetAllValidatorsPublicKeysCalled    func() map[uint32][][]byte
}

// GetNumTotalEligible -
func (ncm *NodesCoordinatorMock) GetNumTotalEligible() uint64 {
	return 1
}

// GetAllValidatorsPublicKeys -
func (ncm *NodesCoordinatorMock) GetAllValidatorsPublicKeys() map[uint32][][]byte {
	if ncm.GetAllValidatorsPublicKeysCalled != nil {
		return ncm.GetAllValidatorsPublicKeysCalled()
	}

	return nil
}

// GetValidatorsIndexes -
func (ncm *NodesCoordinatorMock) GetValidatorsIndexes([]string) []uint64 {
	return nil
}

// ComputeValidatorsGroup -
func (ncm *NodesCoordinatorMock) ComputeValidatorsGroup(
	randomness []byte,
	round uint64,
	shardId uint32,
) (validatorsGroup []sharding.Validator, err error) {

	if ncm.ComputeValidatorsGroupCalled != nil {
		return ncm.ComputeValidatorsGroupCalled(randomness, round, shardId)
	}

	var list []sharding.Validator

	return list, nil
}

// ConsensusGroupSize -
func (ncm *NodesCoordinatorMock) ConsensusGroupSize(uint32) int {
	return 1
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

	pubKeys := make([]string, 0)

	for _, v := range validators {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
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
func (ncm *NodesCoordinatorMock) SetNodesPerShards(map[uint32][]sharding.Validator) error {
	return nil
}

// GetSelectedPublicKeys -
func (ncm *NodesCoordinatorMock) GetSelectedPublicKeys([]byte, uint32) (publicKeys []string, err error) {
	panic("implement me")
}

// GetValidatorWithPublicKey -
func (ncm *NodesCoordinatorMock) GetValidatorWithPublicKey(address []byte) (sharding.Validator, uint32, error) {
	if ncm.GetValidatorWithPublicKeyCalled != nil {
		return ncm.GetValidatorWithPublicKeyCalled(address)
	}
	return nil, 0, nil
}

// GetOwnPublicKey -
func (ncm *NodesCoordinatorMock) GetOwnPublicKey() []byte {
	return []byte("key")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ncm *NodesCoordinatorMock) IsInterfaceNil() bool {
	if ncm == nil {
		return true
	}
	return false
}
