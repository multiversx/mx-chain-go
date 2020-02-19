package sharding

import "sync"

func (nc *indexHashedNodesCoordinator) NodesConfig() map[uint32]*epochNodesConfig {
	return nc.nodesConfig
}

func (ihgs *indexHashedNodesCoordinator) SaveState(key []byte) error {
	return ihgs.saveState(key)
}

func (ihgs *indexHashedNodesCoordinator) CurrentEpoch() uint32 {
	return ihgs.currentEpoch
}

func (ihgs *indexHashedNodesCoordinator) NodesCoordinatorToRegistry() *NodesCoordinatorRegistry {
	return ihgs.nodesCoordinatorToRegistry()
}

func (ihgs *indexHashedNodesCoordinator) RegistryToNodesCoordinator(
	config *NodesCoordinatorRegistry,
) (map[uint32]*epochNodesConfig, error) {
	return ihgs.registryToNodesCoordinator(config)
}

func (ihgs *indexHashedNodesCoordinator) ComputeShardForPublicKey(nodesConfig *epochNodesConfig) uint32 {
	return ihgs.computeShardForPublicKey(nodesConfig)
}

func (ihgs *indexHashedNodesCoordinatorWithRater) ExpandEligibleList(validators []Validator, mut *sync.RWMutex) []Validator {
	return ihgs.expandEligibleList(validators, mut)
}

func (enc *epochNodesConfig) ShardId() uint32 {
	return enc.shardId
}

func (enc *epochNodesConfig) NbShards() uint32 {
	return enc.nbShards
}

func (enc *epochNodesConfig) EligibleMap() map[uint32][]Validator {
	return enc.eligibleMap
}

func (enc *epochNodesConfig) WaitingMap() map[uint32][]Validator {
	return enc.waitingMap
}

func (enc *epochNodesConfig) MutNodesMaps() sync.RWMutex {
	return enc.mutNodesMaps
}

func EpochNodesConfigToEpochValidators(config *epochNodesConfig) *EpochValidators {
	return epochNodesConfigToEpochValidators(config)
}

func EpochValidatorsToEpochNodesConfig(config *EpochValidators) (*epochNodesConfig, error) {
	return epochValidatorsToEpochNodesConfig(config)
}

func ValidatorArrayToSerializableValidatorArray(validators []Validator) []*SerializableValidator {
	return validatorArrayToSerializableValidatorArray(validators)
}

func SerializableValidatorArrayToValidatorArray(sValidators []*SerializableValidator) ([]Validator, error) {
	return serializableValidatorArrayToValidatorArray(sValidators)
}

func (msc *multiShardCoordinator) CalculateMasks() (uint32, uint32) {
	return msc.calculateMasks()
}

func (msc *multiShardCoordinator) Masks() (uint32, uint32) {
	return msc.maskHigh, msc.maskLow
}

func (g *Genesis) ProcessConfig() error {
	return g.processConfig()
}

func (ns *NodesSetup) ProcessConfig() error {
	return ns.processConfig()
}

func (ns *NodesSetup) ProcessShardAssignment() {
	ns.processShardAssignment()
}

func (ns *NodesSetup) ProcessMetaChainAssigment() {
	ns.processMetaChainAssigment()
}

func (ns *NodesSetup) CreateInitialNodesInfo() {
	ns.createInitialNodesInfo()
}

func (ns *NodesSetup) Eligible() map[uint32][]*NodeInfo {
	return ns.eligible
}

func (ns *NodesSetup) Waiting() map[uint32][]*NodeInfo {
	return ns.waiting
}
