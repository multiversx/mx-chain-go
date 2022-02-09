package preprocess

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const MaxGasLimitPerBlock = uint64(100000)

func feeHandlerMock() *mock.FeeHandlerStub {
	return &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
			return 0
		},
		MaxGasLimitPerBlockCalled: func() uint64 {
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
		RemoveDataCalled: func(key []byte, cacheID string) {
		},
	}
}

func initDataPool() *dataRetrieverMock.PoolsHolderStub {
	sdp := &dataRetrieverMock.PoolsHolderStub{
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

func createDefaultTransactionsProcessorArgs() ArgsTransactionPreProcessor {
	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}

	return ArgsTransactionPreProcessor{
		TxDataPool:           tdp.Transactions(),
		Store:                &mock.ChainStorerMock{},
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		TxProcessor:          &testscommon.TxProcessorMock{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(3),
		Accounts:             &stateMock.AccountsStub{},
		OnRequestTransaction: requestTransaction,
		EconomicsFee:         feeHandlerMock(),
		GasHandler:           &mock.GasHandlerMock{},
		BlockTracker:         &mock.BlockTrackerMock{},
		BlockType:            block.TxBlock,
		PubkeyConverter:      createMockPubkeyConverter(),
		BlockSizeComputation: &testscommon.BlockSizeComputationStub{},
		BalanceComputation:   &testscommon.BalanceComputationStub{},
		EpochNotifier:        &epochNotifier.EpochNotifierStub{},
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch: 2,
		FrontRunningProtectionEnableEpoch:           30,
		ScheduledMiniBlocksEnableEpoch:              2,
		TxTypeHandler:                               &testscommon.TxTypeHandlerMock{},
		ScheduledTxsExecutionHandler:                &testscommon.ScheduledTxsExecutionStub{},
	}
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilPool(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.TxDataPool = nil
	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilStore(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.Store = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilTxStorage, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilHasher(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.Hasher = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilMarsalizer(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.Marshalizer = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilTxProce(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.TxProcessor = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilTxProcessor, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilShardCoord(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.ShardCoordinator = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilAccounts(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.Accounts = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilRequestFunc(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.OnRequestTransaction = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilFeeHandler(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.EconomicsFee = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilGasHandler(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.GasHandler = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilGasHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilBlockTracker(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.BlockTracker = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilBlockTracker, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilPubkeyConverter(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.PubkeyConverter = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilBlockSizeComputationHandler(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.BlockSizeComputation = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilBalanceComputationHandler(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.BalanceComputation = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilEpochNotifier(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.EpochNotifier = nil

	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilEpochNotifier, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilTxTypeHandler(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.TxTypeHandler = nil
	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilTxTypeHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorNilScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	args.ScheduledTxsExecutionHandler = nil
	txs, err := NewTransactionPreprocessor(args)
	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilScheduledTxsExecutionHandler, err)
}

func TestTxsPreprocessor_NewTransactionPreprocessorOkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()
	txs, err := NewTransactionPreprocessor(args)
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

	dataPool := dataRetrieverMock.NewPoolsHolderMock()

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

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := dataRetrieverMock.NewPoolsHolderMock()
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

	txsRetrieved, txHashesRetrieved, err := txs.getAllTxsFromMiniBlock(mb, func() bool { return true }, func() bool { return false })

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

	totalGasProvided := uint64(0)
	args := createDefaultTransactionsProcessorArgs()
	args.TxDataPool, _ = dataRetrieverMock.CreateTxPool(2, 0)
	args.TxProcessor = &testscommon.TxProcessorMock{
		ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			return 0, nil
		}}
	args.GasHandler = &mock.GasHandlerMock{
		SetGasProvidedCalled: func(gasProvided uint64, hash []byte) {
			totalGasProvided += gasProvided
		},
		TotalGasProvidedCalled: func() uint64 {
			return totalGasProvided
		},
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return 0, 0, nil
		},
		SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		TotalGasRefundedCalled: func() uint64 {
			return 0
		},
	}

	txs, _ := NewTransactionPreprocessor(args)
	assert.NotNil(t, txs)

	sndShardId := uint32(0)
	dstShardId := uint32(1)
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)

	addedTxs := make([]*transaction.Transaction, 0)
	for i := 0; i < 10; i++ {
		newTx := &transaction.Transaction{GasLimit: uint64(i)}

		txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, newTx)
		args.TxDataPool.AddData(txHash, newTx, newTx.Size(), strCache)

		addedTxs = append(addedTxs, newTx)
	}

	sortedTxsAndHashes, _, _ := txs.computeSortedTxs(sndShardId, dstShardId, MaxGasLimitPerBlock, []byte("randomness"))
	miniBlocks, _, err := txs.createAndProcessMiniBlocksFromMeV1(haveTimeTrue, isShardStuckFalse, isMaxBlockSizeReachedFalse, sortedTxsAndHashes)
	assert.Nil(t, err)

	txHashes := 0
	for _, miniBlock := range miniBlocks {
		txHashes += len(miniBlock.TxHashes)
	}

	assert.Equal(t, len(addedTxs), txHashes)
}

func TestTransactions_CreateAndProcessMiniBlockCrossShardGasLimitAddAllAsNoSCCalls(t *testing.T) {
	t.Parallel()

	totalGasProvided := uint64(0)
	args := createDefaultTransactionsProcessorArgs()
	args.TxProcessor = &testscommon.TxProcessorMock{
		ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			return 0, nil
		}}
	args.GasHandler = &mock.GasHandlerMock{
		SetGasProvidedCalled: func(gasProvided uint64, hash []byte) {
			totalGasProvided += gasProvided
		},
		TotalGasProvidedCalled: func() uint64 {
			return totalGasProvided
		},
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return 0, 0, nil
		},
		SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		TotalGasRefundedCalled: func() uint64 {
			return 0
		},
	}
	args.TxDataPool, _ = dataRetrieverMock.CreateTxPool(2, 0)
	txs, _ := NewTransactionPreprocessor(args)
	assert.NotNil(t, txs)

	sndShardId := uint32(0)
	dstShardId := uint32(1)
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)

	gasLimit := MaxGasLimitPerBlock / uint64(5)

	addedTxs := make([]*transaction.Transaction, 0)
	for i := 0; i < 10; i++ {
		newTx := &transaction.Transaction{GasLimit: gasLimit, GasPrice: uint64(i), RcvAddr: []byte("012345678910")}

		txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, newTx)
		args.TxDataPool.AddData(txHash, newTx, newTx.Size(), strCache)

		addedTxs = append(addedTxs, newTx)
	}

	sortedTxsAndHashes, _, _ := txs.computeSortedTxs(sndShardId, dstShardId, MaxGasLimitPerBlock, []byte("randomness"))
	miniBlocks, _, err := txs.createAndProcessMiniBlocksFromMeV1(haveTimeTrue, isShardStuckFalse, isMaxBlockSizeReachedFalse, sortedTxsAndHashes)
	assert.Nil(t, err)

	txHashes := 0
	for _, miniBlock := range miniBlocks {
		txHashes += len(miniBlock.TxHashes)
	}

	assert.Equal(t, len(addedTxs), txHashes)
}

func TestTransactions_CreateAndProcessMiniBlockCrossShardGasLimitAddOnly5asSCCall(t *testing.T) {
	t.Parallel()

	numTxsToAdd := 5
	gasLimit := MaxGasLimitPerBlock / uint64(numTxsToAdd)

	totalGasProvided := uint64(0)
	args := createDefaultTransactionsProcessorArgs()
	args.TxDataPool, _ = dataRetrieverMock.CreateTxPool(2, 0)
	args.TxProcessor = &testscommon.TxProcessorMock{
		ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			return 0, nil
		}}
	args.GasHandler = &mock.GasHandlerMock{
		SetGasProvidedCalled: func(gasProvided uint64, hash []byte) {
			totalGasProvided += gasProvided
		},
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return gasLimit, gasLimit, nil
		},
		TotalGasProvidedCalled: func() uint64 {
			return totalGasProvided
		},
		SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		GasRefundedCalled: func(hash []byte) uint64 {
			return 0
		},
		RemoveGasProvidedCalled: func(hashes [][]byte) {
			totalGasProvided = 0
		},
		RemoveGasRefundedCalled: func(hashes [][]byte) {
		},
	}

	txs, _ := NewTransactionPreprocessor(args)

	assert.NotNil(t, txs)

	sndShardId := uint32(0)
	dstShardId := uint32(1)
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)

	scAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	for i := 0; i < 10; i++ {
		newTx := &transaction.Transaction{GasLimit: gasLimit, GasPrice: uint64(i), RcvAddr: scAddress}

		txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, newTx)
		args.TxDataPool.AddData(txHash, newTx, newTx.Size(), strCache)
	}

	sortedTxsAndHashes, _, _ := txs.computeSortedTxs(sndShardId, dstShardId, MaxGasLimitPerBlock, []byte("randomness"))
	miniBlocks, _, err := txs.createAndProcessMiniBlocksFromMeV1(haveTimeTrue, isShardStuckFalse, isMaxBlockSizeReachedFalse, sortedTxsAndHashes)
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

func TestTransactions_GetTotalGasConsumedShouldWork(t *testing.T) {
	t.Parallel()

	var gasProvided uint64
	var gasRefunded uint64
	var gasPenalized uint64

	args := createDefaultTransactionsProcessorArgs()
	args.GasHandler = &mock.GasHandlerMock{
		TotalGasProvidedCalled: func() uint64 {
			return gasProvided
		},
		TotalGasRefundedCalled: func() uint64 {
			return gasRefunded
		},
		TotalGasPenalizedCalled: func() uint64 {
			return gasPenalized
		},
	}

	preprocessor, _ := NewTransactionPreprocessor(args)

	gasProvided = 10
	gasRefunded = 2
	gasPenalized = 3
	totalGasConsumed := preprocessor.getTotalGasConsumed()
	assert.Equal(t, gasProvided, totalGasConsumed)

	preprocessor.EpochConfirmed(2, 0)

	totalGasConsumed = preprocessor.getTotalGasConsumed()
	assert.Equal(t, gasProvided-gasRefunded-gasPenalized, totalGasConsumed)

	gasRefunded = 5
	gasPenalized = 6
	totalGasConsumed = preprocessor.getTotalGasConsumed()
	assert.Equal(t, gasProvided, totalGasConsumed)
}

func TestTransactions_UpdateGasConsumedWithGasRefundedAndGasPenalizedShouldWork(t *testing.T) {
	t.Parallel()

	var gasRefunded uint64
	var gasPenalized uint64

	args := createDefaultTransactionsProcessorArgs()
	args.GasHandler = &mock.GasHandlerMock{
		GasRefundedCalled: func(_ []byte) uint64 {
			return gasRefunded
		},
		GasPenalizedCalled: func(_ []byte) uint64 {
			return gasPenalized
		},
	}
	args.OptimizeGasUsedInCrossMiniBlocksEnableEpoch = 2

	preprocessor, _ := NewTransactionPreprocessor(args)
	gasInfo := gasConsumedInfo{
		gasConsumedByMiniBlocksInSenderShard:  0,
		gasConsumedByMiniBlockInReceiverShard: 5,
		totalGasConsumedInSelfShard:           10,
	}

	gasRefunded = 2
	gasPenalized = 1
	preprocessor.updateGasConsumedWithGasRefundedAndGasPenalized([]byte("txHash"), &gasInfo)
	assert.Equal(t, uint64(5), gasInfo.gasConsumedByMiniBlockInReceiverShard)
	assert.Equal(t, uint64(10), gasInfo.totalGasConsumedInSelfShard)

	preprocessor.EpochConfirmed(2, 0)

	gasRefunded = 10
	gasPenalized = 1
	preprocessor.updateGasConsumedWithGasRefundedAndGasPenalized([]byte("txHash"), &gasInfo)
	assert.Equal(t, uint64(5), gasInfo.gasConsumedByMiniBlockInReceiverShard)
	assert.Equal(t, uint64(10), gasInfo.totalGasConsumedInSelfShard)

	gasRefunded = 5
	gasPenalized = 1
	preprocessor.updateGasConsumedWithGasRefundedAndGasPenalized([]byte("txHash"), &gasInfo)
	assert.Equal(t, uint64(5), gasInfo.gasConsumedByMiniBlockInReceiverShard)
	assert.Equal(t, uint64(10), gasInfo.totalGasConsumedInSelfShard)

	gasRefunded = 2
	gasPenalized = 1
	preprocessor.updateGasConsumedWithGasRefundedAndGasPenalized([]byte("txHash"), &gasInfo)
	assert.Equal(t, uint64(2), gasInfo.gasConsumedByMiniBlockInReceiverShard)
	assert.Equal(t, uint64(7), gasInfo.totalGasConsumedInSelfShard)
}

func Example_sortTransactionsBySenderAndNonce() {
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

	sortTransactionsBySenderAndNonceLegacy(txs)

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

func Example_sortTransactionsBySenderAndNonce2() {
	txs := []*txcache.WrappedTransaction{
		{Tx: &transaction.Transaction{Nonce: 2, SndAddr: []byte("bbbb")}, TxHash: []byte("w")},
		{Tx: &transaction.Transaction{Nonce: 1, SndAddr: []byte("aaaa")}, TxHash: []byte("x")},
		{Tx: &transaction.Transaction{Nonce: 1, SndAddr: []byte("bbbb")}, TxHash: []byte("y")},
		{Tx: &transaction.Transaction{Nonce: 2, SndAddr: []byte("aaaa")}, TxHash: []byte("z")},
		{Tx: &transaction.Transaction{Nonce: 7, SndAddr: []byte("aabb")}, TxHash: []byte("t")},
		{Tx: &transaction.Transaction{Nonce: 6, SndAddr: []byte("aabb")}, TxHash: []byte("a")},
		{Tx: &transaction.Transaction{Nonce: 1, SndAddr: []byte("ffff")}, TxHash: []byte("b")},
		{Tx: &transaction.Transaction{Nonce: 3, SndAddr: []byte("eeee")}, TxHash: []byte("c")},
		{Tx: &transaction.Transaction{Nonce: 3, SndAddr: []byte("bbbb")}, TxHash: []byte("c")},
	}

	sortTransactionsBySenderAndNonceLegacy(txs)

	for _, item := range txs {
		fmt.Println(item.Tx.GetNonce(), string(item.Tx.GetSndAddr()), string(item.TxHash))
	}

	// Output:
	// 1 aaaa x
	// 2 aaaa z
	// 6 aabb a
	// 7 aabb t
	// 1 bbbb y
	// 2 bbbb w
	// 3 bbbb c
	// 3 eeee c
	// 1 ffff b
}

func Example_sortTransactionsBySenderAndNonceWithFrontRunningProtection() {
	randomness := "randomness"
	txPreproc := transactions{
		basePreProcess: &basePreProcess{
			hasher: &mock.HasherStub{
				ComputeCalled: func(s string) []byte {
					if s == randomness {
						return []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
					}

					return []byte(s)
				},
			},
		},
	}

	nbSenders := 5

	usedRandomness := txPreproc.hasher.Compute(randomness)
	senders := make([][]byte, 0)
	for i := 0; i < nbSenders; i++ {
		sender := make([]byte, len(usedRandomness))
		copy(sender, usedRandomness)
		sender[len(usedRandomness)-1-i] = 0
		senders = append(senders, sender)
	}

	txs := []*txcache.WrappedTransaction{
		{Tx: &transaction.Transaction{Nonce: 2, SndAddr: senders[2]}, TxHash: []byte("w")},
		{Tx: &transaction.Transaction{Nonce: 1, SndAddr: senders[0]}, TxHash: []byte("x")},
		{Tx: &transaction.Transaction{Nonce: 1, SndAddr: senders[2]}, TxHash: []byte("y")},
		{Tx: &transaction.Transaction{Nonce: 2, SndAddr: senders[0]}, TxHash: []byte("z")},
		{Tx: &transaction.Transaction{Nonce: 7, SndAddr: senders[1]}, TxHash: []byte("t")},
		{Tx: &transaction.Transaction{Nonce: 6, SndAddr: senders[1]}, TxHash: []byte("a")},
		{Tx: &transaction.Transaction{Nonce: 1, SndAddr: senders[4]}, TxHash: []byte("b")},
		{Tx: &transaction.Transaction{Nonce: 3, SndAddr: senders[3]}, TxHash: []byte("c")},
		{Tx: &transaction.Transaction{Nonce: 3, SndAddr: senders[2]}, TxHash: []byte("c")},
	}

	txPreproc.sortTransactionsBySenderAndNonceWithFrontRunningProtection(txs, []byte(randomness))

	for _, item := range txs {
		fmt.Println(item.Tx.GetNonce(), hex.EncodeToString(item.Tx.GetSndAddr()), string(item.TxHash))
	}

	// Output:
	// 1 ffffffffffffffffffffffffffffff00 x
	// 2 ffffffffffffffffffffffffffffff00 z
	// 6 ffffffffffffffffffffffffffff00ff a
	// 7 ffffffffffffffffffffffffffff00ff t
	// 1 ffffffffffffffffffffffffff00ffff y
	// 2 ffffffffffffffffffffffffff00ffff w
	// 3 ffffffffffffffffffffffffff00ffff c
	// 3 ffffffffffffffffffffffff00ffffff c
	// 1 ffffffffffffffffffffff00ffffffff b
}

func BenchmarkSortTransactionsByNonceAndSender_WhenReversedNoncesLegacy(b *testing.B) {
	numTx := 30000
	hasher := sha256.NewSha256()
	txs := make([]*txcache.WrappedTransaction, numTx)
	for i := 0; i < numTx; i++ {
		addr := hasher.Compute(fmt.Sprintf("sender-%d", i))
		txs[i] = &txcache.WrappedTransaction{
			Tx: &transaction.Transaction{
				Nonce:   uint64(numTx - i),
				SndAddr: addr,
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortTransactionsBySenderAndNonceLegacy(txs)
	}
}

func BenchmarkSortTransactionsByNonceAndSender_WhenReversedNoncesWithFrontRunningProtection(b *testing.B) {
	numTx := 30000
	hasher := sha256.NewSha256()
	marshaller := &marshal.GogoProtoMarshalizer{}
	txs := make([]*txcache.WrappedTransaction, numTx)
	for i := 0; i < numTx; i++ {
		addr := hasher.Compute(fmt.Sprintf("sender-%d", i))
		txs[i] = &txcache.WrappedTransaction{
			Tx: &transaction.Transaction{
				Nonce:   uint64(numTx - i),
				SndAddr: addr,
			},
		}
	}

	var frontRunProtection atomic.Flag
	_ = frontRunProtection.SetReturningPrevious()
	txpreproc := &transactions{
		basePreProcess: &basePreProcess{
			hasher:                     hasher,
			marshalizer:                marshaller,
			flagFrontRunningProtection: frontRunProtection,
		},
	}
	numRands := 1000
	randomness := make([][]byte, numRands)
	for i := 0; i < numRands; i++ {
		randomness[i] = hasher.Compute(fmt.Sprintf("%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txpreproc.sortTransactionsBySenderAndNonce(txs, randomness[i%numRands])
	}
}

func createGoodPreprocessor(dataPool dataRetriever.PoolsHolder) *transactions {
	args := createDefaultTransactionsProcessorArgs()
	args.TxDataPool = dataPool.Transactions()
	preprocessor, _ := NewTransactionPreprocessor(args)

	return preprocessor
}

func TestTransactionPreprocessor_ProcessTxsToMeShouldUseCorrectSenderAndReceiverShards(t *testing.T) {
	t.Parallel()

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

	args := createDefaultTransactionsProcessorArgs()
	args.ShardCoordinator = shardCoordinatorMock
	args.TxProcessor = &testscommon.TxProcessorMock{
		ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			return 0, nil
		},
	}
	preprocessor, _ := NewTransactionPreprocessor(args)

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

	_ = preprocessor.ProcessTxsToMe(&block.Header{}, &body, haveTimeTrue)

	_, senderShardID, receiverShardID = preprocessor.GetTxInfoForCurrentBlock(txHash)
	assert.Equal(t, uint32(2), senderShardID)
	assert.Equal(t, uint32(0), receiverShardID)
}

func TestTransactionsPreprocessor_ProcessMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	tdp := &dataRetrieverMock.PoolsHolderStub{
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
	nbTxsProcessed := 0
	maxBlockSize := 16
	args := createDefaultTransactionsProcessorArgs()
	args.TxProcessor = &testscommon.TxProcessorMock{
		ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			nbTxsProcessed++
			return vmcommon.Ok, nil
		},
	}
	args.BlockSizeComputation = &testscommon.BlockSizeComputationStub{
		IsMaxBlockSizeWithoutThrottleReachedCalled: func(mbs int, txs int) bool {
			return mbs+txs > maxBlockSize
		},
	}
	args.TxDataPool = tdp.Transactions()
	txs, err := NewTransactionPreprocessor(args)

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
		return nbTxsProcessed + 1, nbTxsProcessed * common.AdditionalScrForEachScCallOrSpecialTx
	}
	txsToBeReverted, numTxsProcessed, err := txs.ProcessMiniBlock(miniBlock, haveTimeTrue, haveAdditionalTimeFalse, f, false)

	assert.Equal(t, process.ErrMaxBlockSizeReached, err)
	assert.Equal(t, 3, len(txsToBeReverted))
	assert.Equal(t, 3, numTxsProcessed)

	f = func() (int, int) {
		if nbTxsProcessed == 0 {
			return 0, 0
		}
		return nbTxsProcessed, nbTxsProcessed * common.AdditionalScrForEachScCallOrSpecialTx
	}
	txsToBeReverted, numTxsProcessed, err = txs.ProcessMiniBlock(miniBlock, haveTimeTrue, haveAdditionalTimeFalse, f, false)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(txsToBeReverted))
	assert.Equal(t, 3, numTxsProcessed)
}

func TestTransactionsPreprocessor_ProcessMiniBlockShouldErrMaxGasLimitUsedForDestMeTxsIsReached(t *testing.T) {
	t.Parallel()

	tdp := &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(id string) (c storage.Cacher) {
					return &testscommon.CacherStub{
						PeekCalled: func(key []byte) (value interface{}, ok bool) {
							if reflect.DeepEqual(key, []byte("tx_hash1")) {
								return &transaction.Transaction{}, true
							}
							return nil, false
						},
					}
				},
			}
		},
	}

	args := createDefaultTransactionsProcessorArgs()
	args.TxDataPool = tdp.Transactions()
	args.GasHandler = &mock.GasHandlerMock{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return 0, MaxGasLimitPerBlock * maxGasLimitPercentUsedForDestMeTxs / 100, nil
		},
		TotalGasProvidedCalled: func() uint64 {
			return 1
		},
	}

	txs, err := NewTransactionPreprocessor(args)
	require.Nil(t, err)

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, []byte("tx_hash1"))

	miniBlock := &block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
		Type:            block.TxBlock,
	}

	txsToBeReverted, numTxsProcessed, err := txs.ProcessMiniBlock(miniBlock, haveTimeTrue, haveAdditionalTimeFalse, getNumOfCrossInterMbsAndTxsZero, false)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(txsToBeReverted))
	assert.Equal(t, 1, numTxsProcessed)

	txs.EpochConfirmed(2, 0)

	txsToBeReverted, numTxsProcessed, err = txs.ProcessMiniBlock(miniBlock, haveTimeTrue, haveAdditionalTimeFalse, getNumOfCrossInterMbsAndTxsZero, false)

	assert.Equal(t, process.ErrMaxGasLimitUsedForDestMeTxsIsReached, err)
	assert.Equal(t, 1, len(txsToBeReverted))
	assert.Equal(t, 0, numTxsProcessed)
}

