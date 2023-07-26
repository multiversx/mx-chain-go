package nodesCoordinator

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

type sovereignIndexHashedNodesCoordinatorWithRater struct {
	*sovereignIndexHashedNodesCoordinator
	chanceComputer ChanceComputer
}

// NewSovereignIndexHashedNodesCoordinatorWithRater creates a new index hashed group selector
func NewSovereignIndexHashedNodesCoordinatorWithRater(
	indexNodesCoordinator *sovereignIndexHashedNodesCoordinator,
	chanceComputer ChanceComputer,
) (*sovereignIndexHashedNodesCoordinatorWithRater, error) {
	if check.IfNil(indexNodesCoordinator) {
		return nil, ErrNilNodesCoordinator
	}
	if check.IfNil(chanceComputer) {
		return nil, ErrNilChanceComputer
	}

	ihncr := &sovereignIndexHashedNodesCoordinatorWithRater{
		sovereignIndexHashedNodesCoordinator: indexNodesCoordinator,
		chanceComputer:                       chanceComputer,
	}

	ihncr.nodesCoordinatorHelper = ihncr

	ihncr.mutNodesConfig.Lock()
	defer ihncr.mutNodesConfig.Unlock()

	nodesConfig, ok := ihncr.nodesConfig[ihncr.currentEpoch]
	if !ok {
		nodesConfig = &epochNodesConfig{}
	}

	nodesConfig.mutNodesMaps.Lock()
	defer nodesConfig.mutNodesMaps.Unlock()

	var err error
	nodesConfig.selectors, err = ihncr.createSelectors(nodesConfig)
	if err != nil {
		return nil, err
	}

	ihncr.epochStartRegistrationHandler.UnregisterHandler(indexNodesCoordinator)
	ihncr.epochStartRegistrationHandler.RegisterHandler(ihncr)
	return ihncr, nil
}

// ComputeAdditionalLeaving - computes the extra leaving validators that have a threshold below the minimum rating
func (ihnc *sovereignIndexHashedNodesCoordinatorWithRater) ComputeAdditionalLeaving(allValidators []*state.ShardValidatorInfo) (map[uint32][]Validator, error) {
	extraLeavingNodesMap := make(map[uint32][]Validator)
	minChances := ihnc.GetChance(0)
	for _, vInfo := range allValidators {
		if vInfo.List == string(common.InactiveList) || vInfo.List == string(common.JailedList) {
			continue
		}
		chances := ihnc.GetChance(vInfo.TempRating)
		if chances < minChances {
			val, err := NewValidator(vInfo.PublicKey, chances, vInfo.Index)
			if err != nil {
				return nil, err
			}
			log.Debug("computed leaving based on rating for validator", "pk", vInfo.GetPublicKey())
			extraLeavingNodesMap[vInfo.ShardId] = append(extraLeavingNodesMap[vInfo.ShardId], val)
		}
	}

	return extraLeavingNodesMap, nil
}

//IsInterfaceNil verifies that the underlying value is nil
func (ihnc *sovereignIndexHashedNodesCoordinatorWithRater) IsInterfaceNil() bool {
	return ihnc == nil
}

// GetChance returns the chance from an actual rating
func (ihnc *sovereignIndexHashedNodesCoordinatorWithRater) GetChance(rating uint32) uint32 {
	return ihnc.chanceComputer.GetChance(rating)
}

// ValidatorsWeights returns the weights/chances for each given validator
func (ihnc *sovereignIndexHashedNodesCoordinatorWithRater) ValidatorsWeights(validators []Validator) ([]uint32, error) {
	minChance := ihnc.GetChance(0)
	weights := make([]uint32, len(validators))

	for i, validatorInShard := range validators {
		weights[i] = validatorInShard.Chances()
		if weights[i] < minChance {
			//default weight if all validators need to be selected
			weights[i] = minChance
		}
	}

	return weights, nil
}

// LoadState loads the nodes coordinator state from the used boot storage
func (ihnc *sovereignIndexHashedNodesCoordinatorWithRater) LoadState(key []byte) error {
	return ihnc.baseLoadState(key)
}
