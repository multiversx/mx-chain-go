package preprocess

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

const MaxGasLimitPerBlock = uint64(100000)

func feeHandlerMock() *mock.FeeHandlerStub {
	return &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return 0
		},
		MaxGasLimitPerBlockCalled: func() uint64 {
			return MaxGasLimitPerBlock
		},
	}
}

func shardedDataCacherNotifier() dataRetriever.ShardedDataCacherNotifier {
	return &testscommon.ShardedDataStub{
		ShardDataStoreCalled: func(id string) (c storage.Cacher) {
			return &testscommon.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &smartContractResult.SmartContractResult{Nonce: 10}, true
					}
					if reflect.DeepEqual(key, []byte("tx2_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
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
		AddDataCalled:                 func(key []byte, data interface{}, sizeInBytes int, cacheId string) {},
		RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {},
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			if reflect.DeepEqual(key, []byte("tx1_hash")) {
				return &smartContractResult.SmartContractResult{Nonce: 10}, true
			}
			if reflect.DeepEqual(key, []byte("tx2_hash")) {
				return &transaction.Transaction{Nonce: 10}, true
			}
			return nil, false
		},
	}
}

func initDataPool() *testscommon.PoolsHolderStub {
	sdp := &testscommon.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return shardedDataCacherNotifier()
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return shardedDataCacherNotifier()
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(id string) (c storage.Cacher) {
					return &testscommon.CacherStub{
						PeekCalled: func(key []byte) (value interface{}, ok bool) {
							if reflect.DeepEqual(key, []byte("tx1_hash")) {
								return &rewardTx.RewardTx{Value: big.NewInt(100)}, true
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
				AddDataCalled:                 func(key []byte, data interface{}, sizeInBytes int, cacheId string) {},
				RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {},
				SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &rewardTx.RewardTx{Value: big.NewInt(100)}, true
					}
					return nil, false
				},
			}
		},
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
			return cs
		},
		HeadersCalled: func() dataRetriever.HeadersPool {
			cs := &mock.HeadersCacherStub{}
			cs.RegisterHandlerCalled = func(i func(header data.HeaderHandler, key []byte)) {
			}
			return cs
		},
	}
	return sdp
}

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilPool(t *testing.T) {
	t.Parallel()

	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		nil,
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilStore(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		nil,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilTxStorage, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilHasher(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		nil,
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilMarsalizer(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		nil,
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilTxProce(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilTxProcessor, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilShardCoord(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		nil,
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilAccounts(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilRequestFunc(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		nil,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilFeeHandler(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		nil,
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilGasHandler(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		nil,
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilGasHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilBlockTracker(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		nil,
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilBlockTracker, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilPubkeyConverter(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		nil,
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilBlockSizeComputationHandler(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		nil,
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilBalanceComputationHandler(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		nil,
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilEpochNotifier(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		nil,
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilEpochNotifier, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilTxTypeHandler(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		nil,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilTxTypeHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorOkValsShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, txs)
	assert.False(t, txs.IsInterfaceNil())
}

func TestTxsPreProcessor_GetTransactionFromPool(t *testing.T) {
	t.Parallel()
	dataPool := initDataPool()
	txs := createGoodPreprocessor(dataPool)
	txHash := []byte("tx2_hash")
	tx, _ := process.GetTransactionHandlerFromPool(1, 1, txHash, dataPool.Transactions(), false)
	assert.NotNil(t, txs)
	assert.NotNil(t, tx)
	assert.Equal(t, uint64(10), tx.(*transaction.Transaction).Nonce)
}

func TestTransactionPreprocessor_RequestTransactionFromNetwork(t *testing.T) {
	t.Parallel()
	dataPool := initDataPool()
	txs := createGoodPreprocessor(dataPool)
	shardID := uint32(1)
	txHash1 := []byte("tx_hash1")
	txHash2 := []byte("tx_hash2")
	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash1)
	txHashes = append(txHashes, txHash2)
	mBlk := block.MiniBlock{ReceiverShardID: shardID, TxHashes: txHashes}
	body.MiniBlocks = append(body.MiniBlocks, &mBlk)
	txsRequested := txs.RequestBlockTransactions(body)
	assert.Equal(t, 2, txsRequested)
}

func TestTransactionPreprocessor_RequestBlockTransactionFromMiniBlockFromNetwork(t *testing.T) {
	t.Parallel()
	dataPool := initDataPool()
	txs := createGoodPreprocessor(dataPool)

	shardID := uint32(1)
	txHash1 := []byte("tx_hash1")
	txHash2 := []byte("tx_hash2")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash1)
	txHashes = append(txHashes, txHash2)
	mb := &block.MiniBlock{ReceiverShardID: shardID, TxHashes: txHashes}
	txsRequested := txs.RequestTransactionsForMiniBlock(mb)
	assert.Equal(t, 2, txsRequested)
}

func TestTransactionPreprocessor_ReceivedTransactionShouldEraseRequested(t *testing.T) {
	t.Parallel()

	dataPool := testscommon.NewPoolsHolderMock()

	shardedDataStub := &testscommon.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &testscommon.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return &transaction.Transaction{}, true
				},
			}
		},
	}

	dataPool.SetTransactions(shardedDataStub)

	txs := createGoodPreprocessor(dataPool)

	//add 3 tx hashes on requested list
	txHash1 := []byte("tx hash 1")
	txHash2 := []byte("tx hash 2")
	txHash3 := []byte("tx hash 3")

	txs.AddTxHashToRequestedList(txHash1)
	txs.AddTxHashToRequestedList(txHash2)
	txs.AddTxHashToRequestedList(txHash3)

	txs.SetMissingTxs(3)

	//received txHash2
	txs.ReceivedTransaction(txHash2, &txcache.WrappedTransaction{Tx: &transaction.Transaction{}})

	assert.True(t, txs.IsTxHashRequested(txHash1))
	assert.False(t, txs.IsTxHashRequested(txHash2))
	assert.True(t, txs.IsTxHashRequested(txHash3))
}

//------- GetAllTxsFromMiniBlock

func computeHash(data interface{}, marshalizer marshal.Marshalizer, hasher hashing.Hasher) []byte {
	buff, _ := marshalizer.Marshal(data)
	return hasher.Compute(string(buff))
}

func TestTransactionPreprocessor_GetAllTxsFromMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := testscommon.NewPoolsHolderMock()
	senderShardId := uint32(0)
	destinationShardId := uint32(1)

	txsSlice := []*transaction.Transaction{
		{Nonce: 1},
		{Nonce: 2},
		{Nonce: 3},
	}
	transactionsHashes := make([][]byte, len(txsSlice))

	//add defined transactions to sender-destination cacher
	for idx, tx := range txsSlice {
		transactionsHashes[idx] = computeHash(tx, marshalizer, hasher)

		dataPool.Transactions().AddData(
			transactionsHashes[idx],
			tx,
			tx.Size(),
			process.ShardCacherIdentifier(senderShardId, destinationShardId),
		)
	}

	//add some random data
	txRandom := &transaction.Transaction{Nonce: 4}
	dataPool.Transactions().AddData(
		computeHash(txRandom, marshalizer, hasher),
		txRandom,
		txRandom.Size(),
		process.ShardCacherIdentifier(3, 4),
	)

	txs := createGoodPreprocessor(dataPool)

	mb := &block.MiniBlock{
		SenderShardID:   senderShardId,
		ReceiverShardID: destinationShardId,
		TxHashes:        transactionsHashes,
	}

	txsRetrieved, txHashesRetrieved, err := txs.getAllTxsFromMiniBlock(mb, func() bool { return true })

	assert.Nil(t, err)
	assert.Equal(t, len(txsSlice), len(txsRetrieved))
	assert.Equal(t, len(txsSlice), len(txHashesRetrieved))
	for idx, tx := range txsSlice {
		//txReceived should be all txs in the same order
		assert.Equal(t, txsRetrieved[idx], tx)
		//verify corresponding transaction hashes
		assert.Equal(t, txHashesRetrieved[idx], computeHash(tx, marshalizer, hasher))
	}
}

func TestTransactionPreprocessor_RemoveBlockDataFromPoolsNilBlockShouldErr(t *testing.T) {
	t.Parallel()
	dataPool := initDataPool()
	txs := createGoodPreprocessor(dataPool)
	err := txs.RemoveBlockDataFromPools(nil, dataPool.MiniBlocks())
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilTxBlockBody)
}

func TestTransactionPreprocessor_RemoveBlockDataFromPoolsOK(t *testing.T) {
	t.Parallel()
	dataPool := initDataPool()
	txs := createGoodPreprocessor(dataPool)
	body := &block.Body{}
	txHash := []byte("txHash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)
	err := txs.RemoveBlockDataFromPools(body, dataPool.MiniBlocks())
	assert.Nil(t, err)
}

func TestTransactions_CreateAndProcessMiniBlockCrossShardGasLimitAddAll(t *testing.T) {
	t.Parallel()

	txPool, _ := testscommon.CreateTxPool(2, 0)
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

	totalGasConsumed := uint64(0)
	txs, _ := NewTransactionPreprocessor(
		txPool,
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		&mock.TxProcessorMock{ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			return 0, nil
		}},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{
			SetGasConsumedCalled: func(gasConsumed uint64, hash []byte) {
				totalGasConsumed += gasConsumed
			},
			TotalGasConsumedCalled: func() uint64 {
				return totalGasConsumed
			},
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			TotalGasRefundedCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)
	assert.NotNil(t, txs)

	sndShardId := uint32(0)
	dstShardId := uint32(1)
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)

	addedTxs := make([]*transaction.Transaction, 0)
	for i := 0; i < 10; i++ {
		newTx := &transaction.Transaction{GasLimit: uint64(i)}

		txHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		txPool.AddData(txHash, newTx, newTx.Size(), strCache)

		addedTxs = append(addedTxs, newTx)
	}

	sortedTxsAndHashes, _ := txs.computeSortedTxs(sndShardId, dstShardId)
	miniBlocks, err := txs.createAndProcessMiniBlocksFromMeV1(haveTimeTrue, isShardStuckFalse, isMaxBlockSizeReachedFalse, sortedTxsAndHashes)
	assert.Nil(t, err)

	txHashes := 0
	for _, miniBlock := range miniBlocks {
		txHashes += len(miniBlock.TxHashes)
	}

	assert.Equal(t, len(addedTxs), txHashes)
}

func TestTransactions_CreateAndProcessMiniBlockCrossShardGasLimitAddAllAsNoSCCalls(t *testing.T) {
	t.Parallel()

	txPool, _ := testscommon.CreateTxPool(2, 0)
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

	totalGasConsumed := uint64(0)
	txs, _ := NewTransactionPreprocessor(
		txPool,
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		&mock.TxProcessorMock{ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			return 0, nil
		}},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{
			SetGasConsumedCalled: func(gasConsumed uint64, hash []byte) {
				totalGasConsumed += gasConsumed
			},
			TotalGasConsumedCalled: func() uint64 {
				return totalGasConsumed
			},
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			TotalGasRefundedCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)
	assert.NotNil(t, txs)

	sndShardId := uint32(0)
	dstShardId := uint32(1)
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)

	gasLimit := MaxGasLimitPerBlock / uint64(5)

	addedTxs := make([]*transaction.Transaction, 0)
	for i := 0; i < 10; i++ {
		newTx := &transaction.Transaction{GasLimit: gasLimit, GasPrice: uint64(i), RcvAddr: []byte("012345678910")}

		txHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		txPool.AddData(txHash, newTx, newTx.Size(), strCache)

		addedTxs = append(addedTxs, newTx)
	}

	sortedTxsAndHashes, _ := txs.computeSortedTxs(sndShardId, dstShardId)
	miniBlocks, err := txs.createAndProcessMiniBlocksFromMeV1(haveTimeTrue, isShardStuckFalse, isMaxBlockSizeReachedFalse, sortedTxsAndHashes)
	assert.Nil(t, err)

	txHashes := 0
	for _, miniBlock := range miniBlocks {
		txHashes += len(miniBlock.TxHashes)
	}

	assert.Equal(t, len(addedTxs), txHashes)
}

func TestTransactions_CreateAndProcessMiniBlockCrossShardGasLimitAddOnly5asSCCall(t *testing.T) {
	t.Parallel()

	txPool, _ := testscommon.CreateTxPool(2, 0)
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

	numTxsToAdd := 5
	gasLimit := MaxGasLimitPerBlock / uint64(numTxsToAdd)

	totalGasConsumed := uint64(0)
	txs, _ := NewTransactionPreprocessor(
		txPool,
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		&mock.TxProcessorMock{ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			return 0, nil
		}},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{
			SetGasConsumedCalled: func(gasConsumed uint64, hash []byte) {
				totalGasConsumed += gasConsumed
			},
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return gasLimit, gasLimit, nil
			},
			TotalGasConsumedCalled: func() uint64 {
				return totalGasConsumed
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			GasRefundedCalled: func(hash []byte) uint64 {
				return 0
			},
			RemoveGasConsumedCalled: func(hashes [][]byte) {
				totalGasConsumed = 0
			},
			RemoveGasRefundedCalled: func(hashes [][]byte) {
			},
		},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)
	assert.NotNil(t, txs)

	sndShardId := uint32(0)
	dstShardId := uint32(1)
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)

	scAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	for i := 0; i < 10; i++ {
		newTx := &transaction.Transaction{GasLimit: gasLimit, GasPrice: uint64(i), RcvAddr: scAddress}

		txHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		txPool.AddData(txHash, newTx, newTx.Size(), strCache)
	}

	sortedTxsAndHashes, _ := txs.computeSortedTxs(sndShardId, dstShardId)
	miniBlocks, err := txs.createAndProcessMiniBlocksFromMeV1(haveTimeTrue, isShardStuckFalse, isMaxBlockSizeReachedFalse, sortedTxsAndHashes)
	assert.Nil(t, err)

	txHashes := 0
	for _, miniBlock := range miniBlocks {
		txHashes += len(miniBlock.TxHashes)
	}

	assert.Equal(t, numTxsToAdd, txHashes)
}

func TestTransactions_IsDataPrepared_NumMissingTxsZeroShouldWork(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool()
	txs := createGoodPreprocessor(dataPool)

	err := txs.IsDataPrepared(0, haveTime)
	assert.Nil(t, err)
}

func TestTransactions_IsDataPrepared_NumMissingTxsGreaterThanZeroTxNotReceivedShouldTimeout(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool()
	txs := createGoodPreprocessor(dataPool)

	haveTimeShorter := func() time.Duration {
		return time.Millisecond
	}
	err := txs.IsDataPrepared(2, haveTimeShorter)
	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestTransactions_IsDataPrepared_NumMissingTxsGreaterThanZeroShouldWork(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool()
	txs := createGoodPreprocessor(dataPool)

	go func() {
		txs.SetRcvdTxChan()
	}()

	err := txs.IsDataPrepared(2, haveTime)
	assert.Nil(t, err)
}

func ExampleSortTransactionsBySenderAndNonce() {
	txs := []*txcache.WrappedTransaction{
		{Tx: &transaction.Transaction{Nonce: 3, SndAddr: []byte("bbbb")}, TxHash: []byte("w")},
		{Tx: &transaction.Transaction{Nonce: 1, SndAddr: []byte("aaaa")}, TxHash: []byte("x")},
		{Tx: &transaction.Transaction{Nonce: 5, SndAddr: []byte("bbbb")}, TxHash: []byte("y")},
		{Tx: &transaction.Transaction{Nonce: 2, SndAddr: []byte("aaaa")}, TxHash: []byte("z")},
		{Tx: &transaction.Transaction{Nonce: 7, SndAddr: []byte("aabb")}, TxHash: []byte("t")},
		{Tx: &transaction.Transaction{Nonce: 6, SndAddr: []byte("aabb")}, TxHash: []byte("a")},
		{Tx: &transaction.Transaction{Nonce: 3, SndAddr: []byte("ffff")}, TxHash: []byte("b")},
		{Tx: &transaction.Transaction{Nonce: 3, SndAddr: []byte("eeee")}, TxHash: []byte("c")},
	}

	SortTransactionsBySenderAndNonce(txs)

	for _, item := range txs {
		fmt.Println(item.Tx.GetNonce(), string(item.Tx.GetSndAddr()), string(item.TxHash))
	}

	// Output:
	// 1 aaaa x
	// 2 aaaa z
	// 6 aabb a
	// 7 aabb t
	// 3 bbbb w
	// 5 bbbb y
	// 3 eeee c
	// 3 ffff b
}

func BenchmarkSortTransactionsByNonceAndSender_WhenReversedNonces(b *testing.B) {
	numTx := 100000
	txs := make([]*txcache.WrappedTransaction, numTx)
	for i := 0; i < numTx; i++ {
		txs[i] = &txcache.WrappedTransaction{
			Tx: &transaction.Transaction{
				Nonce:   uint64(numTx - i),
				SndAddr: []byte(fmt.Sprintf("sender-%d", i)),
			},
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SortTransactionsBySenderAndNonce(txs)
	}
}

func createGoodPreprocessor(dataPool dataRetriever.PoolsHolder) *transactions {
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	return preprocessor
}

func TestTransactionPreprocessor_ProcessTxsToMeShouldUseCorrectSenderAndReceiverShards(t *testing.T) {
	t.Parallel()

	dataPool := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	shardCoordinatorMock := mock.NewMultiShardsCoordinatorMock(3)
	shardCoordinatorMock.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, []byte("0")) {
			return 0
		}
		if bytes.Equal(address, []byte("1")) {
			return 1
		}
		if bytes.Equal(address, []byte("2")) {
			return 2
		}

		return shardCoordinatorMock.SelfId()
	}

	preprocessor, _ := NewTransactionPreprocessor(
		dataPool.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		shardCoordinatorMock,
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	tx := transaction.Transaction{SndAddr: []byte("2"), RcvAddr: []byte("0")}
	txHash, _ := core.CalculateHash(preprocessor.marshalizer, preprocessor.hasher, tx)
	body := block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{txHash},
				SenderShardID:   1,
				ReceiverShardID: 0,
				Type:            block.TxBlock,
			},
		},
	}

	preprocessor.AddTxForCurrentBlock(txHash, &tx, 1, 0)

	_, senderShardID, receiverShardID := preprocessor.GetTxInfoForCurrentBlock(txHash)
	assert.Equal(t, uint32(1), senderShardID)
	assert.Equal(t, uint32(0), receiverShardID)

	_ = preprocessor.ProcessTxsToMe(&body, haveTimeTrue)

	_, senderShardID, receiverShardID = preprocessor.GetTxInfoForCurrentBlock(txHash)
	assert.Equal(t, uint32(2), senderShardID)
	assert.Equal(t, uint32(0), receiverShardID)
}

func TestTransactionsPreprocessor_ProcessMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	tdp := &testscommon.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(id string) (c storage.Cacher) {
					return &testscommon.CacherStub{
						PeekCalled: func(key []byte) (value interface{}, ok bool) {
							if reflect.DeepEqual(key, []byte("tx_hash1")) {
								return &transaction.Transaction{Nonce: 10}, true
							}
							if reflect.DeepEqual(key, []byte("tx_hash2")) {
								return &transaction.Transaction{Nonce: 11}, true
							}
							if reflect.DeepEqual(key, []byte("tx_hash3")) {
								return &transaction.Transaction{Nonce: 12}, true
							}
							return nil, false
						},
					}
				},
			}
		},
	}
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	nbTxsProcessed := 0
	maxBlockSize := 16
	txs, err := NewTransactionPreprocessor(
		tdp.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				nbTxsProcessed++
				return vmcommon.Ok, nil
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		&mock.GasHandlerMock{},
		&mock.BlockTrackerMock{},
		block.TxBlock,
		createMockPubkeyConverter(),
		&mock.BlockSizeComputationStub{
			IsMaxBlockSizeWithoutThrottleReachedCalled: func(mbs int, txs int) bool {
				return mbs+txs > maxBlockSize
			},
		},
		&mock.BalanceComputationStub{},
		&mock.EpochNotifierStub{},
		0,
		&mock.TxTypeHandlerMock{},
	)

	assert.NotNil(t, txs)
	assert.Nil(t, err)

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, []byte("tx_hash1"), []byte("tx_hash2"), []byte("tx_hash3"))

	miniBlock := &block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
		Type:            block.TxBlock,
	}

	f := func() (int, int) {
		if nbTxsProcessed == 0 {
			return 0, 0
		}
		return nbTxsProcessed + 1, nbTxsProcessed * core.AdditionalScrForEachScCallOrSpecialTx
	}
	txsToBeReverted, numTxsProcessed, err := txs.ProcessMiniBlock(miniBlock, haveTimeTrue, f)

	assert.Equal(t, process.ErrMaxBlockSizeReached, err)
	assert.Equal(t, 3, len(txsToBeReverted))
	assert.Equal(t, 3, numTxsProcessed)

	f = func() (int, int) {
		if nbTxsProcessed == 0 {
			return 0, 0
		}
		return nbTxsProcessed, nbTxsProcessed * core.AdditionalScrForEachScCallOrSpecialTx
	}
	txsToBeReverted, numTxsProcessed, err = txs.ProcessMiniBlock(miniBlock, haveTimeTrue, f)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(txsToBeReverted))
	assert.Equal(t, 3, numTxsProcessed)
}
