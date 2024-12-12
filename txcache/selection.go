package txcache

import (
	"container/heap"
	"time"
)

func (cache *TxCache) doSelectTransactions(session SelectionSession, gasRequested uint64, maxNum int, selectionLoopMaximumDuration time.Duration) (bunchOfTransactions, uint64) {
	bunches := cache.acquireBunchesOfTransactions()

	return selectTransactionsFromBunches(session, bunches, gasRequested, maxNum, selectionLoopMaximumDuration)
}

func (cache *TxCache) acquireBunchesOfTransactions() []bunchOfTransactions {
	senders := cache.getSenders()
	bunches := make([]bunchOfTransactions, 0, len(senders))

	for _, sender := range senders {
		bunches = append(bunches, sender.getTxs())
	}

	return bunches
}

// Selection tolerates concurrent transaction additions / removals.
func selectTransactionsFromBunches(session SelectionSession, bunches []bunchOfTransactions, gasRequested uint64, maxNum int, selectionLoopMaximumDuration time.Duration) (bunchOfTransactions, uint64) {
	selectedTransactions := make(bunchOfTransactions, 0, initialCapacityOfSelectionSlice)
	sessionWrapper := newSelectionSessionWrapper(session)

	// Items popped from the heap are added to "selectedTransactions".
	transactionsHeap := newMaxTransactionsHeap(len(bunches))
	heap.Init(transactionsHeap)

	// Initialize the heap with the first transaction of each bunch
	for _, bunch := range bunches {
		item, err := newTransactionsHeapItem(bunch)
		if err != nil {
			continue
		}

		// Items will be reused (see below). Each sender gets one (and only one) item in the heap.
		heap.Push(transactionsHeap, item)
	}

	accumulatedGas := uint64(0)
	selectionLoopStartTime := time.Now()

	// Select transactions (sorted).
	for transactionsHeap.Len() > 0 {
		// Always pick the best transaction.
		item := heap.Pop(transactionsHeap).(*transactionsHeapItem)
		gasLimit := item.currentTransaction.Tx.GetGasLimit()

		if accumulatedGas+gasLimit > gasRequested {
			break
		}
		if len(selectedTransactions) >= maxNum {
			break
		}
		if len(selectedTransactions)%selectionLoopDurationCheckInterval == 0 {
			if time.Since(selectionLoopStartTime) > selectionLoopMaximumDuration {
				logSelect.Debug("TxCache.selectTransactionsFromBunches, selection loop timeout", "duration", time.Since(selectionLoopStartTime))
				break
			}
		}

		shouldSkipSender := detectSkippableSender(sessionWrapper, item)
		if shouldSkipSender {
			// Item was popped from the heap, but not used downstream.
			// Therefore, the sender is completely ignored (from now on) in the current selection session.
			continue
		}

		shouldSkipTransaction := detectSkippableTransaction(sessionWrapper, item)
		if !shouldSkipTransaction {
			accumulatedGas += gasLimit
			selectedTransaction := item.selectCurrentTransaction()
			selectedTransactions = append(selectedTransactions, selectedTransaction)
			sessionWrapper.accumulateConsumedBalance(selectedTransaction)
		}

		// If there are more transactions in the same bunch (same sender as the popped item),
		// add the next one to the heap (to compete with the others).
		// Heap item is reused (same originating sender), pushed back on the heap.
		if item.gotoNextTransaction() {
			heap.Push(transactionsHeap, item)
		}
	}

	return selectedTransactions, accumulatedGas
}

func detectSkippableSender(sessionWrapper *selectionSessionWrapper, item *transactionsHeapItem) bool {
	nonce := sessionWrapper.getNonce(item.sender)

	if item.detectInitialGap(nonce) {
		return true
	}
	if item.detectMiddleGap() {
		return true
	}
	if sessionWrapper.detectWillFeeExceedBalance(item.currentTransaction) {
		return true
	}

	return false
}

func detectSkippableTransaction(sessionWrapper *selectionSessionWrapper, item *transactionsHeapItem) bool {
	nonce := sessionWrapper.getNonce(item.sender)

	if item.detectLowerNonce(nonce) {
		return true
	}
	if item.detectIncorrectlyGuarded(sessionWrapper) {
		return true
	}
	if item.detectNonceDuplicate() {
		return true
	}

	return false
}
