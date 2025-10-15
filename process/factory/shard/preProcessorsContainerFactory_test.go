package shard

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
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

func createMockPubkeyConverter() *testscommon.PubkeyConverterMock {
	return testscommon.NewPubkeyConverterMock(32)
}

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

	args := createPreProcessorsContainerFactoryArgs()
	args.ShardCoordinator = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilStore(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.Store = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilStore, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.Marshalizer = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.Hasher = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilDataPool(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.DataPool = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilAddrConv(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.PubkeyConverter = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilPubkeyConverter, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilAccounts(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.Accounts = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilAccountsProposal(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.AccountsProposal = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.ErrorIs(t, err, process.ErrNilAccountsAdapter)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilTxProcessor(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.TxProcessor = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilTxProcessor, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilSCProcessor(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.ScProcessor = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilSmartContractProcessor, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilSCR(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.ScResultProcessor = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilSmartContractResultProcessor, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilRewardTxProcessor(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.RewardsTxProcessor = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilRewardsTxProcessor, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilRequestHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.RequestHandler = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilFeeHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.EconomicsFee = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilGasHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.GasHandler = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilGasHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBlockTracker(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.BlockTracker = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilBlockTracker, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBlockSizeComputationHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.BlockSizeComputation = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilBalanceComputationHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.BalanceComputation = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.EnableEpochsHandler = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilEnableRoundsHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.EnableRoundsHandler = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilEnableRoundsHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.TxTypeHandler = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilTxTypeHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.ScheduledTxsExecutionHandler = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilScheduledTxsExecutionHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilProcessedMiniBlocksTracker(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.ProcessedMiniBlocksTracker = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilProcessedMiniBlocksTracker, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory_NilTxExecutionOrderHandler(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	args.TxExecutionOrderHandler = nil
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Equal(t, process.ErrNilTxExecutionOrderHandler, err)
	assert.Nil(t, ppcm)
}

func TestNewPreProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	args := createPreProcessorsContainerFactoryArgs()
	ppcm, err := NewPreProcessorsContainerFactory(args)

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

	args := createPreProcessorsContainerFactoryArgs()
	args.DataPool = dataPool
	ppcm, err := NewPreProcessorsContainerFactory(args)

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
	args := createPreProcessorsContainerFactoryArgs()
	args.DataPool = dataPool
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)

	container, err := ppcm.Create()
	assert.Nil(t, container)
	assert.Equal(t, process.ErrNilTransactionPool, err)
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

	args := createPreProcessorsContainerFactoryArgs()
	args.DataPool = dataPool
	ppcm, err := NewPreProcessorsContainerFactory(args)

	assert.Nil(t, err)
	assert.NotNil(t, ppcm)

	container, err := ppcm.Create()
	assert.Nil(t, err)
	assert.Equal(t, 4, container.Len())
}

func createPreProcessorsContainerFactoryArgs() ArgsPreProcessorsContainerFactory {
	return ArgsPreProcessorsContainerFactory{
		ShardCoordinator:             mock.NewMultiShardsCoordinatorMock(3),
		Store:                        &storageStubs.ChainStorerStub{},
		Marshalizer:                  &mock.MarshalizerMock{},
		Hasher:                       &hashingMocks.HasherMock{},
		DataPool:                     dataRetrieverMock.NewPoolsHolderMock(),
		PubkeyConverter:              createMockPubkeyConverter(),
		Accounts:                     &stateMock.AccountsStub{},
		AccountsProposal:             &stateMock.AccountsStub{},
		RequestHandler:               &testscommon.RequestHandlerStub{},
		TxProcessor:                  &testscommon.TxProcessorMock{},
		ScProcessor:                  &testscommon.SCProcessorMock{},
		ScResultProcessor:            &testscommon.SmartContractResultsProcessorMock{},
		RewardsTxProcessor:           &testscommon.RewardTxProcessorMock{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerMock{},
		GasHandler:                   &testscommon.GasHandlerStub{},
		BlockTracker:                 &mock.BlockTrackerMock{},
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
