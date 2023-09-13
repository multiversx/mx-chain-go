package resolverscontainer_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/stretchr/testify/require"
)

func TestSovereignNewShardResolversContainerFactory(t *testing.T) {
	t.Parallel()

	factory := resolverscontainer.NewSovereignShardResolversContainerFactoryCreator()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(resolverscontainer.ShardResolversContainerFactoryCreator), factory)
}

func TestSovereignShardResolversContainerFactory_Create(t *testing.T) {
	t.Parallel()

	factory := resolverscontainer.NewSovereignShardResolversContainerFactoryCreator()

	args := getArgumentsShard()
	container, err := factory.CreateShardResolversContainerFactory(args)
	require.Nil(t, err)
	require.Equal(t, "*resolverscontainer.sovereignShardResolversContainerFactory", fmt.Sprintf("%T", container))
}
