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

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const MaxGasLimitPerBlock = uint64(100000)

var txHash = []byte("tx_hash1")

func FeeHandlerMock() *mock.FeeHandlerStub {
	return &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return 0
		},
		MaxGasLimitPerBlockCalled: func() uint64 {
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

func initDataPool(testHash []byte) *testscommon.PoolsHolderStub {
	tx := &transaction.Transaction{
		Nonce: 10,
		Value: big.NewInt(0),
	}
	sc := &smartContractResult.SmartContractResult{Nonce: 10, SndAddr: []byte("0"), RcvAddr: []byte("1")}
	rTx := &rewardTx.RewardTx{Epoch: 0, Round: 1, RcvAddr: []byte("1")}

	txCalled := createShardedDataChacherNotifier(tx, testHash)
	unsignedTxHandler := createShardedDataChacherNotifier(sc, testHash)
	rewardTxCalled := createShardedDataChacherNotifier(rTx, testHash)

	sdp := &testscommon.PoolsHolderStub{
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
	return store
}

func generateTestCache() storage.Cacher {
	cache, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000, Shards: 1, SizeInBytes: 0})
	return cache
}

func generateTestUnit() storage.Storer {
	storer, _ := storageUnit.NewStorageUnit(
		generateTestCache(),
		memorydb.New(),
	)

	return storer
}

func initAccountsMock() *mock.AccountsStub {
	rootHashCalled := func() ([]byte, error) {
		return []byte("rootHash"), nil
	}
	return &mock.AccountsStub{
		RootHashCalled: rootHashCalled,
	}
}

func createMockTransactionCoordinatorArguments() ArgTransactionCoordinator {
	argsTransactionCoordinator := ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(5),
		Accounts:                          &mock.AccountsStub{},
		MiniBlockPool:                     testscommon.NewPoolsHolderMock().MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     &mock.PreProcessorContainerMock{},
		InterProcessors:                   &mock.InterimProcessorContainerMock{},
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{},
		TxTypeHandler:                     &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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

func TestNewTransactionCoordinator_NilFeeAcumulator(t *testing.T) {
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

func TestNewTransactionCoordinator_OK(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)

	assert.Nil(t, err)
	assert.NotNil(t, tc)
	assert.False(t, tc.IsInterfaceNil())
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
		&mock.HasherMock{},
		initDataPool([]byte("tx_hash0")),
		createMockPubkeyConverter(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)
	container, _ := preFactory.Create()

	return container
}

func createInterimProcessorContainer() process.IntermediateProcessorContainer {
	preFactory, _ := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		createMockPubkeyConverter(),
		initStore(),
		initDataPool([]byte("test_hash1")),
	)
	container, _ := preFactory.Create()

	return container
}

func createPreProcessorContainerWithDataPool(
	dataPool dataRetriever.PoolsHolder,
	feeHandler process.FeeHandler,
) process.PreProcessorsContainer {

	totalGasConsumed := uint64(0)
	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		dataPool,
		createMockPubkeyConverter(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&mock.GasHandlerMock{
			SetGasConsumedCalled: func(gasConsumed uint64, hash []byte) {
				totalGasConsumed += gasConsumed
			},
			TotalGasConsumedCalled: func() uint64 {
				return totalGasConsumed
			},
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
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
						gasConsumedByTxInSenderShard := txGasLimitConsumption
						gasConsumedByTxInReceiverShard := tx.GasLimit - txGasLimitConsumption

						return gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, nil
					}

					return tx.GasLimit, tx.GasLimit, nil
				}

				return txGasLimitConsumption, txGasLimitConsumption, nil
			},
			ComputeGasConsumedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			GasRefundedCalled: func(hash []byte) uint64 {
				return 0
			},
			RemoveGasConsumedCalled: func(hashes [][]byte) {
			},
			RemoveGasRefundedCalled: func(hashes [][]byte) {
			},
		},
		&mock.BlockTrackerMock{},
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)
	container, _ := preFactory.Create()

	return container
}

