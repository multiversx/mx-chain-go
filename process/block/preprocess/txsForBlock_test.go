package preprocess

import (
	"strconv"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func Test_NewTxsForBlock(t *testing.T) {
	t.Parallel()

	t.Run("nil shard coordinator should return error", func(t *testing.T) {
		_, err := NewTxsForBlock(nil)
		require.Equal(t, process.ErrNilShardCoordinator, err)
	})
	t.Run("valid shard coordinator OK", func(t *testing.T) {
		shardCoordinator := &mock.ShardCoordinatorMock{}
		tfb, err := NewTxsForBlock(shardCoordinator)
		require.NoError(t, err)
		require.NotNil(t, tfb)
	})
}

func TestTxsForBlock_Reset(t *testing.T) {
	t.Parallel()

	shardCoordinator := &mock.ShardCoordinatorMock{}
	tfb, _ := NewTxsForBlock(shardCoordinator)

	tfb.missingTxs = 5
	tfb.txHashAndInfo["hash1"] = &txInfo{}
	tfb.Reset()

	require.Equal(t, 0, tfb.missingTxs)
	require.Empty(t, tfb.txHashAndInfo)
}

func TestTxsForBlock_GetTxInfoByHash(t *testing.T) {
	t.Parallel()

	shardCoordinator := &mock.ShardCoordinatorMock{}
	tfb, _ := NewTxsForBlock(shardCoordinator)

	txHash := []byte("hash1")
	tInfo := &txInfo{}
	tfb.txHashAndInfo[string(txHash)] = tInfo

	result, ok := tfb.GetTxInfoByHash(txHash)
	require.True(t, ok)
	require.Equal(t, tInfo, result)

	result, ok = tfb.GetTxInfoByHash([]byte("nonexistent"))
	require.False(t, ok)
	require.Nil(t, result)
}

func TestTxsForBlock_ReceivedTransaction(t *testing.T) {
	t.Parallel()

	shardCoordinator := &mock.ShardCoordinatorMock{}

	t.Run("receive last missing transaction", func(t *testing.T) {
		t.Parallel()

		tfb, _ := NewTxsForBlock(shardCoordinator)
		txHash := []byte("hash1")
		tInfo := &txInfo{txShardInfo: &txShardInfo{}}
		tfb.txHashAndInfo[string(txHash)] = tInfo
		tfb.missingTxs = 1

		tx := &transaction.Transaction{
			Nonce: 1,
		}
		tfb.ReceivedTransaction(txHash, tx)
		require.True(t, <-tfb.chRcvAllTxs)
		require.Equal(t, tx, tfb.txHashAndInfo[string(txHash)].tx)
		require.Equal(t, 0, tfb.missingTxs)
	})
	t.Run("receive transaction when nothing missing", func(t *testing.T) {
		t.Parallel()

		tfb, _ := NewTxsForBlock(shardCoordinator)
		txHash := []byte("hash1")
		tInfo := &txInfo{txShardInfo: &txShardInfo{}}
		tfb.txHashAndInfo[string(txHash)] = tInfo
		tfb.missingTxs = 0

		tx := &transaction.Transaction{
			Nonce: 1,
		}
		tfb.ReceivedTransaction(txHash, tx)

		require.Equal(t, tInfo, tfb.txHashAndInfo[string(txHash)])
		require.Equal(t, 0, tfb.missingTxs)
	})
	t.Run("receive one of multiple missing transactions", func(t *testing.T) {
		t.Parallel()

		tfb, _ := NewTxsForBlock(shardCoordinator)
		txHash1 := []byte("hash1")
		txHash2 := []byte("hash2")
		tInfo1 := &txInfo{txShardInfo: &txShardInfo{}}
		tInfo2 := &txInfo{txShardInfo: &txShardInfo{}}
		tfb.txHashAndInfo[string(txHash1)] = tInfo1
		tfb.txHashAndInfo[string(txHash2)] = tInfo2
		tfb.missingTxs = 2

		tx := &transaction.Transaction{
			Nonce: 1,
		}
		tfb.ReceivedTransaction(txHash1, tx)

		require.Equal(t, tx, tfb.txHashAndInfo[string(txHash1)].tx)
		require.Equal(t, 1, tfb.missingTxs)
	})
}

func TestTxsForBlock_AddTransaction(t *testing.T) {
	t.Parallel()

	t.Run("nil transaction should not be added", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := &mock.ShardCoordinatorMock{}
		tfb, _ := NewTxsForBlock(shardCoordinator)

		tfb.AddTransaction([]byte("hash1"), nil, 1, 2)

		require.Empty(t, tfb.txHashAndInfo)
	})
	t.Run("valid transaction should be added", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := &mock.ShardCoordinatorMock{}
		tfb, _ := NewTxsForBlock(shardCoordinator)

		txHash := []byte("hash1")
		tx := &transaction.Transaction{
			Nonce: 1,
		}
		tfb.AddTransaction(txHash, tx, 1, 2)

		tInfo, ok := tfb.txHashAndInfo[string(txHash)]
		require.True(t, ok)
		require.Equal(t, tx, tInfo.tx)
		require.Equal(t, uint32(1), tInfo.senderShardID)
		require.Equal(t, uint32(2), tInfo.receiverShardID)
	})
}

