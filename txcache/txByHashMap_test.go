package txcache

import (
	"math/big"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newTxByHashMap(t *testing.T) {
	txByHashmap := newTxByHashMap(1)
	require.NotNil(t, txByHashmap)
}

func createMockWrappedTx(txHash []byte) *WrappedTransaction {
	return &WrappedTransaction{
		Tx:     nil,
		TxHash: txHash,
	}
}

func createSliceMockWrappedTxs(txHashes [][]byte) []*WrappedTransaction {
	wrappedTxs := make([]*WrappedTransaction, len(txHashes))
	for i, txHash := range txHashes {
		wrappedTxs[i] = createMockWrappedTx(txHash)
	}

	return wrappedTxs
}

func createMockTxHashes(numberOfTxs int) [][]byte {
	txHashes := make([][]byte, numberOfTxs)
	for i := 0; i < numberOfTxs; i++ {
		txHashes[i] = []byte("txHash" + strconv.Itoa(i))
	}

	return txHashes
}

func addWrappedTxs(txByHash *txByHashMap, wrappedTxs []*WrappedTransaction, numberOfTxs int) {
	var wg sync.WaitGroup
	wg.Add(numberOfTxs)

	for _, wrappedTx := range wrappedTxs {
		wrappedTxCopy := wrappedTx
		go func() {
			defer wg.Done()
			txByHash.addTx(wrappedTxCopy)
		}()
	}

	wg.Wait()
}

func removeWrappedTxs(txByHash *txByHashMap, wrappedTxs []*WrappedTransaction, numberOfTxs int) {
	var wg sync.WaitGroup
	wg.Add(numberOfTxs)

	for _, wrappedTx := range wrappedTxs {
		wrappedTxCopy := wrappedTx
		go func() {
			defer wg.Done()
			txByHash.removeTx(string(wrappedTxCopy.TxHash))
		}()
	}

	wg.Wait()
}

func checkExistenceOfTxs(t *testing.T, txByHash *txByHashMap, wrappedTxs []*WrappedTransaction, shouldExist bool) {
	for _, wrappedTx := range wrappedTxs {
		_, found := txByHash.getTx(string(wrappedTx.TxHash))
		require.Equal(t, shouldExist, found)
	}
}

func checkTxFee(t *testing.T, txByHash *WrappedTransaction, expectedFee *big.Int) {
	require.Equal(t, txByHash.Fee, expectedFee)
}

func Test_addTx(t *testing.T) {
	t.Parallel()

	numberOfTxs := 20
	txHashes := createMockTxHashes(numberOfTxs)
	wrappedTxs := createSliceMockWrappedTxs(txHashes)
	txByHash := newTxByHashMap(1)

	checkExistenceOfTxs(t, txByHash, wrappedTxs, false)
	addWrappedTxs(txByHash, wrappedTxs, numberOfTxs)
	checkExistenceOfTxs(t, txByHash, wrappedTxs, true)
}

func Test_removeTx(t *testing.T) {
	t.Parallel()

	numberOfTxs := 20
	txHashes := createMockTxHashes(numberOfTxs)
	wrappedTxs := createSliceMockWrappedTxs(txHashes)
	txByHash := newTxByHashMap(1)

	addWrappedTxs(txByHash, wrappedTxs, numberOfTxs)
	checkExistenceOfTxs(t, txByHash, wrappedTxs, true)

	removeWrappedTxs(txByHash, wrappedTxs, numberOfTxs)
	checkExistenceOfTxs(t, txByHash, wrappedTxs, false)

}

func Test_RemoveTxsBulk(t *testing.T) {
	t.Parallel()

	numberOfTxs := 20
	txHashes := createMockTxHashes(numberOfTxs)
	wrappedTxs := createSliceMockWrappedTxs(txHashes)
	txByHash := newTxByHashMap(1)

	addWrappedTxs(txByHash, wrappedTxs, numberOfTxs)
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

	addWrappedTxs(txByHash, wrappedTxs, numberOfTxs)
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

	addWrappedTxs(txByHash, wrappedTxs, numberOfTxs)
	checkExistenceOfTxs(t, txByHash, wrappedTxs, true)

	txByHash.forEach(func(txHash []byte, value *WrappedTransaction) {
		checkTxFee(t, value, nil)
	})

	txByHash.forEach(func(txHash []byte, value *WrappedTransaction) {
		value.Fee = big.NewInt(20)
	})

	txByHash.forEach(func(txHash []byte, value *WrappedTransaction) {
		checkTxFee(t, value, big.NewInt(20))
	})
}
