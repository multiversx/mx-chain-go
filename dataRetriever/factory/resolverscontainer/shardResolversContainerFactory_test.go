package resolverscontainer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	triesFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
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

func createDataPoolsForShard() dataRetriever.PoolsHolder {
	pools := dataRetrieverMock.NewPoolsHolderStub()
	pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.PeerChangesBlocksCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.UnsignedTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	pools.RewardTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}

	return pools
}

func createStoreForShard() dataRetriever.StorageService {
	return &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{}, nil
		},
	}
}

func createTriesHolderForShard() common.TriesHolder {
	triesHolder := state.NewDataTriesHolder()
	triesHolder.Put([]byte(triesFactory.UserAccountTrie), &trieMock.TrieStub{})
	triesHolder.Put([]byte(triesFactory.PeerAccountTrie), &trieMock.TrieStub{})
	return triesHolder
}

// ------- NewResolversContainerFactory

func TestNewShardResolversContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.ShardCoordinator = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewShardResolversContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewShardResolversContainerFactory_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Store = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilStore, err)
}

func TestNewShardResolversContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Marshalizer = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewShardResolversContainerFactory_NilMarshalizerAndSizeShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Marshalizer = nil
	args.SizeCheckDelta = 1
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewShardResolversContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.DataPools = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPoolHolder, err)
}

func TestNewShardResolversContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Uint64ByteSliceConverter = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewShardResolversContainerFactory_NilDataPackerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.DataPacker = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
}

func TestNewShardResolversContainerFactory_NilPreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.PreferredPeersHolder = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilPreferredPeersHolder, err)
}

func TestNewShardResolversContainerFactory_NilPeersRatingHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.PeersRatingHandler = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilPeersRatingHandler, err)
}

func TestNewShardResolversContainerFactory_NilTriesContainerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.TriesContainer = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilTrieDataGetter, err)
}

func TestNewShardResolversContainerFactory_InvalidNumTotalPeersShouldErr(t *testing.T) {
	t.Parallel()

	t.Run("NumTotalPeers is lower than NumCrossShardPeers", func(t *testing.T) {
		t.Parallel()

		args := getArgumentsShard()
		args.ResolverConfig.NumTotalPeers = 0
		rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

		assert.Nil(t, rcf)
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
	})
	t.Run("NumTotalPeers is equal to NumCrossShardPeers", func(t *testing.T) {
		t.Parallel()

		args := getArgumentsShard()
		args.ResolverConfig.NumTotalPeers = args.ResolverConfig.NumCrossShardPeers
		rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

		assert.Nil(t, rcf)
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
	})
}

func TestNewShardResolversContainerFactory_InvalidNumCrossShardPeersShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.ResolverConfig.NumCrossShardPeers = 0
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewShardResolversContainerFactory_InvalidNumFullHistoryPeersShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.ResolverConfig.NumFullHistoryPeers = 0
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewShardResolversContainerFactory_NilInputAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.InputAntifloodHandler = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilAntifloodHandler))
}

func TestNewShardResolversContainerFactory_NilOutputAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.OutputAntifloodHandler = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilAntifloodHandler))
}

func TestNewShardResolversContainerFactory_NilCurrentNetworkEpochProviderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.CurrentNetworkEpochProvider = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilCurrentNetworkEpochProvider, err)
}

func TestNewShardResolversContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.NotNil(t, rcf)
	assert.Nil(t, err)
	require.False(t, rcf.IsInterfaceNil())
	assert.Equal(t, int(args.ResolverConfig.NumTotalPeers), rcf.NumTotalPeers())
	assert.Equal(t, int(args.ResolverConfig.NumCrossShardPeers), rcf.NumCrossShardPeers())
	assert.Equal(t, int(args.ResolverConfig.NumFullHistoryPeers), rcf.NumFullHistoryPeers())
}

// ------- Create

func TestShardResolversContainerFactory_CreateRegisterTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", factory.TransactionTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", factory.ShardBlocksTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", factory.MiniBlocksTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterTrieNodesFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", factory.AccountTrieNodesTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterPeerAuthenticationShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", common.PeerAuthenticationTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestShardResolversContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	args := getArgumentsShard()
	args.ShardCoordinator = shardCoordinator
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

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
	totalResolvers := numResolverTxs + numResolverHeaders + numResolverMiniBlocks + numResolverMetaBlockHeaders +
		numResolverSCRs + numResolverRewardTxs + numResolverTrieNodes + numResolverPeerAuth + numResolverValidatorInfo

	assert.Equal(t, totalResolvers, container.Len())
}

func getArgumentsShard() resolverscontainer.FactoryArgs {
	return resolverscontainer.FactoryArgs{
		ShardCoordinator:            mock.NewOneShardCoordinatorMock(),
		Messenger:                   createStubTopicMessageHandlerForShard("", ""),
		Store:                       createStoreForShard(),
		Marshalizer:                 &mock.MarshalizerMock{},
		DataPools:                   createDataPoolsForShard(),
		Uint64ByteSliceConverter:    &mock.Uint64ByteSliceConverterMock{},
		DataPacker:                  &mock.DataPackerStub{},
		TriesContainer:              createTriesHolderForShard(),
		SizeCheckDelta:              0,
		InputAntifloodHandler:       &mock.P2PAntifloodHandlerStub{},
		OutputAntifloodHandler:      &mock.P2PAntifloodHandlerStub{},
		NumConcurrentResolvingJobs:  10,
		CurrentNetworkEpochProvider: &mock.CurrentNetworkEpochProviderStub{},
		PreferredPeersHolder:        &p2pmocks.PeersHolderStub{},
		ResolverConfig: config.ResolverConfig{
			NumCrossShardPeers:  1,
			NumTotalPeers:       3,
			NumFullHistoryPeers: 3,
		},
		PeersRatingHandler: &p2pmocks.PeersRatingHandlerStub{},
		PayloadValidator:   &testscommon.PeerAuthenticationPayloadValidatorStub{},
	}
}
