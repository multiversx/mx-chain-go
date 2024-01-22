package nodesCoordinator

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/state"
)

var _ NodesCoordinatorHelper = (*indexHashedNodesCoordinatorWithRater)(nil)

type indexHashedNodesCoordinatorWithRater struct {
	*indexHashedNodesCoordinator
	*baseNodesCoordinatorWithRater
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
		baseNodesCoordinatorWithRater: &baseNodesCoordinatorWithRater{
			chanceComputer: chanceComputer,
		},
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
func (ihnc *indexHashedNodesCoordinatorWithRater) ComputeAdditionalLeaving(allValidators []*state.ShardValidatorInfo) (map[uint32][]Validator, error) {
	return ihnc.baseNodesCoordinatorWithRater.ComputeAdditionalLeaving(allValidators)
}

// IsInterfaceNil verifies that the underlying value is nil
func (ihnc *indexHashedNodesCoordinatorWithRater) IsInterfaceNil() bool {
	return ihnc == nil
}

// GetChance returns the chance from an actual rating
func (ihnc *indexHashedNodesCoordinatorWithRater) GetChance(rating uint32) uint32 {
	return ihnc.baseNodesCoordinatorWithRater.GetChance(rating)
}

// ValidatorsWeights returns the weights/chances for each given validator
func (ihnc *indexHashedNodesCoordinatorWithRater) ValidatorsWeights(validators []Validator) ([]uint32, error) {
	return ihnc.baseNodesCoordinatorWithRater.ValidatorsWeights(validators)
}

// LoadState loads the nodes coordinator state from the used boot storage
func (ihnc *indexHashedNodesCoordinatorWithRater) LoadState(key []byte) error {
	return ihnc.baseLoadState(key)
}
