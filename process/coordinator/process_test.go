package coordinator

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const MaxGasLimitPerBlock = uint64(100000)

var txHash = []byte("tx_hash1")

func FeeHandlerMock() *economicsmocks.EconomicsHandlerStub {
	return &economicsmocks.EconomicsHandlerStub{
		ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
			return 0
		},
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return MaxGasLimitPerBlock
		},
		MaxGasLimitPerMiniBlockCalled: func() uint64 {
			return MaxGasLimitPerBlock
		},
		MaxGasLimitPerBlockForSafeCrossShardCalled: func() uint64 {
			return MaxGasLimitPerBlock
		},
		MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
			return MaxGasLimitPerBlock
		},
		MaxGasLimitPerTxCalled: func() uint64 {
			return MaxGasLimitPerBlock
		},
	}
}

func createShardedDataChacherNotifier(
	handler data.TransactionHandler,
	testHash []byte,
) func() dataRetriever.ShardedDataCacherNotifier {
	return func() dataRetriever.ShardedDataCacherNotifier {
		return &testscommon.ShardedDataStub{
			RegisterOnAddedCalled: func(i func(key []byte, value interface{})) {},
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return &testscommon.CacherStub{
					PeekCalled: func(key []byte) (value interface{}, ok bool) {
						if reflect.DeepEqual(key, testHash) {
							return handler, true
						}
						return nil, false
					},
					KeysCalled: func() [][]byte {
						return [][]byte{[]byte("key1"), []byte("key2")}
					},
					LenCalled: func() int {
						return 0
					},
				}
			},
			RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {},
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				if reflect.DeepEqual(key, []byte("tx1_hash")) {
					return handler, true
				}
				return nil, false
			},
			AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheId string) {
			},
		}
	}
}

func initDataPool(testHash []byte) *dataRetrieverMock.PoolsHolderStub {
	tx := &transaction.Transaction{
		Nonce: 10,
		Value: big.NewInt(0),
	}
	sc := &smartContractResult.SmartContractResult{Nonce: 10, SndAddr: []byte("0"), RcvAddr: []byte("1")}
	rTx := &rewardTx.RewardTx{Epoch: 0, Round: 1, RcvAddr: []byte("1")}

	txCalled := createShardedDataChacherNotifier(tx, testHash)
	unsignedTxHandler := createShardedDataChacherNotifier(sc, testHash)
	rewardTxCalled := createShardedDataChacherNotifier(rTx, testHash)

	sdp := &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled:         txCalled,
		UnsignedTransactionsCalled: unsignedTxHandler,
		RewardTransactionsCalled:   rewardTxCalled,
		MetaBlocksCalled: func() storage.Cacher {
			return &testscommon.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				KeysCalled: func() [][]byte {
					return nil
				},
				LenCalled: func() int {
					return 0
				},
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				RegisterHandlerCalled: func(i func(key []byte, value interface{})) {},
			}
		},
		MiniBlocksCalled: func() storage.Cacher {
			cs := testscommon.NewCacherStub()
			cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {
			}
			cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("bbb"), key) {
					return make(block.MiniBlockSlice, 0), true
				}

				return nil, false
			}
			cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("bbb"), key) {
					return make(block.MiniBlockSlice, 0), true
				}

				return nil, false
			}
			cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {}
			cs.RemoveCalled = func(key []byte) {}
			cs.PutCalled = func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				return false
			}
			return cs
		},
		HeadersCalled: func() dataRetriever.HeadersPool {
			cs := &mock.HeadersCacherStub{}
			cs.RegisterHandlerCalled = func(i func(header data.HeaderHandler, key []byte)) {
			}
			return cs
		},
		CurrBlockTxsCalled: func() dataRetriever.TransactionCacher {
			return &mock.TxForCurrentBlockStub{}
		},
	}
	return sdp
}

func initStore() *dataRetriever.ChainStorer {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.RewardTransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, generateTestUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ReceiptsUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ScheduledSCRsUnit, generateTestUnit())
	return store
}

func generateTestCache() storage.Cacher {
	cache, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000, Shards: 1, SizeInBytes: 0})
	return cache
}

func generateTestUnit() storage.Storer {
	storer, _ := storageunit.NewStorageUnit(
		generateTestCache(),
		database.NewMemDB(),
	)

	return storer
}

func initAccountsMock() *stateMock.AccountsStub {
	rootHashCalled := func() ([]byte, error) {
		return []byte("rootHash"), nil
	}
	return &stateMock.AccountsStub{
		RootHashCalled: rootHashCalled,
	}
}

func createMockTransactionCoordinatorArguments() ArgTransactionCoordinator {
	argsTransactionCoordinator := ArgTransactionCoordinator{
		Hasher:                       &hashingMocks.HasherMock{},
		Marshalizer:                  &mock.MarshalizerMock{},
		ShardCoordinator:             mock.NewMultiShardsCoordinatorMock(5),
		Accounts:                     &stateMock.AccountsStub{},
		MiniBlockPool:                dataRetrieverMock.NewPoolsHolderMock().MiniBlocks(),
		RequestHandler:               &testscommon.RequestHandlerStub{},
		PreProcessors:                &mock.PreProcessorContainerMock{},
		InterProcessors:              &mock.InterimProcessorContainerMock{},
		GasHandler:                   &testscommon.GasHandlerStub{},
		FeeHandler:                   &mock.FeeAccumulatorStub{},
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}

	return argsTransactionCoordinator
}

func TestNewTransactionCoordinator_NilHasher(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.Hasher = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewTransactionCoordinator_TxLogProcessor(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.TransactionsLogProcessor = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilTxLogsProcessor, err)
}

func TestNewTransactionCoordinator_NilMarshalizer(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.Marshalizer = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewTransactionCoordinator_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewTransactionCoordinator_NilAccountsStub(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.Accounts = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewTransactionCoordinator_NilDataPool(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.MiniBlockPool = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilMiniBlockPool, err)
}

func TestNewTransactionCoordinator_NilRequestHandler(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.RequestHandler = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestNewTransactionCoordinator_NilPreProcessor(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.PreProcessors = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilPreProcessorsContainer, err)
}

func TestNewTransactionCoordinator_NilInterProcessor(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.InterProcessors = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilIntermediateProcessorContainer, err)
}

func TestNewTransactionCoordinator_NilGasHandler(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.GasHandler = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilGasHandler, err)
}

func TestNewTransactionCoordinator_NilFeeAccumulator(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.FeeHandler = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewTransactionCoordinator_NilBlockSizeComputation(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.BlockSizeComputation = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
}

func TestNewTransactionCoordinator_NilBalanceComputation(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.BalanceComputation = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
}

func TestNewTransactionCoordinator_NilEconomicsFee(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.EconomicsFee = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewTransactionCoordinator_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.TxTypeHandler = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilTxTypeHandler, err)
}

func TestNewTransactionCoordinator_NilTxLogsProcessor(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.TransactionsLogProcessor = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilTxLogsProcessor, err)
}

func TestNewTransactionCoordinator_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.EnableEpochsHandler = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewTransactionCoordinator_InvalidEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
}

func TestNewTransactionCoordinator_NilScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ScheduledTxsExecutionHandler = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilScheduledTxsExecutionHandler, err)
}

func TestNewTransactionCoordinator_NilDoubleTransactionsDetector(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.DoubleTransactionsDetector = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.True(t, check.IfNil(tc))
	assert.Equal(t, process.ErrNilDoubleTransactionsDetector, err)
}

func TestNewTransactionCoordinator_NilProcessedMiniBlocksTracker(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ProcessedMiniBlocksTracker = nil
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.True(t, check.IfNil(tc))
	assert.Equal(t, process.ErrNilProcessedMiniBlocksTracker, err)
}

func TestNewTransactionCoordinator_OK(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(tc))
}

func TestTransactionCoordinator_GetAllCurrentLogs(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.TransactionsLogProcessor = &mock.TxLogsProcessorStub{
		GetAllCurrentLogsCalled: func() []*data.LogData {
			return []*data.LogData{}
		},
	}

	tc, _ := NewTransactionCoordinator(argsTransactionCoordinator)

	logs := tc.GetAllCurrentLogs()
	require.NotNil(t, logs)
}

func TestTransactionCoordinator_SeparateBody(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.TxBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.TxBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.TxBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.SmartContractResultBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.SmartContractResultBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.SmartContractResultBlock})
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.SmartContractResultBlock})

	separated := tc.separateBodyByType(body)
	assert.Equal(t, 2, len(separated))
	assert.Equal(t, 3, len(separated[block.TxBlock].MiniBlocks))
	assert.Equal(t, 4, len(separated[block.SmartContractResultBlock].MiniBlocks))
}

func createPreProcessorContainer() process.PreProcessorsContainer {
	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		initDataPool([]byte("tx_hash0")),
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)
	container, _ := preFactory.Create()

	return container
}

func createInterimProcessorContainer() process.IntermediateProcessorContainer {
	argsFactory := shard.ArgsNewIntermediateProcessorsContainerFactory{
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(5),
		Marshalizer:      &mock.MarshalizerMock{},
		Hasher:           &hashingMocks.HasherMock{},
		PubkeyConverter:  createMockPubkeyConverter(),
		Store:            initStore(),
		PoolsHolder:      initDataPool([]byte("test_hash1")),
		EconomicsFee:     &economicsmocks.EconomicsHandlerStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.KeepExecOrderOnCreatedSCRsFlag
			},
		},
	}
	preFactory, _ := shard.NewIntermediateProcessorsContainerFactory(argsFactory)
	container, _ := preFactory.Create()

	return container
}

func createPreProcessorContainerWithDataPool(
	dataPool dataRetriever.PoolsHolder,
	feeHandler process.FeeHandler,
) process.PreProcessorsContainer {

	totalGasProvided := uint64(0)
	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataPool,
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&testscommon.GasHandlerStub{
			SetGasProvidedCalled: func(gasProvided uint64, hash []byte) {
				totalGasProvided += gasProvided
			},
			TotalGasProvidedCalled: func() uint64 {
				return totalGasProvided
			},
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				tx, ok := txHandler.(*transaction.Transaction)
				if !ok {
					return 0, 0, process.ErrWrongTypeAssertion
				}

				txGasLimitConsumption := feeHandler.ComputeGasLimit(tx)
				if tx.GasLimit < txGasLimitConsumption {
					return 0, 0, process.ErrInsufficientGasLimitInTx
				}

				if core.IsSmartContractAddress(tx.RcvAddr) {
					if txSenderShardId != txReceiverShardId {
						gasProvidedByTxInSenderShard := txGasLimitConsumption
						gasProvidedByTxInReceiverShard := tx.GasLimit - txGasLimitConsumption

						return gasProvidedByTxInSenderShard, gasProvidedByTxInReceiverShard, nil
					}

					return tx.GasLimit, tx.GasLimit, nil
				}

				return txGasLimitConsumption, txGasLimitConsumption, nil
			},
			ComputeGasProvidedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			GasRefundedCalled: func(hash []byte) uint64 {
				return 0
			},
			RemoveGasProvidedCalled: func(hashes [][]byte) {
			},
			RemoveGasRefundedCalled: func(hashes [][]byte) {
			},
		},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)
	container, _ := preFactory.Create()

	return container
}

func TestTransactionCoordinator_CreateBlockStarted(t *testing.T) {
	t.Parallel()

	totalGasProvided := uint64(0)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.GasHandler = &testscommon.GasHandlerStub{
		InitCalled: func() {
			totalGasProvided = uint64(0)
		},
		TotalGasProvidedCalled: func() uint64 {
			return totalGasProvided
		},
	}
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainer()
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tc.CreateBlockStarted()

	tc.mutPreProcessor.Lock()
	for _, value := range tc.txPreProcessors {
		txs := value.GetAllCurrentUsedTxs()
		assert.Equal(t, 0, len(txs))
	}
	tc.mutPreProcessor.Unlock()
}

func TestTransactionCoordinator_CreateMarshalizedDataNilBody(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainer()
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	mrTxs := tc.CreateMarshalizedData(nil)
	assert.Equal(t, 0, len(mrTxs))
}

func createMiniBlockWithOneTx(sndId, dstId uint32, blockType block.Type, txHash []byte) *block.MiniBlock {
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	return &block.MiniBlock{Type: blockType, SenderShardID: sndId, ReceiverShardID: dstId, TxHashes: txHashes}
}

func createTestBody() *block.Body {
	body := &block.Body{}

	body.MiniBlocks = append(body.MiniBlocks, createMiniBlockWithOneTx(0, 1, block.TxBlock, []byte("tx_hash1")))
	body.MiniBlocks = append(body.MiniBlocks, createMiniBlockWithOneTx(0, 1, block.TxBlock, []byte("tx_hash2")))
	body.MiniBlocks = append(body.MiniBlocks, createMiniBlockWithOneTx(0, 1, block.TxBlock, []byte("tx_hash3")))
	body.MiniBlocks = append(body.MiniBlocks, createMiniBlockWithOneTx(0, 1, block.SmartContractResultBlock, []byte("tx_hash4")))
	body.MiniBlocks = append(body.MiniBlocks, createMiniBlockWithOneTx(0, 1, block.SmartContractResultBlock, []byte("tx_hash5")))
	body.MiniBlocks = append(body.MiniBlocks, createMiniBlockWithOneTx(0, 1, block.SmartContractResultBlock, []byte("tx_hash6")))

	return body
}