func TestTransactionsPreprocessor_ComputeGasProvidedShouldWork(t *testing.T) {
	t.Parallel()

	maxGasLimit := uint64(1500000000)
	txGasLimitInSender := maxGasLimit + 1
	txGasLimitInReceiver := maxGasLimit
	args := createDefaultTransactionsProcessorArgs()
	args.EconomicsFee = &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return maxGasLimit
		},
	}
	args.GasHandler = &mock.GasHandlerMock{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return txGasLimitInSender, txGasLimitInReceiver, nil
		},
	}

	preprocessor, _ := NewTransactionPreprocessor(args)

	tx := transaction.Transaction{}
	txHash := []byte("hash")

	gasInfo := gasConsumedInfo{
		gasConsumedByMiniBlocksInSenderShard:  uint64(0),
		gasConsumedByMiniBlockInReceiverShard: uint64(0),
		totalGasConsumedInSelfShard:           uint64(0),
	}

	gasProvidedByTxInSelfShard, err := preprocessor.computeGasProvided(
		1,
		0,
		&tx,
		txHash,
		&gasInfo,
	)

	assert.Nil(t, err)
	assert.Equal(t, maxGasLimit+1, gasInfo.gasConsumedByMiniBlocksInSenderShard)
	assert.Equal(t, maxGasLimit, gasInfo.gasConsumedByMiniBlockInReceiverShard)
	assert.Equal(t, maxGasLimit, gasProvidedByTxInSelfShard)
	assert.Equal(t, maxGasLimit, gasInfo.totalGasConsumedInSelfShard)
}

