package disabled

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var errNotImplemented = errors.New("not implemented")

type disabledNodesCoordinator struct {
}

// NewNodesCoordinator returns a new instance of a disabledNodesCoordinator
func NewNodesCoordinator() *disabledNodesCoordinator {
	return &disabledNodesCoordinator{}
}

// ValidatorsWeights will return an error that indicates that the function is not implemented
func (d *disabledNodesCoordinator) ValidatorsWeights(_ []sharding.Validator) ([]uint32, error) {
	return nil, errNotImplemented
}

// ComputeAdditionalLeaving will return an error that indicates that the function is not implemented
func (d *disabledNodesCoordinator) ComputeAdditionalLeaving(_ []*state.ShardValidatorInfo) (map[uint32][]sharding.Validator, error) {
	return nil, errNotImplemented
}

// GetChance will return a zero value
func (d *disabledNodesCoordinator) GetChance(uint32) uint32 {
	return 0
}

// GetValidatorsIndexes will return an error that indicates that the function is not implemented
func (d *disabledNodesCoordinator) GetValidatorsIndexes(_ []string, _ uint32) ([]uint64, error) {
	return nil, errNotImplemented
}

// GetAllEligibleValidatorsPublicKeys will return an error that indicates that the function is not implemented
func (d *disabledNodesCoordinator) GetAllEligibleValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, errNotImplemented
}

// GetAllWaitingValidatorsPublicKeys will return an error that indicates that the function is not implemented
func (d *disabledNodesCoordinator) GetAllWaitingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, errNotImplemented
}

// GetAllLeavingValidatorsPublicKeys will return an error that indicates that the function is not implemented
func (d *disabledNodesCoordinator) GetAllLeavingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, errNotImplemented
}

// GetConsensusValidatorsPublicKeys will return an error that indicates that the function is not implemented
func (d *disabledNodesCoordinator) GetConsensusValidatorsPublicKeys(_ []byte, _ uint64, _ uint32, _ uint32) ([]string, error) {
	return nil, errNotImplemented
}

// GetOwnPublicKey will return an empty byte slice
func (d *disabledNodesCoordinator) GetOwnPublicKey() []byte {
	return make([]byte, 0)
}

// ComputeConsensusGroup will return an error that indicates that the function is not implemented
func (d *disabledNodesCoordinator) ComputeConsensusGroup(_ []byte, _ uint64, _ uint32, _ uint32) (validatorsGroup []sharding.Validator, err error) {
	return nil, errNotImplemented
}

// GetValidatorWithPublicKey will return an error that indicates that the function is not implemented
func (d *disabledNodesCoordinator) GetValidatorWithPublicKey(_ []byte) (validator sharding.Validator, shardId uint32, err error) {
	return nil, 0, errNotImplemented
}

// LoadState will return an error that indicates that the function is not implemented
func (d *disabledNodesCoordinator) LoadState(_ []byte) error {
	return errNotImplemented
}

// GetSavedStateKey will return an empty byte slice
func (d *disabledNodesCoordinator) GetSavedStateKey() []byte {
	return make([]byte, 0)
}

// ShardIdForEpoch will return an error that indicates that the function is not implemented
func (d *disabledNodesCoordinator) ShardIdForEpoch(_ uint32) (uint32, error) {
	return 0, errNotImplemented
}

// ShuffleOutForEpoch won't do anything
func (d *disabledNodesCoordinator) ShuffleOutForEpoch(_ uint32) {
}

// GetConsensusWhitelistedNodes will return an error that indicates that the function is not implemented
func (d *disabledNodesCoordinator) GetConsensusWhitelistedNodes(_ uint32) (map[string]struct{}, error) {
	return nil, errNotImplemented
}

// ConsensusGroupSize will return a zero value
func (d *disabledNodesCoordinator) ConsensusGroupSize(uint32) int {
	return 0
}

// GetNumTotalEligible will return a zero value
func (d *disabledNodesCoordinator) GetNumTotalEligible() uint64 {
	return 0
}

// IsInterfaceNil return true if there is no value under the interface
func (d *disabledNodesCoordinator) IsInterfaceNil() bool {
	return d == nil
}
