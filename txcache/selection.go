package txcache

import (
	"container/heap"
)

func (cache *TxCache) doSelectTransactions(accountNonceProvider AccountNonceProvider, gasRequested uint64, maxNum int) (bunchOfTransactions, uint64) {
	senders := cache.getSenders()
	bunches := make([]bunchOfTransactions, 0, len(senders))

	for _, sender := range senders {
		bunches = append(bunches, sender.getSequentialTxs())
	}

	return selectTransactionsFromBunches(bunches, gasRequested, maxNum)
}

// Selection tolerates concurrent transaction additions / removals.
func selectTransactionsFromBunches(bunches []bunchOfTransactions, gasRequested uint64, maxNum int) (bunchOfTransactions, uint64) {
	selectedTransactions := make(bunchOfTransactions, 0, initialCapacityOfSelectionSlice)

	// Items popped from the heap are added to "selectedTransactions".
	transactionsHeap := newMaxTransactionsHeap(len(bunches))
	heap.Init(transactionsHeap)

	// Initialize the heap with the first transaction of each bunch
	for i, bunch := range bunches {
		if len(bunch) == 0 {
			// Some senders may have no eligible transactions (initial gaps).
			continue
		}

		// Items will be reused (see below). Each sender gets one (and only one) item in the heap.
		heap.Push(transactionsHeap, &transactionsHeapItem{
			senderIndex:      i,
			transactionIndex: 0,
			transaction:      bunch[0],
		})
	}

	accumulatedGas := uint64(0)

	// Select transactions (sorted).
	for transactionsHeap.Len() > 0 {
		// Always pick the best transaction.
		item := heap.Pop(transactionsHeap).(*transactionsHeapItem)
		gasLimit := item.transaction.Tx.GetGasLimit()

		if accumulatedGas+gasLimit > gasRequested {
			break
		}
		if len(selectedTransactions) >= maxNum {
			break
		}

		accumulatedGas += gasLimit
		selectedTransactions = append(selectedTransactions, item.transaction)

		// If there are more transactions in the same bunch (same sender as the popped item),
		// add the next one to the heap (to compete with the others).
		item.transactionIndex++

		if item.transactionIndex < len(bunches[item.senderIndex]) {
			// Item is reused (same originating sender), pushed back on the heap.
			item.transaction = bunches[item.senderIndex][item.transactionIndex]
			heap.Push(transactionsHeap, item)
		}
	}

	return selectedTransactions, accumulatedGas
}
