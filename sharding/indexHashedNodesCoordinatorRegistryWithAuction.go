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

func (ncr *NodesCoordinatorRegistryWithAuction) GetCurrentEpoch() uint32 {
	return ncr.CurrentEpoch
}

func (ncr *NodesCoordinatorRegistryWithAuction) GetEpochsConfig() map[string]EpochValidatorsHandler {
	ret := make(map[string]EpochValidatorsHandler)
	for epoch, config := range ncr.EpochsConfig {
		ret[epoch] = config
	}

	return ret
}

func (ncr *NodesCoordinatorRegistryWithAuction) SetCurrentEpoch(epoch uint32) {
	ncr.CurrentEpoch = epoch
}

func (ncr *NodesCoordinatorRegistryWithAuction) SetEpochsConfig(epochsConfig map[string]EpochValidatorsHandler) {
	ncr.EpochsConfig = make(map[string]*EpochValidatorsWithAuction)

	for epoch, config := range epochsConfig {
		ncr.EpochsConfig[epoch] = &EpochValidatorsWithAuction{
			EpochValidators: &EpochValidators{
				EligibleValidators: config.GetEligibleValidators(),
				WaitingValidators:  config.GetWaitingValidators(),
				LeavingValidators:  config.GetLeavingValidators(),
			},
			ShuffledOutValidators: nil,
		}
	}
}

// nodesCoordinatorToRegistryWithAuction will export the nodesCoordinator data to the registry which contains auction list
func (ihgs *indexHashedNodesCoordinator) nodesCoordinatorToRegistryWithAuction() *NodesCoordinatorRegistryWithAuction {
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
		EpochValidators:       epochNodesConfigToEpochValidators(config),
		ShuffledOutValidators: make(map[string][]*SerializableValidator, len(config.shuffledOutMap)),
	}

	for k, v := range config.leavingMap {
		result.ShuffledOutValidators[fmt.Sprint(k)] = ValidatorArrayToSerializableValidatorArray(v)
	}

	return result
}