func TestTransactionCoordinator_CreateMarshalizedData(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainer()
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	mrTxs := tc.CreateMarshalizedData(createTestBody())
	assert.Equal(t, 0, len(mrTxs))
}

func TestTransactionCoordinator_CreateMarshalizedDataWithTxsAndScr(t *testing.T) {
	t.Parallel()

	interimContainer := createInterimProcessorContainer()
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainer()
	argsTransactionCoordinator.InterProcessors = interimContainer
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	scrs := make([]data.TransactionHandler, 0)
	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, createMiniBlockWithOneTx(0, 1, block.TxBlock, txHash))

	scr := &smartContractResult.SmartContractResult{SndAddr: []byte("snd"), RcvAddr: []byte("rcv"), Value: big.NewInt(99), PrevTxHash: []byte("txHash")}
	scrHash, _ := core.CalculateHash(&mock.MarshalizerMock{}, &hashingMocks.HasherMock{}, scr)
	scrs = append(scrs, scr)
	body.MiniBlocks = append(body.MiniBlocks, createMiniBlockWithOneTx(0, 1, block.SmartContractResultBlock, scrHash))

	scr = &smartContractResult.SmartContractResult{SndAddr: []byte("snd"), RcvAddr: []byte("rcv"), Value: big.NewInt(199), PrevTxHash: []byte("txHash")}
	scrHash, _ = core.CalculateHash(&mock.MarshalizerMock{}, &hashingMocks.HasherMock{}, scr)
	scrs = append(scrs, scr)
	body.MiniBlocks = append(body.MiniBlocks, createMiniBlockWithOneTx(0, 1, block.SmartContractResultBlock, scrHash))

	scr = &smartContractResult.SmartContractResult{SndAddr: []byte("snd"), RcvAddr: []byte("rcv"), Value: big.NewInt(299), PrevTxHash: []byte("txHash")}
	scrHash, _ = core.CalculateHash(&mock.MarshalizerMock{}, &hashingMocks.HasherMock{}, scr)
	scrs = append(scrs, scr)
	body.MiniBlocks = append(body.MiniBlocks, createMiniBlockWithOneTx(0, 1, block.SmartContractResultBlock, scrHash))

	scrInterimProc, _ := interimContainer.Get(block.SmartContractResultBlock)
	_ = scrInterimProc.AddIntermediateTransactions(scrs)

	mrTxs := tc.CreateMarshalizedData(body)
	assert.Equal(t, 1, len(mrTxs))

	marshalizer := &mock.MarshalizerMock{}
	topic := factory.UnsignedTransactionTopic + "_0_1"
	assert.Equal(t, len(scrs), len(mrTxs[topic]))
	for i := 0; i < len(mrTxs[topic]); i++ {
		unMrsScr := &smartContractResult.SmartContractResult{}
		_ = marshalizer.Unmarshal(unMrsScr, mrTxs[topic][i])

		assert.Equal(t, unMrsScr, scrs[i])
	}
}

func TestTransactionCoordinator_CreateMbsAndProcessCrossShardTransactionsDstMeNilHeader(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainer()
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	haveAdditionalTime := func() bool {
		return false
	}
	mbs, txs, finalized, err := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(nil, nil, haveTime, haveAdditionalTime, false)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, uint32(0), txs)
	assert.False(t, finalized)
}

func createTestMetablock() *block.MetaBlock {
	meta := &block.MetaBlock{}

	meta.ShardInfo = make([]block.ShardData, 0)

	shardMbs := make([]block.MiniBlockHeader, 0)
	shardMbs = append(shardMbs, block.MiniBlockHeader{Hash: []byte("mb0"), SenderShardID: 0, ReceiverShardID: 0, TxCount: 1})
	shardMbs = append(shardMbs, block.MiniBlockHeader{Hash: []byte("mb1"), SenderShardID: 0, ReceiverShardID: 1, TxCount: 1})
	shardData := block.ShardData{ShardID: 0, HeaderHash: []byte("header0"), TxCount: 2, ShardMiniBlockHeaders: shardMbs}

	meta.ShardInfo = append(meta.ShardInfo, shardData)

	shardMbs = make([]block.MiniBlockHeader, 0)
	shardMbs = append(shardMbs, block.MiniBlockHeader{Hash: []byte("mb2"), SenderShardID: 1, ReceiverShardID: 0, TxCount: 1})
	shardMbs = append(shardMbs, block.MiniBlockHeader{Hash: []byte("mb3"), SenderShardID: 1, ReceiverShardID: 1, TxCount: 1})
	shardData = block.ShardData{ShardID: 1, HeaderHash: []byte("header0"), TxCount: 2, ShardMiniBlockHeaders: shardMbs}

	meta.ShardInfo = append(meta.ShardInfo, shardData)

	return meta
}

func TestTransactionCoordinator_CreateMbsAndProcessCrossShardTransactionsDstMeNoTime(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainer()
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return false
	}
	haveAdditionalTime := func() bool {
		return false
	}
	mbs, txs, finalized, err := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(createTestMetablock(), nil, haveTime, haveAdditionalTime, false)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, uint32(0), txs)
	assert.False(t, finalized)
}

func TestTransactionCoordinator_CreateMbsAndProcessCrossShardTransactionsNothingInPool(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainer()
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	haveAdditionalTime := func() bool {
		return false
	}
	mbs, txs, finalized, err := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(createTestMetablock(), nil, haveTime, haveAdditionalTime, false)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, uint32(0), txs)
	assert.False(t, finalized)
}

func TestTransactionCoordinator_CreateMbsAndProcessCrossShardTransactions(t *testing.T) {
	t.Parallel()

	tdp := initDataPool(txHash)
	cacherCfg := storageunit.CacheConfig{Capacity: 100, Type: storageunit.LRUCache}
	hdrPool, _ := storageunit.NewCache(cacherCfg)
	tdp.MiniBlocksCalled = func() storage.Cacher {
		return hdrPool
	}

	totalGasProvided := uint64(0)
	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		tdp,
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&testscommon.GasHandlerStub{
			SetGasProvidedCalled: func(gasProvided uint64, hash []byte) {
				totalGasProvided += gasProvided
			},
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasProvidedCalled: func() uint64 {
				return totalGasProvided
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			TotalGasRefundedCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)
	container, _ := preFactory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = container
	argsTransactionCoordinator.GasHandler = &testscommon.GasHandlerStub{
		TotalGasProvidedCalled: func() uint64 {
			return totalGasProvided
		},
	}
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	haveAdditionalTime := func() bool {
		return false
	}
	metaHdr := createTestMetablock()

	for i := 0; i < len(metaHdr.ShardInfo); i++ {
		for j := 0; j < len(metaHdr.ShardInfo[i].ShardMiniBlockHeaders); j++ {
			mbHdr := metaHdr.ShardInfo[i].ShardMiniBlockHeaders[j]
			mb := block.MiniBlock{SenderShardID: mbHdr.SenderShardID, ReceiverShardID: mbHdr.ReceiverShardID, Type: block.TxBlock, TxHashes: [][]byte{txHash}}
			tdp.MiniBlocks().Put(mbHdr.Hash, &mb, mb.Size())
		}
	}

	mbs, txs, finalized, err := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(metaHdr, nil, haveTime, haveAdditionalTime, false)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(mbs))
	assert.Equal(t, uint32(1), txs)
	assert.True(t, finalized)
}

func TestTransactionCoordinator_CreateMbsAndProcessCrossShardTransactionsWithSkippedShard(t *testing.T) {
	t.Parallel()

	mbPool := dataRetrieverMock.NewPoolsHolderMock().MiniBlocks()
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.MiniBlockPool = mbPool
	tc, _ := NewTransactionCoordinator(argsTransactionCoordinator)

	tc.txPreProcessors[block.TxBlock] = &mock.PreProcessorMock{
		RequestTransactionsForMiniBlockCalled: func(miniBlock *block.MiniBlock) int {
			return 0
		},
	}

	haveTime := func() bool {
		return true
	}
	haveAdditionalTime := func() bool {
		return false
	}

	metaBlock := &block.MetaBlock{}
	metaBlock.ShardInfo = make([]block.ShardData, 0)
	shardMbs := make([]block.MiniBlockHeader, 0)
	shardMbs = append(shardMbs, block.MiniBlockHeader{Hash: []byte("mb0"), SenderShardID: 1, ReceiverShardID: 0, TxCount: 1})
	shardMbs = append(shardMbs, block.MiniBlockHeader{Hash: []byte("mb1"), SenderShardID: 1, ReceiverShardID: 0, TxCount: 1})
	shardMbs = append(shardMbs, block.MiniBlockHeader{Hash: []byte("mb2"), SenderShardID: 1, ReceiverShardID: 0, TxCount: 1})
	shardData := block.ShardData{ShardID: 1, HeaderHash: []byte("header0"), TxCount: 3, ShardMiniBlockHeaders: shardMbs}

	metaBlock.ShardInfo = append(metaBlock.ShardInfo, shardData)

	for i := 0; i < len(metaBlock.ShardInfo); i++ {
		for j := 0; j < len(metaBlock.ShardInfo[i].ShardMiniBlockHeaders); j++ {
			mbHdr := metaBlock.ShardInfo[i].ShardMiniBlockHeaders[j]
			if bytes.Equal(mbHdr.Hash, []byte("mb1")) {
				continue
			}

			hash := fmt.Sprintf("tx_hash_from_%s", mbHdr.Hash)
			mb := block.MiniBlock{SenderShardID: mbHdr.SenderShardID, ReceiverShardID: mbHdr.ReceiverShardID, Type: block.TxBlock, TxHashes: [][]byte{[]byte(hash)}}
			mbPool.Put(mbHdr.Hash, &mb, mb.Size())
		}
	}

	mbs, txs, finalized, err := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(metaBlock, nil, haveTime, haveAdditionalTime, false)
	assert.Nil(t, err)
	require.Equal(t, 1, len(mbs))
	assert.Equal(t, uint32(1), txs)
	assert.False(t, finalized)
	require.Equal(t, 1, len(mbs[0].TxHashes))
	assert.Equal(t, []byte("tx_hash_from_mb0"), mbs[0].TxHashes[0])
}

func TestTransactionCoordinator_HandleProcessMiniBlockInit(t *testing.T) {
	mbHash := []byte("miniblock hash")
	numResetGasHandler := 0
	numInitInterimProc := 0
	shardCoord := testscommon.NewMultiShardsCoordinatorMock(1)
	tc := &transactionCoordinator{
		accounts: initAccountsMock(),
		gasHandler: &testscommon.GasHandlerStub{
			ResetCalled: func(key []byte) {
				assert.Equal(t, key, mbHash)
				numResetGasHandler++
			},
		},
		keysInterimProcs: []block.Type{block.SmartContractResultBlock},
		interimProcessors: map[block.Type]process.IntermediateTransactionHandler{
			block.SmartContractResultBlock: &mock.IntermediateTransactionHandlerStub{
				InitProcessedResultsCalled: func(key []byte) {
					assert.Equal(t, mbHash, key)
					numInitInterimProc++
				},
			},
		},
		shardCoordinator: shardCoord,
	}

	t.Run("shard 0 should call init", func(t *testing.T) {
		numResetGasHandler = 0
		numInitInterimProc = 0
		shardCoord.CurrentShard = 0

		tc.handleProcessMiniBlockInit(mbHash)
		assert.Equal(t, 1, numResetGasHandler)
		assert.Equal(t, 1, numInitInterimProc)
	})
	t.Run("shard meta should call init", func(t *testing.T) {
		numResetGasHandler = 0
		numInitInterimProc = 0
		shardCoord.CurrentShard = core.MetachainShardId

		tc.handleProcessMiniBlockInit(mbHash)
		assert.Equal(t, 1, numResetGasHandler)
		assert.Equal(t, 1, numInitInterimProc)
	})
}

