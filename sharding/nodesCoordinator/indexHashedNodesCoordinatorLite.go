package nodesCoordinator

import (
	"github.com/ElrondNetwork/elrond-go/state"
)

// SetNodesConfigFromValidatorsInfo sets epoch config based on validators list configuration
func (ihgs *indexHashedNodesCoordinator) SetNodesConfigFromValidatorsInfo(epoch uint32, randomness []byte, validatorsInfo []*state.ShardValidatorInfo) error {
	copiedPrevious := &epochNodesConfig{}
	newNodesConfig, err := ihgs.computeNodesConfigFromList(copiedPrevious, validatorsInfo)
	if err != nil {
		return err
	}

	additionalLeavingMap, err := ihgs.nodesCoordinatorHelper.ComputeAdditionalLeaving(validatorsInfo)
	if err != nil {
		return err
	}

	unStakeLeavingList := ihgs.createSortedListFromMap(newNodesConfig.leavingMap)
	additionalLeavingList := ihgs.createSortedListFromMap(additionalLeavingMap)

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

	resUpdateNodes, err := ihgs.shuffler.UpdateNodeLists(shufflerArgs)
	if err != nil {
		return err
	}

	leavingNodesMap, _ := createActuallyLeavingPerShards(
		newNodesConfig.leavingMap,
		additionalLeavingMap,
		resUpdateNodes.Leaving,
	)

	err = ihgs.setNodesPerShards(resUpdateNodes.Eligible, resUpdateNodes.Waiting, leavingNodesMap, epoch)
	if err != nil {
		return err
	}

	return nil
}

// IsEpochInConfig checks wether the specified epoch is already in map
func (ihgs *indexHashedNodesCoordinator) IsEpochInConfig(epoch uint32) bool {
	ihgs.mutNodesConfig.Lock()
	_, status := ihgs.nodesConfig[epoch]
	defer ihgs.mutNodesConfig.Unlock()

	return status
}