func TestTxsForBlock_HasMissingTransactions(t *testing.T) {
	t.Parallel()

	shardCoordinator := &mock.ShardCoordinatorMock{}
	tfb, _ := NewTxsForBlock(shardCoordinator)

	require.False(t, tfb.HasMissingTransactions())

	tfb.missingTxs = 1
	require.True(t, tfb.HasMissingTransactions())

	tfb.missingTxs = 10
	require.True(t, tfb.HasMissingTransactions())
}

func TestTxsForBlock_ComputeExistingAndRequestMissing(t *testing.T) {
	t.Parallel()

	t.Run("nil body should return 0 missing transactions", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := &mock.ShardCoordinatorMock{}
		tfb, _ := NewTxsForBlock(shardCoordinator)

		missingTxs := tfb.ComputeExistingAndRequestMissing(nil, func(block.Type) bool { return true }, nil, nil)
		require.Equal(t, 0, missingTxs)
	})
	t.Run("no missing transactions, as Smart contract result in pool", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := &mock.ShardCoordinatorMock{}
		tfb, _ := NewTxsForBlock(shardCoordinator)

		body := createBlockBody(block.SmartContractResultBlock, 1)

		txPool := testscommon.NewShardedDataCacheNotifierMock()
		txPool.AddData(body.MiniBlocks[0].TxHashes[0], &transaction.Transaction{Nonce: 1}, 100, "0")

		onRequestTxs := func(shardID uint32, txHashes [][]byte) {
			require.Fail(t, "should not request transactions when none are missing")
		}

		missingTxs := tfb.ComputeExistingAndRequestMissing(body, func(block.Type) bool { return true }, txPool, onRequestTxs)
		require.Equal(t, 0, missingTxs)
	})
	t.Run("no missing transactions, as invalid transaction in pool", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := &mock.ShardCoordinatorMock{}
		tfb, _ := NewTxsForBlock(shardCoordinator)

		body := createBlockBody(block.InvalidBlock, 1)

		txPool := testscommon.NewShardedDataCacheNotifierMock()
		txPool.AddData(body.MiniBlocks[0].TxHashes[0], &transaction.Transaction{Nonce: 1}, 100, "0")

		onRequestTxs := func(shardID uint32, txHashes [][]byte) {
			require.Fail(t, "should not request transactions when none are missing")
		}

		missingTxs := tfb.ComputeExistingAndRequestMissing(body, func(block.Type) bool { return true }, txPool, onRequestTxs)
		require.Equal(t, 0, missingTxs)
	})
	t.Run("no missing transactions, as no transactions in block", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := &mock.ShardCoordinatorMock{}
		tfb, _ := NewTxsForBlock(shardCoordinator)

		body := createBlockBody(block.SmartContractResultBlock, 0)

		txPool := testscommon.NewShardedDataCacheNotifierMock()
		onRequestTxs := func(shardID uint32, txHashes [][]byte) {
			require.Fail(t, "should not request transactions when none are missing")
		}

		missingTxs := tfb.ComputeExistingAndRequestMissing(body, func(block.Type) bool { return true }, txPool, onRequestTxs)
		require.Equal(t, 0, missingTxs)
	})
	t.Run("missing transactions should be requested", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := &mock.ShardCoordinatorMock{}
		tfb, _ := NewTxsForBlock(shardCoordinator)
		body := createBlockBody(block.TxBlock, 1)
		bodyWithDuplicatedTxs := createBlockBody(block.TxBlock, 1)
		bodyWithDuplicatedTxs.MiniBlocks[0].TxHashes = append(bodyWithDuplicatedTxs.MiniBlocks[0].TxHashes, body.MiniBlocks[0].TxHashes[0])

		txPool := testscommon.NewShardedDataCacheNotifierMock()
		onRequestTxs := func(shardID uint32, txHashes [][]byte) {
			require.Equal(t, uint32(0), shardID)
			require.Equal(t, body.MiniBlocks[0].TxHashes, txHashes)
		}

		missingTxs := tfb.ComputeExistingAndRequestMissing(bodyWithDuplicatedTxs, func(block.Type) bool { return true }, txPool, onRequestTxs)
		require.Equal(t, 1, missingTxs)
		require.Equal(t, 1, tfb.missingTxs)
	})
}

