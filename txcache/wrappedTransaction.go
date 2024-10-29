package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
)

// WrappedTransaction contains a transaction, its hash and extra information
type WrappedTransaction struct {
	Tx              data.TransactionHandler
	TxHash          []byte
	SenderShardID   uint32
	ReceiverShardID uint32
	Size            int64

	TxFee *big.Int
}

// computeFee computes (and caches) the transaction fee.
func (wrappedTx *WrappedTransaction) computeFee(txGasHandler TxGasHandler) {
	wrappedTx.TxFee = txGasHandler.ComputeTxFee(wrappedTx.Tx)
}