func TestTransactionCoordinator_CreateMbsAndProcessCrossShardTransactionsNilPreProcessor(t *testing.T) {
	t.Parallel()

	tdp := initDataPool(txHash)
	cacherCfg := storageunit.CacheConfig{Capacity: 100, Type: storageunit.LRUCache}
	hdrPool, _ := storageunit.NewCache(cacherCfg)
	tdp.MiniBlocksCalled = func() storage.Cacher {
		return hdrPool
	}

	totalGasProvided := uint64(0)
	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		tdp,
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&testscommon.GasHandlerStub{
			SetGasProvidedCalled: func(gasProvided uint64, hash []byte) {
				totalGasProvided += gasProvided
			},
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasProvidedCalled: func() uint64 {
				return totalGasProvided
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			TotalGasRefundedCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)
	container, _ := preFactory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = container
	argsTransactionCoordinator.GasHandler = &testscommon.GasHandlerStub{
		TotalGasProvidedCalled: func() uint64 {
			return totalGasProvided
		},
	}
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	haveAdditionalTime := func() bool {
		return false
	}
	metaHdr := createTestMetablock()

	unknownPreprocessorType := block.Type(254)
	for i := 0; i < len(metaHdr.ShardInfo); i++ {
		for j := 0; j < len(metaHdr.ShardInfo[i].ShardMiniBlockHeaders); j++ {
			mbHdr := metaHdr.ShardInfo[i].ShardMiniBlockHeaders[j]
			mb := block.MiniBlock{SenderShardID: mbHdr.SenderShardID, ReceiverShardID: mbHdr.ReceiverShardID, Type: unknownPreprocessorType, TxHashes: [][]byte{txHash}}
			tdp.MiniBlocks().Put(mbHdr.Hash, &mb, mb.Size())
		}
	}

	mbs, txs, finalized, err := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(metaHdr, nil, haveTime, haveAdditionalTime, false)

	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, process.ErrNilPreProcessor))
	assert.Nil(t, mbs)
	assert.Equal(t, uint32(0), txs)
	assert.False(t, finalized)
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMeNothingToProcess(t *testing.T) {
	t.Parallel()

	shardedCacheMock := &testscommon.ShardedDataStub{
		RegisterOnAddedCalled: func(i func(key []byte, value interface{})) {},
		ShardDataStoreCalled: func(id string) (c storage.Cacher) {
			return &testscommon.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, false
				},
				KeysCalled: func() [][]byte {
					return nil
				},
				LenCalled: func() int {
					return 0
				},
			}
		},
		RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {},
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
		AddDataCalled: func(_ []byte, _ interface{}, _ int, _ string) {
		},
	}

	totalGasProvided := uint64(0)
	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&dataRetrieverMock.PoolsHolderStub{
			TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return shardedCacheMock
			},
			UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return shardedCacheMock
			},
			RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return shardedCacheMock
			},
		},
		createMockPubkeyConverter(),
		&stateMock.AccountsStub{},
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&testscommon.GasHandlerStub{
			TotalGasProvidedCalled: func() uint64 {
				return totalGasProvided
			},
		},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)
	container, _ := preFactory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.MiniBlockPool = dataRetrieverMock.NewPoolsHolderMock().MiniBlocks()
	argsTransactionCoordinator.PreProcessors = container
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime, []byte("randomness"))

	assert.Equal(t, 0, len(mbs))
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMeNoTime(t *testing.T) {
	t.Parallel()
	tdp := initDataPool(txHash)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return false
	}
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime, []byte("randomness"))

	assert.Equal(t, 0, len(mbs))
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMeNoSpace(t *testing.T) {
	t.Parallel()
	totalGasProvided := uint64(0)
	tdp := initDataPool(txHash)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	argsTransactionCoordinator.GasHandler = &testscommon.GasHandlerStub{
		TotalGasProvidedCalled: func() uint64 {
			return totalGasProvided
		},
	}
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime, []byte("randomness"))

	assert.Equal(t, 0, len(mbs))
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMe(t *testing.T) {
	t.Parallel()

	nrShards := uint32(5)
	txPool, _ := dataRetrieverMock.CreateTxPool(nrShards, 0)
	tdp := initDataPool(txHash)
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return txPool
	}

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(nrShards)
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}

	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	for shId := uint32(0); shId < nrShards; shId++ {
		strCache := process.ShardCacherIdentifier(0, shId)
		newTx := &transaction.Transaction{GasLimit: uint64(shId)}

		computedTxHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		txPool.AddData(computedTxHash, newTx, newTx.Size(), strCache)
	}

	// we have one tx per shard.
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime, []byte("randomness"))

	assert.Equal(t, int(nrShards), len(mbs))
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMeMultipleMiniblocks(t *testing.T) {
	t.Parallel()

	nrShards := uint32(5)
	txPool, _ := dataRetrieverMock.CreateTxPool(nrShards, 0)
	tdp := initDataPool(txHash)
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return txPool
	}

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(nrShards)
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}

	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	sndShardId := uint32(0)
	dstShardId := uint32(1)
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)

	numTxsToAdd := 100
	gasLimit := MaxGasLimitPerBlock / uint64(numTxsToAdd)

	scAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")

	allTxs := 100
	for i := 0; i < allTxs; i++ {
		newTx := &transaction.Transaction{GasLimit: gasLimit, GasPrice: uint64(i), RcvAddr: scAddress}

		computedTxHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		txPool.AddData(computedTxHash, newTx, newTx.Size(), strCache)
	}

	// we have one tx per shard.
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime, []byte("randomness"))

	assert.Equal(t, 1, len(mbs))
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMeMultipleMiniblocksShouldApplyGasLimit(t *testing.T) {
	t.Parallel()

	allTxs := 100
	numTxsToAdd := 20
	gasLimit := MaxGasLimitPerBlock / uint64(numTxsToAdd)
	numMiniBlocks := allTxs / numTxsToAdd

	nrShards := uint32(5)
	txPool, _ := dataRetrieverMock.CreateTxPool(nrShards, 0)
	tdp := initDataPool(txHash)
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return txPool
	}

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(nrShards)
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(
		tdp,
		&economicsmocks.EconomicsHandlerStub{
			MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerTxCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return gasLimit / uint64(numMiniBlocks)
			},
		})
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}

	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	sndShardId := uint32(0)
	dstShardId := uint32(1)
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)

	scAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")

	for i := 0; i < allTxs; i++ {
		newTx := &transaction.Transaction{GasLimit: gasLimit + gasLimit/uint64(numMiniBlocks), GasPrice: uint64(i), RcvAddr: scAddress}

		computedTxHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		txPool.AddData(computedTxHash, newTx, newTx.Size(), strCache)
	}

	// we have one tx per shard.
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime, []byte("randomness"))

	assert.Equal(t, 1, len(mbs))
}

func TestTransactionCoordinator_CompactAndExpandMiniblocksShouldWork(t *testing.T) {
	t.Parallel()

	numTxsPerBulk := 100
	numTxsToAdd := 20
	gasLimit := MaxGasLimitPerBlock / uint64(numTxsToAdd)

	nrShards := uint32(5)
	txPool, _ := dataRetrieverMock.CreateTxPool(nrShards, 0)
	tdp := initDataPool(txHash)
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return txPool
	}

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(nrShards)
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(
		tdp,
		&economicsmocks.EconomicsHandlerStub{
			MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			MaxGasLimitPerTxCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 0
			},
		})
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}

	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	// set more identifiers to match both scenarios: intra-shard txs and cross-shard txs.
	var shardCacherIdentifiers []string
	shardCacherIdentifiers = append(shardCacherIdentifiers, process.ShardCacherIdentifier(0, 0))
	shardCacherIdentifiers = append(shardCacherIdentifiers, process.ShardCacherIdentifier(0, 1))
	shardCacherIdentifiers = append(shardCacherIdentifiers, process.ShardCacherIdentifier(0, 2))
	shardCacherIdentifiers = append(shardCacherIdentifiers, process.ShardCacherIdentifier(0, 3))
	shardCacherIdentifiers = append(shardCacherIdentifiers, process.ShardCacherIdentifier(0, 4))

	scAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")

	for _, shardCacher := range shardCacherIdentifiers {
		for i := 0; i < numTxsPerBulk; i++ {
			newTx := &transaction.Transaction{GasLimit: gasLimit, GasPrice: uint64(i), RcvAddr: scAddress}

			computedTxHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
			txPool.AddData(computedTxHash, newTx, newTx.Size(), shardCacher)
		}
	}

	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime, []byte("randomness"))

	assert.Equal(t, 1, len(mbs))
}

func TestTransactionCoordinator_GetAllCurrentUsedTxs(t *testing.T) {
	t.Parallel()

	nrShards := uint32(5)
	txPool, _ := dataRetrieverMock.CreateTxPool(nrShards, 0)
	tdp := initDataPool(txHash)
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return txPool
	}

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(nrShards)
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	argsTransactionCoordinator.GasHandler = &testscommon.GasHandlerStub{
		ComputeGasProvidedByTxCalled: func(txSndShId uint32, txRcvShId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return 0, 0, nil
		},
	}
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	usedTxs := tc.GetAllCurrentUsedTxs(block.TxBlock)
	assert.Equal(t, 0, len(usedTxs))

	// create block to have some txs
	haveTime := func() bool {
		return true
	}

	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	for i := uint32(0); i < nrShards; i++ {
		strCache := process.ShardCacherIdentifier(0, i)
		newTx := &transaction.Transaction{GasLimit: uint64(i)}

		computedTxHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		txPool.AddData(computedTxHash, newTx, newTx.Size(), strCache)
	}

	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime, []byte("randomness"))
	require.Equal(t, 5, len(mbs))

	usedTxs = tc.GetAllCurrentUsedTxs(block.TxBlock)
	require.Equal(t, 5, len(usedTxs))
}

func TestTransactionCoordinator_RequestBlockTransactionsNilBody(t *testing.T) {
	t.Parallel()

	tdp := initDataPool(txHash)
	nrShards := uint32(5)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(nrShards)
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tc.RequestBlockTransactions(nil)

	tc.mutRequestedTxs.Lock()
	for _, value := range tc.requestedTxs {
		assert.Equal(t, 0, value)
	}
	tc.mutRequestedTxs.Unlock()
}

func TestTransactionCoordinator_RequestBlockTransactionsRequestOne(t *testing.T) {
	t.Parallel()

	tdp := initDataPool(txHash)
	nrShards := uint32(5)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(nrShards)
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	body := &block.Body{}
	txHashToAsk := []byte("tx_hashnotinPool")
	miniBlock := &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHash, txHashToAsk}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)
	tc.RequestBlockTransactions(body)

	tc.mutRequestedTxs.Lock()
	assert.Equal(t, 1, tc.requestedTxs[block.TxBlock])
	tc.mutRequestedTxs.Unlock()

	haveTime := func() time.Duration {
		return time.Second
	}
	err = tc.IsDataPreparedForProcessing(haveTime)
	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestTransactionCoordinator_IsDataPreparedForProcessing(t *testing.T) {
	t.Parallel()

	tdp := initDataPool(txHash)
	nrShards := uint32(5)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(nrShards)
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() time.Duration {
		return time.Second
	}
	err = tc.IsDataPreparedForProcessing(haveTime)
	assert.Nil(t, err)
}

func TestTransactionCoordinator_SaveTxsToStorage(t *testing.T) {
	t.Parallel()

	tdp := initDataPool(txHash)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	argsTransactionCoordinator.Accounts = initAccountsMock()
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have panic")
		}
	}()

	tc.SaveTxsToStorage(nil)

	body := &block.Body{}
	miniBlock := &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHash}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	tc.RequestBlockTransactions(body)
	tc.SaveTxsToStorage(body)

	txHashToAsk := []byte("tx_hashnotinPool")
	miniBlock = &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHashToAsk}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	tc.SaveTxsToStorage(body)
}

func TestTransactionCoordinator_RestoreBlockDataFromStorage(t *testing.T) {
	t.Parallel()

	tdp := initDataPool(txHash)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	argsTransactionCoordinator.Accounts = initAccountsMock()
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	nrTxs, err := tc.RestoreBlockDataFromStorage(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, nrTxs)

	body := &block.Body{}
	miniBlock := &block.MiniBlock{SenderShardID: 1, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHash}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	tc.RequestBlockTransactions(body)
	tc.SaveTxsToStorage(body)
	nrTxs, err = tc.RestoreBlockDataFromStorage(body)
	assert.Equal(t, 1, nrTxs)
	assert.Nil(t, err)

	txHashToAsk := []byte("tx_hashnotinPool")
	miniBlock = &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHashToAsk}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	tc.SaveTxsToStorage(body)

	nrTxs, err = tc.RestoreBlockDataFromStorage(body)
	assert.Equal(t, 1, nrTxs)
	assert.NotNil(t, err)
}

func TestTransactionCoordinator_RemoveBlockDataFromPool(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool(txHash)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	argsTransactionCoordinator.Accounts = initAccountsMock()
	argsTransactionCoordinator.MiniBlockPool = dataPool.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock())
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	err = tc.RemoveBlockDataFromPool(nil)
	assert.Nil(t, err)

	body := &block.Body{}
	miniBlock := &block.MiniBlock{SenderShardID: 1, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHash}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	tc.RequestBlockTransactions(body)
	err = tc.RemoveBlockDataFromPool(body)
	assert.Nil(t, err)
}

