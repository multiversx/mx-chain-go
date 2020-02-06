package resolverscontainer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	triesFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/storage"
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

func createDataPoolsForShard() dataRetriever.PoolsHolder {
	pools := &mock.PoolsHolderStub{}
	pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	pools.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	pools.UnsignedTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	pools.RewardTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}

	return pools
}

func createStoreForShard() dataRetriever.StorageService {
	return &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{}
		},
	}
}

func createTriesHolderForShard() state.TriesHolder {
	triesHolder := state.NewDataTriesHolder()
	triesHolder.Put([]byte(triesFactory.UserAccountTrie), &mock.TrieStub{})
	triesHolder.Put([]byte(triesFactory.PeerAccountTrie), &mock.TrieStub{})
	return triesHolder
}

//------- NewResolversContainerFactory

func TestNewShardResolversContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewShardResolversContainerFactory(
		nil,
		createStubTopicMessageHandlerForShard("", ""),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewShardResolversContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		nil,
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewShardResolversContainerFactory_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", ""),
		nil,
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilTxStorage, err)
}

func TestNewShardResolversContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", ""),
		createStoreForShard(),
		nil,
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewShardResolversContainerFactory_NilMarshalizerAndSizeShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", ""),
		createStoreForShard(),
		nil,
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		1,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewShardResolversContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", ""),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		nil,
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPoolHolder, err)
}

func TestNewShardResolversContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", ""),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		nil,
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewShardResolversContainerFactory_NilSliceSplitterShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", ""),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		nil,
		createTriesHolderForShard(),
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
}

func TestNewShardResolversContainerFactory_NilTrieDataGetterShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", ""),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		nil,
		0,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilTrieDataGetter, err)
}

func TestNewShardResolversContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	rcf, err := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", ""),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		1,
	)

	assert.NotNil(t, rcf)
	assert.Nil(t, err)
	require.False(t, rcf.IsInterfaceNil())
}

//------- Create

func TestShardResolversContainerFactory_CreateTopicCreationTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard(factory.TransactionTopic, ""),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateTopicCreationHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard(factory.ShardBlocksTopic, ""),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateTopicCreationMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard(factory.MiniBlocksTopic, ""),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateTopicCreationPeerChBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard(factory.PeerChBodyTopic, ""),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", factory.TransactionTopic),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", factory.ShardBlocksTopic),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", factory.MiniBlocksTopic),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterPeerChBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", factory.PeerChBodyTopic),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterTrieNodesFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", factory.AccountTrieNodesTopic),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandlerForShard("", ""),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

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

	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(
		shardCoordinator,
		createStubTopicMessageHandlerForShard("", ""),
		createStoreForShard(),
		&mock.MarshalizerMock{},
		createDataPoolsForShard(),
		&mock.Uint64ByteSliceConverterMock{},
		&mock.DataPackerStub{},
		createTriesHolderForShard(),
		0,
	)

	container, _ := rcf.Create()

	numResolverSCRs := noOfShards + 1
	numResolverTxs := noOfShards + 1
	numResolverRewardTxs := noOfShards + 1
	numResolverHeaders := 1
	numResolverMiniBlocks := noOfShards + 1
	numResolverPeerChanges := 1
	numResolverMetaBlockHeaders := 1
	numResolverTrieNodes := 2
	totalResolvers := numResolverTxs + numResolverHeaders + numResolverMiniBlocks + numResolverPeerChanges +
		numResolverMetaBlockHeaders + numResolverSCRs + numResolverRewardTxs + numResolverTrieNodes

	assert.Equal(t, totalResolvers, container.Len())
}
