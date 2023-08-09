package sharding

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignShardCoordinatorFactory(t *testing.T) {
	factory := NewSovereignShardCoordinatorFactory()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(ShardCoordinatorFactory), factory)
}

func TestSovereignShardCoordinatorFactory_CreateShardCoordinator(t *testing.T) {
	factory := NewSovereignShardCoordinatorFactory()

	nodesCoordinator, err := factory.CreateShardCoordinator(1, core.SovereignChainShardId)
	require.Nil(t, err)
	require.IsType(t, &sovereignShardCoordinator{}, nodesCoordinator)
}
