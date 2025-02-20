package nodesCoordinator

import (
	"testing"

	"github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/stretchr/testify/require"
)

func TestNewIndexHashedNodesCoordinatorWithRaterFactory(t *testing.T) {
	factory := NewIndexHashedNodesCoordinatorWithRaterFactory()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(NodesCoordinatorWithRaterFactory), factory)
}

func TestIndexHashedNodesCoordinatorWithRaterFactory_CreateNodesCoordinatorWithRater(t *testing.T) {
	factory := NewIndexHashedNodesCoordinatorWithRaterFactory()

	args := &NodesCoordinatorWithRaterArgs{
		ArgNodesCoordinator: createArguments(),
		ChanceComputer:      &mock.RaterMock{},
	}

	nodesCoordinator, err := factory.CreateNodesCoordinatorWithRater(args)
	require.Nil(t, err)
	require.IsType(t, &indexHashedNodesCoordinatorWithRater{}, nodesCoordinator)

	args.ArgNodesCoordinator.EligibleNodes = nil
	nodesCoordinator, err = factory.CreateNodesCoordinatorWithRater(args)
	require.Nil(t, nodesCoordinator)
	require.Equal(t, ErrNilInputNodesMap, err)
}