func TestTransactionCoordinator_CreateBlockStarted(t *testing.T) {
	t.Parallel()

	totalGasConsumed := uint64(0)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.GasHandler = &mock.GasHandlerMock{
		InitCalled: func() {
			totalGasConsumed = uint64(0)
		},
		TotalGasConsumedCalled: func() uint64 {
			return totalGasConsumed
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
	scrHash, _ := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, scr)
	scrs = append(scrs, scr)
	body.MiniBlocks = append(body.MiniBlocks, createMiniBlockWithOneTx(0, 1, block.SmartContractResultBlock, scrHash))

	scr = &smartContractResult.SmartContractResult{SndAddr: []byte("snd"), RcvAddr: []byte("rcv"), Value: big.NewInt(199), PrevTxHash: []byte("txHash")}
	scrHash, _ = core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, scr)
	scrs = append(scrs, scr)
	body.MiniBlocks = append(body.MiniBlocks, createMiniBlockWithOneTx(0, 1, block.SmartContractResultBlock, scrHash))

	scr = &smartContractResult.SmartContractResult{SndAddr: []byte("snd"), RcvAddr: []byte("rcv"), Value: big.NewInt(299), PrevTxHash: []byte("txHash")}
	scrHash, _ = core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, scr)
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
	mbs, txs, finalized, err := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(nil, nil, haveTime)

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
	mbs, txs, finalized, err := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(createTestMetablock(), nil, haveTime)

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
	mbs, txs, finalized, err := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(createTestMetablock(), nil, haveTime)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, uint32(0), txs)
	assert.False(t, finalized)
}

