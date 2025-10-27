package txcache

import (
	"container/heap"
	"time"

	"github.com/multiversx/mx-chain-go/common"
)

func (cache *TxCache) doSelectTransactions(virtualSession *virtualSelectionSession, options common.TxSelectionOptions) (bunchOfTransactions, uint64) {
	bunches := cache.acquireBunchesOfTransactions()

	return selectTransactionsFromBunches(virtualSession, bunches, options)
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
func selectTransactionsFromBunches(
	virtualSession *virtualSelectionSession,
	bunches []bunchOfTransactions,
	options common.TxSelectionOptions,
) (bunchOfTransactions, uint64) {
	gasRequested := options.GetGasRequested()
	maxNumTxs := options.GetMaxNumTxs()
	loopDurationCheckInterval := options.GetLoopDurationCheckInterval()
	selectionLoopMaxDuration := time.Duration(options.GetLoopMaximumDurationMs()) * time.Millisecond

	logSelect.Debug("TxCache.selectTransactionsFromBunches",
		"len(bunches)", len(bunches),
		"gasRequested", gasRequested,
		"maxNumTxs", maxNumTxs,
		"loopDurationCheckInterval", loopDurationCheckInterval,
		"selectionLoopMaxDuration", selectionLoopMaxDuration,
	)

	selectedTransactions := make(bunchOfTransactions, 0, initialCapacityOfSelectionSlice)

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
		if len(selectedTransactions) >= maxNumTxs {
			break
		}
		if len(selectedTransactions)%loopDurationCheckInterval == 0 {
			if time.Since(selectionLoopStartTime) > selectionLoopMaxDuration {
				logSelect.Debug("TxCache.selectTransactionsFromBunches, selection loop timeout", "duration", time.Since(selectionLoopStartTime))
				break
			}
		}

		senderRecord, err := virtualSession.getRecord(item.sender)
		if err != nil {
			log.Debug("TxCache.selectTransactionsFromBunches when getting the virtual record of sender", "err", err,
				"address", item.sender)
			continue
		}

		shouldSkipSender := detectSkippableSender(virtualSession, item, senderRecord)
		if shouldSkipSender {
			// Item was popped from the heap, but not used downstream.
			// Therefore, the sender is completely ignored (from now on) in the current selection session.
			continue
		}

		shouldSkipTransaction := detectSkippableTransaction(virtualSession, item, senderRecord)
		if !shouldSkipTransaction {
			accumulatedGas += gasLimit
			selectedTransaction := item.selectCurrentTransaction()
			selectedTransactions = append(selectedTransactions, selectedTransaction)
			err := virtualSession.accumulateConsumedBalance(selectedTransaction, senderRecord)
			if err != nil {
				// This error is unlikely to occur, as it would have been raised earlier during the detectSkippableSender call.
				// Even if it does occur, it doesn't imply that the transaction should not be selected.
				// Therefore, we only log the error here.
				log.Warn("TxCache.selectTransactionsFromBunches error when accumulating consumed balance",
					"err", err,
					"txHash", selectedTransaction.TxHash)
			}
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

// Note (future micro-optimization): we can merge "detectSkippableSender()" and "detectSkippableTransaction()" into a single function,
// any share the result of "sessionWrapper.getNonceForAccountRecord()".
func detectSkippableSender(virtualSession *virtualSelectionSession, item *transactionsHeapItem, virtualRecord *virtualAccountRecord) bool {
	nonce, err := virtualRecord.getInitialNonce()
	if err != nil {
		// Every virtual record is initialized with the session nonce, to avoid virtual records with no initial nonce.
		// So this error should never appear.
		log.Debug("detectSkippableSender", "err", err)
		return true
	}

	if item.detectInitialGap(nonce) {
		return true
	}
	if item.detectMiddleGap() {
		return true
	}
	if virtualSession.detectWillFeeExceedBalance(item.currentTransaction) {
		return true
	}

	return false
}

func detectSkippableTransaction(virtualSession *virtualSelectionSession, item *transactionsHeapItem, virtualRecord *virtualAccountRecord) bool {
	nonce, err := virtualRecord.getInitialNonce()
	if err != nil {
		// Every virtual record is initialized with the session nonce, to avoid virtual records with no initial nonce.
		// So this error should never appear.
		log.Debug("detectSkippableTransaction", "err", err)
		return true
	}
	if item.detectLowerNonce(nonce) {
		return true
	}
	if item.detectIncorrectlyGuarded(virtualSession) {
		return true
	}
	if item.detectNonceDuplicate() {
		return true
	}

	return false
}
