package txcache

import (
	"bytes"
	"sync/atomic"

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

	PricePerUnit atomic.Uint64
}

// precomputeFields computes (and caches) the (average) price per gas unit.
func (wrappedTx *WrappedTransaction) precomputeFields(txGasHandler TxGasHandler) {
	fee := txGasHandler.ComputeTxFee(wrappedTx.Tx).Uint64()

	gasLimit := wrappedTx.Tx.GetGasLimit()
	if gasLimit == 0 {
		return
	}

	wrappedTx.PricePerUnit.Store(fee / gasLimit)
}

// Equality is out of scope (not possible in our case).
func (wrappedTx *WrappedTransaction) isTransactionMoreValuableForNetwork(otherTransaction *WrappedTransaction) bool {
	// First, compare by price per unit
	ppu := wrappedTx.PricePerUnit.Load()
	ppuOther := otherTransaction.PricePerUnit.Load()
	if ppu != ppuOther {
		return ppu > ppuOther
	}

	// In the end, compare by transaction hash
	return bytes.Compare(wrappedTx.TxHash, otherTransaction.TxHash) < 0
}