func TestTransactionCoordinator_CreateMbsAndProcessCrossShardTransactions(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	tdp := initDataPool(txHash)
	cacherCfg := storageUnit.CacheConfig{Capacity: 100, Type: storageUnit.LRUCache}
	hdrPool, _ := storageUnit.NewCache(cacherCfg)
	tdp.MiniBlocksCalled = func() storage.Cacher {
		return hdrPool
	}

	totalGasConsumed := uint64(0)
	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		tdp,
		createMockPubkeyConverter(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&mock.GasHandlerMock{
			SetGasConsumedCalled: func(gasConsumed uint64, hash []byte) {
				totalGasConsumed += gasConsumed
			},
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasConsumedCalled: func() uint64 {
				return totalGasConsumed
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			TotalGasRefundedCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlockTrackerMock{},
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)
	container, _ := preFactory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = container
	argsTransactionCoordinator.GasHandler = &mock.GasHandlerMock{
		TotalGasConsumedCalled: func() uint64 {
			return totalGasConsumed
		},
	}
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	metaHdr := createTestMetablock()

	for i := 0; i < len(metaHdr.ShardInfo); i++ {
		for j := 0; j < len(metaHdr.ShardInfo[i].ShardMiniBlockHeaders); j++ {
			mbHdr := metaHdr.ShardInfo[i].ShardMiniBlockHeaders[j]
			mb := block.MiniBlock{SenderShardID: mbHdr.SenderShardID, ReceiverShardID: mbHdr.ReceiverShardID, Type: block.TxBlock, TxHashes: [][]byte{txHash}}
			tdp.MiniBlocks().Put(mbHdr.Hash, &mb, mb.Size())
		}
	}

	mbs, txs, finalized, err := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(metaHdr, nil, haveTime)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(mbs))
	assert.Equal(t, uint32(1), txs)
	assert.True(t, finalized)
}

func TestTransactionCoordinator_CreateMbsAndProcessCrossShardTransactionsWithSkippedShard(t *testing.T) {
	t.Parallel()

	mbPool := testscommon.NewPoolsHolderMock().MiniBlocks()
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

	mbs, txs, finalized, err := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(metaBlock, nil, haveTime)
	assert.Nil(t, err)
	require.Equal(t, 1, len(mbs))
	assert.Equal(t, uint32(1), txs)
	assert.False(t, finalized)
	require.Equal(t, 1, len(mbs[0].TxHashes))
	assert.Equal(t, []byte("tx_hash_from_mb0"), mbs[0].TxHashes[0])
}

func TestTransactionCoordinator_CreateMbsAndProcessCrossShardTransactionsNilPreProcessor(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	tdp := initDataPool(txHash)
	cacherCfg := storageUnit.CacheConfig{Capacity: 100, Type: storageUnit.LRUCache}
	hdrPool, _ := storageUnit.NewCache(cacherCfg)
	tdp.MiniBlocksCalled = func() storage.Cacher {
		return hdrPool
	}

	totalGasConsumed := uint64(0)
	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		tdp,
		createMockPubkeyConverter(),
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&mock.GasHandlerMock{
			SetGasConsumedCalled: func(gasConsumed uint64, hash []byte) {
				totalGasConsumed += gasConsumed
			},
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasConsumedCalled: func() uint64 {
				return totalGasConsumed
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			TotalGasRefundedCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlockTrackerMock{},
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)
	container, _ := preFactory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = container
	argsTransactionCoordinator.GasHandler = &mock.GasHandlerMock{
		TotalGasConsumedCalled: func() uint64 {
			return totalGasConsumed
		},
	}
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
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

	mbs, txs, finalized, err := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(metaHdr, nil, haveTime)

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

	totalGasConsumed := uint64(0)
	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&testscommon.PoolsHolderStub{
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
		&mock.AccountsStub{},
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&mock.GasHandlerMock{
			TotalGasConsumedCalled: func() uint64 {
				return totalGasConsumed
			},
		},
		&mock.BlockTrackerMock{},
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)
	container, _ := preFactory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.MiniBlockPool = testscommon.NewPoolsHolderMock().MiniBlocks()
	argsTransactionCoordinator.PreProcessors = container
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime)

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
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime)

	assert.Equal(t, 0, len(mbs))
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMeNoSpace(t *testing.T) {
	t.Parallel()
	totalGasConsumed := uint64(0)
	tdp := initDataPool(txHash)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	argsTransactionCoordinator.GasHandler = &mock.GasHandlerMock{
		TotalGasConsumedCalled: func() uint64 {
			return totalGasConsumed
		},
	}
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime)

	assert.Equal(t, 0, len(mbs))
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMe(t *testing.T) {
	t.Parallel()

	nrShards := uint32(5)
	txPool, _ := testscommon.CreateTxPool(nrShards, 0)
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
	hasher := &mock.HasherMock{}
	for shId := uint32(0); shId < nrShards; shId++ {
		strCache := process.ShardCacherIdentifier(0, shId)
		newTx := &transaction.Transaction{GasLimit: uint64(shId)}

		txHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		txPool.AddData(txHash, newTx, newTx.Size(), strCache)
	}

	// we have one tx per shard.
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime)

	assert.Equal(t, int(nrShards), len(mbs))
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMeMultipleMiniblocks(t *testing.T) {
	t.Parallel()

	nrShards := uint32(5)
	txPool, _ := testscommon.CreateTxPool(nrShards, 0)
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
	hasher := &mock.HasherMock{}

	sndShardId := uint32(0)
	dstShardId := uint32(1)
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)

	numTxsToAdd := 100
	gasLimit := MaxGasLimitPerBlock / uint64(numTxsToAdd)

	scAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")

	allTxs := 100
	for i := 0; i < allTxs; i++ {
		newTx := &transaction.Transaction{GasLimit: gasLimit, GasPrice: uint64(i), RcvAddr: scAddress}

		txHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		txPool.AddData(txHash, newTx, newTx.Size(), strCache)

	}

	// we have one tx per shard.
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime)

	assert.Equal(t, 1, len(mbs))
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMeMultipleMiniblocksShouldApplyGasLimit(t *testing.T) {
	t.Parallel()

	allTxs := 100
	numTxsToAdd := 20
	gasLimit := MaxGasLimitPerBlock / uint64(numTxsToAdd)
	numMiniBlocks := allTxs / numTxsToAdd

	nrShards := uint32(5)
	txPool, _ := testscommon.CreateTxPool(nrShards, 0)
	tdp := initDataPool(txHash)
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return txPool
	}

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(nrShards)
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(
		tdp,
		&mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
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
	hasher := &mock.HasherMock{}

	sndShardId := uint32(0)
	dstShardId := uint32(1)
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)

	scAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")

	for i := 0; i < allTxs; i++ {
		newTx := &transaction.Transaction{GasLimit: gasLimit + gasLimit/uint64(numMiniBlocks), GasPrice: uint64(i), RcvAddr: scAddress}

		txHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		txPool.AddData(txHash, newTx, newTx.Size(), strCache)
	}

	// we have one tx per shard.
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime)

	assert.Equal(t, 1, len(mbs))
}

