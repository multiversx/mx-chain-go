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
	eligible map[uint32][]sharding.Validator,
	waiting map[uint32][]sharding.Validator,
	epoch uint32,
	updateList bool,
) error {
	return nil
}

func (n *nodesCoordinator) ComputeLeaving(allValidators []sharding.Validator) []sharding.Validator {
	return nil
}

func (n *nodesCoordinator) GetValidatorsIndexes(publicKeys []string, epoch uint32) ([]uint64, error) {
	return nil, nil
}

func (n *nodesCoordinator) GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

func (n *nodesCoordinator) GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	return nil, nil
}

func (n *nodesCoordinator) GetConsensusValidatorsPublicKeys(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error) {
	return nil, nil
}

func (n *nodesCoordinator) GetOwnPublicKey() []byte {
	return nil
}

func (n *nodesCoordinator) ComputeConsensusGroup(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []sharding.Validator, err error) {
	return nil, nil
}

func (n *nodesCoordinator) GetValidatorWithPublicKey(publicKey []byte, epoch uint32) (validator sharding.Validator, shardId uint32, err error) {
	return nil, 0, nil
}

func (n *nodesCoordinator) UpdatePeersListAndIndex() error {
	return nil
}

func (n *nodesCoordinator) LoadState(key []byte) error {
	return nil
}

func (n *nodesCoordinator) GetSavedStateKey() []byte {
	return nil
}

func (n *nodesCoordinator) ShardIdForEpoch(epoch uint32) (uint32, error) {
	return 0, nil
}

func (n *nodesCoordinator) GetConsensusWhitelistedNodes(epoch uint32) (map[string]struct{}, error) {
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