func TestTransactionCoordinator_ProcessBlockTransactionProcessTxError(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool(txHash)

	accounts := initAccountsMock()
	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataPool,
		createMockPubkeyConverter(),
		accounts,
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, process.ErrHigherNonceInTransaction
			},
		},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&testscommon.GasHandlerStub{
			ComputeGasProvidedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasProvidedCalled: func() uint64 {
				return 0
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)
	container, _ := preFactory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	argsTransactionCoordinator.Accounts = initAccountsMock()
	argsTransactionCoordinator.MiniBlockPool = dataPool.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = container
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() time.Duration {
		return time.Second
	}
	err = tc.ProcessBlockTransaction(&block.Header{}, &block.Body{}, haveTime)
	assert.Nil(t, err)

	body := &block.Body{}
	miniBlock := &block.MiniBlock{SenderShardID: 1, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHash}}
	miniBlockHash1, _ := core.CalculateHash(tc.marshalizer, tc.hasher, miniBlock)
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	tc.RequestBlockTransactions(body)
	err = tc.ProcessBlockTransaction(&block.Header{MiniBlockHeaders: []block.MiniBlockHeader{{Hash: miniBlockHash1, TxCount: 1}}}, body, haveTime)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)

	noTime := func() time.Duration {
		return 0
	}
	err = tc.ProcessBlockTransaction(&block.Header{MiniBlockHeaders: []block.MiniBlockHeader{{Hash: miniBlockHash1, TxCount: 1}}}, body, noTime)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)

	txHashToAsk := []byte("tx_hashnotinPool")
	miniBlock = &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHashToAsk}}
	miniBlockHash2, _ := core.CalculateHash(tc.marshalizer, tc.hasher, miniBlock)
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)
	err = tc.ProcessBlockTransaction(&block.Header{MiniBlockHeaders: []block.MiniBlockHeader{{Hash: miniBlockHash1, TxCount: 1}, {Hash: miniBlockHash2, TxCount: 1}}}, body, haveTime)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestTransactionCoordinator_ProcessBlockTransaction(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool(txHash)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	argsTransactionCoordinator.Accounts = initAccountsMock()
	argsTransactionCoordinator.MiniBlockPool = dataPool.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock())
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() time.Duration {
		return time.Second
	}
	err = tc.ProcessBlockTransaction(&block.Header{}, &block.Body{}, haveTime)
	assert.Nil(t, err)

	body := &block.Body{}
	miniBlock := &block.MiniBlock{SenderShardID: 1, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHash}}
	miniBlockHash1, _ := core.CalculateHash(tc.marshalizer, tc.hasher, miniBlock)
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	tc.RequestBlockTransactions(body)
	err = tc.ProcessBlockTransaction(&block.Header{MiniBlockHeaders: []block.MiniBlockHeader{{Hash: miniBlockHash1, TxCount: 1}}}, body, haveTime)
	assert.Nil(t, err)

	noTime := func() time.Duration {
		return -1
	}
	err = tc.ProcessBlockTransaction(&block.Header{MiniBlockHeaders: []block.MiniBlockHeader{{Hash: miniBlockHash1, TxCount: 1}}}, body, noTime)
	assert.Equal(t, process.ErrTimeIsOut, err)

	txHashToAsk := []byte("tx_hashnotinPool")
	miniBlock = &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHashToAsk}}
	miniBlockHash2, _ := core.CalculateHash(tc.marshalizer, tc.hasher, miniBlock)
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)
	err = tc.ProcessBlockTransaction(&block.Header{MiniBlockHeaders: []block.MiniBlockHeader{{Hash: miniBlockHash1, TxCount: 1}, {Hash: miniBlockHash2, TxCount: 1}}}, body, haveTime)
	assert.Equal(t, process.ErrMissingTransaction, err)
}

func TestTransactionCoordinator_RequestMiniblocks(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool(txHash)
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	nrCalled := 0
	mutex := sync.Mutex{}

	requestHandler := &testscommon.RequestHandlerStub{
		RequestMiniBlockHandlerCalled: func(destShardID uint32, miniblockHash []byte) {
			mutex.Lock()
			nrCalled++
			mutex.Unlock()
		},
	}

	accounts := initAccountsMock()
	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		dataPool,
		createMockPubkeyConverter(),
		accounts,
		requestHandler,
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)
	container, _ := preFactory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = shardCoordinator
	argsTransactionCoordinator.Accounts = accounts
	argsTransactionCoordinator.MiniBlockPool = dataPool.MiniBlocks()
	argsTransactionCoordinator.RequestHandler = requestHandler
	argsTransactionCoordinator.PreProcessors = container
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tc.RequestMiniBlocks(nil)
	time.Sleep(time.Second)
	mutex.Lock()
	assert.Equal(t, 0, nrCalled)
	mutex.Unlock()

	header := createTestMetablock()
	tc.RequestMiniBlocks(header)

	crossMbs := header.GetMiniBlockHeadersWithDst(shardCoordinator.SelfId())
	time.Sleep(time.Second)
	mutex.Lock()
	assert.Equal(t, len(crossMbs), nrCalled)
	mutex.Unlock()
}

func TestShardProcessor_ProcessMiniBlockCompleteWithOkTxsShouldExecuteThemAndNotRevertAccntState(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := dataRetrieverMock.NewPoolsHolderMock()

	// we will have a miniblock that will have 3 tx hashes
	// all txs will be in datapool and none of them will return err when processed
	// so, tx processor will return nil on processing tx

	txHash1 := []byte("tx hash 1")
	txHash2 := []byte("tx hash 2")
	txHash3 := []byte("tx hash 3")

	senderShardId := uint32(0)
	receiverShardId := uint32(1)

	miniBlock := block.MiniBlock{
		SenderShardID:   senderShardId,
		ReceiverShardID: receiverShardId,
		TxHashes:        [][]byte{txHash1, txHash2, txHash3},
	}

	tx1Nonce := uint64(45)
	tx2Nonce := uint64(46)
	tx3Nonce := uint64(47)

	// put the existing tx inside datapool
	cacheId := process.ShardCacherIdentifier(senderShardId, receiverShardId)
	dataPool.Transactions().AddData(txHash1, &transaction.Transaction{
		Nonce: tx1Nonce,
		Data:  txHash1,
	}, 0, cacheId)
	dataPool.Transactions().AddData(txHash2, &transaction.Transaction{
		Nonce: tx2Nonce,
		Data:  txHash2,
	}, 0, cacheId)
	dataPool.Transactions().AddData(txHash3, &transaction.Transaction{
		Nonce: tx3Nonce,
		Data:  txHash3,
	}, 0, cacheId)

	tx1ExecutionResult := uint64(0)
	tx2ExecutionResult := uint64(0)
	tx3ExecutionResult := uint64(0)

	accounts := &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}

	totalGasProvided := uint64(0)

	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		marshalizer,
		hasher,
		dataPool,
		createMockPubkeyConverter(),
		accounts,
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				// execution, in this context, means moving the tx nonce to itx corresponding execution result variable
				if bytes.Equal(transaction.Data, txHash1) {
					tx1ExecutionResult = transaction.Nonce
				}
				if bytes.Equal(transaction.Data, txHash2) {
					tx2ExecutionResult = transaction.Nonce
				}
				if bytes.Equal(transaction.Data, txHash3) {
					tx3ExecutionResult = transaction.Nonce
				}

				return 0, nil
			},
		},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&testscommon.GasHandlerStub{
			SetGasProvidedCalled: func(gasProvided uint64, hash []byte) {
				totalGasProvided += gasProvided
			},
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasProvidedCalled: func() uint64 {
				return 0
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			TotalGasRefundedCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)
	container, _ := preFactory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	argsTransactionCoordinator.Accounts = accounts
	argsTransactionCoordinator.MiniBlockPool = dataPool.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = container
	argsTransactionCoordinator.GasHandler = &testscommon.GasHandlerStub{
		TotalGasProvidedCalled: func() uint64 {
			return 0
		},
	}
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	haveAdditionalTime := func() bool {
		return false
	}
	preproc := tc.getPreProcessor(block.TxBlock)
	processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
		IndexOfLastTxProcessed: -1,
		FullyProcessed:         false,
	}
	err = tc.processCompleteMiniBlock(preproc, &miniBlock, []byte("hash"), haveTime, haveAdditionalTime, false, processedMbInfo)

	assert.Nil(t, err)
	assert.Equal(t, tx1Nonce, tx1ExecutionResult)
	assert.Equal(t, tx2Nonce, tx2ExecutionResult)
	assert.Equal(t, tx3Nonce, tx3ExecutionResult)
}

func TestShardProcessor_ProcessMiniBlockCompleteWithErrorWhileProcessShouldCallRevertAccntState(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := dataRetrieverMock.NewPoolsHolderMock()

	// we will have a miniblock that will have 3 tx hashes
	// all txs will be in datapool and none of them will return err when processed
	// so, tx processor will return nil on processing tx

	txHash1 := []byte("tx hash 1")
	txHash2 := []byte("tx hash 2 - this will cause the tx processor to err")
	txHash3 := []byte("tx hash 3")

	senderShardId := uint32(0)
	receiverShardId := uint32(1)

	miniBlock := block.MiniBlock{
		SenderShardID:   senderShardId,
		ReceiverShardID: receiverShardId,
		TxHashes:        [][]byte{txHash1, txHash2, txHash3},
	}

	tx1Nonce := uint64(45)
	tx2Nonce := uint64(46)
	tx3Nonce := uint64(47)

	// put the existing tx inside datapool
	cacheId := process.ShardCacherIdentifier(senderShardId, receiverShardId)
	dataPool.Transactions().AddData(txHash1, &transaction.Transaction{
		Nonce: tx1Nonce,
		Data:  txHash1,
	}, 0, cacheId)
	dataPool.Transactions().AddData(txHash2, &transaction.Transaction{
		Nonce: tx2Nonce,
		Data:  txHash2,
	}, 0, cacheId)
	dataPool.Transactions().AddData(txHash3, &transaction.Transaction{
		Nonce: tx3Nonce,
		Data:  txHash3,
	}, 0, cacheId)

	currentJournalLen := 445
	revertAccntStateCalled := false

	accounts := &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			if snapshot == currentJournalLen {
				revertAccntStateCalled = true
			}

			return nil
		},
		JournalLenCalled: func() int {
			return currentJournalLen
		},
	}

	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		marshalizer,
		hasher,
		dataPool,
		createMockPubkeyConverter(),
		accounts,
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				if bytes.Equal(transaction.Data, txHash2) {
					return 0, process.ErrHigherNonceInTransaction
				}
				return 0, nil
			},
		},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&testscommon.GasHandlerStub{
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasProvidedCalled: func() uint64 {
				return 0
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			TotalGasRefundedCalled: func() uint64 {
				return 0
			},
			SetGasProvidedCalled: func(gasProvided uint64, hash []byte) {},
			RemoveGasRefundedCalled: func(hashes [][]byte) {
			},
			RemoveGasProvidedCalled: func(hashes [][]byte) {
			},
		},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
	)
	container, _ := preFactory.Create()

	totalGasProvided := uint64(0)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	argsTransactionCoordinator.Accounts = accounts
	argsTransactionCoordinator.MiniBlockPool = dataPool.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = container
	argsTransactionCoordinator.GasHandler = &testscommon.GasHandlerStub{
		TotalGasProvidedCalled: func() uint64 {
			return totalGasProvided
		},
		SetGasProvidedCalled: func(gasProvided uint64, hash []byte) {
			totalGasProvided = gasProvided
		},
	}
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	haveAdditionalTime := func() bool {
		return false
	}
	preproc := tc.getPreProcessor(block.TxBlock)
	processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
		IndexOfLastTxProcessed: -1,
		FullyProcessed:         false,
	}
	err = tc.processCompleteMiniBlock(preproc, &miniBlock, []byte("hash"), haveTime, haveAdditionalTime, false, processedMbInfo)

	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
	assert.True(t, revertAccntStateCalled)
}

func TestTransactionCoordinator_VerifyCreatedBlockTransactionsNilOrMiss(t *testing.T) {
	t.Parallel()

	tdp := initDataPool(txHash)
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	argsFactory := shard.ArgsNewIntermediateProcessorsContainerFactory{
		ShardCoordinator: shardCoordinator,
		Marshalizer:      &mock.MarshalizerMock{},
		Hasher:           &hashingMocks.HasherMock{},
		PubkeyConverter:  createMockPubkeyConverter(),
		Store:            &storageStubs.ChainStorerStub{},
		PoolsHolder:      tdp,
		EconomicsFee:     &economicsmocks.EconomicsHandlerStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.KeepExecOrderOnCreatedSCRsFlag
			},
		},
	}
	preFactory, _ := shard.NewIntermediateProcessorsContainerFactory(argsFactory)
	container, _ := preFactory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = shardCoordinator
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.InterProcessors = container
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	err = tc.VerifyCreatedBlockTransactions(&block.Header{ReceiptsHash: []byte("receipt")}, &block.Body{})
	assert.Equal(t, process.ErrReceiptsHashMissmatch, err)

	body := &block.Body{MiniBlocks: []*block.MiniBlock{{Type: block.TxBlock}}}
	err = tc.VerifyCreatedBlockTransactions(&block.Header{ReceiptsHash: []byte("receipt")}, body)
	assert.Equal(t, process.ErrReceiptsHashMissmatch, err)

	body = &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				Type:            block.SmartContractResultBlock,
				ReceiverShardID: shardCoordinator.SelfId(),
				SenderShardID:   shardCoordinator.SelfId() + 1},
		},
	}
	err = tc.VerifyCreatedBlockTransactions(&block.Header{ReceiptsHash: []byte("receipt")}, body)
	assert.Equal(t, process.ErrReceiptsHashMissmatch, err)

	body = &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				Type:            block.SmartContractResultBlock,
				ReceiverShardID: shardCoordinator.SelfId() + 1,
			},
		},
	}
	err = tc.VerifyCreatedBlockTransactions(&block.Header{ReceiptsHash: []byte("receipt")}, body)
	assert.Equal(t, process.ErrNilMiniBlocks, err)
}

