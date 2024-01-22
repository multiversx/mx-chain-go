package resolverscontainer_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/stretchr/testify/require"
)

func TestNewShardResolversContainerFactoryCreator(t *testing.T) {
	t.Parallel()

	factory := resolverscontainer.NewShardResolversContainerFactoryCreator()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(resolverscontainer.ShardResolversContainerFactoryCreator), factory)
}

func TestShardResolversContainerFactoryCreator_CreateShardResolversContainerFactory(t *testing.T) {
	t.Parallel()

	factory := resolverscontainer.NewShardResolversContainerFactoryCreator()

	args := getArgumentsShard()
	container, err := factory.CreateShardResolversContainerFactory(args)
	require.Nil(t, err)
	require.Equal(t, "*resolverscontainer.shardResolversContainerFactory", fmt.Sprintf("%T", container))
}
