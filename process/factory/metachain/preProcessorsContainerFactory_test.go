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

func createMockCacheConfig() config.TransactionsPoolConfig {
	return config.TransactionsPoolConfig{
		SelectionGasBandwidthIncreasePercent:          400,
		SelectionGasBandwidthIncreaseScheduledPercent: 260,
	}
}

func createMockSortedTransactionsConfig() config.SortedTransactionsConfig {
	return config.SortedTransactionsConfig{
		TxCacheSelectionGasRequested:        10_000_000_000,
		TxCacheSelectionMaxNumTxs:           30000,
		TxCacheSelectionLoopMaximumDuration: 250,
	}
}

func TestNewPreProcessorsContainerFactory_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		nil,
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilStore(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)

	assert.Equal(t, process.ErrNilStore, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilMarshalizer(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		nil,
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		nil,
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilDataPool(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		nil,
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilAccounts(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		nil,
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilFeeHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		nil,
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)

	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilTxProcessor(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		nil,
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)

	assert.Equal(t, process.ErrNilTxProcessor, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilRequestHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		nil,
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)
	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilGasHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		nil,
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)
	assert.Equal(t, process.ErrNilGasHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBlockTracker(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		nil,
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)
	assert.Equal(t, process.ErrNilBlockTracker, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilPubkeyConverter(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		nil,
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBlockSizeComputationHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		nil,
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)
	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBalanceComputationHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		nil,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)
	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		nil,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		nil,
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)
	assert.Equal(t, process.ErrNilTxTypeHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		nil,
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)
	assert.Equal(t, process.ErrNilScheduledTxsExecutionHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilProcessedMiniBlocksTracker(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		nil,
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)
	assert.Equal(t, process.ErrNilProcessedMiniBlocksTracker, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilTxExecutionOrderHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		nil,
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)

	assert.Equal(t, process.ErrNilTxExecutionOrderHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)
	assert.False(t, ppcm.IsInterfaceNil())
}

func TestNewPreProcessorsContainerFactory_ErrBadSelectionGasBandwidthIncreasePercent(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		config.TransactionsPoolConfig{SelectionGasBandwidthIncreasePercent: 0},
		createMockSortedTransactionsConfig(),
	)

	assert.Error(t, err)
	assert.Equal(t, process.ErrBadSelectionGasBandwidthIncreasePercent, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_ErrBadSelectionGasBandwidthIncreaseScheduledPercent(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		config.TransactionsPoolConfig{
			SelectionGasBandwidthIncreasePercent:          400,
			SelectionGasBandwidthIncreaseScheduledPercent: 0,
		},
		createMockSortedTransactionsConfig(),
	)

	assert.Error(t, err)
	assert.Equal(t, process.ErrBadSelectionGasBandwidthIncreaseScheduledPercent, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_ErrBadTxCacheSelectionMaxNumTxs(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		config.TransactionsPoolConfig{
			SelectionGasBandwidthIncreasePercent:          400,
			SelectionGasBandwidthIncreaseScheduledPercent: 260,
		},
		config.SortedTransactionsConfig{
			TxCacheSelectionGasRequested:        10_000_000_000,
			TxCacheSelectionMaxNumTxs:           0,
			TxCacheSelectionLoopMaximumDuration: 0,
		},
	)

	assert.Error(t, err)
	assert.Equal(t, process.ErrBadTxCacheSelectionMaxNumTxs, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_ErrBadTxCacheSelectionLoopMaximumDuration(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		config.SortedTransactionsConfig{
			TxCacheSelectionGasRequested: 10_000_000_000,
			TxCacheSelectionMaxNumTxs:    30000,
		},
	)

	assert.Error(t, err)
	assert.Equal(t, process.ErrBadTxCacheSelectionLoopMaximumDuration, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_ErrBadTxCacheSelectionGasRequested(t *testing.T) {
	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		config.SortedTransactionsConfig{
			TxCacheSelectionMaxNumTxs:           30000,
			TxCacheSelectionLoopMaximumDuration: 250,
		},
	)

	assert.Error(t, err)
	assert.Equal(t, process.ErrBadTxCacheSelectionGasRequested, err)
	assert.Nil(t, ppcm)
}

func TestPreProcessorsContainerFactory_CreateErrTxPreproc(t *testing.T) {
	t.Parallel()

	dataPool := dataRetrieverMock.NewPoolsHolderStub()
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataPool,
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)

	container, err := ppcm.Create()
	assert.Nil(t, container)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestPreProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	ppcm, err := metachain.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&economicsmocks.EconomicsHandlerMock{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&commonMock.TxExecutionOrderHandlerStub{},
		createMockCacheConfig(),
		createMockSortedTransactionsConfig(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)

	container, err := ppcm.Create()
	assert.Nil(t, err)
	assert.Equal(t, 2, container.Len())
}
