package preprocess

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func haveTime() time.Duration {
	return 2000 * time.Millisecond
}

func haveTimeTrue() bool {
	return true
}

func haveAdditionalTimeFalse() bool {
	return false
}

func isShardStuckFalse(uint32) bool {
	return false
}
func isMaxBlockSizeReachedFalse(int, int) bool {
	return false
}

func getNumOfCrossInterMbsAndTxsZero() (int, int) {
	return 0, 0
}

func getTotalGasConsumedZero() uint64 {
	return 0
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilPool(t *testing.T) {
	t.Parallel()

	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		nil,
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilUTxDataPool, err)
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilStore(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		nil,
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilUTxStorage, err)
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilHasher(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		nil,
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilMarsalizer(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		nil,
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilTxProce(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilTxProcessor, err)
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilShardCoord(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		nil,
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilAccounts(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilRequestFunc(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		nil,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilGasHandler(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		nil,
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilGasHandler, err)
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(txs))
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilPubkeyConverter(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		nil,
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilBlockSizeComputationHandler(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		nil,
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilBalanceComputationHandler(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		nil,
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilBalanceComputationHandler, err)
}

func TestScrsPreprocessor_NewSmartContractResultPreprocessorNilEpochNotifier(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, err := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.GasHandlerMock{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		nil,
		2,
	)

	assert.Nil(t, txs)
	assert.Equal(t, process.ErrNilEpochNotifier, err)
}

func TestScrsPreProcessor_GetTransactionFromPool(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	txHash := []byte("tx1_hash")
	tx, _ := process.GetTransactionHandlerFromPool(1, 1, txHash, tdp.UnsignedTransactions(), false)
	assert.NotNil(t, txs)
	assert.NotNil(t, tx)
	assert.Equal(t, uint64(10), tx.(*smartContractResult.SmartContractResult).Nonce)
}

func TestScrsPreprocessor_RequestTransactionNothingToRequestAsGeneratedAtProcessing(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	shardCoord := mock.NewMultiShardsCoordinatorMock(3)
	txs, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		shardCoord,
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	shardId := uint32(1)
	txHash1 := []byte("tx_hash1")
	txHash2 := []byte("tx_hash2")
	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash1)
	txHashes = append(txHashes, txHash2)
	mBlk := block.MiniBlock{SenderShardID: shardCoord.SelfId(), ReceiverShardID: shardId, TxHashes: txHashes, Type: block.SmartContractResultBlock}
	body.MiniBlocks = append(body.MiniBlocks, &mBlk)

	txsRequested := txs.RequestBlockTransactions(body)

	assert.Equal(t, 0, txsRequested)
}

func TestScrsPreprocessor_RequestTransactionFromNetwork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	shardCoord := mock.NewMultiShardsCoordinatorMock(3)
	txs, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		shardCoord,
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	shardId := uint32(1)
	txHash1 := []byte("tx_hash1")
	txHash2 := []byte("tx_hash2")
	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash1)
	txHashes = append(txHashes, txHash2)
	mBlk := block.MiniBlock{SenderShardID: shardCoord.SelfId() + 1, ReceiverShardID: shardId, TxHashes: txHashes, Type: block.SmartContractResultBlock}
	body.MiniBlocks = append(body.MiniBlocks, &mBlk)

	txsRequested := txs.RequestBlockTransactions(body)

	assert.Equal(t, 2, txsRequested)
}

func TestScrsPreprocessor_RequestBlockTransactionFromMiniBlockFromNetwork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	shardId := uint32(1)
	txHash1 := []byte("tx_hash1")
	txHash2 := []byte("tx_hash2")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash1)
	txHashes = append(txHashes, txHash2)
	mb := &block.MiniBlock{ReceiverShardID: shardId, TxHashes: txHashes, Type: block.SmartContractResultBlock}

	txsRequested := txs.RequestTransactionsForMiniBlock(mb)

	assert.Equal(t, 2, txsRequested)
}

func TestScrsPreprocessor_ReceivedTransactionShouldEraseRequested(t *testing.T) {
	t.Parallel()

	dataPool := dataRetrieverMock.NewPoolsHolderMock()

	shardedDataStub := &testscommon.ShardedDataStub{
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return &testscommon.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return &smartContractResult.SmartContractResult{}, true
				},
			}
		},
	}

	dataPool.SetUnsignedTransactions(shardedDataStub)

	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewSmartContractResultPreprocessor(
		dataPool.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	//add 3 tx hashes on requested list
	txHash1 := []byte("tx hash 1")
	txHash2 := []byte("tx hash 2")
	txHash3 := []byte("tx hash 3")

	txs.AddScrHashToRequestedList(txHash1)
	txs.AddScrHashToRequestedList(txHash2)
	txs.AddScrHashToRequestedList(txHash3)

	txs.SetMissingScr(3)

	//received txHash2
	txs.receivedSmartContractResult(txHash2, &smartContractResult.SmartContractResult{})

	assert.True(t, txs.IsScrHashRequested(txHash1))
	assert.False(t, txs.IsScrHashRequested(txHash2))
	assert.True(t, txs.IsScrHashRequested(txHash3))
}