func TestTransactionsPreprocessor_SplitMiniBlocksIfNeededShouldWork(t *testing.T) {
	t.Parallel()

	var gasLimitPerMiniBlock uint64
	txGasLimit := uint64(100)

	args := createDefaultTransactionsProcessorArgs()
	args.EconomicsFee = &mock.FeeHandlerStub{
		MaxGasLimitPerMiniBlockForSafeCrossShardCalled: func() uint64 {
			return gasLimitPerMiniBlock
		},
		MaxGasLimitPerTxCalled: func() uint64 {
			return gasLimitPerMiniBlock
		},
	}
	args.GasHandler = &mock.GasHandlerMock{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return txGasLimit, txGasLimit, nil
		},
	}

	preprocessor, _ := NewTransactionPreprocessor(args)
	tx1 := transaction.Transaction{Nonce: 0, GasLimit: txGasLimit}
	tx2 := transaction.Transaction{Nonce: 1, GasLimit: txGasLimit}
	tx3 := transaction.Transaction{Nonce: 2, GasLimit: txGasLimit}
	tx4 := transaction.Transaction{Nonce: 3, GasLimit: txGasLimit}
	tx5 := transaction.Transaction{Nonce: 4, GasLimit: txGasLimit}
	tx6 := transaction.Transaction{Nonce: 5, GasLimit: txGasLimit}
	preprocessor.txsForCurrBlock.txHashAndInfo["hash1"] = &txInfo{tx: &tx1}
	preprocessor.txsForCurrBlock.txHashAndInfo["hash2"] = &txInfo{tx: &tx2}
	preprocessor.txsForCurrBlock.txHashAndInfo["hash3"] = &txInfo{tx: &tx3}
	preprocessor.txsForCurrBlock.txHashAndInfo["hash4"] = &txInfo{tx: &tx4}
	preprocessor.txsForCurrBlock.txHashAndInfo["hash5"] = &txInfo{tx: &tx5}
	preprocessor.txsForCurrBlock.txHashAndInfo["hash6"] = &txInfo{tx: &tx6}

	miniBlocks := make([]*block.MiniBlock, 0)

	mb1 := block.MiniBlock{
		ReceiverShardID: 1,
		TxHashes:        [][]byte{[]byte("hash1"), []byte("hash2")},
	}
	miniBlocks = append(miniBlocks, &mb1)

	mb2 := block.MiniBlock{
		ReceiverShardID: 2,
		TxHashes:        [][]byte{[]byte("hash3"), []byte("hash4"), []byte("hash5"), []byte("hash6"), []byte("hash7")},
	}
	miniBlocks = append(miniBlocks, &mb2)

	mb3 := block.MiniBlock{
		ReceiverShardID: 0,
		TxHashes:        [][]byte{[]byte("hash1"), []byte("hash2")},
	}
	miniBlocks = append(miniBlocks, &mb3)

	gasLimitPerMiniBlock = 300

	splitMiniBlocks := preprocessor.splitMiniBlocksBasedOnMaxGasLimitIfNeeded(miniBlocks)
	assert.Equal(t, 3, len(splitMiniBlocks))

	preprocessor.EpochConfirmed(2, 0)

	splitMiniBlocks = preprocessor.splitMiniBlocksBasedOnMaxGasLimitIfNeeded(miniBlocks)
	assert.Equal(t, 4, len(splitMiniBlocks))

	gasLimitPerMiniBlock = 199
	splitMiniBlocks = preprocessor.splitMiniBlocksBasedOnMaxGasLimitIfNeeded(miniBlocks)
	assert.Equal(t, 7, len(splitMiniBlocks))
}

