package txcache

import (
	"container/heap"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransactionsHeap_Len(t *testing.T) {
	t.Parallel()

	txHeap := newMinTransactionsHeap(10)
	require.Equal(t, 0, txHeap.Len())

	mockBunchOfTransactions := createBunchesOfTransactionsWithUniformDistribution(2, 5)
	txHeapItem, err := newTransactionsHeapItem(mockBunchOfTransactions[0])
	require.NoError(t, err)

	txHeap.Push(txHeapItem)
	require.Equal(t, txHeap.Len(), 1)

	txHeapItem, err = newTransactionsHeapItem(mockBunchOfTransactions[1])
	txHeap.Push(txHeapItem)
	require.Equal(t, txHeap.Len(), 2)

	_ = txHeap.Pop()
	require.Equal(t, txHeap.Len(), 1)

	_ = txHeap.Pop()
	require.Equal(t, txHeap.Len(), 0)
}

func createMockBunchOfTxsWithSpecificTxHashes(noOfTxsPerBunch int, noOfBunches int) []bunchOfTransactions {
	bunchesOfTxs := make([]bunchOfTransactions, noOfBunches)
	for i := range bunchesOfTxs {
		bunchesOfTxs[i] = make(bunchOfTransactions, 0, noOfTxsPerBunch)
	}
	for i := 0; i < noOfTxsPerBunch*noOfBunches; i++ {
		sender := i % noOfBunches
		tx := createTx([]byte("txHash"+strconv.Itoa(i)), "sender"+strconv.Itoa(sender), uint64(i))
		bunchesOfTxs[sender] = append(bunchesOfTxs[sender], tx)
	}

	return bunchesOfTxs
}

func createMockBunchOfTxsWithSpecificTxHashesAndGasLimit(noOfBunches int, noOfTxsPerBunch int) []bunchOfTransactions {
	bunchesOfTxs := make([]bunchOfTransactions, noOfBunches)
	for i := range bunchesOfTxs {
		bunchesOfTxs[i] = make(bunchOfTransactions, 0, noOfTxsPerBunch)
	}
	for i := 0; i < noOfTxsPerBunch*noOfBunches; i++ {
		sender := i % noOfBunches
		tx := createTx([]byte("txHash"+strconv.Itoa(i)), "sender"+strconv.Itoa(sender), 0).withGasLimit(uint64(i))
		bunchesOfTxs[sender] = append(bunchesOfTxs[sender], tx)
	}

	return bunchesOfTxs
}

func compareSortingResults(t *testing.T, txHeap *transactionsHeap, expectedOrderOfTxs []string) {
	currentExpectedTx := 0
	for txHeap.Len() > 0 {
		tx := heap.Pop(txHeap).(*transactionsHeapItem)
		require.Equal(t, expectedOrderOfTxs[currentExpectedTx], string(tx.currentTransaction.TxHash))
		currentExpectedTx += 1

		if tx.gotoNextTransaction() {
			heap.Push(txHeap, tx)
		}
	}
}

func TestTransactionsHeap_SortingByTxHashShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should work for minHeap", func(t *testing.T) {
		t.Parallel()

		bunchesOfTxs := createMockBunchOfTxsWithSpecificTxHashes(5, 3)

		minHeap := newMinTransactionsHeap(3)
		heap.Init(minHeap)

		for _, bunch := range bunchesOfTxs {
			item, err := newTransactionsHeapItem(bunch)
			require.NoError(t, err)

			heap.Push(minHeap, item)
		}

		expectedOrderOfTxs := []string{
			"txHash2", "txHash5", "txHash8", "txHash11", "txHash14", // bunch 2
			"txHash1", "txHash4", "txHash7", "txHash10", "txHash13", // after all txs from bunch2 are out, bunch1 will win
			"txHash0", "txHash3", "txHash6", "txHash9", "txHash12", //  finally, bunch 0
		}

		compareSortingResults(t, minHeap, expectedOrderOfTxs)
	})

	t.Run("should work for maxHeap", func(t *testing.T) {
		t.Parallel()

		bunchesOfTxs := createMockBunchOfTxsWithSpecificTxHashes(3, 3)

		maxHeap := newMaxTransactionsHeap(3)
		heap.Init(maxHeap)

		for _, bunch := range bunchesOfTxs {
			item, err := newTransactionsHeapItem(bunch)
			require.NoError(t, err)

			heap.Push(maxHeap, item)
		}

		expectedOrderOfTxs := []string{
			"txHash0", "txHash1", "txHash2", "txHash3", "txHash4",
			"txHash5", "txHash6", "txHash7", "txHash8", "txHash9",
		}

		compareSortingResults(t, maxHeap, expectedOrderOfTxs)
	})

}

func TestTransactionsHeap_SortingByGasLimitShouldWork(t *testing.T) {
	t.Parallel()

	bunchesOfTxs := createMockBunchOfTxsWithSpecificTxHashesAndGasLimit(3, 5)

	t.Run("should work for minHeap", func(t *testing.T) {
		t.Parallel()

		minHeap := newMinTransactionsHeap(3)
		heap.Init(minHeap)

		for _, bunch := range bunchesOfTxs {
			item, err := newTransactionsHeapItem(bunch)
			require.NoError(t, err)

			heap.Push(minHeap, item)
		}

		expectedOrderOfTxs := []string{
			"txHash0", "txHash1", "txHash2", "txHash3", "txHash4",
			"txHash5", "txHash6", "txHash7", "txHash8", "txHash9",
			"txHash10", "txHash11", "txHash12", "txHash13", "txHash14",
		}

		compareSortingResults(t, minHeap, expectedOrderOfTxs)
	})

	t.Run("should work for maxHeap", func(t *testing.T) {
		t.Parallel()

		maxHeap := newMaxTransactionsHeap(3)
		heap.Init(maxHeap)

		for _, bunch := range bunchesOfTxs {
			item, err := newTransactionsHeapItem(bunch)
			require.NoError(t, err)

			heap.Push(maxHeap, item)
		}

		expectedOrderOfTxs := []string{
			"txHash2", "txHash5", "txHash8", "txHash11", "txHash14", // bunch 2
			"txHash1", "txHash4", "txHash7", "txHash10", "txHash13", // after all txs from bunch2 are out, bunch1 will win
			"txHash0", "txHash3", "txHash6", "txHash9", "txHash12", //  finally, bunch 0
		}

		compareSortingResults(t, maxHeap, expectedOrderOfTxs)
	})
}

func TestTransactionsHeap_Swap(t *testing.T) {
	t.Parallel()

	bunchOfTx1 := bunchOfTransactions{
		createTx([]byte("txHashA"), "sender1", 0),
	}
	bunchOfTx2 := bunchOfTransactions{
		createTx([]byte("txHashB"), "sender2", 1),
	}

	minHeap := newMaxTransactionsHeap(3)
	heap.Init(minHeap)

	item, err := newTransactionsHeapItem(bunchOfTx1)
	require.NoError(t, err)
	minHeap.Push(item)

	item, err = newTransactionsHeapItem(bunchOfTx2)
	require.NoError(t, err)
	minHeap.Push(item)

	require.Equal(t, true, minHeap.Less(0, 1))
	minHeap.Swap(0, 1)
	require.Equal(t, false, minHeap.Less(0, 1))
}
