package txcache

import (
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func Test_newTxByHashMap(t *testing.T) {
	txByHashmap := newTxByHashMap(1)
	require.NotNil(t, txByHashmap)
}

func addWrappedTxsConcurrently(txByHash *txByHashMap, wrappedTxs []*WrappedTransaction, numberOfTxs int) {
	var wg sync.WaitGroup
	wg.Add(numberOfTxs)

	for _, wrappedTx := range wrappedTxs {
		go func(wrappedTx *WrappedTransaction) {
			defer wg.Done()
			txByHash.addTx(wrappedTx)
		}(wrappedTx)
	}

	wg.Wait()
}

func removeWrappedTxsConcurrently(txByHash *txByHashMap, wrappedTxs []*WrappedTransaction, numberOfTxs int) {
	var wg sync.WaitGroup
	wg.Add(numberOfTxs)

	for _, wrappedTx := range wrappedTxs {
		go func(wrappedTx *WrappedTransaction) {
			defer wg.Done()
			txByHash.removeTx(string(wrappedTx.TxHash))
		}(wrappedTx)
	}

	wg.Wait()
}

func checkExistenceOfTxs(t *testing.T, txByHash *txByHashMap, wrappedTxs []*WrappedTransaction, shouldExist bool) {
	for _, wrappedTx := range wrappedTxs {
		_, found := txByHash.getTx(string(wrappedTx.TxHash))
		require.Equal(t, shouldExist, found)
	}
}

func Test_addTx(t *testing.T) {
	t.Parallel()

	numberOfTxs := 20
	txHashes := createMockTxHashes(numberOfTxs)
	wrappedTxs := createSliceMockWrappedTxs(txHashes)
	txByHash := newTxByHashMap(1)

	checkExistenceOfTxs(t, txByHash, wrappedTxs, false)
	addWrappedTxsConcurrently(txByHash, wrappedTxs, numberOfTxs)
	checkExistenceOfTxs(t, txByHash, wrappedTxs, true)
}

func Test_removeTx(t *testing.T) {
	t.Parallel()

	numberOfTxs := 20
	txHashes := createMockTxHashes(numberOfTxs)
	wrappedTxs := createSliceMockWrappedTxs(txHashes)
	txByHash := newTxByHashMap(1)

	addWrappedTxsConcurrently(txByHash, wrappedTxs, numberOfTxs)
	checkExistenceOfTxs(t, txByHash, wrappedTxs, true)

	removeWrappedTxsConcurrently(txByHash, wrappedTxs, numberOfTxs)
	checkExistenceOfTxs(t, txByHash, wrappedTxs, false)
}

func Test_RemoveTxsBulkWithCheck(t *testing.T) {
	t.Parallel()

	txByHash := newTxByHashMap(2)
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 10)
	txCache.txByHash = txByHash

	accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
		GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
			return 0, big.NewInt(6 * 100000 * oneBillion), true, nil
		},
	}

	txHashes := createMockTxHashes(10)
	wrappedTxs := createSliceMockWrappedTxsWithSameSender(txHashes, "alice")

	for _, tx := range wrappedTxs {
		txCache.AddTx(tx)
	}

	err := txCache.OnProposedBlock(
		[]byte("hash1"),
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: txHashes[0:5],
				},
			},
		},
		&block.Header{
			Nonce:    uint64(0),
			PrevHash: []byte("hash0"),
			RootHash: []byte("rootHash0"),
		},
		accountsProvider,
		holders.NewBlockchainInfo([]byte("hash0"), []byte("hash0"), 0),
	)
	require.Nil(t, err)

	noOfRemovedTxs := txByHash.RemoveTxsBulkWithCheck(txHashes, txCache.tracker)
	require.Equal(t, uint32(5), noOfRemovedTxs)
}

func Test_RemoveTxsBulk(t *testing.T) {
	t.Parallel()

	numberOfTxs := 20
	txHashes := createMockTxHashes(numberOfTxs)
	wrappedTxs := createSliceMockWrappedTxs(txHashes)
	txByHash := newTxByHashMap(1)

	addWrappedTxsConcurrently(txByHash, wrappedTxs, numberOfTxs)
	checkExistenceOfTxs(t, txByHash, wrappedTxs, true)

	removed := txByHash.RemoveTxsBulk(txHashes)

	checkExistenceOfTxs(t, txByHash, wrappedTxs, false)
	require.Equal(t, uint32(len(wrappedTxs)), removed)
}

func Test_clear(t *testing.T) {
	t.Parallel()

	numberOfTxs := 20
	txHashes := createMockTxHashes(numberOfTxs)
	wrappedTxs := createSliceMockWrappedTxs(txHashes)
	txByHash := newTxByHashMap(1)

	addWrappedTxsConcurrently(txByHash, wrappedTxs, numberOfTxs)
	checkExistenceOfTxs(t, txByHash, wrappedTxs, true)

	txByHash.clear()

	checkExistenceOfTxs(t, txByHash, wrappedTxs, false)
}

func Test_forEach(t *testing.T) {
	t.Parallel()

	numberOfTxs := 20
	txHashes := createMockTxHashes(numberOfTxs)
	wrappedTxs := createSliceMockWrappedTxs(txHashes)
	txByHash := newTxByHashMap(1)

	addWrappedTxsConcurrently(txByHash, wrappedTxs, numberOfTxs)
	checkExistenceOfTxs(t, txByHash, wrappedTxs, true)

	var expectedFee *big.Int

	expectedFee = nil
	txByHash.forEach(func(txHash []byte, value *WrappedTransaction) {
		require.Equal(t, value.Fee, expectedFee)
	})

	txByHash.forEach(func(txHash []byte, value *WrappedTransaction) {
		value.Fee = big.NewInt(20)
	})

	expectedFee = big.NewInt(20)
	txByHash.forEach(func(txHash []byte, value *WrappedTransaction) {
		require.Equal(t, value.Fee, expectedFee)
	})
}
