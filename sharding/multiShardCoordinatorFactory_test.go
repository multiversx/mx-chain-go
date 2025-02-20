package sharding

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestNewMultiShardCoordinatorFactory(t *testing.T) {
	factory := NewMultiShardCoordinatorFactory()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(ShardCoordinatorFactory), factory)
}

func TestMultiShardCoordinatorFactory_CreateShardCoordinator(t *testing.T) {
	factory := NewMultiShardCoordinatorFactory()

	nodesCoordinator, err := factory.CreateShardCoordinator(3, core.MetachainShardId)
	require.Nil(t, err)
	require.IsType(t, &multiShardCoordinator{}, nodesCoordinator)
}
