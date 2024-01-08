package requesterscontainer_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewMetaRequestersContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.ShardCoordinator = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewMetaRequestersContainerFactory_NilMainMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.MainMessenger = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilMessenger))
}

func TestNewMetaRequestersContainerFactory_NilFullArchiveMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.FullArchiveMessenger = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilMessenger))
}

func TestNewMetaRequestersContainerFactory_NilMarshallerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.Marshaller = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewMetaRequestersContainerFactory_NilMarshallerAndSizeCheckShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.Marshaller = nil
	args.SizeCheckDelta = 1
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewMetaRequestersContainerFactory_NilMainPreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.MainPreferredPeersHolder = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilPreferredPeersHolder))
}

func TestNewMetaRequestersContainerFactory_NilFullArchivePreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.FullArchivePreferredPeersHolder = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilPreferredPeersHolder))
}

func TestNewMetaRequestersContainerFactory_NilPeersRatingHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.PeersRatingHandler = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilPeersRatingHandler, err)
}

func TestNewMetaRequestersContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.Uint64ByteSliceConverter = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewMetaRequestersContainerFactory_NilOutputAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.OutputAntifloodHandler = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilAntifloodHandler))
}

func TestNewMetaRequestersContainerFactory_NilCurrentNetworkEpochProviderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.CurrentNetworkEpochProvider = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilCurrentNetworkEpochProvider, err)
}

func TestNewMetaRequestersContainerFactory_InvalidNumCrossShardPeersShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.RequesterConfig.NumCrossShardPeers = 0
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewMetaRequestersContainerFactory_InvalidNumTotalPeersShouldErr(t *testing.T) {
	t.Parallel()

	t.Run("NumTotalPeers is lower than NumCrossShardPeers", func(t *testing.T) {
		args := getArguments()
		args.RequesterConfig.NumTotalPeers = 0
		rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

		assert.Nil(t, rcf)
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
	})
	t.Run("NumTotalPeers is equal to NumCrossShardPeers", func(t *testing.T) {
		args := getArguments()
		args.RequesterConfig.NumTotalPeers = args.RequesterConfig.NumCrossShardPeers
		rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

		assert.Nil(t, rcf)
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
	})
}

func TestNewMetaRequestersContainerFactory_InvalidNumFullHistoryPeersShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.RequesterConfig.NumFullHistoryPeers = 0
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewMetaRequestersContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArguments()
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(rcf))
	assert.Equal(t, int(args.RequesterConfig.NumTotalPeers), rcf.NumTotalPeers())
	assert.Equal(t, int(args.RequesterConfig.NumCrossShardPeers), rcf.NumCrossShardPeers())
	assert.Equal(t, int(args.RequesterConfig.NumFullHistoryPeers), rcf.NumFullHistoryPeers())
}

func TestMetaRequestersContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	args := getArguments()
	rcf, _ := requesterscontainer.NewMetaRequestersContainerFactory(args)

	container, err := rcf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestMetaRequestersContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4
	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	args := getArguments()
	args.ShardCoordinator = shardCoordinator
	rcf, _ := requesterscontainer.NewMetaRequestersContainerFactory(args)

	container, _ := rcf.Create()
	numRequestersShardHeadersForMetachain := noOfShards
	numRequesterMetablocks := 1
	numRequestersMiniBlocks := noOfShards + 2
	numRequestersUnsigned := noOfShards + 1
	numRequestersRewards := noOfShards
	numRequestersTxs := noOfShards + 1
	numRequestersTrieNodes := 2
	numRequestersPeerAuth := 1
	numRequesterValidatorInfo := 1
	totalRequesters := numRequestersShardHeadersForMetachain + numRequesterMetablocks + numRequestersMiniBlocks +
		numRequestersUnsigned + numRequestersTxs + numRequestersTrieNodes + numRequestersRewards + numRequestersPeerAuth + numRequesterValidatorInfo

	assert.Equal(t, totalRequesters, container.Len())

	err := rcf.AddShardTrieNodeRequesters(container)
	assert.Nil(t, err)
	assert.Equal(t, totalRequesters+noOfShards, container.Len())
}
