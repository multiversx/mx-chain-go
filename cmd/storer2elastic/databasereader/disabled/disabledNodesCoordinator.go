package disabled

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type disabledNodesCoordinator struct {
}

func NewNodesCoordinator() *disabledNodesCoordinator {
	return &disabledNodesCoordinator{}
}

func (d *disabledNodesCoordinator) ValidatorsWeights(validators []sharding.Validator) ([]uint32, error) {
	panic("implement me")
}

func (d *disabledNodesCoordinator) ComputeAdditionalLeaving(allValidators []*state.ShardValidatorInfo) (map[uint32][]sharding.Validator, error) {
	panic("implement me")
}

func (d *disabledNodesCoordinator) GetChance(uint32) uint32 {
	panic("implement me")
}

func (d *disabledNodesCoordinator) GetValidatorsIndexes(publicKeys []string, epoch uint32) ([]uint64, error) {
	panic("implement me")
}

func (d *disabledNodesCoordinator) GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	panic("implement me")
}

func (d *disabledNodesCoordinator) GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	panic("implement me")
}

func (d *disabledNodesCoordinator) GetAllLeavingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	panic("implement me")
}

func (d *disabledNodesCoordinator) GetConsensusValidatorsPublicKeys(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error) {
	panic("implement me")
}

func (d *disabledNodesCoordinator) GetOwnPublicKey() []byte {
	panic("implement me")
}

func (d *disabledNodesCoordinator) ComputeConsensusGroup(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []sharding.Validator, err error) {
	panic("implement me")
}

func (d *disabledNodesCoordinator) GetValidatorWithPublicKey(publicKey []byte) (validator sharding.Validator, shardId uint32, err error) {
	panic("implement me")
}

func (d *disabledNodesCoordinator) LoadState(key []byte) error {
	panic("implement me")
}

func (d *disabledNodesCoordinator) GetSavedStateKey() []byte {
	panic("implement me")
}

func (d *disabledNodesCoordinator) ShardIdForEpoch(epoch uint32) (uint32, error) {
	panic("implement me")
}

func (d *disabledNodesCoordinator) ShuffleOutForEpoch(_ uint32) {
	panic("implement me")
}

func (d *disabledNodesCoordinator) GetConsensusWhitelistedNodes(epoch uint32) (map[string]struct{}, error) {
	panic("implement me")
}

func (d *disabledNodesCoordinator) ConsensusGroupSize(uint32) int {
	panic("implement me")
}

func (d *disabledNodesCoordinator) GetNumTotalEligible() uint64 {
	panic("implement me")
}

func (d *disabledNodesCoordinator) IsInterfaceNil() bool {
	return d == nil
}