func TestTxsForBlock_WaitForRequestedData(t *testing.T) {
	t.Parallel()

	t.Run("no missing transaction should immediately return", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := &mock.ShardCoordinatorMock{}
		tfb, _ := NewTxsForBlock(shardCoordinator)

		tfb.missingTxs = 0
		err := tfb.WaitForRequestedData(100 * time.Millisecond)
		require.NoError(t, err)
	})

	t.Run("wait for receiving a missing transaction", func(t *testing.T) {
		shardCoordinator := &mock.ShardCoordinatorMock{}
		tfb, _ := NewTxsForBlock(shardCoordinator)

		go func() {
			time.Sleep(100 * time.Millisecond)
			tfb.chRcvAllTxs <- true
		}()

		tfb.missingTxs = 1
		err := tfb.WaitForRequestedData(200 * time.Millisecond)
		require.NoError(t, err)
	})
	t.Run("timeout while waiting for missing transaction", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := &mock.ShardCoordinatorMock{}
		tfb, _ := NewTxsForBlock(shardCoordinator)

		tfb.missingTxs = 1
		err := tfb.WaitForRequestedData(100 * time.Millisecond)
		require.Equal(t, process.ErrTimeIsOut, err)
	})
}

func TestTxsForBlock_GetMissingTxsCount(t *testing.T) {
	t.Parallel()

	shardCoordinator := &mock.ShardCoordinatorMock{}
	tfb, _ := NewTxsForBlock(shardCoordinator)

	require.Equal(t, 0, tfb.GetMissingTxsCount())

	tfb.missingTxs = 5
	require.Equal(t, 5, tfb.GetMissingTxsCount())

	tfb.missingTxs = 10
	require.Equal(t, 10, tfb.GetMissingTxsCount())
}

func TestTxsForBlock_GetAllCurrentUsedTxs(t *testing.T) {
	t.Parallel()

	shardCoordinator := &mock.ShardCoordinatorMock{}
	tfb, _ := NewTxsForBlock(shardCoordinator)

	txHash1 := []byte("hash1")
	txHash2 := []byte("hash2")
	tfb.txHashAndInfo[string(txHash1)] = &txInfo{tx: &transaction.Transaction{Nonce: 1}}
	tfb.txHashAndInfo[string(txHash2)] = &txInfo{tx: &transaction.Transaction{Nonce: 2}}

	allTxs := tfb.GetAllCurrentUsedTxs()
	require.Len(t, allTxs, 2)
	require.Equal(t, tfb.txHashAndInfo[string(txHash1)].tx, allTxs[string(txHash1)])
	require.Equal(t, tfb.txHashAndInfo[string(txHash2)].tx, allTxs[string(txHash2)])
}

func TestTxsForBlock_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	shardCoordinator := &mock.ShardCoordinatorMock{}
	tfb, _ := NewTxsForBlock(shardCoordinator)

	require.False(t, tfb.IsInterfaceNil())

	var nilTfb *txsForBlock
	require.True(t, nilTfb.IsInterfaceNil())
}

func createBlockBody(blockType block.Type, numTxHashes uint16) *block.Body {
	txHashes := make([][]byte, 0, numTxHashes)
	for i := uint16(0); i < numTxHashes; i++ {
		txHash := []byte("hash" + strconv.Itoa(int(i)))
		txHashes = append(txHashes, txHash)
	}
	return &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				Type:            blockType,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        txHashes,
			},
		},
	}
}