func TestTransactionCoordinator_VerifyCreatedBlockTransactionsOk(t *testing.T) {
	t.Parallel()

	tdp := initDataPool(txHash)
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	argsFactory := shard.ArgsNewIntermediateProcessorsContainerFactory{
		ShardCoordinator: shardCoordinator,
		Marshalizer:      &mock.MarshalizerMock{},
		Hasher:           &hashingMocks.HasherMock{},
		PubkeyConverter:  createMockPubkeyConverter(),
		Store:            &storageStubs.ChainStorerStub{},
		PoolsHolder:      tdp,
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
				return MaxGasLimitPerBlock
			},
		},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.KeepExecOrderOnCreatedSCRsFlag
			},
		},
	}
	interFactory, _ := shard.NewIntermediateProcessorsContainerFactory(argsFactory)
	container, _ := interFactory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = shardCoordinator
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.InterProcessors = container
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	sndAddr := []byte("0")
	rcvAddr := []byte("1")
	scr := &smartContractResult.SmartContractResult{Nonce: 10, SndAddr: sndAddr, RcvAddr: rcvAddr, PrevTxHash: []byte("txHash"), Value: big.NewInt(0)}
	scrHash, _ := core.CalculateHash(&mock.MarshalizerMock{}, &hashingMocks.HasherMock{}, scr)

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, sndAddr) {
			return shardCoordinator.SelfId()
		}
		if bytes.Equal(address, rcvAddr) {
			return shardCoordinator.SelfId() + 1
		}
		return shardCoordinator.SelfId() + 2
	}

	tdp.UnsignedTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &testscommon.ShardedDataStub{
			RegisterOnAddedCalled: func(i func(key []byte, value interface{})) {},
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return &testscommon.CacherStub{
					PeekCalled: func(key []byte) (value interface{}, ok bool) {
						if reflect.DeepEqual(key, scrHash) {
							return scr, true
						}
						return nil, false
					},
					KeysCalled: func() [][]byte {
						return [][]byte{[]byte("key1"), []byte("key2")}
					},
					LenCalled: func() int {
						return 0
					},
				}
			},
			RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {},
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				if reflect.DeepEqual(key, scrHash) {
					return scr, true
				}
				return nil, false
			},
			AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheId string) {
			},
		}
	}

	interProc, _ := container.Get(block.SmartContractResultBlock)
	tx, _ := tdp.UnsignedTransactions().SearchFirstData(scrHash)
	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, tx.(data.TransactionHandler))
	err = interProc.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	body := &block.Body{MiniBlocks: []*block.MiniBlock{{Type: block.SmartContractResultBlock, ReceiverShardID: shardCoordinator.SelfId() + 1, TxHashes: [][]byte{scrHash}}}}
	err = tc.VerifyCreatedBlockTransactions(&block.Header{}, body)
	assert.Equal(t, process.ErrReceiptsHashMissmatch, err)
}

func TestTransactionCoordinator_SaveTxsToStorageCallsSaveIntermediate(t *testing.T) {
	t.Parallel()

	tdp := initDataPool(txHash)
	intermediateTxWereSaved := false
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	argsTransactionCoordinator.Accounts = initAccountsMock()
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	argsTransactionCoordinator.InterProcessors = &mock.InterimProcessorContainerMock{
		KeysCalled: func() []block.Type {
			return []block.Type{block.SmartContractResultBlock}
		},
		GetCalled: func(key block.Type) (handler process.IntermediateTransactionHandler, e error) {
			if key == block.SmartContractResultBlock {
				return &mock.IntermediateTransactionHandlerMock{
					SaveCurrentIntermediateTxToStorageCalled: func() {
						intermediateTxWereSaved = true
					},
				}, nil
			}
			return nil, errors.New("invalid handler type")
		},
	}
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	body := &block.Body{}
	miniBlock := &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHash}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	tc.RequestBlockTransactions(body)

	tc.SaveTxsToStorage(body)
	assert.True(t, intermediateTxWereSaved)
}

func TestTransactionCoordinator_PreprocessorsHasToBeOrderedRewardsAreLast(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool(txHash)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	argsTransactionCoordinator.Accounts = initAccountsMock()
	argsTransactionCoordinator.MiniBlockPool = dataPool.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock())
	argsTransactionCoordinator.InterProcessors = createInterimProcessorContainer()
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	preProcLen := len(tc.keysTxPreProcs)
	lastKey := tc.keysTxPreProcs[preProcLen-1]

	assert.Equal(t, block.RewardsBlock, lastKey)
}

func TestTransactionCoordinator_GetCreatedInShardMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	tc, _ := NewTransactionCoordinator(argsTransactionCoordinator)

	mb1 := &block.MiniBlock{
		Type: block.SmartContractResultBlock,
	}
	mb2 := &block.MiniBlock{
		Type: block.ReceiptBlock,
	}

	tc.keysInterimProcs = append(tc.keysInterimProcs, block.SmartContractResultBlock)
	tc.keysInterimProcs = append(tc.keysInterimProcs, block.ReceiptBlock)

	tc.interimProcessors[block.SmartContractResultBlock] = &mock.IntermediateTransactionHandlerMock{
		GetCreatedInShardMiniBlockCalled: func() *block.MiniBlock {
			return mb1
		},
	}
	tc.interimProcessors[block.ReceiptBlock] = &mock.IntermediateTransactionHandlerMock{
		GetCreatedInShardMiniBlockCalled: func() *block.MiniBlock {
			return mb2
		},
	}

	miniblocks := tc.GetCreatedInShardMiniBlocks()
	assert.Equal(t, []*block.MiniBlock{mb1, mb2}, miniblocks)
}

func TestTransactionCoordinator_GetNumOfCrossInterMbsAndTxsShouldWork(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	tc, _ := NewTransactionCoordinator(argsTransactionCoordinator)

	tc.keysInterimProcs = append(tc.keysInterimProcs, block.SmartContractResultBlock)
	tc.keysInterimProcs = append(tc.keysInterimProcs, block.ReceiptBlock)

	tc.interimProcessors[block.SmartContractResultBlock] = &mock.IntermediateTransactionHandlerMock{
		GetNumOfCrossInterMbsAndTxsCalled: func() (int, int) {
			return 2, 2
		},
	}
	tc.interimProcessors[block.ReceiptBlock] = &mock.IntermediateTransactionHandlerMock{
		GetNumOfCrossInterMbsAndTxsCalled: func() (int, int) {
			return 3, 8
		},
	}

	numMbs, numTxs := tc.GetNumOfCrossInterMbsAndTxs()

	assert.Equal(t, 5, numMbs)
	assert.Equal(t, 10, numTxs)
}

func TestTransactionCoordinator_IsMaxBlockSizeReachedShouldWork(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.BlockSizeComputation = &testscommon.BlockSizeComputationStub{
		IsMaxBlockSizeWithoutThrottleReachedCalled: func(i int, i2 int) bool {
			return i+i2 > 4
		},
	}
	tc, _ := NewTransactionCoordinator(argsTransactionCoordinator)

	tc.keysTxPreProcs = append(tc.keysTxPreProcs, block.TxBlock)

	body := &block.Body{
		MiniBlocks: make([]*block.MiniBlock, 0),
	}

	mb1 := &block.MiniBlock{
		Type:            block.TxBlock,
		ReceiverShardID: 0,
		TxHashes:        [][]byte{[]byte("txHash1")},
	}
	mb2 := &block.MiniBlock{
		Type:            block.TxBlock,
		ReceiverShardID: 1,
		TxHashes:        [][]byte{[]byte("txHash2")},
	}
	body.MiniBlocks = append(body.MiniBlocks, mb1)
	body.MiniBlocks = append(body.MiniBlocks, mb2)

	tc.txPreProcessors[block.TxBlock] = &mock.PreProcessorMock{
		GetAllCurrentUsedTxsCalled: func() map[string]data.TransactionHandler {
			allTxs := make(map[string]data.TransactionHandler)
			allTxs["txHash2"] = &transaction.Transaction{
				RcvAddr: make([]byte, 0),
			}
			return allTxs
		},
	}
	assert.False(t, tc.isMaxBlockSizeReached(body))

	tc.txPreProcessors[block.TxBlock] = &mock.PreProcessorMock{
		GetAllCurrentUsedTxsCalled: func() map[string]data.TransactionHandler {
			allTxs := make(map[string]data.TransactionHandler)
			allTxs["txHash2"] = &transaction.Transaction{
				RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1),
			}
			return allTxs
		},
	}
	assert.True(t, tc.isMaxBlockSizeReached(body))
}

func TestTransactionCoordinator_GetNumOfCrossShardScCallsShouldWork(t *testing.T) {
	t.Parallel()

	mb := &block.MiniBlock{
		Type:     block.TxBlock,
		TxHashes: [][]byte{[]byte("txHash1")},
	}

	allTxs := make(map[string]data.TransactionHandler)

	mb.ReceiverShardID = 0
	assert.Equal(t, 0, getNumOfCrossShardScCallsOrSpecialTxs(mb, allTxs, 0))

	mb.ReceiverShardID = 1
	assert.Equal(t, 1, getNumOfCrossShardScCallsOrSpecialTxs(mb, allTxs, 0))

	allTxs["txHash1"] = &transaction.Transaction{
		RcvAddr:     make([]byte, 0),
		RcvUserName: make([]byte, 0),
	}
	assert.Equal(t, 0, getNumOfCrossShardScCallsOrSpecialTxs(mb, allTxs, 0))

	allTxs["txHash1"] = &transaction.Transaction{
		RcvAddr:     make([]byte, core.NumInitCharactersForScAddress+1),
		RcvUserName: make([]byte, 0),
	}
	assert.Equal(t, 1, getNumOfCrossShardScCallsOrSpecialTxs(mb, allTxs, 0))
}

func TestTransactionCoordinator_GetNumOfCrossShardSpecialTxsShouldWork(t *testing.T) {
	t.Parallel()

	mb := &block.MiniBlock{
		Type:     block.TxBlock,
		TxHashes: [][]byte{[]byte("txHash1")},
	}

	allTxs := make(map[string]data.TransactionHandler)

	mb.ReceiverShardID = 0
	assert.Equal(t, 0, getNumOfCrossShardScCallsOrSpecialTxs(mb, allTxs, 0))

	mb.ReceiverShardID = 1
	assert.Equal(t, 1, getNumOfCrossShardScCallsOrSpecialTxs(mb, allTxs, 0))

	allTxs["txHash1"] = &transaction.Transaction{
		RcvAddr:     make([]byte, 0),
		RcvUserName: make([]byte, 0),
	}
	assert.Equal(t, 0, getNumOfCrossShardScCallsOrSpecialTxs(mb, allTxs, 0))

	allTxs["txHash1"] = &transaction.Transaction{
		RcvAddr:     make([]byte, 0),
		RcvUserName: []byte("username"),
	}
	assert.Equal(t, 1, getNumOfCrossShardScCallsOrSpecialTxs(mb, allTxs, 0))
}

func TestTransactionCoordinator_VerifyCreatedMiniBlocksShouldReturnWhenEpochIsNotEnabled(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:                   &hashingMocks.HasherMock{},
		Marshalizer:              &mock.MarshalizerMock{},
		ShardCoordinator:         mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                 initAccountsMock(),
		MiniBlockPool:            dataPool.MiniBlocks(),
		RequestHandler:           &testscommon.RequestHandlerStub{},
		PreProcessors:            createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:          createInterimProcessorContainer(),
		GasHandler:               &testscommon.GasHandlerStub{},
		FeeHandler:               &mock.FeeAccumulatorStub{},
		BlockSizeComputation:     &testscommon.BlockSizeComputationStub{},
		BalanceComputation:       &testscommon.BalanceComputationStub{},
		EconomicsFee:             &economicsmocks.EconomicsHandlerStub{},
		TxTypeHandler:            &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor: &mock.TxLogsProcessorStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
				if flag == common.BlockGasAndFeesReCheckFlag {
					return 1
				}
				return 0
			},
		},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	header := &block.Header{}
	body := &block.Body{}

	err = tc.VerifyCreatedMiniBlocks(header, body)
	assert.Nil(t, err)
}

func TestTransactionCoordinator_VerifyCreatedMiniBlocksShouldErrMaxGasLimitPerMiniBlockInReceiverShardIsReached(t *testing.T) {
	t.Parallel()

	maxGasLimitPerBlock := uint64(1500000000)
	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return maxGasLimitPerBlock + 1
			},
			MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerMiniBlockCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
		},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}

	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tc.txPreProcessors[block.TxBlock] = &mock.PreProcessorMock{
		GetAllCurrentUsedTxsCalled: func() map[string]data.TransactionHandler {
			allTxs := make(map[string]data.TransactionHandler)
			allTxs[string(txHash)] = &transaction.Transaction{}
			return allTxs
		},
	}

	header := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{{}},
	}
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{txHash},
				ReceiverShardID: 1,
			},
		},
	}

	err = tc.VerifyCreatedMiniBlocks(header, body)
	assert.Equal(t, process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached, err)
}

func TestTransactionCoordinator_VerifyCreatedMiniBlocksShouldErrMaxAccumulatedFeesExceeded(t *testing.T) {
	t.Parallel()

	maxGasLimitPerBlock := uint64(1500000000)
	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerTxCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}

	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tc.txPreProcessors[block.TxBlock] = &mock.PreProcessorMock{
		GetAllCurrentUsedTxsCalled: func() map[string]data.TransactionHandler {
			allTxs := make(map[string]data.TransactionHandler)
			allTxs[string(txHash)] = &transaction.Transaction{
				GasLimit: 100,
				GasPrice: 1,
			}
			return allTxs
		},
	}

	header := &block.Header{
		AccumulatedFees:  big.NewInt(101),
		DeveloperFees:    big.NewInt(10),
		MiniBlockHeaders: []block.MiniBlockHeader{{TxCount: 1}},
	}
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{txHash},
				ReceiverShardID: 1,
			},
		},
	}

	err = tc.VerifyCreatedMiniBlocks(header, body)
	assert.Equal(t, process.ErrMaxAccumulatedFeesExceeded, err)
}

