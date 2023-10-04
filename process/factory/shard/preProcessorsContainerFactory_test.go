package shard

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	customErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	mockCommon "github.com/multiversx/mx-chain-go/testscommon/common"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockPubkeyConverter() *testscommon.PubkeyConverterMock {
	return testscommon.NewPubkeyConverterMock(32)
}

func createMockPreProcessorsContainerFactoryArguments() ArgPreProcessorsContainerFactory {
	return ArgPreProcessorsContainerFactory{
		ShardCoordinator:             mock.NewMultiShardsCoordinatorMock(3),
		Store:                        &storageStubs.ChainStorerStub{},
		Marshaller:                   &mock.MarshalizerMock{},
		Hasher:                       &hashingMocks.HasherMock{},
		DataPool:                     dataRetrieverMock.NewPoolsHolderMock(),
		PubkeyConverter:              createMockPubkeyConverter(),
		Accounts:                     &stateMock.AccountsStub{},
		RequestHandler:               &testscommon.RequestHandlerStub{},
		TxProcessor:                  &testscommon.TxProcessorMock{},
		ScProcessor:                  &testscommon.SCProcessorMock{},
		ScResultProcessor:            &testscommon.SmartContractResultsProcessorMock{},
		RewardsTxProcessor:           &testscommon.RewardTxProcessorMock{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerStub{},
		GasHandler:                   &testscommon.GasHandlerStub{},
		BlockTracker:                 &mock.BlockTrackerMock{},
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		TxExecutionOrderHandler:      &mockCommon.TxExecutionOrderHandlerStub{},
		TxPreprocessorCreator:        preprocess.NewTxPreProcessorCreator(),
		ChainRunType:                 common.ChainRunTypeRegular,
	}
}

func TestNewPreProcessorsContainerFactory_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.ShardCoordinator = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilStore(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.Store = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilStore, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilMarshaller(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.Marshaller = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.Hasher = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilDataPool(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.DataPool = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilAddrConv(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.PubkeyConverter = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilPubkeyConverter, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilAccounts(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.Accounts = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilTxProcessor(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.TxProcessor = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilTxProcessor, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilSCProcessor(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.ScProcessor = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilSmartContractProcessor, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilSCR(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.ScResultProcessor = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilSmartContractResultProcessor, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilRewardTxProcessor(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.RewardsTxProcessor = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilRewardsTxProcessor, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilRequestHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.RequestHandler = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilFeeHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.EconomicsFee = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilGasHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.GasHandler = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilGasHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilBlockTracker(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.BlockTracker = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilBlockTracker, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilBlockSizeComputationHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.BlockSizeComputation = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilBalanceComputationHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.BalanceComputation = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.EnableEpochsHandler = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.TxTypeHandler = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilTxTypeHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.ScheduledTxsExecutionHandler = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilScheduledTxsExecutionHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilProcessedMiniBlocksTracker(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.ProcessedMiniBlocksTracker = nil
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilProcessedMiniBlocksTracker, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Nil(t, err)
	assert.NotNil(t, ppcf)
	assert.False(t, ppcf.IsInterfaceNil())
}

func TestPreProcessorsContainerFactory_CreateErrTxPreproc(t *testing.T) {
	t.Parallel()
	dataPool := dataRetrieverMock.NewPoolsHolderStub()
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}

	args := createMockPreProcessorsContainerFactoryArguments()
	args.DataPool = dataPool
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Nil(t, err)
	assert.NotNil(t, ppcf)

	container, err := ppcf.Create()
	assert.Nil(t, container)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestPreProcessorsContainerFactory_CreateErrScrPreproc(t *testing.T) {
	t.Parallel()
	dataPool := dataRetrieverMock.NewPoolsHolderStub()
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &testscommon.ShardedDataStub{}
	}
	dataPool.UnsignedTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}

	args := createMockPreProcessorsContainerFactoryArguments()
	args.DataPool = dataPool
	ppcf, err := NewPreProcessorsContainerFactory(args)

	assert.Nil(t, err)
	assert.NotNil(t, ppcf)

	container, err := ppcf.Create()
	assert.Nil(t, container)
	assert.Equal(t, process.ErrNilUTxDataPool, err)
}

func TestPreProcessorsContainerFactory_CreateSCRPreprocessor(t *testing.T) {
	t.Run("createSmartContractResultPreProcessor should create a main chain instance", func(t *testing.T) {
		t.Parallel()

		ppcf, err := createMockPreProcessorContainerFactory()
		require.Nil(t, err)
		require.NotNil(t, ppcf)

		ppcf.chainRunType = common.ChainRunTypeRegular

		preProc, errCreate := ppcf.createSmartContractResultPreProcessor()

		assert.NotNil(t, preProc)
		assert.Nil(t, errCreate)
	})

	t.Run("createSmartContractResultPreProcessor should create a sovereign chain instance", func(t *testing.T) {
		t.Parallel()

		ppcf, err := createMockPreProcessorContainerFactory()
		require.Nil(t, err)
		require.NotNil(t, ppcf)

		ppcf.chainRunType = common.ChainRunTypeSovereign

		preProc, errCreate := ppcf.createSmartContractResultPreProcessor()

		assert.NotNil(t, preProc)
		assert.Nil(t, errCreate)
	})

	t.Run("createSmartContractResultPreProcessor should error when chain run type is not implemented", func(t *testing.T) {
		t.Parallel()

		ppcf, err := createMockPreProcessorContainerFactory()
		require.Nil(t, err)
		require.NotNil(t, ppcf)

		ppcf.chainRunType = "invalid"

		preProc, errCreate := ppcf.createSmartContractResultPreProcessor()

		assert.Nil(t, preProc)
		assert.True(t, errors.Is(errCreate, customErrors.ErrUnimplementedChainRunType))
	})
}

func TestPreProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	ppcf, err := createMockPreProcessorContainerFactory()

	assert.Nil(t, err)
	assert.NotNil(t, ppcf)

	container, err := ppcf.Create()
	assert.Nil(t, err)
	assert.Equal(t, 4, container.Len())
}

func TestCreateTxPreProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	ppcf, err := createMockPreProcessorContainerFactory()
	require.Nil(t, err)
	require.NotNil(t, ppcf)

	ppcf.chainRunType = common.ChainRunTypeRegular

	preProc, errCreate := ppcf.createTxPreProcessor()

	assert.NotNil(t, preProc)
	assert.Nil(t, errCreate)

}

func createMockPreProcessorContainerFactory() (*preProcessorsContainerFactory, error) {
	dataPool := dataRetrieverMock.NewPoolsHolderStub()

	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &testscommon.ShardedDataStub{}
	}
	dataPool.UnsignedTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &testscommon.ShardedDataStub{}
	}
	dataPool.RewardTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &testscommon.ShardedDataStub{}
	}

	args := createMockPreProcessorsContainerFactoryArguments()
	args.DataPool = dataPool

	return NewPreProcessorsContainerFactory(args)
}