func TestTransactionCoordinator_CompactAndExpandMiniblocksShouldWork(t *testing.T) {
	t.Parallel()

	numTxsPerBulk := 100
	numTxsToAdd := 20
	gasLimit := MaxGasLimitPerBlock / uint64(numTxsToAdd)

	nrShards := uint32(5)
	txPool, _ := testscommon.CreateTxPool(nrShards, 0)
	tdp := initDataPool(txHash)
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return txPool
	}

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(nrShards)
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(
		tdp,
		&mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
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
	hasher := &mock.HasherMock{}

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

			txHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
			txPool.AddData(txHash, newTx, newTx.Size(), shardCacher)
		}
	}

	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime)

	assert.Equal(t, 1, len(mbs))
}

func TestTransactionCoordinator_GetAllCurrentUsedTxs(t *testing.T) {
	t.Parallel()

	nrShards := uint32(5)
	txPool, _ := testscommon.CreateTxPool(nrShards, 0)
	tdp := initDataPool(txHash)
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return txPool
	}

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(nrShards)
	argsTransactionCoordinator.MiniBlockPool = tdp.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = createPreProcessorContainerWithDataPool(tdp, FeeHandlerMock())
	argsTransactionCoordinator.GasHandler = &mock.GasHandlerMock{
		ComputeGasConsumedByTxCalled: func(txSndShId uint32, txRcvShId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
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
	hasher := &mock.HasherMock{}
	for i := uint32(0); i < nrShards; i++ {
		strCache := process.ShardCacherIdentifier(0, i)
		newTx := &transaction.Transaction{GasLimit: uint64(i)}

		txHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		txPool.AddData(txHash, newTx, newTx.Size(), strCache)
	}

	mbs := tc.CreateMbsAndProcessTransactionsFromMe(haveTime)
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

	err = tc.SaveTxsToStorage(nil)
	assert.Nil(t, err)

	body := &block.Body{}
	miniBlock := &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHash}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	tc.RequestBlockTransactions(body)

	err = tc.SaveTxsToStorage(body)
	assert.Nil(t, err)

	txHashToAsk := []byte("tx_hashnotinPool")
	miniBlock = &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHashToAsk}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	err = tc.SaveTxsToStorage(body)
	assert.Equal(t, process.ErrMissingTransaction, err)
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
	err = tc.SaveTxsToStorage(body)
	assert.Nil(t, err)
	nrTxs, err = tc.RestoreBlockDataFromStorage(body)
	assert.Equal(t, 1, nrTxs)
	assert.Nil(t, err)

	txHashToAsk := []byte("tx_hashnotinPool")
	miniBlock = &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHashToAsk}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	err = tc.SaveTxsToStorage(body)
	assert.Equal(t, process.ErrMissingTransaction, err)

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
		&mock.HasherMock{},
		dataPool,
		createMockPubkeyConverter(),
		accounts,
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, process.ErrHigherNonceInTransaction
			},
		},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&mock.GasHandlerMock{
			ComputeGasConsumedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasConsumedCalled: func() uint64 {
				return 0
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
		&mock.BlockTrackerMock{},
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
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
	err = tc.ProcessBlockTransaction(&block.Body{}, haveTime)
	assert.Nil(t, err)

	body := &block.Body{}
	miniBlock := &block.MiniBlock{SenderShardID: 1, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHash}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	tc.RequestBlockTransactions(body)
	err = tc.ProcessBlockTransaction(body, haveTime)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)

	noTime := func() time.Duration {
		return 0
	}
	err = tc.ProcessBlockTransaction(body, noTime)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)

	txHashToAsk := []byte("tx_hashnotinPool")
	miniBlock = &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHashToAsk}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)
	err = tc.ProcessBlockTransaction(body, haveTime)
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
	err = tc.ProcessBlockTransaction(&block.Body{}, haveTime)
	assert.Nil(t, err)

	body := &block.Body{}
	miniBlock := &block.MiniBlock{SenderShardID: 1, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHash}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	tc.RequestBlockTransactions(body)
	err = tc.ProcessBlockTransaction(body, haveTime)
	assert.Nil(t, err)

	noTime := func() time.Duration {
		return -1
	}
	err = tc.ProcessBlockTransaction(body, noTime)
	assert.Equal(t, process.ErrTimeIsOut, err)

	txHashToAsk := []byte("tx_hashnotinPool")
	miniBlock = &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, Type: block.TxBlock, TxHashes: [][]byte{txHashToAsk}}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)
	err = tc.ProcessBlockTransaction(body, haveTime)
	assert.Equal(t, process.ErrMissingTransaction, err)
}

