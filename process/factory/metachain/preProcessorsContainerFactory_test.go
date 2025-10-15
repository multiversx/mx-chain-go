package metachain_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/metachain"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	commonMock "github.com/multiversx/mx-chain-go/testscommon/common"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func createMockTxCacheSelectionConfig() config.TxCacheSelectionConfig {
	return config.TxCacheSelectionConfig{
		SelectionGasBandwidthIncreasePercent:          400,
		SelectionGasBandwidthIncreaseScheduledPercent: 260,
		SelectionGasRequested:                         10_000_000_000,
		SelectionMaxNumTxs:                            30000,
		SelectionLoopMaximumDuration:                  250,
		SelectionLoopDurationCheckInterval:            10,
	}
}

func TestNewPreProcessorsContainerFactory_NilShardCoordinator(t *testing.T) {
	t.Parallel()
	args := createPreProcessorContainerFactoryArgs()
	args.ShardCoordinator = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilStore(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.Store = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilStore, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.Marshalizer = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.Hasher = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilDataPool(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.DataPool = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilAccounts(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.Accounts = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilAccountsProposal(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.AccountsProposal = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.ErrorIs(t, err, process.ErrNilAccountsAdapter)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilFeeHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.EconomicsFee = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilTxProcessor(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.TxProcessor = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilTxProcessor, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilRequestHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.RequestHandler = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)
	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilGasHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.GasHandler = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)
	assert.Equal(t, process.ErrNilGasHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBlockTracker(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.BlockTracker = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)
	assert.Equal(t, process.ErrNilBlockTracker, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilPubkeyConverter(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.PubkeyConverter = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBlockSizeComputationHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.BlockSizeComputation = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)
	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBalanceComputationHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.BalanceComputation = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)
	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.EnableEpochsHandler = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilEnableRoundsHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.EnableRoundsHandler = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)
	assert.Equal(t, process.ErrNilEnableRoundsHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.TxTypeHandler = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)
	assert.Equal(t, process.ErrNilTxTypeHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.ScheduledTxsExecutionHandler = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)
	assert.Equal(t, process.ErrNilScheduledTxsExecutionHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilProcessedMiniBlocksTracker(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.ProcessedMiniBlocksTracker = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)
	assert.Equal(t, process.ErrNilProcessedMiniBlocksTracker, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilTxExecutionOrderHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	args.TxExecutionOrderHandler = nil
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilTxExecutionOrderHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)
	assert.False(t, ppcm.IsInterfaceNil())
}

func TestPreProcessorsContainerFactory_CreateErrTxPreproc(t *testing.T) {
	t.Parallel()

	dataPool := dataRetrieverMock.NewPoolsHolderStub()
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}
	args := createPreProcessorContainerFactoryArgs()
	args.DataPool = dataPool

	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)

	container, err := ppcm.Create()
	assert.Nil(t, container)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestPreProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	args := createPreProcessorContainerFactoryArgs()
	ppcm, err := metachain.NewPreProcessorsContainerFactory(args)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)

	container, err := ppcm.Create()
	assert.Nil(t, err)
	assert.Equal(t, 2, container.Len())
}

func createPreProcessorContainerFactoryArgs() metachain.ArgsPreProcessorsContainerFactory {
	return metachain.ArgsPreProcessorsContainerFactory{
		ShardCoordinator:             mock.NewMultiShardsCoordinatorMock(3),
		Store:                        &storageStubs.ChainStorerStub{},
		Marshalizer:                  &mock.MarshalizerMock{},
		Hasher:                       &hashingMocks.HasherMock{},
		DataPool:                     dataRetrieverMock.NewPoolsHolderMock(),
		Accounts:                     &stateMock.AccountsStub{},
		AccountsProposal:             &stateMock.AccountsStub{},
		RequestHandler:               &testscommon.RequestHandlerStub{},
		TxProcessor:                  &testscommon.TxProcessorMock{},
		ScResultProcessor:            &testscommon.SmartContractResultsProcessorMock{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerMock{},
		GasHandler:                   &testscommon.GasHandlerStub{},
		BlockTracker:                 &mock.BlockTrackerMock{},
		PubkeyConverter:              createMockPubkeyConverter(),
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		EnableRoundsHandler:          &testscommon.EnableRoundsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		TxExecutionOrderHandler:      &commonMock.TxExecutionOrderHandlerStub{},
		TxCacheSelectionConfig:       createMockTxCacheSelectionConfig(),
	}
}
