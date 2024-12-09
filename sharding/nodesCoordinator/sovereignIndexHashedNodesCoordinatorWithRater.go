package nodesCoordinator

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type sovereignIndexHashedNodesCoordinatorWithRater struct {
	*sovereignIndexHashedNodesCoordinator
	*baseNodesCoordinatorWithRater
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

// IsInterfaceNil verifies that the underlying value is nil
func (ihnc *sovereignIndexHashedNodesCoordinatorWithRater) IsInterfaceNil() bool {
	return ihnc == nil
}