func TestTransactionsPreProcessor_preFilterTransactionsNoBandwidth(t *testing.T) {
	gasHandler := &mock.GasHandlerMock{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return txHandler.GetGasLimit(), txHandler.GetGasLimit(), nil
		},
	}
	economicsFee := &economicsmocks.EconomicsHandlerStub{
		MinGasLimitCalled: func() uint64 {
			return 10
		},
	}

	txsProcessor := &transactions{
		basePreProcess: &basePreProcess{
			gasTracker: gasTracker{
				shardCoordinator: mock.NewMultiShardsCoordinatorMock(3),
				economicsFee:     economicsFee,
				gasHandler:       gasHandler,
			},
		},
	}

	nbMoveBalance := 2
	nbSCCalls := 2
	sender0 := []byte("sender0")
	sender1 := []byte("sender00")
	moveGasCost := uint64(10)
	scCallGasCost := uint64(1000)
	bandwidth := uint64(10)

	txsMoveSender0 := createWrappedMoveBalanceTxs(nbMoveBalance, 0, 1, sender0, moveGasCost)
	txsSCCallsSender0 := createWrappedTxsWithData(nbSCCalls, 0, 1, sender0, scCallGasCost)
	txsMoveSender1 := createWrappedMoveBalanceTxs(nbMoveBalance, 0, 2, sender1, moveGasCost)
	txsSCCallsSender1 := createWrappedTxsWithData(nbSCCalls, 0, 0, sender1, scCallGasCost)

	txs := make([]*txcache.WrappedTransaction, 0)
	txs = append(txs, txsMoveSender0...)
	txs = append(txs, txsSCCallsSender0...)
	txs = append(txs, txsMoveSender1...)
	txs = append(txs, txsSCCallsSender1...)

	var expectedPreFiltered []*txcache.WrappedTransaction
	expectedPreFiltered = append(expectedPreFiltered, txsMoveSender0...)
	expectedPreFiltered = append(expectedPreFiltered, txsMoveSender1...)

	filteredTxs, _ := txsProcessor.preFilterTransactionsWithMoveBalancePriority(txs, bandwidth)
	require.Len(t, filteredTxs, 4)
	require.Equal(t, expectedPreFiltered, filteredTxs)
}

