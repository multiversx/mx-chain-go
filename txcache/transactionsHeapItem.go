package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-storage-go/types"
)

type transactionsHeapItem struct {
	sender []byte
	bunch  bunchOfTransactions

	// The sender's state, as fetched in "requestAccountStateIfNecessary".
	senderState *types.AccountState

	currentTransactionIndex        int
	currentTransaction             *WrappedTransaction
	currentTransactionNonce        uint64
	latestSelectedTransaction      *WrappedTransaction
	latestSelectedTransactionNonce uint64

	consumedBalance *big.Int
}

func newTransactionsHeapItem(bunch bunchOfTransactions) (*transactionsHeapItem, error) {
	if len(bunch) == 0 {
		return nil, errEmptyBunchOfTransactions
	}

	firstTransaction := bunch[0]

	return &transactionsHeapItem{
		sender: firstTransaction.Tx.GetSndAddr(),
		bunch:  bunch,

		senderState: nil,

		currentTransactionIndex:   0,
		currentTransaction:        firstTransaction,
		currentTransactionNonce:   firstTransaction.Tx.GetNonce(),
		latestSelectedTransaction: nil,

		consumedBalance: big.NewInt(0),
	}, nil
}

func (item *transactionsHeapItem) selectCurrentTransaction(session SelectionSession) *WrappedTransaction {
	item.accumulateConsumedBalance(session)

	item.latestSelectedTransaction = item.currentTransaction
	item.latestSelectedTransactionNonce = item.currentTransactionNonce

	return item.currentTransaction
}

func (item *transactionsHeapItem) accumulateConsumedBalance(session SelectionSession) {
	fee := item.currentTransaction.Fee
	if fee != nil {
		item.consumedBalance.Add(item.consumedBalance, fee)
	}

	transferredValue := session.GetTransferredValue(item.currentTransaction.Tx)
	if transferredValue != nil {
		item.consumedBalance.Add(item.consumedBalance, transferredValue)
	}
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

func (item *transactionsHeapItem) detectInitialGap() bool {
	if item.latestSelectedTransaction != nil {
		return false
	}
	if item.senderState == nil {
		return false
	}

	hasInitialGap := item.currentTransactionNonce > item.senderState.Nonce
	if hasInitialGap {
		logSelect.Trace("transactionsHeapItem.detectInitialGap, initial gap",
			"tx", item.currentTransaction.TxHash,
			"nonce", item.currentTransactionNonce,
			"sender", item.sender,
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

func (item *transactionsHeapItem) detectWillFeeExceedBalance() bool {
	if item.senderState == nil {
		return false
	}

	fee := item.currentTransaction.Fee
	if fee == nil {
		return false
	}

	// Here, we are not interested into an eventual transfer of value (we only check if there's enough balance to pay the transaction fee).
	futureConsumedBalance := new(big.Int).Add(item.consumedBalance, fee)
	senderBalance := item.senderState.Balance

	willFeeExceedBalance := futureConsumedBalance.Cmp(senderBalance) > 0
	if willFeeExceedBalance {
		logSelect.Trace("transactionsHeapItem.detectWillFeeExceedBalance",
			"tx", item.currentTransaction.TxHash,
			"sender", item.sender,
			"balance", item.senderState.Balance,
			"consumedBalance", item.consumedBalance,
		)
	}

	return willFeeExceedBalance
}

func (item *transactionsHeapItem) detectLowerNonce() bool {
	if item.senderState == nil {
		return false
	}

	isLowerNonce := item.currentTransactionNonce < item.senderState.Nonce
	if isLowerNonce {
		logSelect.Trace("transactionsHeapItem.detectLowerNonce",
			"tx", item.currentTransaction.TxHash,
			"nonce", item.currentTransactionNonce,
			"sender", item.sender,
			"senderState.Nonce", item.senderState.Nonce,
		)
	}

	return isLowerNonce
}

func (item *transactionsHeapItem) detectBadlyGuarded(session SelectionSession) bool {
	isBadlyGuarded := session.IsBadlyGuarded(item.currentTransaction.Tx)
	if isBadlyGuarded {
		logSelect.Trace("transactionsHeapItem.detectBadlyGuarded",
			"tx", item.currentTransaction.TxHash,
			"sender", item.sender,
		)
	}

	return isBadlyGuarded
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

func (item *transactionsHeapItem) requestAccountStateIfNecessary(session SelectionSession) error {
	if item.senderState != nil {
		return nil
	}

	senderState, err := session.GetAccountState(item.sender)
	if err != nil {
		return err
	}

	item.senderState = senderState
	return nil
}

func (item *transactionsHeapItem) isCurrentTransactionMoreValuableForNetwork(other *transactionsHeapItem) bool {
	return item.currentTransaction.isTransactionMoreValuableForNetwork(other.currentTransaction)
}
