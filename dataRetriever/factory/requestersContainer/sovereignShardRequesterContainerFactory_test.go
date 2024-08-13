package requesterscontainer_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/stretchr/testify/require"
)

func createSovArgs() requesterscontainer.FactoryArgs {
	args := getArguments()
	args.ShardCoordinator = sharding.NewSovereignShardCoordinator()

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

	numRequesterSCRs := 1
	numRequesterTxs := 1
	numRequesterRewardTxs := 0
	numRequesterHeaders := 1
	numRequesterMiniBlocks := 1
	numRequesterMetaBlockHeaders := 0
	numRequesterTrieNodes := 2
	numRequesterPeerAuth := 1
	numRequesterValidatorInfo := 1
	numRequesterExtendedHeader := 1
	numRequesters := numRequesterTxs + numRequesterHeaders + numRequesterMiniBlocks + numRequesterMetaBlockHeaders +
		numRequesterSCRs + numRequesterRewardTxs + numRequesterTrieNodes + numRequesterPeerAuth + numRequesterValidatorInfo + numRequesterExtendedHeader

	require.Equal(t, numRequesters, container.Len()) // only one added container for extended header

	topicDelim := "_"
	topicDelimCt := 0
	iterateFunc := func(key string, requester dataRetriever.Requester) bool {
		require.False(t, strings.Contains(strings.ToLower(key), "meta"))

		if strings.Contains(key, topicDelim) {
			keyTokens := strings.Split(key, topicDelim)
			require.Len(t, keyTokens, 2)
			require.Equal(t, fmt.Sprintf("%d", core.SovereignChainShardId), keyTokens[1])
			topicDelimCt++
		}

		return true
	}

	container.Iterate(iterateFunc)
	require.Equal(t, numRequesters-2, topicDelimCt) // without peerAuthentication + validatorInfo topics
}

func TestSovereignShardRequestersContainerFactory_NumPeers(t *testing.T) {
	t.Parallel()

	args := createSovArgs()
	shardContainer, _ := requesterscontainer.NewShardRequestersContainerFactory(args)
	sovShardContainer, _ := requesterscontainer.NewSovereignShardRequestersContainerFactory(shardContainer)

	require.Equal(t, sovShardContainer.NumCrossShardPeers(), 0)
	require.Equal(t, int(args.RequesterConfig.NumTotalPeers), sovShardContainer.NumTotalPeers())
}
