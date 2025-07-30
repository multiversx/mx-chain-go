package txcache

type transactionsHeapItem struct {
	sender []byte
	bunch  bunchOfTransactions

	currentTransactionIndex        int
	currentTransaction             *WrappedTransaction
	currentTransactionNonce        uint64
	latestSelectedTransaction      *WrappedTransaction
	latestSelectedTransactionNonce uint64
}

func newTransactionsHeapItem(bunch bunchOfTransactions) (*transactionsHeapItem, error) {
	if len(bunch) == 0 {
		return nil, errEmptyBunchOfTransactions
	}

	firstTransaction := bunch[0]

	return &transactionsHeapItem{
		sender: firstTransaction.Tx.GetSndAddr(),
		bunch:  bunch,

		currentTransactionIndex:   0,
		currentTransaction:        firstTransaction,
		currentTransactionNonce:   firstTransaction.Tx.GetNonce(),
		latestSelectedTransaction: nil,
	}, nil
}

func (item *transactionsHeapItem) selectCurrentTransaction() *WrappedTransaction {
	item.latestSelectedTransaction = item.currentTransaction
	item.latestSelectedTransactionNonce = item.currentTransactionNonce

	return item.currentTransaction
}

func (item *transactionsHeapItem) gotoNextTransaction() bool {
	if item.currentTransactionIndex+1 >= len(item.bunch) {
		return false
	}

	item.currentTransactionIndex++
	item.currentTransaction = item.bunch[item.currentTransactionIndex]
	item.currentTransactionNonce = item.currentTransaction.Tx.GetNonce()
	return true
}

func (item *transactionsHeapItem) detectInitialGap(senderNonce uint64) bool {
	if item.latestSelectedTransaction != nil {
		return false
	}

	hasInitialGap := item.currentTransactionNonce > senderNonce
	if hasInitialGap {
		logSelect.Trace("transactionsHeapItem.detectInitialGap, initial gap",
			"tx", item.currentTransaction.TxHash,
			"nonce", item.currentTransactionNonce,
			"sender", item.sender,
			"senderNonce", senderNonce,
		)
	}

	return hasInitialGap
}

func (item *transactionsHeapItem) detectMiddleGap() bool {
	if item.latestSelectedTransaction == nil {
		return false
	}

	// Detect middle gap.
	hasMiddleGap := item.currentTransactionNonce > item.latestSelectedTransactionNonce+1
	if hasMiddleGap {
		logSelect.Trace("transactionsHeapItem.detectMiddleGap, middle gap",
			"tx", item.currentTransaction.TxHash,
			"nonce", item.currentTransactionNonce,
			"sender", item.sender,
			"previousSelectedNonce", item.latestSelectedTransactionNonce,
		)
	}

	return hasMiddleGap
}

func (item *transactionsHeapItem) detectLowerNonce(senderNonce uint64) bool {
	isLowerNonce := item.currentTransactionNonce < senderNonce
	if isLowerNonce {
		logSelect.Trace("transactionsHeapItem.detectLowerNonce",
			"tx", item.currentTransaction.TxHash,
			"nonce", item.currentTransactionNonce,
			"sender", item.sender,
			"senderNonce", senderNonce,
		)
	}

	return isLowerNonce
}

func (item *transactionsHeapItem) detectIncorrectlyGuarded(sessionWrapper *selectionSessionWrapper) bool {
	isIncorrectlyGuarded := sessionWrapper.isIncorrectlyGuarded(item.currentTransaction.Tx)
	if isIncorrectlyGuarded {
		logSelect.Trace("transactionsHeapItem.detectIncorrectlyGuarded",
			"tx", item.currentTransaction.TxHash,
			"sender", item.sender,
		)
	}

	return isIncorrectlyGuarded
}

func (item *transactionsHeapItem) detectNonceDuplicate() bool {
	if item.latestSelectedTransaction == nil {
		return false
	}

	isDuplicate := item.currentTransactionNonce == item.latestSelectedTransactionNonce
	if isDuplicate {
		logSelect.Trace("transactionsHeapItem.detectNonceDuplicate",
			"tx", item.currentTransaction.TxHash,
			"sender", item.sender,
			"nonce", item.currentTransactionNonce,
		)
	}

	return isDuplicate
}

func (item *transactionsHeapItem) isCurrentTransactionMoreValuableForNetwork(other *transactionsHeapItem) bool {
	return item.currentTransaction.isTransactionMoreValuableForNetwork(other.currentTransaction)
}
