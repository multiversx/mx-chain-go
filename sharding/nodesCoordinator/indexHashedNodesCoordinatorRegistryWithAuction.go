package nodesCoordinator

import (
	"fmt"
)

// nodesCoordinatorToRegistryWithAuction will export the nodesCoordinator data to the registry which contains auction list
func (ihnc *indexHashedNodesCoordinator) nodesCoordinatorToRegistryWithAuction() *NodesCoordinatorRegistryWithAuction {
	ihnc.mutNodesConfig.RLock()
	defer ihnc.mutNodesConfig.RUnlock()

	registry := &NodesCoordinatorRegistryWithAuction{
		CurrentEpoch:            ihnc.currentEpoch,
		EpochsConfigWithAuction: make(map[string]*EpochValidatorsWithAuction),
	}

	minEpoch, lastEpoch := ihnc.getMinAndLastEpoch()
	for epoch := minEpoch; epoch <= lastEpoch; epoch++ {
		epochNodesData, ok := ihnc.nodesConfig[epoch]
		if !ok {
			continue
		}

		registry.EpochsConfigWithAuction[fmt.Sprint(epoch)] = epochNodesConfigToEpochValidatorsWithAuction(epochNodesData)
	}

	return registry
}

func epochNodesConfigToEpochValidatorsWithAuction(config *epochNodesConfig) *EpochValidatorsWithAuction {
	result := &EpochValidatorsWithAuction{
		Eligible:       make(map[string]Validators, len(config.eligibleMap)),
		Waiting:        make(map[string]Validators, len(config.waitingMap)),
		Leaving:        make(map[string]Validators, len(config.leavingMap)),
		ShuffledOut:    make(map[string]Validators, len(config.shuffledOutMap)),
		LowWaitingList: config.lowWaitingList,
	}

	for k, v := range config.eligibleMap {
		result.Eligible[fmt.Sprint(k)] = Validators{Data: ValidatorArrayToSerializableValidatorArray(v)}
	}

	for k, v := range config.waitingMap {
		result.Waiting[fmt.Sprint(k)] = Validators{Data: ValidatorArrayToSerializableValidatorArray(v)}
	}

	for k, v := range config.leavingMap {
		result.Leaving[fmt.Sprint(k)] = Validators{Data: ValidatorArrayToSerializableValidatorArray(v)}
	}

	for k, v := range config.shuffledOutMap {
		result.ShuffledOut[fmt.Sprint(k)] = Validators{Data: ValidatorArrayToSerializableValidatorArray(v)}
	}

	return result
}
