package sharding

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type indexHashedNodesCoordinatorWithRater struct {
	*indexHashedNodesCoordinator
	chanceComputer ChanceComputer
}

// NewIndexHashedNodesCoordinatorWithRater creates a new index hashed group selector
func NewIndexHashedNodesCoordinatorWithRater(
	indexNodesCoordinator *indexHashedNodesCoordinator,
	chanceComputer ChanceComputer,
) (*indexHashedNodesCoordinatorWithRater, error) {
	if check.IfNil(indexNodesCoordinator) {
		return nil, ErrNilNodesCoordinator
	}
	if check.IfNil(chanceComputer) {
		return nil, ErrNilChanceComputer
	}

	ihncr := &indexHashedNodesCoordinatorWithRater{
		indexHashedNodesCoordinator: indexNodesCoordinator,
		chanceComputer:              chanceComputer,
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

// ComputeLeaving - computes the validators that have a threshold below the minimum rating
func (ihgs *indexHashedNodesCoordinatorWithRater) ComputeLeaving(allValidators []*state.ShardValidatorInfo) ([]Validator, error) {
	leavingList := make([]Validator, 0)
	minChances := ihgs.GetChance(0)
	for _, vInfo := range allValidators {
		if vInfo.List == string(core.InactiveList) {
			log.Debug("inactive validator", "pk", vInfo.GetPublicKey())
			continue
		}

		chances := ihgs.GetChance(vInfo.TempRating)
		if chances < minChances || vInfo.List == string(core.LeavingList) {
			val, err := NewValidator(vInfo.PublicKey, chances, vInfo.Index)
			if err != nil {
				return nil, err
			}

			leavingList = append(leavingList, val)
		}
	}

	return leavingList, nil
}

//IsInterfaceNil verifies that the underlying value is nil
func (ihgs *indexHashedNodesCoordinatorWithRater) IsInterfaceNil() bool {
	return ihgs == nil
}

// GetChance returns the chance from an actual rating
func (ihgs *indexHashedNodesCoordinatorWithRater) GetChance(rating uint32) uint32 {
	return ihgs.chanceComputer.GetChance(rating)
}

// ValidatorsWeights returns the weights/chances for each given validator
func (ihgs *indexHashedNodesCoordinatorWithRater) ValidatorsWeights(validators []Validator) ([]uint32, error) {
	minChance := ihgs.GetChance(0)
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
