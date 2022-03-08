package sharding

import "fmt"

// EpochValidatorsWithAuction holds one epoch configuration for a nodes coordinator
type EpochValidatorsWithAuction struct {
	*EpochValidators
	ShuffledOutValidators map[string][]*SerializableValidator `json:"shuffledOutValidators"`
}

// NodesCoordinatorRegistryWithAuction holds the data that can be used to initialize a nodes coordinator
type NodesCoordinatorRegistryWithAuction struct {
	EpochsConfig map[string]*EpochValidatorsWithAuction `json:"epochConfigs"`
	CurrentEpoch uint32                                 `json:"currentEpoch"`
}

// NodesCoordinatorToRegistryWithAuction will export the nodesCoordinator data to the registry which contains auction list
func (ihgs *indexHashedNodesCoordinator) NodesCoordinatorToRegistryWithAuction() *NodesCoordinatorRegistryWithAuction {
	ihgs.mutNodesConfig.RLock()
	defer ihgs.mutNodesConfig.RUnlock()

	registry := &NodesCoordinatorRegistryWithAuction{
		CurrentEpoch: ihgs.currentEpoch,
		EpochsConfig: make(map[string]*EpochValidatorsWithAuction),
	}
	// todo: extract this into a common func with NodesCoordinatorToRegistry
	minEpoch := 0
	lastEpoch := ihgs.getLastEpochConfig()
	if lastEpoch >= nodesCoordinatorStoredEpochs {
		minEpoch = int(lastEpoch) - nodesCoordinatorStoredEpochs + 1
	}

	for epoch := uint32(minEpoch); epoch <= lastEpoch; epoch++ {
		epochNodesData, ok := ihgs.nodesConfig[epoch]
		if !ok {
			continue
		}

		registry.EpochsConfig[fmt.Sprint(epoch)] = epochNodesConfigToEpochValidatorsWithAuction(epochNodesData)
	}

	return registry
}

func epochNodesConfigToEpochValidatorsWithAuction(config *epochNodesConfig) *EpochValidatorsWithAuction {
	result := &EpochValidatorsWithAuction{
		EpochValidators: &EpochValidators{
			EligibleValidators: make(map[string][]*SerializableValidator, len(config.eligibleMap)),
			WaitingValidators:  make(map[string][]*SerializableValidator, len(config.waitingMap)),
			LeavingValidators:  make(map[string][]*SerializableValidator, len(config.leavingMap)),
		},
		ShuffledOutValidators: make(map[string][]*SerializableValidator, len(config.shuffledOutMap)),
	}

	for k, v := range config.eligibleMap {
		result.EligibleValidators[fmt.Sprint(k)] = ValidatorArrayToSerializableValidatorArray(v)
	}

	for k, v := range config.waitingMap {
		result.WaitingValidators[fmt.Sprint(k)] = ValidatorArrayToSerializableValidatorArray(v)
	}

	for k, v := range config.leavingMap {
		result.LeavingValidators[fmt.Sprint(k)] = ValidatorArrayToSerializableValidatorArray(v)
	}

	for k, v := range config.leavingMap {
		result.ShuffledOutValidators[fmt.Sprint(k)] = ValidatorArrayToSerializableValidatorArray(v)
	}

	return result
}
