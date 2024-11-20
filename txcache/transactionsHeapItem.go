package txcache

import (
	"bytes"
	"math/big"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-storage-go/types"
)

type transactionsHeapItem struct {
	// Whether the sender's state has been requested within a selection session.
	senderStateRequested bool
	// Whether the sender's state has been requested and provided (with success) within a selection session.
	senderStateProvided bool
	// The sender's state (if requested and provided).
	senderState *types.AccountState

	bunch                     bunchOfTransactions
	currentTransactionIndex   int
	currentTransaction        *WrappedTransaction
	latestSelectedTransaction *WrappedTransaction

	accumulatedFee *big.Int
}

func newTransactionsHeapItem(bunch bunchOfTransactions) *transactionsHeapItem {
	return &transactionsHeapItem{
		senderStateRequested: false,
		senderStateProvided:  false,
		senderState:          nil,

		bunch:                   bunch,
		currentTransactionIndex: 0,
		currentTransaction:      bunch[0],
		accumulatedFee:          big.NewInt(0),
	}
}

func (item *transactionsHeapItem) selectTransaction() *WrappedTransaction {
	item.accumulateFee()
	item.latestSelectedTransaction = item.currentTransaction
	return item.currentTransaction
}

func (item *transactionsHeapItem) accumulateFee() {
	fee := item.currentTransaction.Fee.Load()
	if fee == nil {
		// This should never happen during selection.
		return
	}

	item.accumulatedFee.Add(item.accumulatedFee, fee)
}

func (item *transactionsHeapItem) gotoNextTransaction() bool {
	item.currentTransactionIndex++

	if item.currentTransactionIndex >= len(item.bunch) {
		return false
	}

	item.currentTransaction = item.bunch[item.currentTransactionIndex]
	return true
}

func (item *transactionsHeapItem) detectInitialGap() bool {
	if item.latestSelectedTransaction != nil {
		return false
	}
	if !item.senderStateProvided {
		// This should never happen during selection.
		return false
	}

	hasInitialGap := item.currentTransaction.Tx.GetNonce() > item.senderState.Nonce
	if hasInitialGap && logSelect.GetLevel() <= logger.LogTrace {
		logSelect.Trace("transactionsHeapItem.detectGap, initial gap",
			"tx", item.currentTransaction.TxHash,
			"nonce", item.currentTransaction.Tx.GetNonce(),
			"sender", item.currentTransaction.Tx.GetSndAddr(),
			"senderState.Nonce", item.senderState.Nonce,
		)
	}

	return hasInitialGap
}

func (item *transactionsHeapItem) detectMiddleGap() bool {
	if item.latestSelectedTransaction == nil {
		return false
	}

	// Detect middle gap.
	previouslySelectedTransactionNonce := item.latestSelectedTransaction.Tx.GetNonce()

	hasMiddleGap := item.currentTransaction.Tx.GetNonce() > previouslySelectedTransactionNonce+1
	if hasMiddleGap && logSelect.GetLevel() <= logger.LogTrace {
		logSelect.Trace("transactionsHeapItem.detectGap, middle gap",
			"tx", item.currentTransaction.TxHash,
			"nonce", item.currentTransaction.Tx.GetNonce(),
			"sender", item.currentTransaction.Tx.GetSndAddr(),
			"previousSelectedNonce", previouslySelectedTransactionNonce,
		)
	}

	return hasMiddleGap
}

func (item *transactionsHeapItem) detectFeeExceededBalance() bool {
	if !item.senderStateProvided {
		// This should never happen during selection.
		return false
	}

	hasFeeExceededBalance := item.accumulatedFee.Cmp(item.senderState.Balance) > 0
	if hasFeeExceededBalance && logSelect.GetLevel() <= logger.LogTrace {
		logSelect.Trace("transactionsHeapItem.detectFeeExceededBalance",
			"tx", item.currentTransaction.TxHash,
			"sender", item.currentTransaction.Tx.GetSndAddr(),
			"balance", item.senderState.Balance,
			"accumulatedFee", item.accumulatedFee,
		)
	}

	return hasFeeExceededBalance
}

func (item *transactionsHeapItem) detectLowerNonce() bool {
	if !item.senderStateProvided {
		// This should never happen during selection.
		return false
	}

	isLowerNonce := item.currentTransaction.Tx.GetNonce() < item.senderState.Nonce
	if isLowerNonce && logSelect.GetLevel() <= logger.LogTrace {
		logSelect.Trace("transactionsHeapItem.detectLowerNonce",
			"tx", item.currentTransaction.TxHash,
			"nonce", item.currentTransaction.Tx.GetNonce(),
			"sender", item.currentTransaction.Tx.GetSndAddr(),
			"senderState.Nonce", item.senderState.Nonce,
		)
	}

	return isLowerNonce
}

func (item *transactionsHeapItem) detectBadlyGuarded() bool {
	if !item.senderStateProvided {
		// This should never happen during selection.
		return false
	}

	transactionGuardian := *item.currentTransaction.Guardian.Load()
	accountGuardian := item.senderState.Guardian
	isBadlyGuarded := bytes.Compare(transactionGuardian, accountGuardian) != 0
	if isBadlyGuarded && logSelect.GetLevel() <= logger.LogTrace {
		logSelect.Trace("transactionsHeapItem.detectBadlyGuarded",
			"tx", item.currentTransaction.TxHash,
			"sender", item.currentTransaction.Tx.GetSndAddr(),
			"transactionGuardian", transactionGuardian,
			"accountGuardian", accountGuardian,
		)
	}

	return isBadlyGuarded
}

func (item *transactionsHeapItem) detectNonceDuplicate() bool {
	if item.latestSelectedTransaction == nil {
		return false
	}
	if !item.senderStateProvided {
		// This should never happen during selection.
		return false
	}

	isDuplicate := item.currentTransaction.Tx.GetNonce() == item.latestSelectedTransaction.Tx.GetNonce()
	if isDuplicate && logSelect.GetLevel() <= logger.LogTrace {
		logSelect.Trace("transactionsHeapItem.detectNonceDuplicate",
			"tx", item.currentTransaction.TxHash,
			"sender", item.currentTransaction.Tx.GetSndAddr(),
			"nonce", item.currentTransaction.Tx.GetNonce(),
		)
	}

	return isDuplicate
}
