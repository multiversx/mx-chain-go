package interceptorscontainer_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/stretchr/testify/require"
)

func createSovInterceptorsContainerArgs() interceptorscontainer.ArgsSovereignShardInterceptorsContainerFactory {
	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.ShardCoordinator = sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
	shardContainer, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	return interceptorscontainer.ArgsSovereignShardInterceptorsContainerFactory{
		ShardContainer:           shardContainer,
		IncomingHeaderSubscriber: &sovereign.IncomingHeaderSubscriberStub{},
	}
}

func TestNewSovereignShardInterceptorsContainerFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil shard interceptor container, should return error", func(t *testing.T) {
		args := createSovInterceptorsContainerArgs()
		args.ShardContainer = nil
		sovContainer, err := interceptorscontainer.NewSovereignShardInterceptorsContainerFactory(args)
		require.Equal(t, errors.ErrNilShardInterceptorsContainerFactory, err)
		require.Nil(t, sovContainer)
	})
	t.Run("nil incoming header subscriber, should return error", func(t *testing.T) {
		args := createSovInterceptorsContainerArgs()
		args.IncomingHeaderSubscriber = nil
		sovContainer, err := interceptorscontainer.NewSovereignShardInterceptorsContainerFactory(args)
		require.Equal(t, errors.ErrNilIncomingHeaderSubscriber, err)
		require.Nil(t, sovContainer)
	})
	t.Run("should work", func(t *testing.T) {
		args := createSovInterceptorsContainerArgs()
		sovContainer, err := interceptorscontainer.NewSovereignShardInterceptorsContainerFactory(args)
		require.Nil(t, err)
		require.False(t, sovContainer.IsInterfaceNil())
	})
}

func TestSovereignShardInterceptorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	args := createSovInterceptorsContainerArgs()

	sovContainer, err := interceptorscontainer.NewSovereignShardInterceptorsContainerFactory(args)
	mainContainer, fullArchiveContainer, err := sovContainer.Create()

	noOfShards := 1
	totalInterceptors := calcNumShardInterceptors(noOfShards) + 1 // one extra for shard extended header
	require.Nil(t, err)
	require.Equal(t, totalInterceptors, mainContainer.Len())
	require.Equal(t, 0, fullArchiveContainer.Len())

	shardCoord := sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
	_, err = mainContainer.Get(factory.ExtendedHeaderProofTopic + shardCoord.CommunicationIdentifier(shardCoord.SelfId()))
	require.Nil(t, err)
}
