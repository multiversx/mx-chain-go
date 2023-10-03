package resolverscontainer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

var errExpected = errors.New("expected error")

func createMessengerStubForShard(matchStrToErrOnCreate string, matchStrToErrOnRegister string) p2p.Messenger {
	stub := &p2pmocks.MessengerStub{}

	stub.CreateTopicCalled = func(name string, createChannelForTopic bool) error {
		if matchStrToErrOnCreate == "" {
			return nil
		}

		if strings.Contains(name, matchStrToErrOnCreate) {
			return errExpected
		}

		return nil
	}

	stub.RegisterMessageProcessorCalled = func(topic string, identifier string, handler p2p.MessageProcessor) error {
		if matchStrToErrOnRegister == "" {
			return nil
		}

		if strings.Contains(topic, matchStrToErrOnRegister) {
			return errExpected
		}

		return nil
	}

	return stub
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
	triesHolder.Put([]byte(dataRetriever.UserAccountsUnit.String()), &trieMock.TrieStub{})
	triesHolder.Put([]byte(dataRetriever.PeerAccountsUnit.String()), &trieMock.TrieStub{})
	return triesHolder
}

// ------- NewResolversContainerFactory

func TestNewShardResolversContainerFactory_NewNumGoRoutinesThrottlerFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.NumConcurrentResolvingJobs = 0
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, core.ErrNotPositiveValue, err)
}

func TestNewShardResolversContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.ShardCoordinator = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewShardResolversContainerFactory_NilMainMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.MainMessenger = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilMessenger))
}

func TestNewShardResolversContainerFactory_NilFullArchiveMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.FullArchiveMessenger = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilMessenger))
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

func TestNewShardResolversContainerFactory_NilMainPreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.MainPreferredPeersHolder = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilPreferredPeersHolder))
}

func TestNewShardResolversContainerFactory_NilFullArchivePreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.FullArchivePreferredPeersHolder = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilPreferredPeersHolder))
}

func TestNewShardResolversContainerFactory_NilTriesContainerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.TriesContainer = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilTrieDataGetter, err)
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

// ------- Create

func TestShardResolversContainerFactory_CreateRegisterTxFailsOnMainNetworkShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.MainMessenger = createMessengerStubForShard("", factory.TransactionTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterTxFailsOnFullArchiveNetworkShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.FullArchiveMessenger = createMessengerStubForShard("", factory.TransactionTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterHdrFailsOnMainNetworkShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.MainMessenger = createMessengerStubForShard("", factory.ShardBlocksTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterHdrFailsOnFullArchiveNetworkShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.FullArchiveMessenger = createMessengerStubForShard("", factory.ShardBlocksTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterMiniBlocksFailsOnMainNetworkShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.MainMessenger = createMessengerStubForShard("", factory.MiniBlocksTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterMiniBlocksFailsOnFullArchiveNetworkShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.FullArchiveMessenger = createMessengerStubForShard("", factory.MiniBlocksTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterTrieNodesFailsOnMainNetworkShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.MainMessenger = createMessengerStubForShard("", factory.AccountTrieNodesTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterTrieNodesFailsOnFullArchiveNetworkShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.FullArchiveMessenger = createMessengerStubForShard("", factory.AccountTrieNodesTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterPeerAuthenticationOnMainNetworkShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.MainMessenger = createMessengerStubForShard("", common.PeerAuthenticationTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterPeerAuthenticationOnFullArchiveNetworkShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.FullArchiveMessenger = createMessengerStubForShard("", common.PeerAuthenticationTopic)
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
	assert.Equal(t, totalResolvers, registerMainCnt)
	assert.Equal(t, totalResolvers, registerFullArchiveCnt)
}

func TestShardResolversContainerFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.ShardCoordinator = nil
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)
	assert.True(t, rcf.IsInterfaceNil())

	rcf, _ = resolverscontainer.NewShardResolversContainerFactory(getArgumentsMeta())
	assert.False(t, rcf.IsInterfaceNil())
}

func getArgumentsShard() resolverscontainer.FactoryArgs {
	return resolverscontainer.FactoryArgs{
		ShardCoordinator:                mock.NewOneShardCoordinatorMock(),
		MainMessenger:                   createMessengerStubForShard("", ""),
		FullArchiveMessenger:            createMessengerStubForShard("", ""),
		Store:                           createStoreForShard(),
		Marshalizer:                     &mock.MarshalizerMock{},
		DataPools:                       createDataPoolsForShard(),
		Uint64ByteSliceConverter:        &mock.Uint64ByteSliceConverterMock{},
		DataPacker:                      &mock.DataPackerStub{},
		TriesContainer:                  createTriesHolderForShard(),
		SizeCheckDelta:                  0,
		InputAntifloodHandler:           &mock.P2PAntifloodHandlerStub{},
		OutputAntifloodHandler:          &mock.P2PAntifloodHandlerStub{},
		NumConcurrentResolvingJobs:      10,
		MainPreferredPeersHolder:        &p2pmocks.PeersHolderStub{},
		FullArchivePreferredPeersHolder: &p2pmocks.PeersHolderStub{},
		PayloadValidator:                &testscommon.PeerAuthenticationPayloadValidatorStub{},
	}
}