func TestTransactionCoordinator_VerifyCreatedMiniBlocksShouldErrMaxDeveloperFeesExceeded(t *testing.T) {
	t.Parallel()

	maxGasLimitPerBlock := uint64(1500000000)
	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerTxCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}

	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tc.txPreProcessors[block.TxBlock] = &mock.PreProcessorMock{
		GetAllCurrentUsedTxsCalled: func() map[string]data.TransactionHandler {
			allTxs := make(map[string]data.TransactionHandler)
			allTxs[string(txHash)] = &transaction.Transaction{
				GasLimit: 100,
				GasPrice: 1,
			}
			return allTxs
		},
	}

	header := &block.Header{
		AccumulatedFees:  big.NewInt(100),
		DeveloperFees:    big.NewInt(11),
		MiniBlockHeaders: []block.MiniBlockHeader{{TxCount: 1}},
	}
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{txHash},
				ReceiverShardID: 1,
			},
		},
	}

	err = tc.VerifyCreatedMiniBlocks(header, body)
	assert.Equal(t, process.ErrMaxDeveloperFeesExceeded, err)
}

func TestTransactionCoordinator_VerifyCreatedMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	maxGasLimitPerBlock := uint64(1500000000)
	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerTxCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}

	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tc.txPreProcessors[block.TxBlock] = &mock.PreProcessorMock{
		GetAllCurrentUsedTxsCalled: func() map[string]data.TransactionHandler {
			allTxs := make(map[string]data.TransactionHandler)
			allTxs[string(txHash)] = &transaction.Transaction{
				GasLimit: 100,
				GasPrice: 1,
			}
			return allTxs
		},
	}

	header := &block.Header{
		AccumulatedFees:  big.NewInt(100),
		DeveloperFees:    big.NewInt(10),
		MiniBlockHeaders: []block.MiniBlockHeader{{TxCount: 1}},
	}
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{txHash},
				ReceiverShardID: 1,
			},
		},
	}

	err = tc.VerifyCreatedMiniBlocks(header, body)
	assert.Nil(t, err)
}

func TestTransactionCoordinator_GetAllTransactionsShouldWork(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:                       &hashingMocks.HasherMock{},
		Marshalizer:                  &mock.MarshalizerMock{},
		ShardCoordinator:             mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                     initAccountsMock(),
		MiniBlockPool:                dataPool.MiniBlocks(),
		RequestHandler:               &testscommon.RequestHandlerStub{},
		PreProcessors:                createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:              createInterimProcessorContainer(),
		GasHandler:                   &testscommon.GasHandlerStub{},
		FeeHandler:                   &mock.FeeAccumulatorStub{},
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1}
	tx2 := &transaction.Transaction{Nonce: 2}
	tx3 := &transaction.Transaction{Nonce: 3}

	txHash1 := "hash1"
	txHash2 := "hash2"
	txHash3 := "hash3"

	tc.txPreProcessors[block.TxBlock] = &mock.PreProcessorMock{
		GetAllCurrentUsedTxsCalled: func() map[string]data.TransactionHandler {
			allTxs := make(map[string]data.TransactionHandler)
			allTxs[txHash1] = tx1
			allTxs[txHash2] = tx2
			allTxs[txHash3] = tx3
			return allTxs
		},
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{},
		},
	}

	mapMiniBlockTypeAllTxs := tc.getAllTransactions(body)

	require.NotNil(t, mapMiniBlockTypeAllTxs)
	require.Equal(t, 1, len(mapMiniBlockTypeAllTxs))

	mapAllTxs := mapMiniBlockTypeAllTxs[block.TxBlock]

	require.NotNil(t, mapAllTxs)
	require.Equal(t, 3, len(mapAllTxs))

	assert.Equal(t, tx1, mapAllTxs[txHash1])
	assert.Equal(t, tx2, mapAllTxs[txHash2])
	assert.Equal(t, tx3, mapAllTxs[txHash3])
}

func TestTransactionCoordinator_VerifyGasLimitShouldErrMaxGasLimitPerMiniBlockInReceiverShardIsReached(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100000000)
	tx2GasLimit := uint64(200000000)
	tx3GasLimit := uint64(300000001)

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
				return tx1GasLimit + tx2GasLimit + tx3GasLimit - 1
			},
			MaxGasLimitPerMiniBlockCalled: func() uint64 {
				return tx1GasLimit + tx2GasLimit + tx3GasLimit - 1
			},
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
		},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: tx1GasLimit}
	tx2 := &transaction.Transaction{Nonce: 2, GasLimit: tx2GasLimit}
	tx3 := &transaction.Transaction{Nonce: 3, GasLimit: tx3GasLimit}

	txHash1 := "hash1"
	txHash2 := "hash2"
	txHash3 := "hash3"

	tc.txPreProcessors[block.TxBlock] = &mock.PreProcessorMock{
		GetAllCurrentUsedTxsCalled: func() map[string]data.TransactionHandler {
			allTxs := make(map[string]data.TransactionHandler)
			allTxs[txHash1] = tx1
			allTxs[txHash2] = tx2
			allTxs[txHash3] = tx3
			return allTxs
		},
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				Type:            block.TxBlock,
				ReceiverShardID: 0,
			},
			{
				Type:            block.SmartContractResultBlock,
				ReceiverShardID: 1,
			},
			{
				Type:            block.TxBlock,
				TxHashes:        [][]byte{[]byte(txHash1), []byte(txHash2), []byte(txHash3)},
				ReceiverShardID: 1,
			},
		},
	}

	mapMiniBlockTypeAllTxs := make(map[block.Type]map[string]data.TransactionHandler)

	mapAllTxs := make(map[string]data.TransactionHandler)
	mapAllTxs[txHash1] = tx1
	mapAllTxs[txHash2] = tx2
	mapAllTxs[txHash3] = tx3

	mapMiniBlockTypeAllTxs[block.TxBlock] = mapAllTxs

	err = tc.verifyGasLimit(&block.Header{MiniBlockHeaders: []block.MiniBlockHeader{{}, {}, {}}}, body, mapMiniBlockTypeAllTxs)
	assert.Equal(t, process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached, err)
}

func TestTransactionCoordinator_VerifyGasLimitShouldWork(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)
	tx2GasLimit := uint64(200)
	tx3GasLimit := uint64(300)

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
				return tx1GasLimit + tx2GasLimit + tx3GasLimit
			},
			MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
				return tx1GasLimit + tx2GasLimit + tx3GasLimit
			},
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
		},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: tx1GasLimit}
	tx2 := &transaction.Transaction{Nonce: 2, GasLimit: tx2GasLimit}
	tx3 := &transaction.Transaction{Nonce: 3, GasLimit: tx3GasLimit}

	txHash1 := "hash1"
	txHash2 := "hash2"
	txHash3 := "hash3"

	tc.txPreProcessors[block.TxBlock] = &mock.PreProcessorMock{
		GetAllCurrentUsedTxsCalled: func() map[string]data.TransactionHandler {
			allTxs := make(map[string]data.TransactionHandler)
			allTxs[txHash1] = tx1
			allTxs[txHash2] = tx2
			allTxs[txHash3] = tx3
			return allTxs
		},
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				Type:            block.TxBlock,
				ReceiverShardID: 0,
			},
			{
				Type:            block.SmartContractResultBlock,
				ReceiverShardID: 1,
			},
			{
				Type:            block.TxBlock,
				TxHashes:        [][]byte{[]byte(txHash1), []byte(txHash2), []byte(txHash3)},
				ReceiverShardID: 1,
			},
		},
	}

	mapMiniBlockTypeAllTxs := make(map[block.Type]map[string]data.TransactionHandler)

	mapAllTxs := make(map[string]data.TransactionHandler)
	mapAllTxs[txHash1] = tx1
	mapAllTxs[txHash2] = tx2
	mapAllTxs[txHash3] = tx3

	mapMiniBlockTypeAllTxs[block.TxBlock] = mapAllTxs

	err = tc.verifyGasLimit(&block.Header{MiniBlockHeaders: []block.MiniBlockHeader{{}, {}, {}}}, body, mapMiniBlockTypeAllTxs)
	assert.Nil(t, err)
}

func TestTransactionCoordinator_CheckGasProvidedByMiniBlockInReceiverShardShouldErrMissingTransaction(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:                       &hashingMocks.HasherMock{},
		Marshalizer:                  &mock.MarshalizerMock{},
		ShardCoordinator:             mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                     initAccountsMock(),
		MiniBlockPool:                dataPool.MiniBlocks(),
		RequestHandler:               &testscommon.RequestHandlerStub{},
		PreProcessors:                createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:              createInterimProcessorContainer(),
		GasHandler:                   &testscommon.GasHandlerStub{},
		FeeHandler:                   &mock.FeeAccumulatorStub{},
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	mb := &block.MiniBlock{
		Type:            block.TxBlock,
		TxHashes:        [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
		ReceiverShardID: 1,
	}

	err = tc.checkGasProvidedByMiniBlockInReceiverShard(mb, nil)
	assert.Equal(t, err, process.ErrMissingTransaction)
}

func TestTransactionCoordinator_CheckGasProvidedByMiniBlockInReceiverShardShouldErrSubtractionOverflow(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit() + 1
			},
		},
		TxTypeHandler: &testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.MoveBalance, process.SCInvoking
			},
		},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: tx1GasLimit}
	txHash1 := "hash1"

	mapAllTxs := make(map[string]data.TransactionHandler)
	mapAllTxs[txHash1] = tx1

	mb := &block.MiniBlock{
		Type:            block.TxBlock,
		TxHashes:        [][]byte{[]byte(txHash1)},
		ReceiverShardID: 1,
	}

	err = tc.checkGasProvidedByMiniBlockInReceiverShard(mb, mapAllTxs)
	assert.Equal(t, err, core.ErrSubtractionOverflow)
}

func TestTransactionCoordinator_CheckGasProvidedByMiniBlockInReceiverShardShouldErrAdditionOverflow(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(math.MaxUint64)
	tx2GasLimit := uint64(1)

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 0
			},
		},
		TxTypeHandler: &testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.MoveBalance, process.SCInvoking
			},
		},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: tx1GasLimit}
	tx2 := &transaction.Transaction{Nonce: 2, GasLimit: tx2GasLimit}

	txHash1 := "hash1"
	txHash2 := "hash2"

	mapAllTxs := make(map[string]data.TransactionHandler)
	mapAllTxs[txHash1] = tx1
	mapAllTxs[txHash2] = tx2

	mb := &block.MiniBlock{
		Type:            block.TxBlock,
		TxHashes:        [][]byte{[]byte(txHash1), []byte(txHash2)},
		ReceiverShardID: 1,
	}

	err = tc.checkGasProvidedByMiniBlockInReceiverShard(mb, mapAllTxs)
	assert.Equal(t, err, core.ErrAdditionOverflow)
}

func TestTransactionCoordinator_CheckGasProvidedByMiniBlockInReceiverShardShouldErrMaxGasLimitPerMiniBlockInReceiverShardIsReached(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100000000)
	tx2GasLimit := uint64(200000000)
	tx3GasLimit := uint64(300000001)

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
				return tx1GasLimit + tx2GasLimit + tx3GasLimit - 1
			},
			MaxGasLimitPerMiniBlockCalled: func() uint64 {
				return tx1GasLimit + tx2GasLimit + tx3GasLimit - 1
			},
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
		},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: tx1GasLimit}
	tx2 := &transaction.Transaction{Nonce: 2, GasLimit: tx2GasLimit}
	tx3 := &transaction.Transaction{Nonce: 3, GasLimit: tx3GasLimit}

	txHash1 := "hash1"
	txHash2 := "hash2"
	txHash3 := "hash3"

	mapAllTxs := make(map[string]data.TransactionHandler)
	mapAllTxs[txHash1] = tx1
	mapAllTxs[txHash2] = tx2
	mapAllTxs[txHash3] = tx3

	mb := &block.MiniBlock{
		Type:            block.TxBlock,
		TxHashes:        [][]byte{[]byte(txHash1), []byte(txHash2), []byte(txHash3)},
		ReceiverShardID: 1,
	}

	err = tc.checkGasProvidedByMiniBlockInReceiverShard(mb, mapAllTxs)
	assert.Equal(t, err, process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached)
}

func TestTransactionCoordinator_CheckGasProvidedByMiniBlockInReceiverShardShouldWork(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)
	tx2GasLimit := uint64(200)
	tx3GasLimit := uint64(300)

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
				return tx1GasLimit + tx2GasLimit + tx3GasLimit
			},
			MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
				return tx1GasLimit + tx2GasLimit + tx3GasLimit
			},
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
		},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}

	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: tx1GasLimit}
	tx2 := &transaction.Transaction{Nonce: 2, GasLimit: tx2GasLimit}
	tx3 := &transaction.Transaction{Nonce: 3, GasLimit: tx3GasLimit}

	txHash1 := "hash1"
	txHash2 := "hash2"
	txHash3 := "hash3"

	mapAllTxs := make(map[string]data.TransactionHandler)
	mapAllTxs[txHash1] = tx1
	mapAllTxs[txHash2] = tx2
	mapAllTxs[txHash3] = tx3

	mb := &block.MiniBlock{
		Type:            block.TxBlock,
		TxHashes:        [][]byte{[]byte(txHash1), []byte(txHash2), []byte(txHash3)},
		ReceiverShardID: 1,
	}

	err = tc.checkGasProvidedByMiniBlockInReceiverShard(mb, mapAllTxs)
	assert.Nil(t, err)
}

