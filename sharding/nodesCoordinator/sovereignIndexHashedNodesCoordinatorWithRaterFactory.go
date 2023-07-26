package nodesCoordinator

type sovereignIndexHashedNodesCoordinatorWithRaterFactory struct {
}

// NewSovereignIndexHashedNodesCoordinatorWithRaterFactory creates a nodes coordinator factory for a sovereign chain
func NewSovereignIndexHashedNodesCoordinatorWithRaterFactory() NodesCoordinatorWithRaterFactory {
	return &sovereignIndexHashedNodesCoordinatorWithRaterFactory{}
}

// CreateNodesCoordinatorWithRater creates a nodes coordinator for a sovereign chain
func (ncf *sovereignIndexHashedNodesCoordinatorWithRaterFactory) CreateNodesCoordinatorWithRater(args *NodesCoordinatorWithRaterArgs) (NodesCoordinator, error) {
	baseNodesCoordinator, err := NewSovereignIndexHashedNodesCoordinator(args.ArgNodesCoordinator)
	if err != nil {
		return nil, err
	}

	return NewSovereignIndexHashedNodesCoordinatorWithRater(baseNodesCoordinator, args.ChanceComputer)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ncf *sovereignIndexHashedNodesCoordinatorWithRaterFactory) IsInterfaceNil() bool {
	return ncf == nil
}
