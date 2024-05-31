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

func createStubMessengerForMeta(matchStrToErrOnCreate string, matchStrToErrOnRegister string) p2p.Messenger {
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

func createDataPoolsForMeta() dataRetriever.PoolsHolder {
	pools := &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{}
		},
		MiniBlocksCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
	}

	return pools
}

func createStoreForMeta() dataRetriever.StorageService {
	return &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{}, nil
		},
	}
}

func createTriesHolderForMeta() common.TriesHolder {
	triesHolder := state.NewDataTriesHolder()
	triesHolder.Put([]byte(dataRetriever.UserAccountsUnit.String()), &trieMock.TrieStub{})
	triesHolder.Put([]byte(dataRetriever.PeerAccountsUnit.String()), &trieMock.TrieStub{})
	return triesHolder
}

// ------- NewResolversContainerFactory

func TestNewMetaResolversContainerFactory_NewNumGoRoutinesThrottlerFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.NumConcurrentResolvingJobs = 0

	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)
	assert.Nil(t, rcf)
	assert.Equal(t, core.ErrNotPositiveValue, err)

	args.NumConcurrentResolvingJobs = 10
	args.NumConcurrentResolvingTrieNodesJobs = 0

	rcf, err = resolverscontainer.NewMetaResolversContainerFactory(args)
	assert.Nil(t, rcf)
	assert.Equal(t, core.ErrNotPositiveValue, err)
}

func TestNewMetaResolversContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.ShardCoordinator = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewMetaResolversContainerFactory_NilMainMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.MainMessenger = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilMessenger))
}

func TestNewMetaResolversContainerFactory_NilFullArchiveMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.FullArchiveMessenger = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilMessenger))
}

func TestNewMetaResolversContainerFactory_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Store = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilStore, err)
}

func TestNewMetaResolversContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Marshalizer = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewMetaResolversContainerFactory_NilMarshalizerAndSizeCheckShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Marshalizer = nil
	args.SizeCheckDelta = 1
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewMetaResolversContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.DataPools = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPoolHolder, err)
}

func TestNewMetaResolversContainerFactory_NilMainPreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.MainPreferredPeersHolder = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilPreferredPeersHolder))
}

func TestNewMetaResolversContainerFactory_NilFullArchivePreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.FullArchivePreferredPeersHolder = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilPreferredPeersHolder))
}

func TestNewMetaResolversContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Uint64ByteSliceConverter = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewMetaResolversContainerFactory_NilDataPackerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.DataPacker = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
}

func TestNewMetaResolversContainerFactory_NilTrieDataGetterShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.TriesContainer = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilTrieDataGetter, err)
}

func TestNewMetaResolversContainerFactory_NilInputAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.InputAntifloodHandler = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilAntifloodHandler))
}

func TestNewMetaResolversContainerFactory_NilOutputAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.OutputAntifloodHandler = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilAntifloodHandler))
}

// ------- Create

func TestMetaResolversContainerFactory_CreateRegisterShardHeadersForMetachainOnMainNetworkFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.MainMessenger = createStubMessengerForMeta("", factory.ShardBlocksTopic)
	rcf, _ := resolverscontainer.NewMetaResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestMetaResolversContainerFactory_CreateRegisterShardHeadersForMetachainOnFullArchiveNetworkFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.FullArchiveMessenger = createStubMessengerForMeta("", factory.ShardBlocksTopic)
	rcf, _ := resolverscontainer.NewMetaResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestMetaResolversContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	rcf, _ := resolverscontainer.NewMetaResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestMetaResolversContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4
	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	args := getArgumentsMeta()
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
	rcf, _ := resolverscontainer.NewMetaResolversContainerFactory(args)

	container, _ := rcf.Create()
	numResolversShardHeadersForMetachain := noOfShards
	numResolverMetablocks := 1
	numResolversMiniBlocks := noOfShards + 2
	numResolversUnsigned := noOfShards + 1
	numResolversRewards := noOfShards
	numResolversTxs := noOfShards + 1
	numResolversTrieNodes := 2
	numResolversPeerAuth := 1
	numResolverValidatorInfo := 1
	totalResolvers := numResolversShardHeadersForMetachain + numResolverMetablocks + numResolversMiniBlocks +
		numResolversUnsigned + numResolversTxs + numResolversTrieNodes + numResolversRewards + numResolversPeerAuth + numResolverValidatorInfo

	assert.Equal(t, totalResolvers, container.Len())
	assert.Equal(t, totalResolvers, registerMainCnt)
	assert.Equal(t, totalResolvers, registerFullArchiveCnt)

	err := rcf.AddShardTrieNodeResolvers(container)
	assert.Nil(t, err)
	assert.Equal(t, totalResolvers+noOfShards, container.Len())
}

func TestMetaResolversContainerFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.ShardCoordinator = nil
	rcf, _ := resolverscontainer.NewMetaResolversContainerFactory(args)
	assert.True(t, rcf.IsInterfaceNil())

	rcf, _ = resolverscontainer.NewMetaResolversContainerFactory(getArgumentsMeta())
	assert.False(t, rcf.IsInterfaceNil())
}

func getArgumentsMeta() resolverscontainer.FactoryArgs {
	return resolverscontainer.FactoryArgs{
		ShardCoordinator:                    mock.NewOneShardCoordinatorMock(),
		MainMessenger:                       createStubMessengerForMeta("", ""),
		FullArchiveMessenger:                createStubMessengerForMeta("", ""),
		Store:                               createStoreForMeta(),
		Marshalizer:                         &mock.MarshalizerMock{},
		DataPools:                           createDataPoolsForMeta(),
		Uint64ByteSliceConverter:            &mock.Uint64ByteSliceConverterMock{},
		DataPacker:                          &mock.DataPackerStub{},
		TriesContainer:                      createTriesHolderForMeta(),
		SizeCheckDelta:                      0,
		InputAntifloodHandler:               &mock.P2PAntifloodHandlerStub{},
		OutputAntifloodHandler:              &mock.P2PAntifloodHandlerStub{},
		NumConcurrentResolvingJobs:          10,
		NumConcurrentResolvingTrieNodesJobs: 3,
		MainPreferredPeersHolder:            &p2pmocks.PeersHolderStub{},
		FullArchivePreferredPeersHolder:     &p2pmocks.PeersHolderStub{},
		PayloadValidator:                    &testscommon.PeerAuthenticationPayloadValidatorStub{},
	}
}
