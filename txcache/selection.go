package txcache

import (
	"container/heap"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
)

func (cache *TxCache) doSelectTransactions(virtualSession *virtualSelectionSession, options common.TxSelectionOptions) (bunchOfTransactions, uint64) {
	bunches := cache.acquireBunchesOfTransactions()

	gracePeriodMs := cache.config.TxCacheBoundsConfig.PropagationGracePeriodMs
	propagationGracePeriod := time.Duration(gracePeriodMs) * time.Millisecond

	return selectTransactionsFromBunches(virtualSession, bunches, options, propagationGracePeriod)
}

func (cache *TxCache) acquireBunchesOfTransactions() []bunchOfTransactions {
	senders := cache.getSenders()
	bunches := make([]bunchOfTransactions, 0, len(senders))

	for _, sender := range senders {
		bunch := sender.getTxsForSelection()
		if len(bunch) > 0 {
			bunches = append(bunches, bunch)
		}
	}

	return bunches
}

// Selection tolerates concurrent transaction additions / removals.
func selectTransactionsFromBunches(
	virtualSession *virtualSelectionSession,
	bunches []bunchOfTransactions,
	options common.TxSelectionOptions,
	propagationGracePeriod time.Duration,
) (bunchOfTransactions, uint64) {
	gasRequested := options.GetGasRequested()
	maxNumTxs := options.GetMaxNumTxs()
	loopDurationCheckInterval := options.GetLoopDurationCheckInterval()

	logSelect.Debug("TxCache.selectTransactionsFromBunches",
		"len(bunches)", len(bunches),
		"gasRequested", gasRequested,
		"maxNumTxs", maxNumTxs,
		"loopDurationCheckInterval", loopDurationCheckInterval,
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
	selectionCutoff := selectionLoopStartTime.Add(-propagationGracePeriod)
	uniqueSenderCount := 0

	var currentTransaction *WrappedTransaction
	var processedTxs int
	// Select transactions (sorted).
	for transactionsHeap.Len() > 0 {
		processedTxs++
		// Always pick the best transaction.
		item := heap.Pop(transactionsHeap).(*transactionsHeapItem)
		gasLimit := item.currentTransaction.Tx.GetGasLimit()

		if accumulatedGas+gasLimit > gasRequested {
			break
		}
		if len(selectedTransactions) >= maxNumTxs {
			break
		}
		if processedTxs%loopDurationCheckInterval == 0 {
			if !options.HaveTimeForSelection() {
				logSelect.Debug("TxCache.selectTransactionsFromBunches, selection loop timeout", "duration", time.Since(selectionLoopStartTime))
				break
			}
		}

		// Skip transactions that haven't had enough time to propagate
		if propagationGracePeriod > 0 && item.currentTransaction.ReceivedAt.After(selectionCutoff) {
			continue
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
			// first, we get the transaction that might be selected
			currentTransaction = item.getCurrentTransaction()
			err = virtualSession.accumulateConsumedBalance(currentTransaction, senderRecord)
			if err != nil {
				// This error is unlikely to occur, as it would have been raised earlier during the detectSkippableSender call.
				// However, we should not select the transaction if anything fails here.
				log.Warn("TxCache.selectTransactionsFromBunches error when accumulating consumed balance",
					"err", err,
					"txHash", currentTransaction.TxHash)
			} else {
				// Track unique senders to match OnProposedBlock's maxAccountsPerBlock limit.
				isNewSender := item.latestSelectedTransaction == nil
				if isNewSender {
					uniqueSenderCount++
					if uniqueSenderCount > maxAccountsPerBlock {
						logSelect.Debug("TxCache.selectTransactionsFromBunches, unique sender limit reached",
							"limit", maxAccountsPerBlock)
						break
					}
				}

				// only if there isn't any error, we select the transaction
				accumulatedGas += gasLimit
				item.selectCurrentTransaction()
				selectedTransactions = append(selectedTransactions, currentTransaction)
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
		// This error is expected for accounts with discontinuous global breadcrumbs,
		// which get a blocked virtual record (initialNonce.HasValue=false).
		// In this case, the sender is correctly skipped to avoid selecting transactions
		// that would fail OnProposedBlock validation.
		log.Debug("detectSkippableSender", "err", err)
		return true
	}

	if virtualRecord.hasPendingChangeGuardian() {
		return true
	}
	if item.detectInitialGap(nonce) {
		return true
	}
	if item.detectMiddleGap() {
		return true
	}
	if virtualSession.detectWillBalanceBeExceeded(item.currentTransaction) {
		return true
	}

	isSetGuardianCall := process.IsSetGuardianCall(item.currentTransaction.Tx.GetData())
	if !isSetGuardianCall {
		return false
	}

	// if an instant change guardian is detected, further transactions for this sender should be skipped as they are probably guarded by the old one
	// allow this one though
	virtualSession.setChangeGuardianIfNeeded(item.currentTransaction.Tx)

	return false
}

func detectSkippableTransaction(virtualSession *virtualSelectionSession, item *transactionsHeapItem, virtualRecord *virtualAccountRecord) bool {
	nonce, err := virtualRecord.getInitialNonce()
	if err != nil {
		// This error is expected for accounts with discontinuous global breadcrumbs,
		// which get a blocked virtual record (initialNonce.HasValue=false).
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