func TestTransactionCoordinator_RequestMiniblocks(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool(txHash)
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	nrCalled := 0
	mutex := sync.Mutex{}

	requestHandler := &mock.RequestHandlerStub{
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
		&mock.HasherMock{},
		dataPool,
		createMockPubkeyConverter(),
		accounts,
		requestHandler,
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
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

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := testscommon.NewPoolsHolderMock()

	//we will have a miniblock that will have 3 tx hashes
	//all txs will be in datapool and none of them will return err when processed
	//so, tx processor will return nil on processing tx

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

	//put the existing tx inside datapool
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

	accounts := &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}

	totalGasConsumed := uint64(0)

	preFactory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		marshalizer,
		hasher,
		dataPool,
		createMockPubkeyConverter(),
		accounts,
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				//execution, in this context, means moving the tx nonce to itx corresponding execution result variable
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
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&mock.GasHandlerMock{
			SetGasConsumedCalled: func(gasConsumed uint64, hash []byte) {
				totalGasConsumed += gasConsumed
			},
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasConsumedCalled: func() uint64 {
				return 0
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			TotalGasRefundedCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlockTrackerMock{},
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)
	container, _ := preFactory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	argsTransactionCoordinator.Accounts = accounts
	argsTransactionCoordinator.MiniBlockPool = dataPool.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = container
	argsTransactionCoordinator.GasHandler = &mock.GasHandlerMock{
		TotalGasConsumedCalled: func() uint64 {
			return 0
		},
	}
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	preproc := tc.getPreProcessor(block.TxBlock)
	err = tc.processCompleteMiniBlock(preproc, &miniBlock, []byte("hash"), haveTime)

	assert.Nil(t, err)
	assert.Equal(t, tx1Nonce, tx1ExecutionResult)
	assert.Equal(t, tx2Nonce, tx2ExecutionResult)
	assert.Equal(t, tx3Nonce, tx3ExecutionResult)
}

func TestShardProcessor_ProcessMiniBlockCompleteWithErrorWhileProcessShouldCallRevertAccntState(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := testscommon.NewPoolsHolderMock()

	//we will have a miniblock that will have 3 tx hashes
	//all txs will be in datapool and none of them will return err when processed
	//so, tx processor will return nil on processing tx

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

	//put the existing tx inside datapool
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

	accounts := &mock.AccountsStub{
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
		&mock.RequestHandlerStub{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				if bytes.Equal(transaction.Data, txHash2) {
					return 0, process.ErrHigherNonceInTransaction
				}
				return 0, nil
			},
		},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		FeeHandlerMock(),
		&mock.GasHandlerMock{
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasConsumedCalled: func() uint64 {
				return 0
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			TotalGasRefundedCalled: func() uint64 {
				return 0
			},
			SetGasConsumedCalled: func(gasConsumed uint64, hash []byte) {},
			RemoveGasRefundedCalled: func(hashes [][]byte) {
			},
			RemoveGasConsumedCalled: func(hashes [][]byte) {
			},
		},
		&mock.BlockTrackerMock{},
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)
	container, _ := preFactory.Create()

	totalGasConsumed := uint64(0)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	argsTransactionCoordinator.Accounts = accounts
	argsTransactionCoordinator.MiniBlockPool = dataPool.MiniBlocks()
	argsTransactionCoordinator.PreProcessors = container
	argsTransactionCoordinator.GasHandler = &mock.GasHandlerMock{
		TotalGasConsumedCalled: func() uint64 {
			return totalGasConsumed
		},
		SetGasConsumedCalled: func(gasConsumed uint64, hash []byte) {
			totalGasConsumed = gasConsumed
		},
	}
	tc, err := NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	haveTime := func() bool {
		return true
	}
	preproc := tc.getPreProcessor(block.TxBlock)
	err = tc.processCompleteMiniBlock(preproc, &miniBlock, []byte("hash"), haveTime)

	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
	assert.True(t, revertAccntStateCalled)
}

func TestTransactionCoordinator_VerifyCreatedBlockTransactionsNilOrMiss(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	tdp := initDataPool(txHash)
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	preFactory, _ := shard.NewIntermediateProcessorsContainerFactory(
		shardCoordinator,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		tdp,
	)
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

	txHash := []byte("txHash")
	tdp := initDataPool(txHash)
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	interFactory, _ := shard.NewIntermediateProcessorsContainerFactory(
		shardCoordinator,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		tdp,
	)
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
	scrHash, _ := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, scr)

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

func TestTransactionCoordinator_SaveTxsToStorageSaveIntermediateTxsErrors(t *testing.T) {
	t.Parallel()

	tdp := initDataPool(txHash)
	retError := errors.New("save error")
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
					SaveCurrentIntermediateTxToStorageCalled: func() error {
						return retError
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

	err = tc.SaveTxsToStorage(body)
	assert.Equal(t, retError, err)
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
					SaveCurrentIntermediateTxToStorageCalled: func() error {
						intermediateTxWereSaved = true
						return nil
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

	err = tc.SaveTxsToStorage(body)
	assert.Nil(t, err)

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

func TestTransactionCoordinator_CreateMarshalizedReceiptsShouldWork(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	tc, _ := NewTransactionCoordinator(argsTransactionCoordinator)

	mb1 := &block.MiniBlock{
		Type: block.SmartContractResultBlock,
	}
	mb2 := &block.MiniBlock{
		Type: block.ReceiptBlock,
	}

	mbsBatch := &batch.Batch{}
	marshalizedMb1, _ := tc.marshalizer.Marshal(mb1)
	marshalizedMb2, _ := tc.marshalizer.Marshal(mb2)
	mbsBatch.Data = append(mbsBatch.Data, marshalizedMb1)
	mbsBatch.Data = append(mbsBatch.Data, marshalizedMb2)
	expectedMarshalizedReceipts, _ := tc.marshalizer.Marshal(mbsBatch)

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

	marshalizedReceipts, err := tc.CreateMarshalizedReceipts()

	assert.Nil(t, err)
	assert.Equal(t, expectedMarshalizedReceipts, marshalizedReceipts)
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

	numMbs, numTxs := tc.getNumOfCrossInterMbsAndTxs()

	assert.Equal(t, 5, numMbs)
	assert.Equal(t, 10, numTxs)
}

func TestTransactionCoordinator_IsMaxBlockSizeReachedShouldWork(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	argsTransactionCoordinator.BlockSizeComputation = &mock.BlockSizeComputationStub{
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

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs := ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{},
		TxTypeHandler:                     &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 1,
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
	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return maxGasLimitPerBlock + 1
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
		},
		TxTypeHandler:                     &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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

	header := &block.Header{}
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
	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                     &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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
		AccumulatedFees: big.NewInt(101),
		DeveloperFees:   big.NewInt(10),
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
	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                     &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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
		AccumulatedFees: big.NewInt(100),
		DeveloperFees:   big.NewInt(11),
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
	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                     &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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
		AccumulatedFees: big.NewInt(100),
		DeveloperFees:   big.NewInt(10),
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

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{},
		TxTypeHandler:                     &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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

	tx1GasLimit := uint64(100)
	tx2GasLimit := uint64(200)
	tx3GasLimit := uint64(300)

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return tx1GasLimit + tx2GasLimit + tx3GasLimit - 1
			},
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
		},
		TxTypeHandler:                     &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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

	err = tc.verifyGasLimit(body, mapMiniBlockTypeAllTxs)
	assert.Equal(t, process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached, err)
}

func TestTransactionCoordinator_VerifyGasLimitShouldWork(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)
	tx2GasLimit := uint64(200)
	tx3GasLimit := uint64(300)

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return tx1GasLimit + tx2GasLimit + tx3GasLimit
			},
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
		},
		TxTypeHandler:                     &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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

	err = tc.verifyGasLimit(body, mapMiniBlockTypeAllTxs)
	assert.Nil(t, err)
}

