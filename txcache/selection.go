package txcache

import (
	"container/heap"
	"time"
)

func (cache *TxCache) doSelectTransactions(accountStateProvider AccountStateProvider, gasRequested uint64, maxNum int, selectionLoopMaximumDuration time.Duration) (bunchOfTransactions, uint64) {
	senders := cache.getSenders()
	bunches := make([]bunchOfTransactions, 0, len(senders))

	for _, sender := range senders {
		bunches = append(bunches, sender.getTxs())
	}

	return selectTransactionsFromBunches(accountStateProvider, bunches, gasRequested, maxNum, selectionLoopMaximumDuration)
}

// Selection tolerates concurrent transaction additions / removals.
func selectTransactionsFromBunches(accountStateProvider AccountStateProvider, bunches []bunchOfTransactions, gasRequested uint64, maxNum int, selectionLoopMaximumDuration time.Duration) (bunchOfTransactions, uint64) {
	selectedTransactions := make(bunchOfTransactions, 0, initialCapacityOfSelectionSlice)

	// Items popped from the heap are added to "selectedTransactions".
	transactionsHeap := newMaxTransactionsHeap(len(bunches))
	heap.Init(transactionsHeap)

	// Initialize the heap with the first transaction of each bunch
	for _, bunch := range bunches {
		if len(bunch) == 0 {
			// Some senders may have no transactions (hazardous).
			continue
		}

		// Items will be reused (see below). Each sender gets one (and only one) item in the heap.
		heap.Push(transactionsHeap, newTransactionsHeapItem(bunch))
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

		requestAccountStateIfNecessary(accountStateProvider, item)

		shouldSkipSender := detectSkippableSender(item)
		if shouldSkipSender {
			// Item was popped from the heap, but not used downstream.
			// Therefore, the sender is completely ignored (from now on) in the current selection session.
			continue
		}

		shouldSkipTransaction := detectSkippableTransaction(item)
		if shouldSkipTransaction {
			// Transaction isn't selected, but the sender is still in the game (will contribute with other transactions).
		} else {
			accumulatedGas += gasLimit
			selectedTransactions = append(selectedTransactions, item.selectTransaction())
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

func requestAccountStateIfNecessary(accountStateProvider AccountStateProvider, item *transactionsHeapItem) {
	if item.senderStateRequested {
		return
	}

	item.senderStateRequested = true

	sender := item.currentTransaction.Tx.GetSndAddr()
	senderState, err := accountStateProvider.GetAccountState(sender)
	if err != nil {
		// Hazardous; should never happen.
		logSelect.Debug("TxCache.requestAccountStateIfNecessary: nonce not available", "sender", sender, "err", err)
		return
	}

	item.senderStateProvided = true
	item.senderState = senderState
}

func detectSkippableSender(item *transactionsHeapItem) bool {
	if item.detectInitialGap() {
		return true
	}
	if item.detectMiddleGap() {
		return true
	}
	if item.detectFeeExceededBalance() {
		return true
	}

	return false
}

func detectSkippableTransaction(item *transactionsHeapItem) bool {
	if item.detectLowerNonce() {
		return true
	}
	if item.detectBadlyGuarded() {
		return true
	}
	if item.detectNonceDuplicate() {
		return true
	}

	return false
}
