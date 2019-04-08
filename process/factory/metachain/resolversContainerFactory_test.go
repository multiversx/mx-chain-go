package metachain_test

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
)

func createStubTopicMessageHandler(matchStrToErrOnCreate string, matchStrToErrOnRegister string) process.TopicMessageHandler {
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

func createDataPools() data.MetaPoolsHolder {
	pools := &mock.MetaPoolsHolderStub{
		MetaBlockNoncesCalled: func() data.Uint64Cacher {
			return &mock.Uint64CacherStub{}
		},
		ShardHeadersCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
		MiniBlockHashesCalled: func() data.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		MetaChainBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
	}

	return pools
}

func createBlockchain() data.ChainHandler {
	return &mock.BlockChainMock{
		StorageService: &mock.ChainStorerMock{
			GetStorerCalled: func(unitType data.UnitType) storage.Storer {
				return &mock.StorerStub{}
			},
		},
	}
}

//------- NewResolversContainerFactory

func TestNewResolversContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		nil,
		createStubTopicMessageHandler("", ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewResolversContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		nil,
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewResolversContainerFactory_NilBlockchainShouldErr(t *testing.T) {
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
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewResolversContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createBlockchain(),
		nil,
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewResolversContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		nil,
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewResolversContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPools(),
		nil,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, process.ErrNilUint64ByteSliceConverter, err)
}

func TestNewResolversContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	rcf, err := metachain.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createBlockchain(),
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
		createBlockchain(),
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
		createBlockchain(),
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
		createBlockchain(),
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
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPools(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, _ := rcf.Create()
	numInterceptorsShardHeadersForMetachain := noOfShards

	assert.Equal(t, numInterceptorsShardHeadersForMetachain, container.Len())
}
