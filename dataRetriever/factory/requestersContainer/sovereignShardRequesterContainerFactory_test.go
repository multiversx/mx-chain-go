package requesterscontainer_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/stretchr/testify/require"
)

func createSovArgs() requesterscontainer.FactoryArgs {
	args := getArguments()
	args.ShardCoordinator = sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)

	return args
}

func TestNewSovereignShardRequestersContainerFactory(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		sovShardContainer, err := requesterscontainer.NewSovereignShardRequestersContainerFactory(nil)
		require.Equal(t, errors.ErrNilShardRequesterContainerFactory, err)
		require.Nil(t, sovShardContainer)
	})

	t.Run("should work", func(t *testing.T) {
		args := createSovArgs()
		shardContainer, _ := requesterscontainer.NewShardRequestersContainerFactory(args)
		sovShardContainer, err := requesterscontainer.NewSovereignShardRequestersContainerFactory(shardContainer)
		require.Nil(t, err)
		require.False(t, sovShardContainer.IsInterfaceNil())
	})
}

func TestSovereignShardRequestersContainerFactory_Create(t *testing.T) {
	t.Parallel()

	args := createSovArgs()
	shardContainer, _ := requesterscontainer.NewShardRequestersContainerFactory(args)
	sovShardContainer, _ := requesterscontainer.NewSovereignShardRequestersContainerFactory(shardContainer)

	container, err := sovShardContainer.Create()
	require.Nil(t, err)

	shardCoord := args.ShardCoordinator
	extendedHeaderKey := factory.ExtendedHeaderProofTopic + shardCoord.CommunicationIdentifier(shardCoord.SelfId())
	_, err = container.Get(extendedHeaderKey)
	require.Nil(t, err)

	numShards := int(shardCoord.NumberOfShards())
	require.Equal(t, getNumRequesters(numShards)+1, container.Len()) // only one added container for extended header
}

func TestSovereignShardRequestersContainerFactory_NumPeers(t *testing.T) {
	t.Parallel()

	args := createSovArgs()
	shardContainer, _ := requesterscontainer.NewShardRequestersContainerFactory(args)
	sovShardContainer, _ := requesterscontainer.NewSovereignShardRequestersContainerFactory(shardContainer)

	require.Equal(t, sovShardContainer.NumCrossShardPeers(), 0)
	require.Equal(t, int(args.RequesterConfig.NumTotalPeers), sovShardContainer.NumTotalPeers())
}