func TestScrsPreprocessor_GetAllTxsFromMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := dataRetrieverMock.NewPoolsHolderMock()
	senderShardId := uint32(0)
	destinationShardId := uint32(1)

	txsSlice := []*smartContractResult.SmartContractResult{
		{Nonce: 1},
		{Nonce: 2},
		{Nonce: 3},
	}
	transactionsHashes := make([][]byte, len(txsSlice))

	//add defined transactions to sender-destination cacher
	for idx, tx := range txsSlice {
		transactionsHashes[idx] = computeHash(tx, marshalizer, hasher)

		dataPool.UnsignedTransactions().AddData(
			transactionsHashes[idx],
			tx,
			tx.Size(),
			process.ShardCacherIdentifier(senderShardId, destinationShardId),
		)
	}

	//add some random data
	txRandom := &smartContractResult.SmartContractResult{Nonce: 4}
	dataPool.UnsignedTransactions().AddData(
		computeHash(txRandom, marshalizer, hasher),
		txRandom,
		txRandom.Size(),
		process.ShardCacherIdentifier(3, 4),
	)

	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewSmartContractResultPreprocessor(
		dataPool.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	mb := &block.MiniBlock{
		SenderShardID:   senderShardId,
		ReceiverShardID: destinationShardId,
		TxHashes:        transactionsHashes,
		Type:            block.SmartContractResultBlock,
	}

	txsRetrieved, txHashesRetrieved, err := txs.getAllScrsFromMiniBlock(mb, haveTimeTrue)

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

func TestScrsPreprocessor_RemoveBlockDataFromPoolsNilBlockShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	err := txs.RemoveBlockDataFromPools(nil, tdp.MiniBlocks())

	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilTxBlockBody)
}

func TestScrsPreprocessor_RemoveBlockDataFromPoolsOK(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

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

	err := txs.RemoveBlockDataFromPools(body, tdp.MiniBlocks())

	assert.Nil(t, err)

}

func TestScrsPreprocessor_IsDataPreparedErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}

	txs, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	err := txs.IsDataPrepared(1, haveTime)

	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestScrsPreprocessor_IsDataPrepared(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}

	txs, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	go func() {
		time.Sleep(50 * time.Millisecond)
		txs.chRcvAllScrs <- true
	}()

	err := txs.IsDataPrepared(1, haveTime)

	assert.Nil(t, err)
}

func TestScrsPreprocessor_SaveTxsToStorage(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}

	txs, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

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

	err := txs.SaveTxsToStorage(body)
	assert.Nil(t, err)
}

func TestScrsPreprocessor_SaveTxsToStorageMissingTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	body := &block.Body{}

	txHash := []byte(nil)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
		Type:            block.SmartContractResultBlock,
	}

	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	err := txs.SaveTxsToStorage(body)

	assert.Equal(t, process.ErrMissingTransaction, err)
}

