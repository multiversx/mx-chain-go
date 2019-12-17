package preprocess

import (
	"bytes"
	"encoding/hex"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/txpool"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func miniBlocksCompacterMock() *mock.MiniBlocksCompacterMock {
	return &mock.MiniBlocksCompacterMock{
		CompactCalled: func(miniBlocks block.MiniBlockSlice, mpaHashesAndTxs map[string]data.TransactionHandler) block.MiniBlockSlice {
			return miniBlocks
		},
		ExpandCalled: func(miniBlocks block.MiniBlockSlice, mapHashesAntTxs map[string]data.TransactionHandler) (block.MiniBlockSlice, error) {
			return miniBlocks, nil
		},
	}
}

func initPoolsHolder() *mock.PoolsHolderStub {
	sdp := &mock.PoolsHolderStub{
		TransactionsTxPool:         txpool.NewShardedTxPool(storageUnit.CacheConfig{Shards: 16}),
		UnsignedTransactionsTxPool: txpool.NewShardedTxPool(storageUnit.CacheConfig{Shards: 16}),
		RewardTransactionsTxPool:   txpool.NewShardedTxPool(storageUnit.CacheConfig{Shards: 16}),

		HeadersNoncesCalled: func() dataRetriever.Uint64SyncMapCacher {
			return &mock.Uint64SyncMapCacherStub{}
		},
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
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
				RegisterHandlerCalled: func(i func(key []byte)) {},
			}
		},
		MiniBlocksCalled: func() storage.Cacher {
			cs := &mock.CacherStub{}
			cs.RegisterHandlerCalled = func(i func(key []byte)) {
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
			cs.RegisterHandlerCalled = func(i func(key []byte)) {}
			cs.RemoveCalled = func(key []byte) {}
			return cs
		},
		HeadersCalled: func() storage.Cacher {
			cs := &mock.CacherStub{}
			cs.RegisterHandlerCalled = func(i func(key []byte)) {
			}
			return cs
		},
	}
	return sdp
}

// TODO: remove non-test
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
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

// TODO: remove non-test
func TestTxsPreprocessor_NewTransactionPreprocessorNilStore(t *testing.T) {
	t.Parallel()

	tdp := initPoolsHolder()
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
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilTxStorage, err)
}

// TODO: remove non-test
func TestTxsPreprocessor_NewTransactionPreprocessorNilHasher(t *testing.T) {
	t.Parallel()

	tdp := initPoolsHolder()
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
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilHasher, err)
}

// TODO: remove non-test
func TestTxsPreprocessor_NewTransactionPreprocessorNilMarsalizer(t *testing.T) {
	t.Parallel()

	tdp := initPoolsHolder()
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
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

// TODO: remove non-test
func TestTxsPreprocessor_NewTransactionPreprocessorNilTxProce(t *testing.T) {
	t.Parallel()

	tdp := initPoolsHolder()
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
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilTxProcessor, err)
}