func TestTransactionCoordinator_VerifyFeesShouldErrMissingTransaction(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:                       &hashingMocks.HasherMock{},
		Marshalizer:                  &mock.MarshalizerMock{},
		ShardCoordinator:             mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                     initAccountsMock(),
		MiniBlockPool:                dataPool.MiniBlocks(),
		RequestHandler:               &testscommon.RequestHandlerStub{},
		PreProcessors:                createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:              createInterimProcessorContainer(),
		GasHandler:                   &testscommon.GasHandlerStub{},
		FeeHandler:                   &mock.FeeAccumulatorStub{},
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}

	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	txHash1 := "hash1"

	header := &block.Header{
		AccumulatedFees:  big.NewInt(100),
		DeveloperFees:    big.NewInt(10),
		MiniBlockHeaders: []block.MiniBlockHeader{{TxCount: 1}},
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				Type:            block.TxBlock,
				TxHashes:        [][]byte{[]byte(txHash1)},
				ReceiverShardID: 1,
			},
		},
	}

	err = tc.verifyFees(header, body, nil)
	assert.Equal(t, process.ErrMissingTransaction, err)
}

func TestTransactionCoordinator_VerifyFeesShouldErrMaxAccumulatedFeesExceeded(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}

	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: tx1GasLimit, GasPrice: 1}
	txHash1 := "hash1"

	mapMiniBlockTypeAllTxs := make(map[block.Type]map[string]data.TransactionHandler)

	mapAllTxs := make(map[string]data.TransactionHandler)
	mapAllTxs[txHash1] = tx1
	mapMiniBlockTypeAllTxs[block.TxBlock] = mapAllTxs

	header := &block.Header{
		AccumulatedFees:  big.NewInt(101),
		DeveloperFees:    big.NewInt(10),
		MiniBlockHeaders: []block.MiniBlockHeader{{TxCount: 1}, {TxCount: 1}},
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				Type: block.PeerBlock,
			},
			{
				Type:            block.TxBlock,
				TxHashes:        [][]byte{[]byte(txHash1)},
				ReceiverShardID: 1,
			},
		},
	}

	err = tc.verifyFees(header, body, mapMiniBlockTypeAllTxs)
	assert.Equal(t, process.ErrMaxAccumulatedFeesExceeded, err)
}

func TestTransactionCoordinator_VerifyFeesShouldErrMaxDeveloperFeesExceeded(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: tx1GasLimit, GasPrice: 1}
	txHash1 := "hash1"

	mapMiniBlockTypeAllTxs := make(map[block.Type]map[string]data.TransactionHandler)

	mapAllTxs := make(map[string]data.TransactionHandler)
	mapAllTxs[txHash1] = tx1
	mapMiniBlockTypeAllTxs[block.TxBlock] = mapAllTxs

	header := &block.Header{
		AccumulatedFees:  big.NewInt(100),
		DeveloperFees:    big.NewInt(11),
		MiniBlockHeaders: []block.MiniBlockHeader{{}, {TxCount: 1}},
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				Type: block.PeerBlock,
			},
			{
				Type:            block.TxBlock,
				TxHashes:        [][]byte{[]byte(txHash1)},
				ReceiverShardID: 1,
			},
		},
	}

	err = tc.verifyFees(header, body, mapMiniBlockTypeAllTxs)
	assert.Equal(t, process.ErrMaxDeveloperFeesExceeded, err)
}

func TestTransactionCoordinator_VerifyFeesShouldErrMaxAccumulatedFeesExceededWhenScheduledFeesAreNotAdded(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)

	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:            &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor: &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:      enableEpochsHandlerStub,
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{
			GetScheduledGasAndFeesCalled: func() scheduled.GasAndFees {
				return scheduled.GasAndFees{
					AccumulatedFees: big.NewInt(1),
					DeveloperFees:   big.NewInt(0),
				}
			},
		},
		DoubleTransactionsDetector: &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker: &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: tx1GasLimit, GasPrice: 1}
	txHash1 := "hash1"

	mapMiniBlockTypeAllTxs := make(map[block.Type]map[string]data.TransactionHandler)

	mapAllTxs := make(map[string]data.TransactionHandler)
	mapAllTxs[txHash1] = tx1
	mapMiniBlockTypeAllTxs[block.TxBlock] = mapAllTxs

	header := &block.Header{
		AccumulatedFees:  big.NewInt(101),
		DeveloperFees:    big.NewInt(10),
		MiniBlockHeaders: []block.MiniBlockHeader{{}, {TxCount: 1}},
	}
	for index := range header.MiniBlockHeaders {
		_ = header.MiniBlockHeaders[index].SetProcessingType(int32(block.Normal))
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				Type: block.PeerBlock,
			},
			{
				Type:            block.TxBlock,
				TxHashes:        [][]byte{[]byte(txHash1)},
				ReceiverShardID: 1,
			},
		},
	}

	err = tc.verifyFees(header, body, mapMiniBlockTypeAllTxs)
	assert.Equal(t, process.ErrMaxAccumulatedFeesExceeded, err)

	enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return flag == common.ScheduledMiniBlocksFlag || flag == common.MiniBlockPartialExecutionFlag
	}

	err = tc.verifyFees(header, body, mapMiniBlockTypeAllTxs)
	assert.Nil(t, err)
}

func TestTransactionCoordinator_VerifyFeesShouldErrMaxDeveloperFeesExceededWhenScheduledFeesAreNotAdded(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)

	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:            &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor: &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:      enableEpochsHandlerStub,
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{
			GetScheduledGasAndFeesCalled: func() scheduled.GasAndFees {
				return scheduled.GasAndFees{
					AccumulatedFees: big.NewInt(0),
					DeveloperFees:   big.NewInt(1),
				}
			},
		},
		DoubleTransactionsDetector: &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker: &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: tx1GasLimit, GasPrice: 1}
	txHash1 := "hash1"

	mapMiniBlockTypeAllTxs := make(map[block.Type]map[string]data.TransactionHandler)

	mapAllTxs := make(map[string]data.TransactionHandler)
	mapAllTxs[txHash1] = tx1
	mapMiniBlockTypeAllTxs[block.TxBlock] = mapAllTxs

	header := &block.Header{
		AccumulatedFees:  big.NewInt(100),
		DeveloperFees:    big.NewInt(11),
		MiniBlockHeaders: []block.MiniBlockHeader{{}, {TxCount: 1}},
	}
	for index := range header.MiniBlockHeaders {
		_ = header.MiniBlockHeaders[index].SetProcessingType(int32(block.Normal))
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				Type: block.PeerBlock,
			},
			{
				Type:            block.TxBlock,
				TxHashes:        [][]byte{[]byte(txHash1)},
				ReceiverShardID: 1,
			},
		},
	}

	err = tc.verifyFees(header, body, mapMiniBlockTypeAllTxs)
	assert.Equal(t, process.ErrMaxDeveloperFeesExceeded, err)

	enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return flag == common.ScheduledMiniBlocksFlag || flag == common.MiniBlockPartialExecutionFlag
	}

	err = tc.verifyFees(header, body, mapMiniBlockTypeAllTxs)
	assert.Nil(t, err)
}

func TestTransactionCoordinator_VerifyFeesShouldWork(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)

	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:            &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor: &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:      enableEpochsHandlerStub,
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{
			GetScheduledGasAndFeesCalled: func() scheduled.GasAndFees {
				return scheduled.GasAndFees{
					AccumulatedFees: big.NewInt(1),
					DeveloperFees:   big.NewInt(1),
				}
			},
		},
		DoubleTransactionsDetector: &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker: &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: tx1GasLimit, GasPrice: 1}
	txHash1 := "hash1"

	mapMiniBlockTypeAllTxs := make(map[block.Type]map[string]data.TransactionHandler)

	mapAllTxs := make(map[string]data.TransactionHandler)
	mapAllTxs[txHash1] = tx1
	mapMiniBlockTypeAllTxs[block.TxBlock] = mapAllTxs

	header := &block.Header{
		AccumulatedFees:  big.NewInt(100),
		DeveloperFees:    big.NewInt(10),
		MiniBlockHeaders: []block.MiniBlockHeader{{}, {TxCount: 1}},
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				Type: block.PeerBlock,
			},
			{
				Type:            block.TxBlock,
				TxHashes:        [][]byte{[]byte(txHash1)},
				ReceiverShardID: 1,
			},
		},
	}

	err = tc.verifyFees(header, body, mapMiniBlockTypeAllTxs)
	assert.Nil(t, err)

	enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return flag == common.ScheduledMiniBlocksFlag || flag == common.MiniBlockPartialExecutionFlag
	}

	header = &block.Header{
		AccumulatedFees:  big.NewInt(101),
		DeveloperFees:    big.NewInt(11),
		MiniBlockHeaders: []block.MiniBlockHeader{{}, {TxCount: 1}},
	}
	for index := range header.MiniBlockHeaders {
		_ = header.MiniBlockHeaders[index].SetProcessingType(int32(block.Normal))
	}

	err = tc.verifyFees(header, body, mapMiniBlockTypeAllTxs)
	assert.Nil(t, err)
}

func TestTransactionCoordinator_GetMaxAccumulatedAndDeveloperFeesShouldErr(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:                       &hashingMocks.HasherMock{},
		Marshalizer:                  &mock.MarshalizerMock{},
		ShardCoordinator:             mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                     initAccountsMock(),
		MiniBlockPool:                dataPool.MiniBlocks(),
		RequestHandler:               &testscommon.RequestHandlerStub{},
		PreProcessors:                createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:              createInterimProcessorContainer(),
		GasHandler:                   &testscommon.GasHandlerStub{},
		FeeHandler:                   &mock.FeeAccumulatorStub{},
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	txHash1 := "hash1"

	mb := &block.MiniBlock{
		Type:            block.TxBlock,
		TxHashes:        [][]byte{[]byte(txHash1)},
		ReceiverShardID: 1,
	}

	mbh := &block.MiniBlockHeader{
		TxCount: 1,
	}

	accumulatedFees, developerFees, errGetMaxFees := tc.getMaxAccumulatedAndDeveloperFees(mbh, mb, nil)
	assert.Equal(t, process.ErrMissingTransaction, errGetMaxFees)
	assert.Nil(t, accumulatedFees)
	assert.Nil(t, developerFees)
}

func TestTransactionCoordinator_GetMaxAccumulatedAndDeveloperFeesShouldWork(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)
	tx2GasLimit := uint64(200)
	tx3GasLimit := uint64(300)

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             initAccountsMock(),
		MiniBlockPool:        dataPool.MiniBlocks(),
		RequestHandler:       &testscommon.RequestHandlerStub{},
		PreProcessors:        createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:      createInterimProcessorContainer(),
		GasHandler:           &testscommon.GasHandlerStub{},
		FeeHandler:           &mock.FeeAccumulatorStub{},
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: tx1GasLimit, GasPrice: 1}
	tx2 := &transaction.Transaction{Nonce: 2, GasLimit: tx2GasLimit, GasPrice: 1}
	tx3 := &transaction.Transaction{Nonce: 3, GasLimit: tx3GasLimit, GasPrice: 1}

	txHash1 := "hash1"
	txHash2 := "hash2"
	txHash3 := "hash3"

	mapAllTxs := make(map[string]data.TransactionHandler)
	mapAllTxs[txHash1] = tx1
	mapAllTxs[txHash2] = tx2
	mapAllTxs[txHash3] = tx3

	mb := &block.MiniBlock{
		Type:            block.TxBlock,
		TxHashes:        [][]byte{[]byte(txHash1), []byte(txHash2), []byte(txHash3)},
		ReceiverShardID: 1,
	}

	mbh := &block.MiniBlockHeader{
		TxCount: 3,
	}

	accumulatedFees, developerFees, errGetMaxFees := tc.getMaxAccumulatedAndDeveloperFees(mbh, mb, mapAllTxs)
	assert.Nil(t, errGetMaxFees)
	assert.Equal(t, big.NewInt(600), accumulatedFees)
	assert.Equal(t, big.NewInt(60), developerFees)
}

func TestTransactionCoordinator_RevertIfNeededShouldWork(t *testing.T) {
	t.Parallel()

	restoreGasSinceLastResetCalled := false
	numTxsFeesReverted := 0

	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:           &hashingMocks.HasherMock{},
		Marshalizer:      &mock.MarshalizerMock{},
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(3),
		Accounts:         initAccountsMock(),
		MiniBlockPool:    dataPool.MiniBlocks(),
		RequestHandler:   &testscommon.RequestHandlerStub{},
		PreProcessors:    createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:  createInterimProcessorContainer(),
		GasHandler: &mock.GasHandlerMock{
			RestoreGasSinceLastResetCalled: func(key []byte) {
				restoreGasSinceLastResetCalled = true
			},
		},
		FeeHandler: &mock.FeeAccumulatorStub{
			RevertFeesCalled: func(txHashes [][]byte) {
				numTxsFeesReverted += len(txHashes)
			},
		},
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
	}

	txHashes := make([][]byte, 0)

	txCoordinatorArgs.ShardCoordinator = &mock.CoordinatorStub{
		SelfIdCalled: func() uint32 {
			return 0
		},
	}
	tc, _ := NewTransactionCoordinator(txCoordinatorArgs)

	createMBDestMeExecutionInfo := &createMiniBlockDestMeExecutionInfo{
		processedTxHashes: txHashes,
	}
	tc.revertIfNeeded(createMBDestMeExecutionInfo, []byte("key"))
	assert.False(t, restoreGasSinceLastResetCalled)
	assert.Equal(t, 0, numTxsFeesReverted)

	txCoordinatorArgs.ShardCoordinator = &mock.CoordinatorStub{
		SelfIdCalled: func() uint32 {
			return core.MetachainShardId
		},
	}
	tc, _ = NewTransactionCoordinator(txCoordinatorArgs)

	createMBDestMeExecutionInfo = &createMiniBlockDestMeExecutionInfo{
		processedTxHashes: txHashes,
	}
	tc.revertIfNeeded(createMBDestMeExecutionInfo, []byte("key"))
	assert.False(t, restoreGasSinceLastResetCalled)
	assert.Equal(t, 0, numTxsFeesReverted)

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	txHashes = append(txHashes, txHash1)
	txHashes = append(txHashes, txHash2)

	createMBDestMeExecutionInfo = &createMiniBlockDestMeExecutionInfo{
		processedTxHashes: txHashes,
	}
	tc.revertIfNeeded(createMBDestMeExecutionInfo, []byte("key"))
	assert.True(t, restoreGasSinceLastResetCalled)
	assert.Equal(t, len(txHashes), numTxsFeesReverted)
}