func TestTransactionCoordinator_CheckGasConsumedByMiniBlockInReceiverShardShouldErrMissingTransaction(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{},
		TxTypeHandler:                     &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
	}
	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	mb := &block.MiniBlock{
		Type:            block.TxBlock,
		TxHashes:        [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
		ReceiverShardID: 1,
	}

	err = tc.checkGasConsumedByMiniBlockInReceiverShard(mb, nil)
	assert.Equal(t, err, process.ErrMissingTransaction)
}

func TestTransactionCoordinator_CheckGasConsumedByMiniBlockInReceiverShardShouldErrSubtractionOverflow(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit() + 1
			},
		},
		TxTypeHandler:                     &mock.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.MoveBalance, process.SCInvoking
			},
		},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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

	err = tc.checkGasConsumedByMiniBlockInReceiverShard(mb, mapAllTxs)
	assert.Equal(t, err, core.ErrSubtractionOverflow)
}

func TestTransactionCoordinator_CheckGasConsumedByMiniBlockInReceiverShardShouldErrAdditionOverflow(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(math.MaxUint64)
	tx2GasLimit := uint64(1)

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 0
			},
		},
		TxTypeHandler:                     &mock.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.MoveBalance, process.SCInvoking
			},
		},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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

	err = tc.checkGasConsumedByMiniBlockInReceiverShard(mb, mapAllTxs)
	assert.Equal(t, err, core.ErrAdditionOverflow)
}

