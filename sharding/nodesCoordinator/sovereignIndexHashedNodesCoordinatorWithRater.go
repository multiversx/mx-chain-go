package nodesCoordinator

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

type sovereignIndexHashedNodesCoordinatorWithRater struct {
	*indexHashedNodesCoordinatorWithRater
}

// NewSovereignIndexHashedNodesCoordinatorWithRater creates a new index hashed group selector
func NewSovereignIndexHashedNodesCoordinatorWithRater(
	indexNodesCoordinator *sovereignIndexHashedNodesCoordinator,
	chanceComputer ChanceComputer,
) (*sovereignIndexHashedNodesCoordinatorWithRater, error) {
	if check.IfNil(indexNodesCoordinator) {
		return nil, errors.New("nil sovereign index hashed nodes coordinator")
	}
	ihnc, err := NewIndexHashedNodesCoordinatorWithRater(indexNodesCoordinator.indexHashedNodesCoordinator, chanceComputer)
	if err != nil {
		return nil, err
	}

	return &sovereignIndexHashedNodesCoordinatorWithRater{
		indexHashedNodesCoordinatorWithRater: ihnc,
	}, nil
}
