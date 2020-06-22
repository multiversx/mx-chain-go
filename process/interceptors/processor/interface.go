package processor

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// InterceptedTransactionHandler defines an intercepted data wrapper over transaction handler that has
// receiver and sender shard getters
type InterceptedTransactionHandler interface {
	SenderShardId() uint32
	ReceiverShardId() uint32
	Nonce() uint64
	SenderAddress() []byte
	Fee() *big.Int
	Transaction() data.TransactionHandler
}

// ShardedPool is a perspective of the sharded data pool
type ShardedPool interface {
	AddData(key []byte, data interface{}, sizeInBytes int, cacheID string)
}
