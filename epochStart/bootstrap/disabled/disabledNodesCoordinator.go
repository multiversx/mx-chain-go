package disabled

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// nodesCoordinator -
type nodesCoordinator struct {
}

// NewNodesCoordinator returns a new instance of nodesCoordinator
func NewNodesCoordinator() *nodesCoordinator {
	return &nodesCoordinator{}
}

func (n *nodesCoordinator) SetNodesPerShards(
	_ map[uint32][]sharding.Validator,
	_ map[uint32][]sharding.Validator,
	_ uint32,
	_ bool,
) error {
	return nil
}

// SaveNodesCoordinatorRegistry -
func (n *nodesCoordinator) SaveNodesCoordinatorRegistry(_ *sharding.NodesCoordinatorRegistry) error {
	return nil
}

func (n *nodesCoordinator) ComputeLeaving(_ []sharding.Validator) []sharding.Validator {
	return nil
}

func (n *nodesCoordinator) GetValidatorsIndexes(_ []string, _ uint32) ([]uint64, error) {
	return nil, nil
}

func (n *nodesCoordinator) GetAllEligibleValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

func (n *nodesCoordinator) GetAllWaitingValidatorsPublicKeys(_ uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

func (n *nodesCoordinator) GetConsensusValidatorsPublicKeys(_ []byte, _ uint64, _ uint32, _ uint32) ([]string, error) {
	return nil, nil
}

func (n *nodesCoordinator) GetOwnPublicKey() []byte {
	return nil
}

func (n *nodesCoordinator) ComputeConsensusGroup(_ []byte, _ uint64, _ uint32, _ uint32) (validatorsGroup []sharding.Validator, err error) {
	return nil, nil
}

func (n *nodesCoordinator) GetValidatorWithPublicKey(_ []byte, _ uint32) (validator sharding.Validator, shardId uint32, err error) {
	return nil, 0, nil
}

func (n *nodesCoordinator) UpdatePeersListAndIndex() error {
	return nil
}

func (n *nodesCoordinator) LoadState(_ []byte) error {
	return nil
}

func (n *nodesCoordinator) GetSavedStateKey() []byte {
	return nil
}

func (n *nodesCoordinator) ShardIdForEpoch(_ uint32) (uint32, error) {
	return 0, nil
}

func (n *nodesCoordinator) GetConsensusWhitelistedNodes(_ uint32) (map[string]struct{}, error) {
	return nil, nil
}

func (n *nodesCoordinator) ConsensusGroupSize(uint32) int {
	return 0
}

func (n *nodesCoordinator) GetNumTotalEligible() uint64 {
	return 0
}

func (n *nodesCoordinator) IsInterfaceNil() bool {
	return n == nil
}
