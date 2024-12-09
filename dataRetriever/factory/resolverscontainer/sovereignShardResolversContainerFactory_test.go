package resolverscontainer_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/factory"
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

	args.ShardCoordinator = sharding.NewSovereignShardCoordinator()
	shardContainer, _ := resolverscontainer.NewShardResolversContainerFactory(args)
	sovContainer, _ := resolverscontainer.NewSovereignShardResolversContainerFactory(shardContainer)

	container, _ := sovContainer.Create()

	numResolverSCRs := 1
	numResolverTxs := 1
	numResolverRewardTxs := 0
	numResolverHeaders := 1
	numResolverMiniBlocks := 1
	numResolverMetaBlockHeaders := 0
	numResolverTrieNodes := 2
	numResolverPeerAuth := 1
	numResolverValidatorInfo := 1
	numResolverExtendedHeader := 1
	totalResolvers := numResolverTxs + numResolverHeaders + numResolverMiniBlocks + numResolverMetaBlockHeaders +
		numResolverSCRs + numResolverRewardTxs + numResolverTrieNodes + numResolverPeerAuth + numResolverValidatorInfo + numResolverExtendedHeader

	require.Equal(t, totalResolvers, container.Len())
	require.Equal(t, totalResolvers, registerMainCnt)
	require.Equal(t, totalResolvers, registerFullArchiveCnt)

	sovShardIDStr := fmt.Sprintf("_%d", core.SovereignChainShardId)
	allKeys := map[string]struct{}{
		factory.TransactionTopic + sovShardIDStr:         {},
		factory.UnsignedTransactionTopic + sovShardIDStr: {},
		factory.ShardBlocksTopic + sovShardIDStr:         {},
		factory.MiniBlocksTopic + sovShardIDStr:          {},
		factory.ValidatorTrieNodesTopic + sovShardIDStr:  {},
		factory.AccountTrieNodesTopic + sovShardIDStr:    {},
		common.PeerAuthenticationTopic:                   {},
		common.ValidatorInfoTopic + sovShardIDStr:        {},
		factory.ExtendedHeaderProofTopic + sovShardIDStr: {},
	}

	iterateFunc := func(key string, resolver dataRetriever.Resolver) bool {
		require.False(t, strings.Contains(strings.ToLower(key), "meta"))
		delete(allKeys, key)
		return true
	}

	container.Iterate(iterateFunc)
	require.Empty(t, allKeys)
}
