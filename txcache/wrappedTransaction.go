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

	PricePerGasUnitQuotient  uint64
	PricePerGasUnitRemainder uint64
}

// computePricePerGasUnit computes (and caches) the (average) price per gas unit.
func (wrappedTx *WrappedTransaction) computePricePerGasUnit(txGasHandler TxGasHandler) {
	fee := txGasHandler.ComputeTxFee(wrappedTx.Tx)
	gasLimit := big.NewInt(0).SetUint64(wrappedTx.Tx.GetGasLimit())

	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient, remainder = quotient.QuoRem(fee, gasLimit, remainder)

	wrappedTx.PricePerGasUnitQuotient = quotient.Uint64()
	wrappedTx.PricePerGasUnitRemainder = remainder.Uint64()
}
