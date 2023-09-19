package resolverscontainer_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignShardResolversContainerFactoryCreator(t *testing.T) {
	t.Parallel()

	factory := resolverscontainer.NewSovereignShardResolversContainerFactoryCreator()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(resolverscontainer.ShardResolversContainerFactoryCreator), factory)
}

func TestSovereignShardResolversContainerFactoryCreator_CreateShardResolversContainerFactory(t *testing.T) {
	t.Parallel()

	factory := resolverscontainer.NewSovereignShardResolversContainerFactoryCreator()

	t.Run("nil shard coordinator, should return error", func(t *testing.T) {
		args := getArgumentsShard()
		args.ShardCoordinator = nil
		container, err := factory.CreateShardResolversContainerFactory(args)
		require.NotNil(t, err)
		require.Nil(t, container)
	})

	t.Run("should work", func(t *testing.T) {
		args := getArgumentsShard()
		container, err := factory.CreateShardResolversContainerFactory(args)
		require.Nil(t, err)
		require.Equal(t, "*resolverscontainer.sovereignShardResolversContainerFactory", fmt.Sprintf("%T", container))
	})

}