// TODO: remove non-test
func TestTxsPreprocessor_NewTransactionPreprocessorNilShardCoord(t *testing.T) {
	t.Parallel()

	tdp := initPoolsHolder()
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
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

// TODO: remove non-test
func TestTxsPreprocessor_NewTransactionPreprocessorNilAccounts(t *testing.T) {
	t.Parallel()

	tdp := initPoolsHolder()
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
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

// TODO: remove non-test
func TestTxsPreprocessor_NewTransactionPreprocessorNilRequestFunc(t *testing.T) {
	t.Parallel()

	tdp := initPoolsHolder()
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
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

// TODO: remove non-test
func TestTxsPreprocessor_NewTransactionPreprocessorNilFeeHandler(t *testing.T) {
	t.Parallel()

	tdp := initPoolsHolder()
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
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

// TODO: remove non-test
func TestTxsPreprocessor_NewTransactionPreprocessorNilMiniBlocksCompacter(t *testing.T) {
	t.Parallel()

	tdp := initPoolsHolder()
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
		&mock.GasHandlerMock{},
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilMiniBlocksCompacter, err)
}

// TODO: remove non-test
func TestTxsPreprocessor_NewTransactionPreprocessorNilGasHandler(t *testing.T) {
	t.Parallel()

	tdp := initPoolsHolder()
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
		miniBlocksCompacterMock(),
		nil,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilGasHandler, err)
}

func TestTxsPreProcessor_GetTransactionFromPool(t *testing.T) {
	t.Parallel()

	poolsHolder := initPoolsHolder()
	poolsHolder.TransactionsTxPool.AddTx([]byte("tx1_hash"), &transaction.Transaction{Nonce: 10}, "1")

	tx, err := process.GetTransactionHandlerFromPool(1, 1, []byte("tx1_hash"), poolsHolder.Transactions())

	require.Nil(t, err)
	require.NotNil(t, tx)
	require.Equal(t, uint64(10), tx.GetNonce())
}

func TestTransactionPreprocessor_RequestTransactionFromNetwork(t *testing.T) {
	t.Parallel()
	poolsHolder := initPoolsHolder()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewTransactionPreprocessor(
		poolsHolder.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	txHashes := [][]byte{[]byte("tx_hash1"), []byte("tx_hash2")}
	miniBlock := block.MiniBlock{ReceiverShardID: 1, TxHashes: txHashes}
	body := block.Body{&miniBlock}
	txsRequested := txs.RequestBlockTransactions(body)

	require.Equal(t, 2, txsRequested)
}

func TestTransactionPreprocessor_RequestBlockTransactionFromMiniBlockFromNetwork(t *testing.T) {
	t.Parallel()
	poolsHolder := initPoolsHolder()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewTransactionPreprocessor(
		poolsHolder.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	txHashes := [][]byte{[]byte("tx_hash1"), []byte("tx_hash2")}
	miniBlock := &block.MiniBlock{ReceiverShardID: 1, TxHashes: txHashes}
	txsRequested := txs.RequestTransactionsForMiniBlock(miniBlock)

	require.Equal(t, 2, txsRequested)
}

func TestTransactionPreprocessor_ReceivedTransactionShouldEraseRequested(t *testing.T) {
	t.Parallel()

	poolsHolder := initPoolsHolder()
	poolsHolder.TransactionsTxPool.AddTx([]byte("tx hash 2"), &transaction.Transaction{Nonce: 10}, "0")

	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewTransactionPreprocessor(
		poolsHolder.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	// Add 3 tx hashes on requested list
	txHash1 := []byte("tx hash 1")
	txHash2 := []byte("tx hash 2")
	txHash3 := []byte("tx hash 3")

	txs.AddTxHashToRequestedList(txHash1)
	txs.AddTxHashToRequestedList(txHash2)
	txs.AddTxHashToRequestedList(txHash3)

	txs.SetMissingTxs(3)

	// Received txHash2
	txs.ReceivedTransaction(txHash2)

	require.True(t, txs.IsTxHashRequested(txHash1))
	require.False(t, txs.IsTxHashRequested(txHash2))
	require.True(t, txs.IsTxHashRequested(txHash3))
}

func replacementForComputeHash(data interface{}) []byte {
	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	return computeHash(data, marshalizer, hasher)
}

// TODO: use replacement only instead
func computeHash(data interface{}, marshalizer marshal.Marshalizer, hasher hashing.Hasher) []byte {
	buff, _ := marshalizer.Marshal(data)
	return hasher.Compute(string(buff))
}

func TestTransactionPreprocessor_GetAllTxsFromMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	poolsHolder := initPoolsHolder()
	senderShardID := uint32(0)
	destinationShardID := uint32(1)

	transactions := []*transaction.Transaction{
		{Nonce: 1},
		{Nonce: 2},
		{Nonce: 3},
	}
	transactionsHashes := make([][]byte, len(transactions))

	// Add defined transactions to sender-destination cache
	for i, tx := range transactions {
		txHash := replacementForComputeHash(tx)
		cacheID := process.ShardCacherIdentifier(senderShardID, destinationShardID)
		poolsHolder.Transactions().AddTx(txHash, tx, cacheID)

		transactionsHashes[i] = txHash
	}

	// Add a random transaction to another cache
	randomTx := &transaction.Transaction{Nonce: 4}
	randomTxHash := replacementForComputeHash(randomTx)
	randomTxCacheID := process.ShardCacherIdentifier(3, 4)
	poolsHolder.Transactions().AddTx(randomTxHash, randomTx, randomTxCacheID)

	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewTransactionPreprocessor(
		poolsHolder.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	miniBlock := &block.MiniBlock{
		SenderShardID:   senderShardID,
		ReceiverShardID: destinationShardID,
		TxHashes:        transactionsHashes,
	}

	txsRetrieved, txHashesRetrieved, err := txs.getAllTxsFromMiniBlock(miniBlock, func() bool { return true })

	require.Nil(t, err)
	require.Equal(t, len(transactions), len(txsRetrieved))
	require.Equal(t, len(transactions), len(txHashesRetrieved))

	for idx, tx := range transactions {
		//txReceived should be all txs in the same order
		require.Equal(t, txsRetrieved[idx], tx)
		//verify corresponding transaction hashes
		require.Equal(t, txHashesRetrieved[idx], replacementForComputeHash(tx))
	}
}

func TestTransactionPreprocessor_RemoveBlockTxsFromPoolNilBlockShouldErr(t *testing.T) {
	t.Parallel()

	poolsHolder := initPoolsHolder()

	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewTransactionPreprocessor(
		poolsHolder.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)

	err := txs.RemoveTxBlockFromPools(nil, poolsHolder.MiniBlocks())

	require.NotNil(t, err)
	require.Equal(t, err, process.ErrNilTxBlockBody)
}

func TestTransactionPreprocessor_RemoveBlockTxsFromPoolOK(t *testing.T) {
	t.Parallel()

	poolsHolder := initPoolsHolder()

	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewTransactionPreprocessor(
		poolsHolder.Transactions(),
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{},
	)
	body := make(block.Body, 0)
	txHash := []byte("txHash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)
	err := txs.RemoveTxBlockFromPools(body, poolsHolder.MiniBlocks())
	require.Nil(t, err)
}

func TestTransactions_CreateAndProcessMiniBlockCrossShardGasLimitAddAll(t *testing.T) {
	t.Parallel()

	poolsHolder := initPoolsHolder()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

	totalGasConsumed := uint64(0)
	txs, _ := NewTransactionPreprocessor(
		poolsHolder.Transactions(),
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		&mock.TxProcessorMock{ProcessTransactionCalled: func(transaction *transaction.Transaction) error {
			return nil
		}},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		miniBlocksCompacterMock(),
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
	)
	assert.NotNil(t, txs)

	senderShardID := uint32(0)
	destinationShardID := uint32(1)
	strCache := process.ShardCacherIdentifier(senderShardID, destinationShardID)

	addedTxs := make([]*transaction.Transaction, 0)
	for i := 0; i < 10; i++ {
		newTx := &transaction.Transaction{GasLimit: uint64(i)}

		txHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		poolsHolder.Transactions().AddTx(txHash, newTx, strCache)

		addedTxs = append(addedTxs, newTx)
	}

	miniBlock, err := txs.CreateAndProcessMiniBlock(senderShardID, destinationShardID, process.MaxItemsInBlock, haveTimeTrue)
	require.Nil(t, err)
	require.Equal(t, len(addedTxs), len(miniBlock.TxHashes))
}

func TestTransactions_CreateAndProcessMiniBlockCrossShardGasLimitAddAllAsNoSCCalls(t *testing.T) {
	t.Parallel()

	txPool := txpool.NewShardedTxPool(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

	totalGasConsumed := uint64(0)
	txs, _ := NewTransactionPreprocessor(
		txPool,
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		&mock.TxProcessorMock{ProcessTransactionCalled: func(transaction *transaction.Transaction) error {
			return nil
		}},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		miniBlocksCompacterMock(),
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
		txPool.AddTx(txHash, newTx, strCache)

		addedTxs = append(addedTxs, newTx)
	}

	mb, err := txs.CreateAndProcessMiniBlock(sndShardId, dstShardId, process.MaxItemsInBlock, haveTimeTrue)
	assert.Nil(t, err)

	assert.Equal(t, len(addedTxs), len(mb.TxHashes))
}

func TestTransactions_CreateAndProcessMiniBlockCrossShardGasLimitAddOnly5asSCCall(t *testing.T) {
	t.Parallel()

	txPool := txpool.NewShardedTxPool(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})
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
		&mock.TxProcessorMock{ProcessTransactionCalled: func(transaction *transaction.Transaction) error {
			return nil
		}},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		miniBlocksCompacterMock(),
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
	)
	assert.NotNil(t, txs)

	sndShardId := uint32(0)
	dstShardId := uint32(1)
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)

	scAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	addedTxs := make([]*transaction.Transaction, 0)
	for i := 0; i < 10; i++ {
		newTx := &transaction.Transaction{GasLimit: gasLimit, GasPrice: uint64(i), RcvAddr: scAddress}

		txHash, _ := core.CalculateHash(marshalizer, hasher, newTx)
		txPool.AddTx(txHash, newTx, strCache)

		addedTxs = append(addedTxs, newTx)
	}

	mb, err := txs.CreateAndProcessMiniBlock(sndShardId, dstShardId, process.MaxItemsInBlock, haveTimeTrue)
	assert.Nil(t, err)

	assert.Equal(t, numTxsToAdd, len(mb.TxHashes))
}

var r *rand.Rand
var mutex sync.Mutex

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func createRandTx(rand *rand.Rand) ([]byte, *transaction.Transaction) {
	mutex.Lock()
	nonce := rand.Uint64()
	mutex.Unlock()
	tx := &transaction.Transaction{
		Nonce: nonce,
	}
	marshalizer := &mock.MarshalizerMock{}
	buffTx, _ := marshalizer.Marshal(tx)
	hash := mock.HasherMock{}.Compute(string(buffTx))
	return hash, tx
}

func hashInSlice(hash []byte, hashes [][]byte) bool {
	for _, h := range hashes {
		if bytes.Equal(h, hash) {
			return true
		}
	}
	return false
}

func txInSlice(tx *transaction.Transaction, transactions []*transaction.Transaction) bool {
	for _, t := range transactions {
		if reflect.DeepEqual(tx, t) {
			return true
		}
	}
	return false
}

func TestMiniBlocksCompaction_CompactAndExpandMiniBlocksShouldResultTheSameMiniBlocks(t *testing.T) {
	t.Parallel()

	totalGasConsumed := uint64(0)
	txPool := txpool.NewShardedTxPool(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewTransactionPreprocessor(
		txPool,
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) error {
				return nil
			},
		},
		mock.NewMultiShardsCoordinatorMock(2),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		miniBlocksCompacterMock(),
		&mock.GasHandlerMock{
			InitCalled: func() {
				totalGasConsumed = 0
			},
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
			GasRefundedCalled: func(hash []byte) uint64 {
				return 0
			},
		},
	)

	keygen := signing.NewKeyGenerator(kyber.NewBlakeSHA256Ed25519())
	_, accPk := keygen.GeneratePair()
	pkBytes, _ := accPk.ToByteArray()

	strCache00 := process.ShardCacherIdentifier(0, 0)
	strCache01 := process.ShardCacherIdentifier(0, 1)

	txHashesInMb1 := [][]byte{[]byte("tx00"), []byte("tx01"), []byte("tx02")}
	txHashesInMb2 := [][]byte{[]byte("tx10"), []byte("tx11"), []byte("tx12")}
	txHashesInMb3 := [][]byte{[]byte("tx20"), []byte("tx21"), []byte("tx22")}
	txHashesInMb4 := [][]byte{[]byte("tx30"), []byte("tx31"), []byte("tx32")}

	mapHashesAndTxs := map[string]data.TransactionHandler{
		string(txHashesInMb1[0]): &transaction.Transaction{Nonce: 0, SndAddr: pkBytes},
		string(txHashesInMb1[1]): &transaction.Transaction{Nonce: 1, SndAddr: pkBytes},
		string(txHashesInMb1[2]): &transaction.Transaction{Nonce: 2, SndAddr: pkBytes},
		string(txHashesInMb2[0]): &transaction.Transaction{Nonce: 3, SndAddr: pkBytes},
		string(txHashesInMb2[1]): &transaction.Transaction{Nonce: 4, SndAddr: pkBytes},
		string(txHashesInMb2[2]): &transaction.Transaction{Nonce: 5, SndAddr: pkBytes},
		string(txHashesInMb3[0]): &transaction.Transaction{Nonce: 6, SndAddr: pkBytes},
		string(txHashesInMb3[1]): &transaction.Transaction{Nonce: 7, SndAddr: pkBytes},
		string(txHashesInMb3[2]): &transaction.Transaction{Nonce: 8, SndAddr: pkBytes},
		string(txHashesInMb4[0]): &transaction.Transaction{Nonce: 9, SndAddr: pkBytes},
		string(txHashesInMb4[1]): &transaction.Transaction{Nonce: 10, SndAddr: pkBytes},
		string(txHashesInMb4[2]): &transaction.Transaction{Nonce: 11, SndAddr: pkBytes},
	}

	txPool.AddTx(txHashesInMb1[0], mapHashesAndTxs[string(txHashesInMb1[0])], strCache00)
	txPool.AddTx(txHashesInMb1[1], mapHashesAndTxs[string(txHashesInMb1[1])], strCache00)
	txPool.AddTx(txHashesInMb1[2], mapHashesAndTxs[string(txHashesInMb1[2])], strCache00)
	mb1 := block.MiniBlock{
		TxHashes:        txHashesInMb1,
		ReceiverShardID: 0,
		SenderShardID:   0,
		Type:            0,
	}

	txPool.AddTx(txHashesInMb2[0], mapHashesAndTxs[string(txHashesInMb2[0])], strCache01)
	txPool.AddTx(txHashesInMb2[1], mapHashesAndTxs[string(txHashesInMb2[1])], strCache01)
	txPool.AddTx(txHashesInMb2[2], mapHashesAndTxs[string(txHashesInMb2[2])], strCache01)
	mb2 := block.MiniBlock{
		TxHashes:        txHashesInMb2,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            0,
	}

	txPool.AddTx(txHashesInMb3[0], mapHashesAndTxs[string(txHashesInMb3[0])], strCache00)
	txPool.AddTx(txHashesInMb3[1], mapHashesAndTxs[string(txHashesInMb3[1])], strCache00)
	txPool.AddTx(txHashesInMb3[2], mapHashesAndTxs[string(txHashesInMb3[2])], strCache00)
	mb3 := block.MiniBlock{
		TxHashes:        txHashesInMb3,
		ReceiverShardID: 0,
		SenderShardID:   0,
		Type:            0,
	}

	txPool.AddTx(txHashesInMb4[0], mapHashesAndTxs[string(txHashesInMb4[0])], strCache01)
	txPool.AddTx(txHashesInMb4[1], mapHashesAndTxs[string(txHashesInMb4[1])], strCache01)
	txPool.AddTx(txHashesInMb4[2], mapHashesAndTxs[string(txHashesInMb4[2])], strCache01)
	mb4 := block.MiniBlock{
		TxHashes:        txHashesInMb4,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            0,
	}

	txs.gasHandler.Init()
	_ = txs.ProcessMiniBlock(&mb1, haveTimeTrue)
	txs.gasHandler.Init()
	_ = txs.ProcessMiniBlock(&mb2, haveTimeTrue)
	txs.gasHandler.Init()
	_ = txs.ProcessMiniBlock(&mb3, haveTimeTrue)
	txs.gasHandler.Init()
	_ = txs.ProcessMiniBlock(&mb4, haveTimeTrue)

	mbsOrig := block.MiniBlockSlice{}
	mbsOrig = append(mbsOrig, &mb1, &mb2, &mb3, &mb4)

	mbsValues := make([]block.MiniBlock, 0)
	for _, mb := range mbsOrig {
		mbsValues = append(mbsValues, *mb)
	}

	compactedMbs := txs.miniBlocksCompacter.Compact(mbsOrig, mapHashesAndTxs)
	expandedMbs, err := txs.miniBlocksCompacter.Expand(compactedMbs, mapHashesAndTxs)
	assert.Nil(t, err)

	assert.Equal(t, len(mbsValues), len(expandedMbs))
	for i := 0; i < len(mbsValues); i++ {
		assert.True(t, reflect.DeepEqual(mbsValues[i], *expandedMbs[i]))
	}
}
