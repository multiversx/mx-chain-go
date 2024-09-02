package interceptorscontainer_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/stretchr/testify/require"
)

func createSovInterceptorsContainerArgs() interceptorscontainer.ArgsSovereignShardInterceptorsContainerFactory {
	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.ShardCoordinator = sharding.NewSovereignShardCoordinator()
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

	sovContainer, _ := interceptorscontainer.NewSovereignShardInterceptorsContainerFactory(args)
	mainContainer, fullArchiveContainer, err := sovContainer.Create()
	require.Nil(t, err)

	numInterceptorTxs := 1
	numInterceptorsUnsignedTxs := 1
	numInterceptorsRewardTxs := 0
	numInterceptorHeaders := 1
	numInterceptorMiniBlocks := 1
	numInterceptorMetachainHeaders := 0
	numInterceptorTrieNodes := 2
	numInterceptorPeerAuth := 1
	numInterceptorHeartbeat := 1
	numInterceptorsShardValidatorInfo := 1
	numInterceptorValidatorInfo := 1
	numInterceptorExtendedHeader := 1
	totalInterceptors := numInterceptorTxs + numInterceptorsUnsignedTxs + numInterceptorsRewardTxs +
		numInterceptorHeaders + numInterceptorMiniBlocks + numInterceptorMetachainHeaders + numInterceptorTrieNodes +
		numInterceptorPeerAuth + numInterceptorHeartbeat + numInterceptorsShardValidatorInfo + numInterceptorValidatorInfo +
		numInterceptorExtendedHeader

	require.Equal(t, totalInterceptors, mainContainer.Len())
	require.Equal(t, 0, fullArchiveContainer.Len())

	topicDelim := "_"
	topicDelimCt := 0
	iterateFunc := func(key string, interceptor process.Interceptor) bool {
		require.False(t, strings.Contains(strings.ToLower(key), "meta"))
		if strings.Contains(key, topicDelim) {
			keyTokens := strings.Split(key, topicDelim)
			require.Len(t, keyTokens, 2)
			require.Equal(t, fmt.Sprintf("%d", core.SovereignChainShardId), keyTokens[1])
			topicDelimCt++
		}

		return true
	}

	mainContainer.Iterate(iterateFunc)

	shardCoord := sharding.NewSovereignShardCoordinator()
	_, err = mainContainer.Get(factory.ExtendedHeaderProofTopic + shardCoord.CommunicationIdentifier(shardCoord.SelfId()))
	require.Nil(t, err)
}