func TestTransactionsPreProcessor_preFilterTransactionsLimitedBandwidthMultipleTxs(t *testing.T) {
	gasHandler := &mock.GasHandlerMock{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return txHandler.GetGasLimit(), txHandler.GetGasLimit(), nil
		},
	}
	economicsFee := &economicsmocks.EconomicsHandlerStub{
		MinGasLimitCalled: func() uint64 {
			return 10
		},
	}
	txsProcessor := &transactions{
		basePreProcess: &basePreProcess{
			gasTracker: gasTracker{
				shardCoordinator: mock.NewMultiShardsCoordinatorMock(3),
				economicsFee:     economicsFee,
				gasHandler:       gasHandler,
			},
		},
	}

	nbMoveBalance := 2
	nbSCCalls := 2
	sender0 := []byte("sender0")
	sender1 := []byte("sender00")
	moveGasCost := uint64(10)
	scCallGasCost := uint64(1000)
	bandwidth := uint64(1000)

	txsMoveSender0 := createWrappedMoveBalanceTxs(nbMoveBalance, 0, 1, sender0, moveGasCost)
	txsSCCallsSender0 := createWrappedTxsWithData(nbSCCalls, 0, 1, sender0, scCallGasCost)
	txsMoveSender1 := createWrappedMoveBalanceTxs(nbMoveBalance, 0, 2, sender1, moveGasCost)
	txsSCCallsSender1 := createWrappedTxsWithData(nbSCCalls, 0, 0, sender1, scCallGasCost)

	txs := make([]*txcache.WrappedTransaction, 0)
	txs = append(txs, txsMoveSender0...)
	txs = append(txs, txsSCCallsSender0...)
	txs = append(txs, txsMoveSender1...)
	txs = append(txs, txsSCCallsSender1...)

	var expectedPreFiltered []*txcache.WrappedTransaction
	expectedPreFiltered = append(expectedPreFiltered, txsMoveSender0...)
	expectedPreFiltered = append(expectedPreFiltered, txsMoveSender1...)

	filteredTxs, _ := txsProcessor.preFilterTransactionsWithMoveBalancePriority(txs, bandwidth)
	require.Len(t, filteredTxs, 4)
	require.Equal(t, expectedPreFiltered, filteredTxs)

	bandwidth = uint64(2000)
	expectedPreFiltered = append(expectedPreFiltered, txsSCCallsSender0[0])
	filteredTxs, _ = txsProcessor.preFilterTransactionsWithMoveBalancePriority(txs, bandwidth)
	require.Len(t, filteredTxs, 5)
	require.Equal(t, expectedPreFiltered, filteredTxs)

	bandwidth = uint64(4000)
	expectedPreFiltered = append(expectedPreFiltered, txsSCCallsSender0[1], txsSCCallsSender1[0])
	filteredTxs, _ = txsProcessor.preFilterTransactionsWithMoveBalancePriority(txs, bandwidth)
	require.Len(t, filteredTxs, 7)
	require.Equal(t, expectedPreFiltered, filteredTxs)
}

