package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data"
	nodesCoord "github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
)

// nodesCoordinator -
type nodesCoordinator struct {
}

// NewNodesCoordinator returns a new instance of nodesCoordinator
func NewNodesCoordinator() *nodesCoordinator {
	return &nodesCoordinator{}
}

// GetChance -
func (n *nodesCoordinator) GetChance(uint32) uint32 {
	return 1
}

// GetAllLeavingValidatorsPublicKeys -
func (n *nodesCoordinator) GetAllLeavingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// ValidatorsWeights -
func (n *nodesCoordinator) ValidatorsWeights(validators []nodesCoord.Validator) ([]uint32, error) {
	return make([]uint32, len(validators)), nil
}

// ComputeAdditionalLeaving -
func (n *nodesCoordinator) ComputeAdditionalLeaving(_ []*state.ShardValidatorInfo) (map[uint32][]nodesCoord.Validator, error) {
	return nil, nil
}

// GetValidatorsIndexes -
func (n *nodesCoordinator) GetValidatorsIndexes(_ []string, _ uint32) ([]uint64, error) {
	return nil, nil
}

// GetAllEligibleValidatorsPublicKeys -
func (n *nodesCoordinator) GetAllEligibleValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// GetAllWaitingValidatorsPublicKeys -
func (n *nodesCoordinator) GetAllWaitingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

// GetConsensusValidatorsPublicKeys -
func (n *nodesCoordinator) GetConsensusValidatorsPublicKeys(_ []byte, _ uint64, _ uint32, _ uint32) ([]string, error) {
	return nil, nil
}

// GetOwnPublicKey -
func (n *nodesCoordinator) GetOwnPublicKey() []byte {
	return nil
}

// ComputeConsensusGroup -
func (n *nodesCoordinator) ComputeConsensusGroup(_ []byte, _ uint64, _ uint32, _ uint32) (validatorsGroup []nodesCoord.Validator, err error) {
	return nil, nil
}

// GetValidatorWithPublicKey -
func (n *nodesCoordinator) GetValidatorWithPublicKey(_ []byte) (validator nodesCoord.Validator, shardId uint32, err error) {
	return nil, 0, nil
}

// LoadState -
func (n *nodesCoordinator) LoadState(_ []byte) error {
	return nil
}

// GetSavedStateKey -
func (n *nodesCoordinator) GetSavedStateKey() []byte {
	return nil
}

// ShardIdForEpoch -
func (n *nodesCoordinator) ShardIdForEpoch(_ uint32) (uint32, error) {
	return 0, nil
}

// ShuffleOutForEpoch verifies if the shards changed in the new epoch and calls the shuffleOutHandler
func (n *nodesCoordinator) ShuffleOutForEpoch(_ uint32) {
}

// GetConsensusWhitelistedNodes -
func (n *nodesCoordinator) GetConsensusWhitelistedNodes(_ uint32) (map[string]struct{}, error) {
	return nil, nil
}

// ConsensusGroupSize -
func (n *nodesCoordinator) ConsensusGroupSize(uint32) int {
	return 0
}

// GetNumTotalEligible -
func (n *nodesCoordinator) GetNumTotalEligible() uint64 {
	return 0
}

// EpochStartPrepare -
func (n *nodesCoordinator) EpochStartPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
}

// NodesCoordinatorToRegistry -
func (n *nodesCoordinator) NodesCoordinatorToRegistry() *nodesCoord.NodesCoordinatorRegistry {
	return nil
}

// IsInterfaceNil -
func (n *nodesCoordinator) IsInterfaceNil() bool {
	return n == nil
}
