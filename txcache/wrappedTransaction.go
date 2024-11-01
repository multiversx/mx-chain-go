package txcache

import (
	"bytes"
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
func (transaction *WrappedTransaction) computePricePerGasUnit(txGasHandler TxGasHandler) {
	fee := txGasHandler.ComputeTxFee(transaction.Tx)
	gasLimit := big.NewInt(0).SetUint64(transaction.Tx.GetGasLimit())

	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient, remainder = quotient.QuoRem(fee, gasLimit, remainder)

	transaction.PricePerGasUnitQuotient = quotient.Uint64()
	transaction.PricePerGasUnitRemainder = remainder.Uint64()
}

// Equality is out of scope (not possible in our case).
func (transaction *WrappedTransaction) isTransactionMoreDesirableByProtocol(otherTransaction *WrappedTransaction) bool {
	// First, compare by price per unit
	ppuQuotient := transaction.PricePerGasUnitQuotient
	ppuQuotientOther := otherTransaction.PricePerGasUnitQuotient
	if ppuQuotient != ppuQuotientOther {
		return ppuQuotient > ppuQuotientOther
	}

	ppuRemainder := transaction.PricePerGasUnitRemainder
	ppuRemainderOther := otherTransaction.PricePerGasUnitRemainder
	if ppuRemainder != ppuRemainderOther {
		return ppuRemainder > ppuRemainderOther
	}

	// Then, compare by gas price (to promote the practice of a higher gas price)
	gasPrice := transaction.Tx.GetGasPrice()
	gasPriceOther := otherTransaction.Tx.GetGasPrice()
	if gasPrice != gasPriceOther {
		return gasPrice > gasPriceOther
	}

	// Then, compare by gas limit (promote the practice of lower gas limit)
	// Compare Gas Limits (promote lower gas limit)
	gasLimit := transaction.Tx.GetGasLimit()
	gasLimitOther := otherTransaction.Tx.GetGasLimit()
	if gasLimit != gasLimitOther {
		return gasLimit < gasLimitOther
	}

	// In the end, compare by transaction hash
	return bytes.Compare(transaction.TxHash, otherTransaction.TxHash) > 0
}
