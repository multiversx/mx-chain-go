package txcache

import (
	"container/heap"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransactionsHeap_Len(t *testing.T) {
	t.Parallel()

	txHeap := newMinTransactionsHeap(10)
	require.Equal(t, 0, txHeap.Len())

	mockBunchOfTxs := createBunchesOfTransactionsWithUniformDistribution(2, 5)
	item, err := newTransactionsHeapItem(mockBunchOfTxs[0])
	require.NoError(t, err)

	txHeap.Push(item)
	require.Equal(t, txHeap.Len(), 1)

	item, err = newTransactionsHeapItem(mockBunchOfTxs[1])
	require.NoError(t, err)
	txHeap.Push(item)
	require.Equal(t, 2, txHeap.Len())

	_ = txHeap.Pop()
	require.Equal(t, 1, txHeap.Len())

	_ = txHeap.Pop()
	require.Equal(t, 0, txHeap.Len())
}

func createMockBunchOfTxsWithSpecificTxHashes(numOfTxsPerBunch int, numOfBunches int) []bunchOfTransactions {
	bunchesOfTxs := make([]bunchOfTransactions, numOfBunches)
	for i := range bunchesOfTxs {
		bunchesOfTxs[i] = make(bunchOfTransactions, 0, numOfTxsPerBunch)
		for j := 0; j < numOfTxsPerBunch; j++ {
			tx := createTx([]byte(fmt.Sprintf("txHash%d", i*numOfTxsPerBunch+j)), fmt.Sprintf("sender%d", i), uint64(i))
			bunchesOfTxs[i] = append(bunchesOfTxs[i], tx)
		}
	}

	return bunchesOfTxs
}

func createMockBunchOfTxsWithSpecificTxHashesAndGasLimit(numOfBunches int, numOfTxsPerBunch int) []bunchOfTransactions {
	bunchesOfTxs := make([]bunchOfTransactions, numOfBunches)
	for i := range bunchesOfTxs {
		bunchesOfTxs[i] = make(bunchOfTransactions, 0, numOfTxsPerBunch)

		for j := 0; j < numOfTxsPerBunch; j++ {
			tx := createTx([]byte(fmt.Sprintf("txHash%d", i*numOfTxsPerBunch+j)), fmt.Sprintf("sender%d", i), 0).withGasLimit(uint64(i))
			bunchesOfTxs[i] = append(bunchesOfTxs[i], tx)
		}
	}

	return bunchesOfTxs
}

func getSortedTxsFromHeap(txHeap *transactionsHeap) []string {
	sortedTxs := make([]string, 0, 15)

	for txHeap.Len() > 0 {
		tx := heap.Pop(txHeap).(*transactionsHeapItem)
		sortedTxs = append(sortedTxs, string(tx.currentTransaction.TxHash))
		if tx.gotoNextTransaction() {
			heap.Push(txHeap, tx)
		}
	}

	return sortedTxs
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
			"txHash5", "txHash6", "txHash7", "txHash8", "txHash9",
			"txHash10", "txHash11", "txHash12", "txHash13", "txHash14",
			"txHash0", "txHash1", "txHash2", "txHash3", "txHash4",
		}

		sortedTxs := getSortedTxsFromHeap(minHeap)
		currentExpectedTx := 0
		for _, tx := range sortedTxs {
			require.Equal(t, expectedOrderOfTxs[currentExpectedTx], tx)
			currentExpectedTx += 1
		}
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

		sortedTxs := getSortedTxsFromHeap(maxHeap)
		currentExpectedTx := 0
		for _, tx := range sortedTxs {
			require.Equal(t, expectedOrderOfTxs[currentExpectedTx], tx)
			currentExpectedTx += 1
		}
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

		sortedTxs := getSortedTxsFromHeap(minHeap)
		currentExpectedTx := 0
		for _, tx := range sortedTxs {
			require.Equal(t, expectedOrderOfTxs[currentExpectedTx], tx)
			currentExpectedTx += 1
		}
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
			"txHash10", "txHash11", "txHash12", "txHash13", "txHash14", // bunch 2
			"txHash5", "txHash6", "txHash7", "txHash8", "txHash9",
			"txHash0", "txHash1", "txHash2", "txHash3", "txHash4",
		}

		sortedTxs := getSortedTxsFromHeap(maxHeap)
		currentExpectedTx := 0
		for _, tx := range sortedTxs {
			require.Equal(t, expectedOrderOfTxs[currentExpectedTx], tx)
			currentExpectedTx += 1
		}
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
