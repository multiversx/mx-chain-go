package preprocess

import "github.com/ElrondNetwork/elrond-go/data"

// SortedTransactionsProvider defines the public API of the transactions cache
type SortedTransactionsProvider interface {
	GetTransactions(numRequested int, batchSizePerSender int) ([]data.TransactionHandler, [][]byte)
}
