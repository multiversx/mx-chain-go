package requesterscontainer_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignShardRequestersContainerFactoryCreator(t *testing.T) {
	t.Parallel()

	factory := requesterscontainer.NewSovereignShardRequestersContainerFactoryCreator()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(requesterscontainer.RequesterContainerFactoryCreator), factory)
}

func TestSovereignShardRequestersContainerFactoryCreator_CreateRequesterContainerFactory(t *testing.T) {
	t.Parallel()

	factory := requesterscontainer.NewSovereignShardRequestersContainerFactoryCreator()

	args := createSovArgs()
	t.Run("should work", func(t *testing.T) {
		container, err := factory.CreateRequesterContainerFactory(args)
		require.Nil(t, err)
		require.IsType(t, container, requesterscontainer.SovereignShardRequestersContainerFactory)
	})
	t.Run("invalid args, should return error", func(t *testing.T) {
		args.Marshaller = nil
		container, err := factory.CreateRequesterContainerFactory(args)
		require.NotNil(t, err)
		require.Nil(t, container)
	})
}
