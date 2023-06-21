package requesterscontainer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errExpected = errors.New("expected error")

func createStubTopicMessageHandler(matchStrToErrOnCreate string) dataRetriever.TopicMessageHandler {
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

func TestNewShardRequestersContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.ShardCoordinator = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewShardRequestersContainerFactory_NilMainMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.MainMessenger = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilMessenger))
}

func TestNewShardRequestersContainerFactory_NilFullArchiveMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.FullArchiveMessenger = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilMessenger))
}

func TestNewShardRequestersContainerFactory_NilMarshallerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.Marshaller = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewShardRequestersContainerFactory_NilMarshallerAndSizeShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.Marshaller = nil
	args.SizeCheckDelta = 1
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewShardRequestersContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.Uint64ByteSliceConverter = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewShardRequestersContainerFactory_NilMainPreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.MainPreferredPeersHolder = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilPreferredPeersHolder))
}

func TestNewShardRequestersContainerFactory_NilFullArchivePreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.FullArchivePreferredPeersHolder = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilPreferredPeersHolder))
}

func TestNewShardRequestersContainerFactory_NilMainPeersRatingHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.MainPeersRatingHandler = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilPeersRatingHandler))
}

func TestNewShardRequestersContainerFactory_NilFullArchivePeersRatingHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.FullArchivePeersRatingHandler = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilPeersRatingHandler))
}

func TestNewShardRequestersContainerFactory_InvalidNumTotalPeersShouldErr(t *testing.T) {
	t.Parallel()

	t.Run("NumTotalPeers is lower than NumCrossShardPeers", func(t *testing.T) {
		t.Parallel()

		args := getArguments()
		args.RequesterConfig.NumTotalPeers = 0
		rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

		assert.Nil(t, rcf)
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
	})
	t.Run("NumTotalPeers is equal to NumCrossShardPeers", func(t *testing.T) {
		t.Parallel()

		args := getArguments()
		args.RequesterConfig.NumTotalPeers = args.RequesterConfig.NumCrossShardPeers
		rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

		assert.Nil(t, rcf)
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
	})
}

func TestNewShardRequestersContainerFactory_InvalidNumCrossShardPeersShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.RequesterConfig.NumCrossShardPeers = 0
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewShardRequestersContainerFactory_InvalidNumFullHistoryPeersShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.RequesterConfig.NumFullHistoryPeers = 0
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewShardRequestersContainerFactory_NilOutputAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.OutputAntifloodHandler = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilAntifloodHandler))
}

func TestNewShardRequestersContainerFactory_NilCurrentNetworkEpochProviderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.CurrentNetworkEpochProvider = nil
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilCurrentNetworkEpochProvider, err)
}

func TestNewShardRequestersContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArguments()
	rcf, err := requesterscontainer.NewShardRequestersContainerFactory(args)

	assert.NotNil(t, rcf)
	assert.Nil(t, err)
	require.False(t, rcf.IsInterfaceNil())
	assert.Equal(t, int(args.RequesterConfig.NumTotalPeers), rcf.NumTotalPeers())
	assert.Equal(t, int(args.RequesterConfig.NumCrossShardPeers), rcf.NumCrossShardPeers())
	assert.Equal(t, int(args.RequesterConfig.NumFullHistoryPeers), rcf.NumFullHistoryPeers())
}

func TestShardRequestersContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	args := getArguments()
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

	args := getArguments()
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

func getArguments() requesterscontainer.FactoryArgs {
	return requesterscontainer.FactoryArgs{
		RequesterConfig: config.RequesterConfig{
			NumCrossShardPeers:  1,
			NumTotalPeers:       3,
			NumFullHistoryPeers: 3,
		},
		ShardCoordinator:                mock.NewOneShardCoordinatorMock(),
		MainMessenger:                   createStubTopicMessageHandler(""),
		FullArchiveMessenger:            createStubTopicMessageHandler(""),
		Marshaller:                      &mock.MarshalizerMock{},
		Uint64ByteSliceConverter:        &mock.Uint64ByteSliceConverterMock{},
		OutputAntifloodHandler:          &mock.P2PAntifloodHandlerStub{},
		CurrentNetworkEpochProvider:     &mock.CurrentNetworkEpochProviderStub{},
		MainPreferredPeersHolder:        &p2pmocks.PeersHolderStub{},
		FullArchivePreferredPeersHolder: &p2pmocks.PeersHolderStub{},
		MainPeersRatingHandler:          &p2pmocks.PeersRatingHandlerStub{},
		FullArchivePeersRatingHandler:   &p2pmocks.PeersRatingHandlerStub{},
		SizeCheckDelta:                  0,
	}
}
