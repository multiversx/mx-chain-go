package resolverscontainer_test

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	triesFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
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

	tmhs.RegisterMessageProcessorCalled = func(topic string, handler p2p.MessageProcessor) error {
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

func createDataPoolsForMeta() dataRetriever.PoolsHolder {
	pools := &testscommon.PoolsHolderStub{
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
	return &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{}
		},
	}
}

func createTriesHolderForMeta() state.TriesHolder {
	triesHolder := state.NewDataTriesHolder()
	triesHolder.Put([]byte(triesFactory.UserAccountTrie), &testscommon.TrieStub{})
	triesHolder.Put([]byte(triesFactory.PeerAccountTrie), &testscommon.TrieStub{})
	return triesHolder
}

//------- NewResolversContainerFactory

func TestNewMetaResolversContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.ShardCoordinator = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewMetaResolversContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Messenger = nil
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
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

func TestNewMetaResolversContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(args)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(rcf))
}

//------- Create

func TestMetaResolversContainerFactory_CreateRegisterShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Messenger = createStubTopicMessageHandlerForMeta("", factory.ShardBlocksTopic)
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
	totalResolvers := numResolversShardHeadersForMetachain + numResolverMetablocks + numResolversMiniBlocks +
		numResolversUnsigned + numResolversTxs + numResolversTrieNodes + numResolversRewards

	assert.Equal(t, totalResolvers, container.Len())

	err := rcf.AddShardTrieNodeResolvers(container)
	assert.Nil(t, err)
	assert.Equal(t, totalResolvers+noOfShards, container.Len())
}

func getArgumentsMeta() resolverscontainer.FactoryArgs {
	return resolverscontainer.FactoryArgs{
		ShardCoordinator:           mock.NewOneShardCoordinatorMock(),
		Messenger:                  createStubTopicMessageHandlerForMeta("", ""),
		Store:                      createStoreForMeta(),
		Marshalizer:                &mock.MarshalizerMock{},
		DataPools:                  createDataPoolsForMeta(),
		Uint64ByteSliceConverter:   &mock.Uint64ByteSliceConverterMock{},
		DataPacker:                 &mock.DataPackerStub{},
		TriesContainer:             createTriesHolderForMeta(),
		SizeCheckDelta:             0,
		InputAntifloodHandler:      &mock.P2PAntifloodHandlerStub{},
		OutputAntifloodHandler:     &mock.P2PAntifloodHandlerStub{},
		NumConcurrentResolvingJobs: 10,
	}
}