func TestTransactionsPreProcessor_preFilterTransactionsLimitedBandwidthMultipleTxsWithSkippedMoveBalance(t *testing.T) {
	gasHandler := &mock.GasHandlerMock{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return txHandler.GetGasLimit(), txHandler.GetGasLimit(), nil
		},
	}
	economicsFee := &economicsmocks.EconomicsHandlerStub{
		MinGasLimitCalled: func() uint64 {
			return 10
		},
	}
	txsProcessor := &transactions{
		basePreProcess: &basePreProcess{
			gasTracker: gasTracker{
				shardCoordinator: mock.NewMultiShardsCoordinatorMock(3),
				economicsFee:     economicsFee,
				gasHandler:       gasHandler,
			},
		},
	}

	nbMoveBalance := 2
	nbSCCalls := 2
	sender0 := []byte("sender0")
	sender1 := []byte("sender00")
	moveGasCost := uint64(10)
	scCallGasCost := uint64(1000)
	bandwidth := uint64(1000)

	txsMoveSender0 := createWrappedMoveBalanceTxs(nbMoveBalance, 0, 1, sender0, moveGasCost)
	txsMoveBatch2Sender0 := createWrappedMoveBalanceTxs(nbMoveBalance, 0, 1, sender0, moveGasCost)
	txsSCCallsSender0 := createWrappedTxsWithData(nbSCCalls, 0, 1, sender0, scCallGasCost)
	txsMoveSender1 := createWrappedMoveBalanceTxs(nbMoveBalance, 0, 2, sender1, moveGasCost)
	txsSCCallsSender1 := createWrappedTxsWithData(nbSCCalls, 0, 0, sender1, scCallGasCost)

	txs := make([]*txcache.WrappedTransaction, 0)
	txs = append(txs, txsMoveSender0...)
	txs = append(txs, txsSCCallsSender0...)
	txs = append(txs, txsMoveBatch2Sender0...)
	// second sender has sc calls before move balance
	txs = append(txs, txsSCCallsSender1...)
	txs = append(txs, txsMoveSender1...)

	var expectedPreFiltered []*txcache.WrappedTransaction
	expectedPreFiltered = append(expectedPreFiltered, txsMoveSender0...)

	filteredTxs, _ := txsProcessor.preFilterTransactionsWithMoveBalancePriority(txs, bandwidth)
	require.Len(t, filteredTxs, 2)
	require.Equal(t, expectedPreFiltered, filteredTxs)

	bandwidth = uint64(2000)
	expectedPreFiltered = append(expectedPreFiltered, txsSCCallsSender0[0])
	filteredTxs, _ = txsProcessor.preFilterTransactionsWithMoveBalancePriority(txs, bandwidth)
	require.Len(t, filteredTxs, 3)
	require.Equal(t, expectedPreFiltered, filteredTxs)

	bandwidth = uint64(4000)
	expectedPreFiltered = append(
		expectedPreFiltered,
		txsSCCallsSender0[1],
		txsMoveBatch2Sender0[0],
		txsMoveBatch2Sender0[1],
		txsSCCallsSender1[0],
	)
	filteredTxs, _ = txsProcessor.preFilterTransactionsWithMoveBalancePriority(txs, bandwidth)
	require.Len(t, filteredTxs, 7)
	require.Equal(t, expectedPreFiltered, filteredTxs)
}

