package txcache

import (
	"math"

	"github.com/multiversx/mx-chain-core-go/core"
)

// accountRange defines a structure which contains the minimum and the maximum nonce of an account
type accountRange struct {
	minNonce core.OptionalUint64
	maxNonce core.OptionalUint64
}

// transactionsTracker contains a map of accounts and for each account corresponds a (minNonce, maxNonce) range
type transactionsTracker struct {
	accountsWithRange map[string]*accountRange
}

// newTransactionsTracker creates a new transactions tracker
func newTransactionsTracker(tracker *selectionTracker, transactions []*WrappedTransaction) *transactionsTracker {
	txTracker := &transactionsTracker{
		accountsWithRange: make(map[string]*accountRange),
	}
	txTracker.createAccountsWithDefaultRange(transactions)
	txTracker.updateAccountsWithRange(tracker)

	return txTracker
}

// createAccountsWithDefaultRange initializes all the required accounts with a default range
func (txTracker *transactionsTracker) createAccountsWithDefaultRange(transactions []*WrappedTransaction) {
	for _, tx := range transactions {
		sender := tx.Tx.GetSndAddr()
		_, ok := txTracker.accountsWithRange[string(sender)]
		if !ok {
			txTracker.accountsWithRange[string(sender)] = &accountRange{
				minNonce: core.OptionalUint64{Value: math.MaxUint64, HasValue: false},
				maxNonce: core.OptionalUint64{Value: 0, HasValue: false},
			}
		}
	}
}

// updateAccountsWithRange updates all the saved accounts with the range extracted from breadcrumbs
func (txTracker *transactionsTracker) updateAccountsWithRange(tracker *selectionTracker) {
	tracker.mutTracker.RLock()
	defer tracker.mutTracker.RUnlock()

	for _, tb := range tracker.blocks {
		txTracker.updateRangesWithBreadcrumbs(tb)
	}
}

// updateRangesWithBreadcrumbs updates a specific account with the range extracted from breadcrumbs
func (txTracker *transactionsTracker) updateRangesWithBreadcrumbs(tb *trackedBlock) {
	for sender, senderBreadcrumb := range tb.breadcrumbsByAddress {
		rangeOfSender, isSenderInBatchOfTxs := txTracker.accountsWithRange[sender]
		if !isSenderInBatchOfTxs {
			// skip this sender because it is not part of the batch of transactions
			continue
		}

		err := txTracker.updateRangeWithBreadcrumb(rangeOfSender, senderBreadcrumb)
		if err != nil {
			continue
		}
	}
}

// updateRangeWithBreadcrumb updates a specific account with the range extracted from a specific breadcrumb
func (txTracker *transactionsTracker) updateRangeWithBreadcrumb(rangeOfSender *accountRange, senderBreadcrumb *accountBreadcrumb) error {
	firstNonce := senderBreadcrumb.firstNonce
	lastNonce := senderBreadcrumb.lastNonce

	if !firstNonce.HasValue || !lastNonce.HasValue {
		// it means sender was part of that tracked block, but only as a fee payer
		return errBreadcrumbOfFeePayer
	}

	rangeOfSender.minNonce.Value = min(firstNonce.Value, rangeOfSender.minNonce.Value)
	rangeOfSender.minNonce.HasValue = true

	rangeOfSender.maxNonce.Value = max(lastNonce.Value, rangeOfSender.maxNonce.Value)
	rangeOfSender.maxNonce.HasValue = true

	return nil
}

// IsTransactionTracked checks if a transaction is still in the tracked blocks of the SelectionTracker
func (txTracker *transactionsTracker) IsTransactionTracked(transaction *WrappedTransaction) bool {
	sender := transaction.Tx.GetSndAddr()
	txNonce := transaction.Tx.GetNonce()

	senderRange, ok := txTracker.accountsWithRange[string(sender)]
	if !ok {
		return false
	}

	minNonce := senderRange.minNonce
	maxNonce := senderRange.maxNonce

	if !minNonce.HasValue || !maxNonce.HasValue {
		return false
	}

	if txNonce < minNonce.Value || txNonce > maxNonce.Value {
		return false
	}

	return true
}
