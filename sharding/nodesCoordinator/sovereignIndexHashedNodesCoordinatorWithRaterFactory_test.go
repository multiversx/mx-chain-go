package nodesCoordinator

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/stretchr/testify/require"
)

func createSovereignArgs() ArgNodesCoordinator {
	args := createArguments()
	args.EligibleNodes[core.SovereignChainShardId] = createDummyNodesList(1, "eligible")
	return args
}

func TestNewSovereignIndexHashedNodesCoordinatorWithRaterFactory(t *testing.T) {
	factory := NewSovereignIndexHashedNodesCoordinatorWithRaterFactory()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(NodesCoordinatorWithRaterFactory), factory)
}

func TestSovereignIndexHashedNodesCoordinatorWithRaterFactory_CreateNodesCoordinatorWithRater(t *testing.T) {
	factory := NewSovereignIndexHashedNodesCoordinatorWithRaterFactory()

	args := &NodesCoordinatorWithRaterArgs{
		ArgNodesCoordinator: createSovereignArgs(),
		ChanceComputer:      &mock.RaterMock{},
	}
	nodesCoordinator, err := factory.CreateNodesCoordinatorWithRater(args)
	require.Nil(t, err)
	require.IsType(t, &sovereignIndexHashedNodesCoordinatorWithRater{}, nodesCoordinator)

	args.ArgNodesCoordinator.EligibleNodes = nil
	nodesCoordinator, err = factory.CreateNodesCoordinatorWithRater(args)
	require.Nil(t, nodesCoordinator)
	require.Equal(t, ErrNilInputNodesMap, err)
}
