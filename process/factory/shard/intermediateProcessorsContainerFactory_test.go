package shard_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func createDataPools() dataRetriever.PoolsHolder {
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
	pools.MetaBlocksCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.UnsignedTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	pools.RewardTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	pools.TrieNodesCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.CurrBlockTxsCalled = func() dataRetriever.TransactionCacher {
		return &mock.TxForCurrentBlockStub{}
	}
	return pools
}

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

func TestNewIntermediateProcessorsContainerFactory_NilShardCoord(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		nil,
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		&storageStubs.ChainStorerStub{},
		dPool,
		&economicsmocks.EconomicsHandlerStub{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilMarshalizer(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		&storageStubs.ChainStorerStub{},
		dPool,
		&economicsmocks.EconomicsHandlerStub{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		nil,
		createMockPubkeyConverter(),
		&storageStubs.ChainStorerStub{},
		dPool,
		&economicsmocks.EconomicsHandlerStub{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilAdrConv(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		nil,
		&storageStubs.ChainStorerStub{},
		dPool,
		&economicsmocks.EconomicsHandlerStub{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilStorer(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		nil,
		dPool,
		&economicsmocks.EconomicsHandlerStub{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilPoolsHolder(t *testing.T) {
	t.Parallel()

	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		&storageStubs.ChainStorerStub{},
		nil,
		&economicsmocks.EconomicsHandlerStub{},
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilEconomicsFeeHandler(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		&storageStubs.ChainStorerStub{},
		dPool,
		nil,
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewIntermediateProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		&storageStubs.ChainStorerStub{},
		dPool,
		&economicsmocks.EconomicsHandlerStub{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ipcf)
	assert.False(t, ipcf.IsInterfaceNil())
}

func TestIntermediateProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		createMockPubkeyConverter(),
		&storageStubs.ChainStorerStub{},
		dPool,
		&economicsmocks.EconomicsHandlerStub{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ipcf)

	container, err := ipcf.Create()
	assert.Nil(t, err)
	assert.Equal(t, 3, container.Len())
}
