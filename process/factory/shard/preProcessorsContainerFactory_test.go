package shard

import (
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func createMockPubkeyConverter() *testscommon.PubkeyConverterMock {
	return testscommon.NewPubkeyConverterMock(32)
}

func TestNewPreProcessorsContainerFactory_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		nil,
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilStore(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilStore, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilMarshalizer(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		nil,
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		nil,
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilDataPool(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		nil,
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilAddrConv(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		nil,
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilPubkeyConverter, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilAccounts(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		nil,
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilTxProcessor(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		nil,
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilTxProcessor, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilSCProcessor(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		nil,
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilSmartContractProcessor, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilSCR(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		nil,
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilSmartContractResultProcessor, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilRewardTxProcessor(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		nil,
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilRewardsTxProcessor, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilRequestHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		nil,
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilFeeHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		nil,
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilGasHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		nil,
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilGasHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBlockTracker(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		nil,
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilBlockTracker, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBlockSizeComputationHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		nil,
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBalanceComputationHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		nil,
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		nil,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		nil,
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilTxTypeHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		nil,
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Equal(t, process.ErrNilScheduledTxsExecutionHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilProcessedMiniBlocksTracker(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		nil,
	)

	assert.Equal(t, process.ErrNilProcessedMiniBlocksTracker, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataRetrieverMock.NewPoolsHolderMock(),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

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

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataPool,
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)

	container, err := ppcm.Create()
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

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataPool,
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)

	container, err := ppcm.Create()
	assert.Nil(t, container)
	assert.Equal(t, process.ErrNilUTxDataPool, err)
}

func TestPreProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()
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

	ppcm, err := NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&storageStubs.ChainStorerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataPool,
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&testscommon.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)

	container, err := ppcm.Create()
	assert.Nil(t, err)
	assert.Equal(t, 4, container.Len())
}
