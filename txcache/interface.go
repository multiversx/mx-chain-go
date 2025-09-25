package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
)

// MempoolHost provides blockchain information for mempool operations
type MempoolHost interface {
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	GetTransferredValue(tx data.TransactionHandler) *big.Int
	IsInterfaceNil() bool
}

// SelectionSession provides blockchain information for transaction selection
type SelectionSession interface {
	GetAccountNonceAndBalance(accountKey []byte) (uint64, *big.Int, bool, error)
	GetRootHash() ([]byte, error)
	IsIncorrectlyGuarded(tx data.TransactionHandler) bool
	IsInterfaceNil() bool
}

// ForEachTransaction is an iterator callback
type ForEachTransaction func(txHash []byte, value *WrappedTransaction)

// txCacheForSelectionTracker provides the TxCache methods used in SelectionTracker
type txCacheForSelectionTracker interface {
	GetByTxHash(txHash []byte) (*WrappedTransaction, bool)
	IsInterfaceNil() bool
}

// TransactionsTracker provides the transactionsTracker methods used by other components
type TransactionsTracker interface {
	IsTransactionTracked(transaction *WrappedTransaction) bool
	IsInterfaceNil() bool
}
