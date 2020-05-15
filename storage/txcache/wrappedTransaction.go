package txcache

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/data"
)

// WrappedTransaction contains a transaction, its hash and extra information
type WrappedTransaction struct {
	Tx              data.TransactionHandler
	TxHash          []byte
	SenderShardID   uint32
	ReceiverShardID uint32
}

func (wrappedTx *WrappedTransaction) sameAs(another *WrappedTransaction) bool {
	return bytes.Equal(wrappedTx.TxHash, another.TxHash)
}