func TestTransactionCoordinator_CheckGasConsumedByMiniBlockInReceiverShardShouldErrMaxGasLimitPerMiniBlockInReceiverShardIsReached(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)
	tx2GasLimit := uint64(200)
	tx3GasLimit := uint64(300)

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return tx1GasLimit + tx2GasLimit + tx3GasLimit - 1
			},
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
		},
		TxTypeHandler:                    &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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

	err = tc.checkGasConsumedByMiniBlockInReceiverShard(mb, mapAllTxs)
	assert.Equal(t, err, process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached)
}

func TestTransactionCoordinator_CheckGasConsumedByMiniBlockInReceiverShardShouldWork(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)
	tx2GasLimit := uint64(200)
	tx3GasLimit := uint64(300)

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return tx1GasLimit + tx2GasLimit + tx3GasLimit
			},
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
		},
		TxTypeHandler:                    &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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

	err = tc.checkGasConsumedByMiniBlockInReceiverShard(mb, mapAllTxs)
	assert.Nil(t, err)
}

func TestTransactionCoordinator_VerifyFeesShouldErrMissingTransaction(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{},
		TxTypeHandler:                    &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
	}

	tc, err := NewTransactionCoordinator(txCoordinatorArgs)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	txHash1 := "hash1"

	header := &block.Header{
		AccumulatedFees: big.NewInt(100),
		DeveloperFees:   big.NewInt(10),
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

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                    &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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
		AccumulatedFees: big.NewInt(101),
		DeveloperFees:   big.NewInt(10),
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

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                    &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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
		AccumulatedFees: big.NewInt(100),
		DeveloperFees:   big.NewInt(11),
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

func TestTransactionCoordinator_VerifyFeesShouldWork(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                    &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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
		AccumulatedFees: big.NewInt(100),
		DeveloperFees:   big.NewInt(10),
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
}

func TestTransactionCoordinator_GetMaxAccumulatedAndDeveloperFeesShouldErr(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{},
		TxTypeHandler:                    &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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

	accumulatedFees, developerFees, errGetMaxFees := tc.getMaxAccumulatedAndDeveloperFees(mb, nil)
	assert.Equal(t, process.ErrMissingTransaction, errGetMaxFees)
	assert.Equal(t, big.NewInt(0), accumulatedFees)
	assert.Equal(t, big.NewInt(0), developerFees)
}

func TestTransactionCoordinator_GetMaxAccumulatedAndDeveloperFeesShouldWork(t *testing.T) {
	t.Parallel()

	tx1GasLimit := uint64(100)
	tx2GasLimit := uint64(200)
	tx3GasLimit := uint64(300)

	txHash := []byte("tx_hash1")
	dataPool := initDataPool(txHash)
	txCoordinatorArgs:= ArgTransactionCoordinator{
		Hasher:                            &mock.HasherMock{},
		Marshalizer:                       &mock.MarshalizerMock{},
		ShardCoordinator:                  mock.NewMultiShardsCoordinatorMock(3),
		Accounts:                          initAccountsMock(),
		MiniBlockPool:                     dataPool.MiniBlocks(),
		RequestHandler:                    &mock.RequestHandlerStub{},
		PreProcessors:                     createPreProcessorContainerWithDataPool(dataPool, FeeHandlerMock()),
		InterProcessors:                   createInterimProcessorContainer(),
		GasHandler:                        &mock.GasHandlerMock{},
		FeeHandler:                        &mock.FeeAccumulatorStub{},
		BlockSizeComputation:              &mock.BlockSizeComputationStub{},
		BalanceComputation:                &mock.BalanceComputationStub{},
		EconomicsFee:                      &mock.FeeHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.1
			},
		},
		TxTypeHandler:                    &mock.TxTypeHandlerMock{},
		BlockGasAndFeesReCheckEnableEpoch: 0,
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

	accumulatedFees, developerFees, errGetMaxFees := tc.getMaxAccumulatedAndDeveloperFees(mb, mapAllTxs)
	assert.Nil(t, errGetMaxFees)
	assert.Equal(t, big.NewInt(600), accumulatedFees)
	assert.Equal(t, big.NewInt(60), developerFees)
}
