package metachain_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/metachain"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/common"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func createMockPreProcessorsContainerFactoryArguments() metachain.ArgPreProcessorsContainerFactory {
	return metachain.ArgPreProcessorsContainerFactory{
		ShardCoordinator:             mock.NewMultiShardsCoordinatorMock(3),
		Store:                        &storageStubs.ChainStorerStub{},
		Marshaller:                   &mock.MarshalizerMock{},
		Hasher:                       &hashingMocks.HasherMock{},
		DataPool:                     dataRetrieverMock.NewPoolsHolderMock(),
		Accounts:                     &stateMock.AccountsStub{},
		RequestHandler:               &testscommon.RequestHandlerStub{},
		TxProcessor:                  &testscommon.TxProcessorMock{},
		ScResultProcessor:            &testscommon.SmartContractResultsProcessorMock{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerStub{},
		GasHandler:                   &testscommon.GasHandlerStub{},
		BlockTracker:                 &mock.BlockTrackerMock{},
		PubkeyConverter:              createMockPubkeyConverter(),
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		TxExecutionOrderHandler:      &common.TxExecutionOrderHandlerStub{},
	}
}

func TestNewPreProcessorsContainerFactory_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.ShardCoordinator = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilStore(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.Store = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilStore, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilMarshaller(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.Marshaller = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.Hasher = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilDataPool(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.DataPool = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilAccounts(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.Accounts = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilFeeHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.EconomicsFee = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilTxProcessor(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.TxProcessor = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilTxProcessor, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilRequestHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.RequestHandler = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilGasHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.GasHandler = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilGasHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilBlockTracker(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.BlockTracker = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilBlockTracker, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilPubkeyConverter(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.PubkeyConverter = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilPubkeyConverter, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilBlockSizeComputationHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.BlockSizeComputation = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilBalanceComputationHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.BalanceComputation = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.EnableEpochsHandler = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.TxTypeHandler = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilTxTypeHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.ScheduledTxsExecutionHandler = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilScheduledTxsExecutionHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory_NilProcessedMiniBlocksTracker(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.ProcessedMiniBlocksTracker = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilProcessedMiniBlocksTracker, err)
	assert.Nil(t, ppcf)
}

func TestPreProcessorsContainerFactory_CreateErrTxExecutionOrderHandler(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	args.TxExecutionOrderHandler = nil
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilTxExecutionOrderHandler, err)
	assert.Nil(t, ppcf)
}

func TestNewPreProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

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
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Nil(t, err)
	assert.NotNil(t, ppcf)

	container, err := ppcf.Create()
	assert.Nil(t, container)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestPreProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	ppcf, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Nil(t, err)
	assert.NotNil(t, ppcf)

	container, err := ppcf.Create()
	assert.Nil(t, err)
	assert.Equal(t, 2, container.Len())
}
