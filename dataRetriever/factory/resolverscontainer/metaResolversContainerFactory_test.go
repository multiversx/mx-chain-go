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
	pools := &mock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{}
		},
		MiniBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
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
	triesHolder.Put([]byte(triesFactory.UserAccountTrie), &mock.TrieStub{})
	triesHolder.Put([]byte(triesFactory.PeerAccountTrie), &mock.TrieStub{})
	return triesHolder
}

//------- NewResolversContainerFactory

func TestNewMetaResolversContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(
		nil,
		createStubTopicMessageHandlerForMeta("", ""),
		createStoreForMeta(),
		&mock.MarshalizerMock{},
		createDataPoolsForMeta(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForMeta(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewMetaResolversContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		nil,
		createStoreForMeta(),
		&mock.MarshalizerMock{},
		createDataPoolsForMeta(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForMeta(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewMetaResolversContainerFactory_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForMeta("", ""),
		nil,
		&mock.MarshalizerMock{},
		createDataPoolsForMeta(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForMeta(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilStore, err)
}

func TestNewMetaResolversContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForMeta("", ""),
		createStoreForMeta(),
		nil,
		createDataPoolsForMeta(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForMeta(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewMetaResolversContainerFactory_NilMarshalizerAndSizeCheckShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForMeta("", ""),
		createStoreForMeta(),
		nil,
		createDataPoolsForMeta(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForMeta(),
		1,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewMetaResolversContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForMeta("", ""),
		createStoreForMeta(),
		&mock.MarshalizerMock{},
		nil,
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForMeta(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPoolHolder, err)
}

func TestNewMetaResolversContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForMeta("", ""),
		createStoreForMeta(),
		&mock.MarshalizerMock{},
		createDataPoolsForMeta(),
		nil,
		&mock.DataPackerStub{},
		createTriesHolderForMeta(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewMetaResolversContainerFactory_NilDataPackerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForMeta("", ""),
		createStoreForMeta(),
		&mock.MarshalizerMock{},
		createDataPoolsForMeta(),
		&mock.Uint64ByteSliceConverterMock{},
		nil,
		createTriesHolderForMeta(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
}

func TestNewMetaResolversContainerFactory_NilTrieDataGetterShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForMeta("", ""),
		createStoreForMeta(),
		&mock.MarshalizerMock{},
		createDataPoolsForMeta(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		nil,
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilTrieDataGetter, err)
}

func TestNewMetaResolversContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewMetaResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForMeta("", ""),
		createStoreForMeta(),
		&mock.MarshalizerMock{},
		createDataPoolsForMeta(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForMeta(),
		0,
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(rcf))
}

//------- Create

func TestMetaResolversContainerFactory_CreateTopicShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewMetaResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForMeta(factory.ShardBlocksTopic, ""),
		createStoreForMeta(),
		&mock.MarshalizerMock{},
		createDataPoolsForMeta(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForMeta(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestMetaResolversContainerFactory_CreateRegisterShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewMetaResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForMeta("", factory.ShardBlocksTopic),
		createStoreForMeta(),
		&mock.MarshalizerMock{},
		createDataPoolsForMeta(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForMeta(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestMetaResolversContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewMetaResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForMeta("", ""),
		createStoreForMeta(),
		&mock.MarshalizerMock{},
		createDataPoolsForMeta(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForMeta(),
		0,
	)

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

	rcf, _ := resolverscontainer.NewMetaResolversContainerFactory(
		shardCoordinator,
		createStubTopicMessageHandlerForMeta("", ""),
		createStoreForMeta(),
		&mock.MarshalizerMock{},
		createDataPoolsForMeta(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForMeta(),
		0,
	)

	container, _ := rcf.Create()
	numResolversShardHeadersForMetachain := noOfShards
	numResolverMetablocks := 1
	numResolversMiniBlocks := noOfShards + 1
	numResolversUnsigned := noOfShards + 1
	numResolversTxs := noOfShards + 1
	numResolversTrieNodes := (noOfShards + 1) * 2
	totalResolvers := numResolversShardHeadersForMetachain + numResolverMetablocks + numResolversMiniBlocks +
		numResolversUnsigned + numResolversTxs + numResolversTrieNodes

	assert.Equal(t, totalResolvers, container.Len())
}
