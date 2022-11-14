package requesterscontainer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/requestersContainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func createStubTopicMessageHandlerForMeta(matchStrToErrOnCreate string, matchStrToErrOnRegister string) dataRetriever.TopicMessageHandler {
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

func TestNewMetaRequestersContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.ShardCoordinator = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewMetaRequestersContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Messenger = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewMetaRequestersContainerFactory_NilMarshallerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Marshaller = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshaller, err)
}

func TestNewMetaRequestersContainerFactory_NilMarshallerAndSizeCheckShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Marshaller = nil
	args.SizeCheckDelta = 1
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshaller, err)
}

func TestNewMetaRequestersContainerFactory_NilPreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.PreferredPeersHolder = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilPreferredPeersHolder, err)
}

func TestNewMetaRequestersContainerFactory_NilPeersRatingHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.PeersRatingHandler = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilPeersRatingHandler, err)
}

func TestNewMetaRequestersContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Uint64ByteSliceConverter = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewMetaRequestersContainerFactory_NilOutputAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.OutputAntifloodHandler = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilAntifloodHandler))
}

func TestNewMetaRequestersContainerFactory_NilCurrentNetworkEpochProviderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.CurrentNetworkEpochProvider = nil
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilCurrentNetworkEpochProvider, err)
}

func TestNewMetaRequestersContainerFactory_InvalidNumCrossShardPeersShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.RequesterConfig.NumCrossShardPeers = 0
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewMetaRequestersContainerFactory_InvalidNumTotalPeersShouldErr(t *testing.T) {
	t.Parallel()

	t.Run("NumTotalPeers is lower than NumCrossShardPeers", func(t *testing.T) {
		args := getArgumentsMeta()
		args.RequesterConfig.NumTotalPeers = 0
		rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

		assert.Nil(t, rcf)
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
	})
	t.Run("NumTotalPeers is equal to NumCrossShardPeers", func(t *testing.T) {
		args := getArgumentsMeta()
		args.RequesterConfig.NumTotalPeers = args.RequesterConfig.NumCrossShardPeers
		rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

		assert.Nil(t, rcf)
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
	})
}

func TestNewMetaRequestersContainerFactory_InvalidNumFullHistoryPeersShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.RequesterConfig.NumFullHistoryPeers = 0
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewMetaRequestersContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	rcf, err := requesterscontainer.NewMetaRequestersContainerFactory(args)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(rcf))
	assert.Equal(t, int(args.RequesterConfig.NumTotalPeers), rcf.NumTotalPeers())
	assert.Equal(t, int(args.RequesterConfig.NumCrossShardPeers), rcf.NumCrossShardPeers())
	assert.Equal(t, int(args.RequesterConfig.NumFullHistoryPeers), rcf.NumFullHistoryPeers())
}

// ------- Create

func TestMetaRequestersContainerFactory_CreateRegisterShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Messenger = createStubTopicMessageHandlerForMeta("", factory.ShardBlocksTopic)
	rcf, _ := requesterscontainer.NewMetaRequestersContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestMetaRequestersContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
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

	args := getArgumentsMeta()
	args.ShardCoordinator = shardCoordinator
	rcf, _ := requesterscontainer.NewMetaRequestersContainerFactory(args)

	container, _ := rcf.Create()
	numRequestersShardHeadersForMetachain := noOfShards
	numResolverMetablocks := 1
	numRequestersMiniBlocks := noOfShards + 2
	numRequestersUnsigned := noOfShards + 1
	numRequestersRewards := noOfShards
	numRequestersTxs := noOfShards + 1
	numRequestersTrieNodes := 2
	numRequestersPeerAuth := 1
	numResolverValidatorInfo := 1
	totalRequesters := numRequestersShardHeadersForMetachain + numResolverMetablocks + numRequestersMiniBlocks +
		numRequestersUnsigned + numRequestersTxs + numRequestersTrieNodes + numRequestersRewards + numRequestersPeerAuth + numResolverValidatorInfo

	assert.Equal(t, totalRequesters, container.Len())

	err := rcf.AddShardTrieNodeRequesters(container)
	assert.Nil(t, err)
	assert.Equal(t, totalRequesters+noOfShards, container.Len())
}

func getArgumentsMeta() requesterscontainer.FactoryArgs {
	return requesterscontainer.FactoryArgs{
		RequesterConfig: config.RequesterConfig{
			NumCrossShardPeers:  1,
			NumTotalPeers:       3,
			NumFullHistoryPeers: 3,
		},
		ShardCoordinator:            mock.NewOneShardCoordinatorMock(),
		Messenger:                   createStubTopicMessageHandlerForMeta("", ""),
		Marshaller:                  &mock.MarshalizerMock{},
		Uint64ByteSliceConverter:    &mock.Uint64ByteSliceConverterMock{},
		OutputAntifloodHandler:      &mock.P2PAntifloodHandlerStub{},
		CurrentNetworkEpochProvider: &mock.CurrentNetworkEpochProviderStub{},
		PreferredPeersHolder:        &p2pmocks.PeersHolderStub{},
		PeersRatingHandler:          &p2pmocks.PeersRatingHandlerStub{},
		SizeCheckDelta:              0,
	}
}