func TestTransactionsPreProcessor_preFilterTransactionsHighBandwidth(t *testing.T) {
	gasHandler := &mock.GasHandlerMock{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return txHandler.GetGasLimit(), txHandler.GetGasLimit(), nil
		},
	}
	economicsFee := &economicsmocks.EconomicsHandlerStub{
		MinGasLimitCalled: func() uint64 {
			return 10
		},
	}
	txsProcessor := &transactions{
		basePreProcess: &basePreProcess{
			gasTracker: gasTracker{
				shardCoordinator: mock.NewMultiShardsCoordinatorMock(3),
				economicsFee:     economicsFee,
				gasHandler:       gasHandler,
			},
		},
	}

	nbMoveBalance := 2
	nbSCCalls := 2
	sender0 := []byte("sender0")
	sender1 := []byte("sender00")
	moveGasCost := uint64(10)
	scCallGasCost := uint64(1000)
	bandwidth := uint64(10000)

	txsMoveSender0 := createWrappedMoveBalanceTxs(nbMoveBalance, 0, 1, sender0, moveGasCost)
	txsMoveBatch2Sender0 := createWrappedMoveBalanceTxs(nbMoveBalance, 0, 1, sender0, moveGasCost)
	txsSCCallsSender0 := createWrappedTxsWithData(nbSCCalls, 0, 1, sender0, scCallGasCost)
	txsMoveSender1 := createWrappedMoveBalanceTxs(nbMoveBalance, 0, 2, sender1, moveGasCost)
	txsMoveBatch2Sender1 := createWrappedMoveBalanceTxs(nbMoveBalance, 0, 2, sender1, moveGasCost)
	txsSCCallsSender1 := createWrappedTxsWithData(nbSCCalls, 0, 0, sender1, scCallGasCost)

	txs := make([]*txcache.WrappedTransaction, 0)
	txs = append(txs, txsMoveSender0...)
	txs = append(txs, txsSCCallsSender0...)
	txs = append(txs, txsMoveBatch2Sender0...)
	txs = append(txs, txsMoveSender1...)
	txs = append(txs, txsSCCallsSender1...)
	txs = append(txs, txsMoveBatch2Sender1...)

	var expectedPreFiltered []*txcache.WrappedTransaction
	expectedPreFiltered = append(expectedPreFiltered, txsMoveSender0...)
	expectedPreFiltered = append(expectedPreFiltered, txsMoveSender1...)
	expectedPreFiltered = append(expectedPreFiltered, txsSCCallsSender0...)
	expectedPreFiltered = append(expectedPreFiltered, txsMoveBatch2Sender0...)
	expectedPreFiltered = append(expectedPreFiltered, txsSCCallsSender1...)
	expectedPreFiltered = append(expectedPreFiltered, txsMoveBatch2Sender1...)

	filteredTxs, _ := txsProcessor.preFilterTransactionsWithMoveBalancePriority(txs, bandwidth)
	require.Len(t, filteredTxs, 12)
	require.Equal(t, expectedPreFiltered, filteredTxs)
}

