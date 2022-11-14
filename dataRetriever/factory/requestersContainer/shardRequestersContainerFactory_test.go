package requesterscontainer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	requesterscontainer "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/requestersContainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errExpected = errors.New("expected error")

func createStubTopicMessageHandlerForShard(matchStrToErrOnCreate string, matchStrToErrOnRegister string) dataRetriever.TopicMessageHandler {
	tmhs := mock.NewTopicMessageHandlerStub()

	tmhs.CreateTopicCalled = func(name string, createChannelForTopic bool) error {
		if matchStrToErrOnCreate == "" {
			return nil
		}

		if strings.Contains(name, matchStrToErrOnCreate) {
			return errExpected
		}

		return nil
	}

	tmhs.RegisterMessageProcessorCalled = func(topic string, identifier string, handler p2p.MessageProcessor) error {
		if matchStrToErrOnRegister == "" {
			return nil
		}

		if strings.Contains(topic, matchStrToErrOnRegister) {
			return errExpected
		}

		return nil
	}

	return tmhs
}

// ------- NewRequestersContainerFactory

func TestNewShardRequestersContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.ShardCoordinator = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewShardRequestersContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewShardRequestersContainerFactory_NilMarshallerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Marshaller = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshaller, err)
}

func TestNewShardRequestersContainerFactory_NilMarshallerAndSizeShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Marshaller = nil
	args.SizeCheckDelta = 1
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshaller, err)
}

func TestNewShardRequestersContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Uint64ByteSliceConverter = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewShardRequestersContainerFactory_NilPreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.PreferredPeersHolder = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilPreferredPeersHolder, err)
}

func TestNewShardRequestersContainerFactory_NilPeersRatingHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.PeersRatingHandler = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilPeersRatingHandler, err)
}

func TestNewShardRequestersContainerFactory_InvalidNumTotalPeersShouldErr(t *testing.T) {
	t.Parallel()

	t.Run("NumTotalPeers is lower than NumCrossShardPeers", func(t *testing.T) {
		t.Parallel()

		args := getArgumentsShard()
		args.RequesterConfig.NumTotalPeers = 0
		rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

		assert.Nil(t, rcf)
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
	})
	t.Run("NumTotalPeers is equal to NumCrossShardPeers", func(t *testing.T) {
		t.Parallel()

		args := getArgumentsShard()
		args.RequesterConfig.NumTotalPeers = args.RequesterConfig.NumCrossShardPeers
		rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

		assert.Nil(t, rcf)
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
	})
}

func TestNewShardRequestersContainerFactory_InvalidNumCrossShardPeersShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.RequesterConfig.NumCrossShardPeers = 0
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewShardRequestersContainerFactory_InvalidNumFullHistoryPeersShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.RequesterConfig.NumFullHistoryPeers = 0
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewShardRequestersContainerFactory_NilOutputAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.OutputAntifloodHandler = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilAntifloodHandler))
}

func TestNewShardRequestersContainerFactory_NilCurrentNetworkEpochProviderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.CurrentNetworkEpochProvider = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilCurrentNetworkEpochProvider, err)
}

func TestNewShardRequestersContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.NotNil(t, rcf)
	assert.Nil(t, err)
	require.False(t, rcf.IsInterfaceNil())
	assert.Equal(t, int(args.RequesterConfig.NumTotalPeers), rcf.NumTotalPeers())
	assert.Equal(t, int(args.RequesterConfig.NumCrossShardPeers), rcf.NumCrossShardPeers())
	assert.Equal(t, int(args.RequesterConfig.NumFullHistoryPeers), rcf.NumFullHistoryPeers())
}

// ------- Create

func TestShardRequestersContainerFactory_CreateRegisterTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", factory.TransactionTopic)
	rcf, _ := requesterscontainer.NewShardRequestersContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardRequestersContainerFactory_CreateRegisterHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", factory.ShardBlocksTopic)
	rcf, _ := requesterscontainer.NewShardRequestersContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardRequestersContainerFactory_CreateRegisterMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", factory.MiniBlocksTopic)
	rcf, _ := requesterscontainer.NewShardRequestersContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardRequestersContainerFactory_CreateRegisterTrieNodesFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", factory.AccountTrieNodesTopic)
	rcf, _ := requesterscontainer.NewShardRequestersContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardRequestersContainerFactory_CreateRegisterPeerAuthenticationShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", common.PeerAuthenticationTopic)
	rcf, _ := requesterscontainer.NewShardRequestersContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardRequestersContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	rcf, _ := requesterscontainer.NewShardRequestersContainerFactory(args)

	container, err := rcf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestShardRequestersContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	args := getArgumentsShard()
	args.ShardCoordinator = shardCoordinator
	rcf, _ := requesterscontainer.NewShardRequestersContainerFactory(args)

	container, _ := rcf.Create()

	numResolverSCRs := noOfShards + 1
	numResolverTxs := noOfShards + 1
	numResolverRewardTxs := 1
	numResolverHeaders := 1
	numResolverMiniBlocks := noOfShards + 2
	numResolverMetaBlockHeaders := 1
	numResolverTrieNodes := 1
	numResolverPeerAuth := 1
	numResolverValidatorInfo := 1
	totalRequesters := numResolverTxs + numResolverHeaders + numResolverMiniBlocks + numResolverMetaBlockHeaders +
		numResolverSCRs + numResolverRewardTxs + numResolverTrieNodes + numResolverPeerAuth + numResolverValidatorInfo

	assert.Equal(t, totalRequesters, container.Len())
}

func getArgumentsShard() requesterscontainer.FactoryArgs {
	return requesterscontainer.FactoryArgs{
		RequesterConfig: config.RequesterConfig{
			NumCrossShardPeers:  1,
			NumTotalPeers:       3,
			NumFullHistoryPeers: 3,
		},
		ShardCoordinator:            mock.NewOneShardCoordinatorMock(),
		Messenger:                   createStubTopicMessageHandlerForShard("", ""),
		Marshaller:                  &mock.MarshalizerMock{},
		Uint64ByteSliceConverter:    &mock.Uint64ByteSliceConverterMock{},
		OutputAntifloodHandler:      &mock.P2PAntifloodHandlerStub{},
		CurrentNetworkEpochProvider: &mock.CurrentNetworkEpochProviderStub{},
		PreferredPeersHolder:        &p2pmocks.PeersHolderStub{},
		PeersRatingHandler:          &p2pmocks.PeersRatingHandlerStub{},
		SizeCheckDelta:              0,
	}
}