func TestScrsPreprocessor_ProcessBlockTransactionsShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	scrPreproc, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{
			ProcessSmartContractResultCalled: func(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	body := &block.Body{}

	txHash := []byte("txHash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
		Type:            block.SmartContractResultBlock,
	}

	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	scrPreproc.AddScrHashToRequestedList([]byte("txHash"))
	txshardInfo := txShardInfo{0, 0}
	scr := smartContractResult.SmartContractResult{
		Nonce: 1,
		Data:  []byte("tx"),
	}

	scrPreproc.scrForBlock.txHashAndInfo["txHash"] = &txInfo{&scr, &txshardInfo}

	err := scrPreproc.ProcessBlockTransactions(&block.Header{}, body, haveTimeTrue)

	assert.Nil(t, err)
}

func TestScrsPreprocessor_ProcessBlockTransactionsShouldErrMaxGasLimitPerBlockInSelfShardIsReached(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	scrPreproc, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{
			ProcessSmartContractResultCalled: func(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&mock.GasHandlerMock{
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, MaxGasLimitPerBlock + 1, nil
			},
		},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	body := &block.Body{}

	txHash := []byte("txHash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
		Type:            block.SmartContractResultBlock,
	}

	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	scrPreproc.AddScrHashToRequestedList([]byte("txHash"))
	txshardInfo := txShardInfo{0, 0}
	scr := smartContractResult.SmartContractResult{
		Nonce: 1,
		Data:  []byte("tx"),
	}

	scrPreproc.scrForBlock.txHashAndInfo["txHash"] = &txInfo{&scr, &txshardInfo}

	err := scrPreproc.ProcessBlockTransactions(&block.Header{}, body, haveTimeTrue)
	assert.Nil(t, err)

	scrPreproc.EpochConfirmed(2, 0)

	err = scrPreproc.ProcessBlockTransactions(&block.Header{}, body, haveTimeTrue)
	assert.Equal(t, process.ErrMaxGasLimitPerBlockInSelfShardIsReached, err)
}

func TestScrsPreprocessor_ProcessMiniBlock(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &testscommon.ShardedDataStub{
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return &testscommon.CacherStub{
					PeekCalled: func(key []byte) (value interface{}, ok bool) {
						if reflect.DeepEqual(key, []byte("tx1_hash")) {
							return &smartContractResult.SmartContractResult{Nonce: 10}, true
						}
						return nil, false
					},
				}
			},
		}
	}

	requestTransaction := func(shardID uint32, txHashes [][]byte) {}

	scr, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{
			ProcessSmartContractResultCalled: func(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
				return 0, nil
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	txHash := []byte("tx1_hash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
		Type:            block.SmartContractResultBlock,
	}

	_, _, err := scr.ProcessMiniBlock(&miniblock, haveTimeTrue, haveAdditionalTimeFalse, getNumOfCrossInterMbsAndTxsZero, false)

	assert.Nil(t, err)
}

func TestScrsPreprocessor_ProcessMiniBlockWrongTypeMiniblockShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}

	scr, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
	}

	_, _, err := scr.ProcessMiniBlock(&miniblock, haveTimeTrue, haveAdditionalTimeFalse, getNumOfCrossInterMbsAndTxsZero, false)

	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrWrongTypeInMiniBlock)
}

func TestScrsPreprocessor_RestoreBlockDataIntoPools(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	scrstorage := mock.ChainStorerMock{}
	scrstorage.AddStorer(1, &storageStubs.StorerStub{})
	err := scrstorage.Put(1, txHash, txHash)
	assert.Nil(t, err)

	scrstorage.GetAllCalled = func(unitType dataRetriever.UnitType, keys [][]byte) (bytes map[string][]byte, e error) {
		par := make(map[string][]byte)
		tx := smartContractResult.SmartContractResult{}
		par["txHash"], _ = json.Marshal(tx)
		return par, nil
	}
	scrstorage.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		return &storageStubs.StorerStub{
			RemoveCalled: func(key []byte) error {
				return nil
			},
		}
	}

	dataPool := dataRetrieverMock.NewPoolsHolderMock()

	shardedDataStub := &testscommon.ShardedDataStub{}

	dataPool.SetUnsignedTransactions(shardedDataStub)
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}

	scr, _ := NewSmartContractResultPreprocessor(
		dataPool.UnsignedTransactions(),
		&scrstorage,
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	body := &block.Body{}

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
		Type:            block.SmartContractResultBlock,
	}

	body.MiniBlocks = append(body.MiniBlocks, &miniblock)
	miniblockPool := testscommon.NewCacherMock()
	scrRestored, err := scr.RestoreBlockDataIntoPools(body, miniblockPool)

	assert.Equal(t, scrRestored, 1)
	assert.Nil(t, err)
}

func TestScrsPreprocessor_RestoreBlockDataIntoPoolsNilMiniblockPoolShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}

	scr, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	body := &block.Body{}

	miniblockPool := storage.Cacher(nil)

	_, err := scr.RestoreBlockDataIntoPools(body, miniblockPool)

	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilMiniBlockPool)
}

func TestSmartContractResults_CreateBlockStartedShouldEmptyTxHashAndInfo(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}

	scr, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	scr.CreateBlockStarted()
	assert.Equal(t, 0, len(scr.scrForBlock.txHashAndInfo))
}

func TestSmartContractResults_GetAllCurrentUsedTxs(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}

	scrPreproc, _ := NewSmartContractResultPreprocessor(
		tdp.UnsignedTransactions(),
		&mock.ChainStorerMock{},
		&hashingMocks.HasherMock{},
		&mock.MarshalizerMock{},
		&testscommon.TxProcessorMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&stateMock.AccountsStub{},
		requestTransaction,
		&testscommon.GasHandlerStub{},
		feeHandlerMock(),
		createMockPubkeyConverter(),
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
	)

	txshardInfo := txShardInfo{0, 3}
	scr := smartContractResult.SmartContractResult{
		Nonce: 1,
		Data:  []byte("tx"),
	}
	scrPreproc.scrForBlock.txHashAndInfo["txHash"] = &txInfo{&scr, &txshardInfo}

	retMap := scrPreproc.GetAllCurrentUsedTxs()
	assert.NotNil(t, retMap)
}
