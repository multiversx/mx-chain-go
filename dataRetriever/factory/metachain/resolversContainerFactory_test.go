package metachain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

var errExpected = errors.New("expected error")

func createStubTopicMessageHandler(matchStrToErrOnCreate string, matchStrToErrOnRegister string) dataRetriever.TopicMessageHandler {
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

func createDataPools() dataRetriever.MetaPoolsHolder {
	pools := &mock.MetaPoolsHolderStub{
		MetaBlockNoncesCalled: func() dataRetriever.Uint64Cacher {
			return &mock.Uint64CacherStub{}
		},
		ShardHeadersCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
		MiniBlockHashesCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		MetaChainBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
	}

	return pools
}

func createStore() dataRetriever.StorageService {
	return &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{}
		},
	}
}

//------- NewResolversContainerFactory

func TestNewResolversContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		nil,
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewResolversContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		nil,
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewResolversContainerFactory_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		nil,
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilStore, err)
}

func TestNewResolversContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		nil,
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewResolversContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		nil,
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPoolHolder, err)
}

func TestNewResolversContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		nil,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewResolversContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.NotNil(t, rcf)
	assert.Nil(t, err)
}

//------- Create

func TestResolversContainerFactory_CreateTopicShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler(factory.ShardHeadersForMetachainTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateRegisterShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", factory.ShardHeadersForMetachainTopic),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	rcf, _ := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, err := rcf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestResolversContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4
	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	rcf, _ := metachain.NewResolversContainerFactory(
		shardCoordinator,
		createStubTopicMessageHandler("", ""),
		createStore(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, _ := rcf.Create()
	numResolversShardHeadersForMetachain := noOfShards
	numResolverMetablocks := 1
	totalResolvers := numResolversShardHeadersForMetachain + numResolverMetablocks

	assert.Equal(t, totalResolvers, container.Len())
}
