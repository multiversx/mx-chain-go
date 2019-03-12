package factory_test

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
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

func createDataPool() data.TransientDataHolder {
	mockTransientDataPool := &mock.TransientDataPoolMock{}
	mockTransientDataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	mockTransientDataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	mockTransientDataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	mockTransientDataPool.MiniBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	mockTransientDataPool.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	return mockTransientDataPool
}

func createBlockchain() *blockchain.BlockChain {
	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{})

	return blkc
}

//------- NewResolversContainerFactory

func TestNewResolversContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := factory.NewResolversContainerFactory(
		nil,
		createStubTopicMessageHandler("", ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewResolversContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		nil,
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewResolversContainerFactory_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		nil,
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewResolversContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createBlockchain(),
		nil,
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, rcf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewResolversContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	rcf, err := factory.NewResolversContainerFactory(
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

	rcf, err := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		nil,
	)

	assert.Nil(t, rcf)
	assert.Equal(t, process.ErrNilUint64ByteSliceConverter, err)
}

func TestNewResolversContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	rcf, err := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.NotNil(t, rcf)
	assert.Nil(t, err)
}

//------- Create

func TestResolversContainerFactory_CreateTopicCreationTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler(factory.TransactionTopic, ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateTopicCreationHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler(factory.HeadersTopic, ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateTopicCreationMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler(factory.MiniBlocksTopic, ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateTopicCreationPeerChBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler(factory.PeerChBodyTopic, ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateRegisterTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", factory.TransactionTopic),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateRegisterHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", factory.HeadersTopic),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateRegisterMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", factory.MiniBlocksTopic),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateRegisterPeerChBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	rcf, _ := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", factory.PeerChBodyTopic),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestResolversContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	rcf, _ := factory.NewResolversContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicMessageHandler("", ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
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

	rcf, _ := factory.NewResolversContainerFactory(
		shardCoordinator,
		createStubTopicMessageHandler("", ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		createDataPool(),
		&mock.Uint64ByteSliceConverterMock{},
	)

	container, _ := rcf.Create()

	assert.Equal(t, noOfShards+1+noOfShards+1, container.Len())
}
