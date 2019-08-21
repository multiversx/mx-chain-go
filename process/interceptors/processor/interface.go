package processor

import "github.com/ElrondNetwork/elrond-go/data"

// InterceptedTransactionHandler defines a transaction-like intercepted data that has
// receiver and sender shard getters
type InterceptedTransactionHandler interface {
	SndShard() uint32
	RcvShard() uint32
	Hash() []byte
	UnderlyingTransaction() data.TransactionHandler
}