func TestTransactionCoordinator_getFinalCrossMiniBlockInfos(t *testing.T) {
	t.Parallel()

	hash1, hash2 := "hash1", "hash2"

	t.Run("scheduledMiniBlocks flag not set", func(t *testing.T) {
		t.Parallel()

		tc, _ := NewTransactionCoordinator(createMockTransactionCoordinatorArguments())

		var crossMiniBlockInfos []*data.MiniBlockInfo

		mbInfos := tc.getFinalCrossMiniBlockInfos(crossMiniBlockInfos, &block.Header{})
		assert.Equal(t, crossMiniBlockInfos, mbInfos)
	})

	t.Run("should work, miniblocks info found for final miniBlock header", func(t *testing.T) {
		t.Parallel()

		args := createMockTransactionCoordinatorArguments()
		enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
		args.EnableEpochsHandler = enableEpochsHandlerStub
		tc, _ := NewTransactionCoordinator(args)
		enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
			return flag == common.ScheduledMiniBlocksFlag
		}

		mbInfo1 := &data.MiniBlockInfo{Hash: []byte(hash1)}
		mbInfo2 := &data.MiniBlockInfo{Hash: []byte(hash2)}
		crossMiniBlockInfos := []*data.MiniBlockInfo{mbInfo1, mbInfo2}

		mbh1 := block.MiniBlockHeader{Hash: []byte(hash1)}
		mbhReserved1 := block.MiniBlockHeaderReserved{State: block.Proposed}
		mbh1.Reserved, _ = mbhReserved1.Marshal()

		mbh2 := block.MiniBlockHeader{Hash: []byte(hash2)}
		mbhReserved2 := block.MiniBlockHeaderReserved{State: block.Final}
		mbh2.Reserved, _ = mbhReserved2.Marshal()

		header := &block.MetaBlock{
			MiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
				mbh2,
			},
		}

		expectedMbInfos := []*data.MiniBlockInfo{mbInfo2}

		mbInfos := tc.getFinalCrossMiniBlockInfos(crossMiniBlockInfos, header)
		assert.Equal(t, expectedMbInfos, mbInfos)
	})
}

func TestTransactionCoordinator_AddIntermediateTransactions(t *testing.T) {
	t.Parallel()

	args := createMockTransactionCoordinatorArguments()

	t.Run("nil interim processor", func(t *testing.T) {
		t.Parallel()

		tc, _ := NewTransactionCoordinator(args)

		tc.keysInterimProcs = append(tc.keysInterimProcs, block.SmartContractResultBlock)
		tc.interimProcessors[block.SmartContractResultBlock] = nil

		mapSCRs := map[block.Type][]data.TransactionHandler{
			block.SmartContractResultBlock: {
				&smartContractResult.SmartContractResult{
					Nonce: 1,
				},
			},
		}

		err := tc.AddIntermediateTransactions(mapSCRs)
		assert.Equal(t, process.ErrNilIntermediateProcessor, err)
	})

	t.Run("failed to add intermediate transactions", func(t *testing.T) {
		t.Parallel()

		tc, _ := NewTransactionCoordinator(args)

		expectedErr := errors.New("expected err")
		tc.keysInterimProcs = append(tc.keysInterimProcs, block.SmartContractResultBlock)
		tc.interimProcessors[block.SmartContractResultBlock] = &mock.IntermediateTransactionHandlerMock{
			AddIntermediateTransactionsCalled: func(txs []data.TransactionHandler) error {
				return expectedErr
			},
		}

		mapSCRs := map[block.Type][]data.TransactionHandler{
			block.SmartContractResultBlock: {
				&smartContractResult.SmartContractResult{
					Nonce: 1,
				},
			},
		}

		err := tc.AddIntermediateTransactions(mapSCRs)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tc, _ := NewTransactionCoordinator(args)

		expectedTxs := []data.TransactionHandler{
			&smartContractResult.SmartContractResult{
				Nonce: 1,
			},
		}

		tc.keysInterimProcs = append(tc.keysInterimProcs, block.SmartContractResultBlock)
		tc.interimProcessors[block.SmartContractResultBlock] = &mock.IntermediateTransactionHandlerMock{
			AddIntermediateTransactionsCalled: func(txs []data.TransactionHandler) error {
				assert.Equal(t, expectedTxs, txs)
				return nil
			},
		}

		mapSCRs := map[block.Type][]data.TransactionHandler{
			block.SmartContractResultBlock: {
				&smartContractResult.SmartContractResult{
					Nonce: 1,
				},
			},
		}

		err := tc.AddIntermediateTransactions(mapSCRs)
		assert.Nil(t, err)
	})
}

func TestTransactionCoordinator_GetAllIntermediateTxs(t *testing.T) {
	t.Parallel()

	args := createMockTransactionCoordinatorArguments()
	tc, _ := NewTransactionCoordinator(args)

	expectedTxs := map[string]data.TransactionHandler{
		"txHash1": &smartContractResult.SmartContractResult{
			Nonce: 1,
		},
		"txHash2": &smartContractResult.SmartContractResult{
			Nonce: 2,
		},
	}

	tc.keysInterimProcs = append(tc.keysInterimProcs, block.SmartContractResultBlock)
	tc.keysInterimProcs = append(tc.keysInterimProcs, block.PeerBlock)
	tc.interimProcessors[block.ReceiptBlock] = nil
	tc.interimProcessors[block.SmartContractResultBlock] = &mock.IntermediateTransactionHandlerMock{
		GetAllCurrentFinishedTxsCalled: func() map[string]data.TransactionHandler {
			return expectedTxs
		},
	}

	expectedAllIntermediateTxs := map[block.Type]map[string]data.TransactionHandler{
		block.SmartContractResultBlock: expectedTxs,
	}

	txs := tc.GetAllIntermediateTxs()
	assert.Equal(t, expectedAllIntermediateTxs, txs)
}

func TestTransactionCoordinator_AddTxsFromMiniBlocks(t *testing.T) {
	args := createMockTransactionCoordinatorArguments()

	t.Run("adding txs from miniBlocks", func(t *testing.T) {
		mb1 := &block.MiniBlock{
			TxHashes:        [][]byte{[]byte("tx1"), []byte("tx2")},
			ReceiverShardID: 0,
			SenderShardID:   1,
			Type:            block.TxBlock,
			Reserved:        nil,
		}

		mb2 := &block.MiniBlock{
			TxHashes:        [][]byte{[]byte("tx3"), []byte("tx4")},
			ReceiverShardID: 1,
			SenderShardID:   0,
			Type:            block.SmartContractResultBlock,
			Reserved:        nil,
		}

		mb3 := &block.MiniBlock{
			TxHashes:        [][]byte{[]byte("tx5"), []byte("tx6")},
			ReceiverShardID: 1,
			SenderShardID:   0,
			Type:            block.InvalidBlock,
			Reserved:        nil,
		}

		tc, _ := NewTransactionCoordinator(args)
		tc.keysInterimProcs = []block.Type{block.TxBlock}
		tc.txPreProcessors[block.TxBlock] = &mock.PreProcessorMock{
			AddTxsFromMiniBlocksCalled: func(miniBlocks block.MiniBlockSlice) {
				require.Equal(t, miniBlocks, block.MiniBlockSlice{mb1})
			},
		}

		tc.txPreProcessors[block.SmartContractResultBlock] = &mock.PreProcessorMock{
			AddTxsFromMiniBlocksCalled: func(miniBlocks block.MiniBlockSlice) {
				require.Equal(t, miniBlocks, block.MiniBlockSlice{mb2})
			},
		}

		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, fmt.Sprintf("should have not paniced %v", r))
			}
		}()

		tc.AddTxsFromMiniBlocks(block.MiniBlockSlice{mb1, mb2, mb3})
	})
}

func TestTransactionCoordinator_AddTransactions(t *testing.T) {
	args := createMockTransactionCoordinatorArguments()

	txGasLimit := uint64(50000)
	tx1 := &transaction.Transaction{Nonce: 1, GasLimit: txGasLimit, GasPrice: 1}
	tx2 := &transaction.Transaction{Nonce: 2, GasLimit: txGasLimit, GasPrice: 1}
	tx3 := &transaction.Transaction{Nonce: 3, GasLimit: txGasLimit, GasPrice: 1}
	txs := []data.TransactionHandler{tx1, tx2, tx3}

	t.Run("missing preprocessor should not panic", func(t *testing.T) {
		tc, _ := NewTransactionCoordinator(args)
		tc.keysInterimProcs = append(tc.keysInterimProcs, block.InvalidBlock)
		tc.interimProcessors[block.InvalidBlock] = nil

		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, fmt.Sprintf("should have not paniced %v", r))
			}
		}()

		tc.AddTransactions(txs, block.InvalidBlock)
	})

	t.Run("valid preprocessor should add", func(t *testing.T) {
		tc, _ := NewTransactionCoordinator(args)
		addTransactionsCalled := &atomic.Flag{}
		tc.keysTxPreProcs = append(tc.keysTxPreProcs, block.TxBlock)
		tc.txPreProcessors[block.TxBlock] = &mock.PreProcessorMock{
			AddTransactionsCalled: func(txHandlers []data.TransactionHandler) {
				require.Equal(t, txs, txHandlers)
				addTransactionsCalled.SetValue(true)
			},
		}

		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, fmt.Sprintf("should have not paniced %v", r))
			}
		}()

		tc.AddTransactions(txs, block.TxBlock)
		require.True(t, addTransactionsCalled.IsSet())
	})
}

func TestGetProcessedMiniBlockInfo_ShouldWork(t *testing.T) {
	processedMiniBlocksInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)

	processedMbInfo := getProcessedMiniBlockInfo(nil, []byte("hash1"))
	assert.False(t, processedMbInfo.FullyProcessed)
	assert.Equal(t, int32(-1), processedMbInfo.IndexOfLastTxProcessed)

	processedMbInfo = getProcessedMiniBlockInfo(processedMiniBlocksInfo, []byte("hash1"))
	assert.False(t, processedMbInfo.FullyProcessed)
	assert.Equal(t, int32(-1), processedMbInfo.IndexOfLastTxProcessed)
	assert.Equal(t, 1, len(processedMiniBlocksInfo))

	processedMbInfo.IndexOfLastTxProcessed = 69
	processedMbInfo.FullyProcessed = true

	processedMbInfo = getProcessedMiniBlockInfo(processedMiniBlocksInfo, []byte("hash1"))
	assert.True(t, processedMbInfo.FullyProcessed)
	assert.Equal(t, int32(69), processedMbInfo.IndexOfLastTxProcessed)
	assert.Equal(t, 1, len(processedMiniBlocksInfo))
	assert.True(t, processedMiniBlocksInfo["hash1"].FullyProcessed)
	assert.Equal(t, int32(69), processedMiniBlocksInfo["hash1"].IndexOfLastTxProcessed)
}

func TestTransactionCoordinator_getIndexesOfLastTxProcessed(t *testing.T) {
	t.Parallel()

	t.Run("calculating hash error should not get indexes", func(t *testing.T) {
		t.Parallel()

		args := createMockTransactionCoordinatorArguments()
		args.Marshalizer = &marshallerMock.MarshalizerMock{
			Fail: true,
		}
		tc, _ := NewTransactionCoordinator(args)

		miniBlock := &block.MiniBlock{}
		miniBlockHeader := &block.MiniBlockHeader{}

		pi, err := tc.getIndexesOfLastTxProcessed(miniBlock, miniBlockHeader)
		assert.Nil(t, pi)
		assert.Equal(t, marshallerMock.ErrMockMarshalizer, err)
	})

	t.Run("should get indexes", func(t *testing.T) {
		t.Parallel()

		args := createMockTransactionCoordinatorArguments()
		args.Marshalizer = &marshallerMock.MarshalizerMock{
			Fail: false,
		}
		tc, _ := NewTransactionCoordinator(args)

		miniBlock := &block.MiniBlock{}
		miniBlockHash, _ := core.CalculateHash(tc.marshalizer, tc.hasher, miniBlock)
		mbh := &block.MiniBlockHeader{
			Hash:    miniBlockHash,
			TxCount: 6,
		}
		_ = mbh.SetIndexOfFirstTxProcessed(2)
		_ = mbh.SetIndexOfLastTxProcessed(4)

		pi, err := tc.getIndexesOfLastTxProcessed(miniBlock, mbh)
		assert.Nil(t, err)
		assert.Equal(t, int32(-1), pi.indexOfLastTxProcessed)
		assert.Equal(t, mbh.GetIndexOfLastTxProcessed(), pi.indexOfLastTxProcessedByProposer)
	})
}
