package testscommon

import (
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
)

// NodesCoordinatorFactoryMock -
type NodesCoordinatorFactoryMock struct {
	CreateNodesCoordinatorWithRaterCalled func(args *nodesCoordinator.NodesCoordinatorWithRaterArgs) (nodesCoordinator.NodesCoordinator, error)
}

// CreateNodesCoordinatorWithRater -
func (n *NodesCoordinatorFactoryMock) CreateNodesCoordinatorWithRater(args *nodesCoordinator.NodesCoordinatorWithRaterArgs) (nodesCoordinator.NodesCoordinator, error) {
	if n.CreateNodesCoordinatorWithRaterCalled != nil {
		return n.CreateNodesCoordinatorWithRater(args)
	}
	return &shardingMocks.NodesCoordinatorMock{}, nil
}

// IsInterfaceNil -
func (n *NodesCoordinatorFactoryMock) IsInterfaceNil() bool {
	return n == nil
}
