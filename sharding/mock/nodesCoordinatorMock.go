package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NodesCoordinatorMock -
type NodesCoordinatorMock struct {
	ComputeValidatorsGroupCalled  func([]byte) ([]sharding.Validator, error)
	GetValidatorsPublicKeysCalled func(randomness []byte) ([]string, error)
}

// ComputeValidatorsGroup -
func (ncm NodesCoordinatorMock) ComputeValidatorsGroup(randomness []byte) (validatorsGroup []sharding.Validator, err error) {
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

	validators, err := ncm.ComputeValidatorsGroup(randomness)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for _, v := range validators {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

// GetNumTotalEligible -
func (ncm *NodesCoordinatorMock) GetNumTotalEligible() uint64 {
	return 1
}

// ConsensusGroupSize -
func (ncm NodesCoordinatorMock) ConsensusGroupSize() int {
	return 1
}

// SetNodesPerShards -
func (ncm NodesCoordinatorMock) SetNodesPerShards(_ map[uint32][]sharding.Validator) error {
	return nil
}

// SetConsensusGroupSize -
func (ncm NodesCoordinatorMock) SetConsensusGroupSize(_ int) error {
	return nil
}

// GetSelectedPublicKeys -
func (ncm NodesCoordinatorMock) GetSelectedPublicKeys(_ []byte) (publicKeys []string, err error) {
	return nil, nil
}

// GetValidatorWithPublicKey -
func (ncm NodesCoordinatorMock) GetValidatorWithPublicKey(_ []byte) (sharding.Validator, uint32, error) {
	return nil, 0, nil
}
