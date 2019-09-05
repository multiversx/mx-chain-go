package processor

import "github.com/ElrondNetwork/elrond-go/data"

// InterceptedTransactionHandler defines an intercepted data wrapper over transaction handler that has
// receiver and sender shard getters
type InterceptedTransactionHandler interface {
	SndShard() uint32
	RcvShard() uint32
	Hash() []byte
	Transaction() data.TransactionHandler
}
