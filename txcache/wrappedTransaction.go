package txcache

import (
	"bytes"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
)

// bunchOfTransactions is a slice of WrappedTransaction pointers
type bunchOfTransactions []*WrappedTransaction

// WrappedTransaction contains a transaction, its hash and extra information
type WrappedTransaction struct {
	Tx              data.TransactionHandler
	TxHash          []byte
	SenderShardID   uint32
	ReceiverShardID uint32
	Size            int64

	// These fields are only set within "precomputeFields".
	// We don't need to protect them with a mutex, since "precomputeFields" is called only once for each transaction.
	// Additional note: "WrappedTransaction" objects are created by the Node, in dataRetriever/txpool/shardedTxPool.go.
	Fee          *big.Int
	PricePerUnit uint64
	Guardian     []byte
}

// precomputeFields computes (and caches) the (average) price per gas unit.
func (wrappedTx *WrappedTransaction) precomputeFields(txGasHandler TxGasHandler) {
	wrappedTx.Fee = txGasHandler.ComputeTxFee(wrappedTx.Tx)

	gasLimit := wrappedTx.Tx.GetGasLimit()
	if gasLimit != 0 {
		wrappedTx.PricePerUnit = wrappedTx.Fee.Uint64() / gasLimit
	}

	txAsGuardedTransaction, ok := wrappedTx.Tx.(data.GuardedTransactionHandler)
	if ok {
		wrappedTx.Guardian = txAsGuardedTransaction.GetGuardianAddr()
	}
}

// Equality is out of scope (not possible in our case).
func (wrappedTx *WrappedTransaction) isTransactionMoreValuableForNetwork(otherTransaction *WrappedTransaction) bool {
	// First, compare by price per unit
	ppu := wrappedTx.PricePerUnit
	ppuOther := otherTransaction.PricePerUnit
	if ppu != ppuOther {
		return ppu > ppuOther
	}

	// In the end, compare by transaction hash
	return bytes.Compare(wrappedTx.TxHash, otherTransaction.TxHash) < 0
}
