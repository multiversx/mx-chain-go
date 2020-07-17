package scwatcher

import "github.com/ElrondNetwork/elrond-go/data"

// ScWatcherDriver defines the interface of the Smart Contracts Watcher
type Driver interface {
	DigestBlock(body data.BodyHandler, header data.HeaderHandler, transactions TransactionsToDigest)
}

// TransactionsToDigest holds current transactions to digest
type TransactionsToDigest struct {
	RegularTxs  map[string]data.TransactionHandler
	ScResults   map[string]data.TransactionHandler
	InvalidTxs  map[string]data.TransactionHandler
	ReceiptsTxs map[string]data.TransactionHandler
}
