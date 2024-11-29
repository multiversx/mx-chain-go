package requesterscontainer_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/stretchr/testify/require"
)

func TestNewShardRequestersContainerFactoryCreator(t *testing.T) {
	t.Parallel()

	factory := requesterscontainer.NewShardRequestersContainerFactoryCreator()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(requesterscontainer.RequesterContainerFactoryCreator), factory)
}

func TestShardRequestersContainerFactoryCreator_CreateRequesterContainerFactory(t *testing.T) {
	t.Parallel()

	factory := requesterscontainer.NewShardRequestersContainerFactoryCreator()

	args := getArguments()
	container, err := factory.CreateRequesterContainerFactory(args)
	require.Nil(t, err)
	require.IsType(t, container, requesterscontainer.ShardRequestersContainerFactory)
}
