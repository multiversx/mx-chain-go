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
	TxFee           float64
}

// computeFee computes the transaction fee.
// The returned fee is also held on the transaction object.
func (wrappedTx *WrappedTransaction) computeFee(txGasHandler TxGasHandler) float64 {
	fee := txGasHandler.ComputeTxFee(wrappedTx.Tx)
	feeAsFloat, _ := new(big.Float).SetInt(fee).Float64()
	wrappedTx.TxFee = feeAsFloat
	return feeAsFloat
}
