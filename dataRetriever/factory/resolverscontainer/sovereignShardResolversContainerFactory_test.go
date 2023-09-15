package resolverscontainer_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignShardResolversContainerFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil arg, should return error", func(t *testing.T) {
		factory, err := resolverscontainer.NewSovereignShardResolversContainerFactory(nil)
		require.Equal(t, errors.ErrNilShardResolversContainerFactory, err)
		require.Nil(t, factory)
	})
	t.Run("should work", func(t *testing.T) {
		args := getArgumentsShard()
		shardContainer, _ := resolverscontainer.NewShardResolversContainerFactory(args)
		factory, err := resolverscontainer.NewSovereignShardResolversContainerFactory(shardContainer)
		require.Nil(t, err)
		require.False(t, factory.IsInterfaceNil())
	})

}

func TestSovereignShardResolversContainerFactory_Create(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	registerMainCnt := 0
	args.MainMessenger = &p2pmocks.MessengerStub{
		RegisterMessageProcessorCalled: func(topic string, identifier string, handler p2p.MessageProcessor) error {
			registerMainCnt++
			return nil
		},
	}
	registerFullArchiveCnt := 0
	args.FullArchiveMessenger = &p2pmocks.MessengerStub{
		RegisterMessageProcessorCalled: func(topic string, identifier string, handler p2p.MessageProcessor) error {
			registerFullArchiveCnt++
			return nil
		},
	}

	args.ShardCoordinator = sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
	shardContainer, _ := resolverscontainer.NewShardResolversContainerFactory(args)
	sovContainer, _ := resolverscontainer.NewSovereignShardResolversContainerFactory(shardContainer)

	container, _ := sovContainer.Create()
	totalResolvers := getNumShardResolvers(int(args.ShardCoordinator.NumberOfShards())) + 1 // one extra topic fo extended header
	require.Equal(t, totalResolvers, container.Len())
	require.Equal(t, totalResolvers, registerMainCnt)
	require.Equal(t, totalResolvers, registerFullArchiveCnt)
}
