package nodesCoordinator

// NodesCoordinatorWithRaterArgs is a struct placeholder for necessary arguments to create a nodes coordinator with rater
type NodesCoordinatorWithRaterArgs struct {
	ArgNodesCoordinator ArgNodesCoordinator
	ChanceComputer      ChanceComputer
}

type indexHashedNodesCoordinatorWithRaterFactory struct {
}

// NewIndexHashedNodesCoordinatorWithRaterFactory creates a nodes coordinator factory for regular chain running(shards + metachain)
func NewIndexHashedNodesCoordinatorWithRaterFactory() NodesCoordinatorWithRaterFactory {
	return &indexHashedNodesCoordinatorWithRaterFactory{}
}

// CreateNodesCoordinatorWithRater creates a nodes coordinator for regular chain running(shards + metachain)
func (ncf *indexHashedNodesCoordinatorWithRaterFactory) CreateNodesCoordinatorWithRater(args *NodesCoordinatorWithRaterArgs) (NodesCoordinator, error) {
	baseNodesCoordinator, err := NewIndexHashedNodesCoordinator(args.ArgNodesCoordinator)
	if err != nil {
		return nil, err
	}

	return NewIndexHashedNodesCoordinatorWithRater(baseNodesCoordinator, args.ChanceComputer)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ncf *indexHashedNodesCoordinatorWithRaterFactory) IsInterfaceNil() bool {
	return ncf == nil
}
