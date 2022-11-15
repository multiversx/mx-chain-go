package requesterscontainer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/requestersContainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errExpected = errors.New("expected error")

func createStubTopicMessageHandlerForShard(matchStrToErrOnCreate string) dataRetriever.TopicMessageHandler {
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
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewShardRequestersContainerFactory_NilMarshallerAndSizeShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Marshaller = nil
	args.SizeCheckDelta = 1
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
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

	numRequesterSCRs := noOfShards + 1
	numRequesterTxs := noOfShards + 1
	numRequesterRewardTxs := 1
	numRequesterHeaders := 1
	numRequesterMiniBlocks := noOfShards + 2
	numRequesterMetaBlockHeaders := 1
	numRequesterTrieNodes := 1
	numRequesterPeerAuth := 1
	numRequesterValidatorInfo := 1
	totalRequesters := numRequesterTxs + numRequesterHeaders + numRequesterMiniBlocks + numRequesterMetaBlockHeaders +
		numRequesterSCRs + numRequesterRewardTxs + numRequesterTrieNodes + numRequesterPeerAuth + numRequesterValidatorInfo

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
		Messenger:                   createStubTopicMessageHandlerForShard(""),
		Marshaller:                  &mock.MarshalizerMock{},
		Uint64ByteSliceConverter:    &mock.Uint64ByteSliceConverterMock{},
		OutputAntifloodHandler:      &mock.P2PAntifloodHandlerStub{},
		CurrentNetworkEpochProvider: &mock.CurrentNetworkEpochProviderStub{},
		PreferredPeersHolder:        &p2pmocks.PeersHolderStub{},
		PeersRatingHandler:          &p2pmocks.PeersRatingHandlerStub{},
		SizeCheckDelta:              0,
	}
}