func TestTransactionsPreProcessor_getRemainingGasPerBlock(t *testing.T) {
	totalGasProvided := uint64(1000)
	maxGasPerBlock := uint64(100000)
	expectedGasPerBlock := maxGasPerBlock - totalGasProvided
	gasHandler := &mock.GasHandlerMock{
		TotalGasProvidedCalled: func() uint64 {
			return totalGasProvided
		},
	}
	economicsFee := &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(shardID uint32) uint64 {
			return maxGasPerBlock
		},
	}

	txsProcessor := &transactions{
		basePreProcess: &basePreProcess{
			gasTracker: gasTracker{
				shardCoordinator: mock.NewMultiShardsCoordinatorMock(3),
				economicsFee:     economicsFee,
				gasHandler:       gasHandler,
			},
		},
	}

	gasPerBlock := txsProcessor.getRemainingGasPerBlock()
	require.Equal(t, expectedGasPerBlock, gasPerBlock)
}

func TestTransactionsPreProcessor_getRemainingGasPerBlockAsScheduled(t *testing.T) {
	totalGasProvided := uint64(1000)
	maxGasPerBlock := uint64(100000)
	expectedGasPerBlock := maxGasPerBlock - totalGasProvided
	gasHandler := &mock.GasHandlerMock{
		TotalGasProvidedAsScheduledCalled: func() uint64 {
			return totalGasProvided
		},
	}
	economicsFee := &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(shardID uint32) uint64 {
			return maxGasPerBlock
		},
	}

	txsProcessor := &transactions{
		basePreProcess: &basePreProcess{
			gasTracker: gasTracker{
				shardCoordinator: mock.NewMultiShardsCoordinatorMock(3),
				economicsFee:     economicsFee,
				gasHandler:       gasHandler,
			},
		},
	}

	gasPerBlock := txsProcessor.getRemainingGasPerBlockAsScheduled()
	require.Equal(t, expectedGasPerBlock, gasPerBlock)
}

func createWrappedMoveBalanceTxs(nb int, srcShard uint32, rcvShard uint32, sender []byte, gasCost uint64) []*txcache.WrappedTransaction {
	txs := make([]*txcache.WrappedTransaction, nb)

	for i := 0; i < nb; i++ {
		txs[i] = &txcache.WrappedTransaction{
			Tx: &transaction.Transaction{
				Nonce:    uint64(i),
				GasLimit: gasCost,
				SndAddr:  sender,
				RcvAddr:  []byte("receiver"),
				Data:     nil,
			},
			TxHash:          []byte(fmt.Sprintf("transactionHash%d", i)),
			SenderShardID:   srcShard,
			ReceiverShardID: rcvShard,
		}
	}

	return txs
}

func createWrappedTxsWithData(nb int, srcShard uint32, rcvShard uint32, sender []byte, gasCost uint64) []*txcache.WrappedTransaction {
	txs := make([]*txcache.WrappedTransaction, nb)

	for i := 0; i < nb; i++ {
		txs[i] = &txcache.WrappedTransaction{
			Tx: &transaction.Transaction{
				Nonce:    uint64(i),
				GasLimit: gasCost,
				SndAddr:  sender,
				RcvAddr:  []byte("receiver1"),
				Data:     []byte("callfunc@@@"),
			},
			TxHash:          []byte(fmt.Sprintf("transactionHash%d", i)),
			SenderShardID:   srcShard,
			ReceiverShardID: rcvShard,
		}
	}

	return txs
}
