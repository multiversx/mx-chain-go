package shard_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
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

func createMockPubkeyConverter() *testscommon.PubkeyConverterMock {
	return testscommon.NewPubkeyConverterMock(32)
}

func createMockArgsNewIntermediateProcessorsFactory() shard.ArgsNewIntermediateProcessorsContainerFactory {
	args := shard.ArgsNewIntermediateProcessorsContainerFactory{
		Hasher:              &hashingMocks.HasherMock{},
		Marshalizer:         &mock.MarshalizerMock{},
		ShardCoordinator:    mock.NewMultiShardsCoordinatorMock(5),
		PubkeyConverter:     createMockPubkeyConverter(),
		Store:               &storageStubs.ChainStorerStub{},
		PoolsHolder:         createDataPools(),
		EconomicsFee:        &economicsmocks.EconomicsHandlerStub{},
		EnableEpochsHandler: enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.KeepExecOrderOnCreatedSCRsFlag),
	}
	return args
}

func TestNewIntermediateProcessorsContainerFactory_NilShardCoord(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewIntermediateProcessorsFactory()
	args.ShardCoordinator = nil
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(args)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewIntermediateProcessorsFactory()
	args.Marshalizer = nil
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(args)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewIntermediateProcessorsFactory()
	args.Hasher = nil
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(args)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilAdrConv(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewIntermediateProcessorsFactory()
	args.PubkeyConverter = nil
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(args)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilStorer(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewIntermediateProcessorsFactory()
	args.Store = nil
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(args)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilPoolsHolder(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewIntermediateProcessorsFactory()
	args.PoolsHolder = nil
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(args)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilEconomicsFeeHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewIntermediateProcessorsFactory()
	args.EconomicsFee = nil
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(args)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewIntermediateProcessorsFactory()
	args.EnableEpochsHandler = nil
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(args)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewIntermediateProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewIntermediateProcessorsFactory()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(args)

	assert.Nil(t, err)
	assert.NotNil(t, ipcf)
	assert.False(t, ipcf.IsInterfaceNil())
}

func TestIntermediateProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewIntermediateProcessorsFactory()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(args)
	assert.Nil(t, err)
	assert.NotNil(t, ipcf)

	container, err := ipcf.Create()
	assert.Nil(t, err)
	assert.Equal(t, 3, container.Len())
}
