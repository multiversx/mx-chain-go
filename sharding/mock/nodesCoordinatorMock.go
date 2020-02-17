package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NodesCoordinatorMock -
type NodesCoordinatorMock struct {
	ComputeValidatorsGroupCalled    func([]byte) ([]sharding.Validator, error)
	GetValidatorsPublicKeysCalled   func(randomness []byte) ([]string, error)
	GetValidatorWithPublicKeyCalled func(publicKey []byte) (sharding.Validator, uint32, error)
}

func (ncm *NodesCoordinatorMock) GetValidatorsIndexes(publicKeys []string) []uint64 {
	panic("implement me")
}

func (ncm *NodesCoordinatorMock) GetAllValidatorsPublicKeys() map[uint32][][]byte {
	panic("implement me")
}

func (ncm *NodesCoordinatorMock) GetValidatorsRewardsAddresses(randomness []byte, round uint64, shardId uint32) ([]string, error) {
	panic("implement me")
}

// ComputeValidatorsGroup -
func (ncm NodesCoordinatorMock) ComputeValidatorsGroup(randomness []byte) (validatorsGroup []sharding.Validator, err error) {
func (ncm *NodesCoordinatorMock) GetOwnPublicKey() []byte {
	panic("implement me")
}

func (ncm *NodesCoordinatorMock) IsInterfaceNil() bool {
	return ncm == nil
}

func (ncm *NodesCoordinatorMock) ComputeValidatorsGroup(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
	if ncm.ComputeValidatorsGroupCalled != nil {
		return ncm.ComputeValidatorsGroupCalled(randomness)
	}

	list := []sharding.Validator{
		NewValidatorMock(big.NewInt(0), 0, []byte("A"), []byte("AA")),
		NewValidatorMock(big.NewInt(0), 0, []byte("B"), []byte("BB")),
		NewValidatorMock(big.NewInt(0), 0, []byte("C"), []byte("CC")),
		NewValidatorMock(big.NewInt(0), 0, []byte("D"), []byte("DD")),
		NewValidatorMock(big.NewInt(0), 0, []byte("E"), []byte("EE")),
		NewValidatorMock(big.NewInt(0), 0, []byte("F"), []byte("FF")),
		NewValidatorMock(big.NewInt(0), 0, []byte("G"), []byte("GG")),
		NewValidatorMock(big.NewInt(0), 0, []byte("H"), []byte("HH")),
		NewValidatorMock(big.NewInt(0), 0, []byte("I"), []byte("II")),
	}

	return list, nil
}

// GetValidatorsPublicKeys -
func (ncm NodesCoordinatorMock) GetValidatorsPublicKeys(randomness []byte) ([]string, error) {
	if ncm.GetValidatorsPublicKeysCalled != nil {
		return ncm.GetValidatorsPublicKeysCalled(randomness)
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

// ConsensusGroupSize -
func (ncm NodesCoordinatorMock) ConsensusGroupSize() int {
	panic("implement me")
}

// SetNodesPerShards -
func (ncm NodesCoordinatorMock) SetNodesPerShards(map[uint32][]sharding.Validator) error {
	return nil
}

// SetConsensusGroupSize -
func (ncm NodesCoordinatorMock) SetConsensusGroupSize(int) error {
	panic("implement me")
}

// GetSelectedPublicKeys -
func (ncm NodesCoordinatorMock) GetSelectedPublicKeys(selection []byte) (publicKeys []string, err error) {
	panic("implement me")
}

// GetValidatorWithPublicKey -
func (ncm NodesCoordinatorMock) GetValidatorWithPublicKey(publicKey []byte) (sharding.Validator, uint32, error) {
	return ncm.GetValidatorWithPublicKeyCalled(publicKey)
}
