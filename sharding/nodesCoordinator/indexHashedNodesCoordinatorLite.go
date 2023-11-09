package nodesCoordinator

import (
	"github.com/multiversx/mx-chain-go/state"
)

// SetNodesConfigFromValidatorsInfo sets epoch config based on validators list configuration
func (ihnc *indexHashedNodesCoordinator) SetNodesConfigFromValidatorsInfo(epoch uint32, randomness []byte, validatorsInfo []*state.ShardValidatorInfo) error {
	newNodesConfig, err := ihnc.computeNodesConfigFromList(&epochNodesConfig{}, validatorsInfo)
	if err != nil {
		return err
	}

	additionalLeavingMap, err := ihnc.nodesCoordinatorHelper.ComputeAdditionalLeaving(validatorsInfo)
	if err != nil {
		return err
	}

	unStakeLeavingList := ihnc.createSortedListFromMap(newNodesConfig.leavingMap)
	additionalLeavingList := ihnc.createSortedListFromMap(additionalLeavingMap)

	shufflerArgs := ArgsUpdateNodes{
		Eligible:          newNodesConfig.eligibleMap,
		Waiting:           newNodesConfig.waitingMap,
		NewNodes:          newNodesConfig.newList,
		UnStakeLeaving:    unStakeLeavingList,
		AdditionalLeaving: additionalLeavingList,
		Rand:              randomness,
		NbShards:          newNodesConfig.nbShards,
		Epoch:             epoch,
	}

	resUpdateNodes, err := ihnc.shuffler.UpdateNodeLists(shufflerArgs)
	if err != nil {
		return err
	}

	leavingNodesMap, _ := createActuallyLeavingPerShards(
		newNodesConfig.leavingMap,
		additionalLeavingMap,
		resUpdateNodes.Leaving,
	)

	err = ihnc.setNodesPerShards(resUpdateNodes.Eligible, resUpdateNodes.Waiting, leavingNodesMap, epoch)
	if err != nil {
		return err
	}

	ihnc.removeOlderEpochs(epoch, nodesCoordinatorStoredEpochs)

	return nil
}

// IsEpochInConfig checks whether the specified epoch is already in map
func (ihnc *indexHashedNodesCoordinator) IsEpochInConfig(epoch uint32) bool {
	ihnc.mutNodesConfig.RLock()
	_, exists := ihnc.nodesConfig[epoch]
	ihnc.mutNodesConfig.RUnlock()

	return exists
}

func (ihnc *indexHashedNodesCoordinator) removeOlderEpochs(epoch uint32, maxDelta uint32) {
	ihnc.mutNodesConfig.Lock()
	if len(ihnc.nodesConfig) >= int(maxDelta) {
		for currEpoch := range ihnc.nodesConfig {
			if currEpoch <= epoch-maxDelta {
				delete(ihnc.nodesConfig, currEpoch)
			}
		}
	}
	ihnc.mutNodesConfig.Unlock()
}
