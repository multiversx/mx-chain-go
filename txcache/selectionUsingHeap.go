package txcache

import "container/heap"

func (cache *TxCache) selectTransactionsUsingHeap(gasRequested uint64) BunchOfTransactions {
	senders := cache.getSenders()
	bunches := make([]BunchOfTransactions, 0, len(senders))

	for _, sender := range senders {
		bunches = append(bunches, sender.getTxsWithoutGaps())
	}

	return selectTransactionsFromBunchesUsingHeap(bunches, gasRequested)
}

func selectTransactionsFromBunchesUsingHeap(bunches []BunchOfTransactions, gasRequested uint64) BunchOfTransactions {
	selectedTransactions := make(BunchOfTransactions, 0, 30000)

	// Items popped from the heap are added to "selectedTransactions".
	transactionsHeap := &TransactionHeap{}
	heap.Init(transactionsHeap)

	// Initialize the heap with the first transaction of each bunch
	for i, bunch := range bunches {
		if len(bunch) == 0 {
			// Some senders may have no eligible transactions (initial gaps).
			continue
		}

		heap.Push(transactionsHeap, &HeapItem{
			bunchIndex:       i,
			transactionIndex: 0,
			transaction:      bunch[0],
		})
	}

	accumulatedGas := uint64(0)

	// Select transactions (sorted).
	for transactionsHeap.Len() > 0 {
		// Always pick the best transaction.
		item := heap.Pop(transactionsHeap).(*HeapItem)

		accumulatedGas += item.transaction.Tx.GetGasLimit()
		if accumulatedGas > gasRequested {
			break
		}

		selectedTransactions = append(selectedTransactions, item.transaction)

		// If there are more transactions in the same bunch (same sender as the popped item),
		// add the next one to the heap (to compete with the others).
		item.transactionIndex++

		if item.transactionIndex < len(bunches[item.bunchIndex]) {
			item.transaction = bunches[item.bunchIndex][item.transactionIndex]
			heap.Push(transactionsHeap, item)
		}
	}

	return selectedTransactions
}

type HeapItem struct {
	bunchIndex       int
	transactionIndex int
	transaction      *WrappedTransaction
}

type TransactionHeap []*HeapItem

func (h TransactionHeap) Len() int { return len(h) }

func (h TransactionHeap) Less(i, j int) bool {
	return h[i].transaction.isTransactionMoreDesirableByProtocol(h[j].transaction)
}

func (h TransactionHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *TransactionHeap) Push(x interface{}) {
	*h = append(*h, x.(*HeapItem))
}

func (h *TransactionHeap) Pop() interface{} {
	// Standard code when storing the heap in a slice:
	// https://pkg.go.dev/container/heap
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}
