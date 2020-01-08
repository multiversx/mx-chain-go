package preprocess

import "github.com/ElrondNetwork/elrond-go/data"

// SortedTransactionsProvider defines the public API of the transactions cache
type SortedTransactionsProvider interface {
	GetSortedTransactions() ([]data.TransactionHandler, [][]byte)
	IsInterfaceNil() bool
}
