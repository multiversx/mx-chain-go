package processor

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// InterceptedTransactionHandler defines an intercepted data wrapper over transaction handler that has
// receiver and sender shard getters
type InterceptedTransactionHandler interface {
	SenderShardId() uint32
	ReceiverShardId() uint32
	Nonce() uint64
	SenderAddress() state.AddressContainer
	TotalValue() *big.Int
	Hash() []byte
	Transaction() data.TransactionHandler
}
