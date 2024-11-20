package txcache

import (
	"bytes"
	"math/big"
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

	Fee          atomic.Pointer[big.Int]
	PricePerUnit atomic.Uint64
	Guardian     atomic.Pointer[[]byte]
}

// precomputeFields computes (and caches) the (average) price per gas unit.
func (wrappedTx *WrappedTransaction) precomputeFields(txGasHandler TxGasHandler) {
	fee := txGasHandler.ComputeTxFee(wrappedTx.Tx)

	gasLimit := wrappedTx.Tx.GetGasLimit()
	if gasLimit == 0 {
		return
	}

	wrappedTx.Fee.Store(fee)
	wrappedTx.PricePerUnit.Store(fee.Uint64() / gasLimit)

	txAsGuardedTransaction, ok := wrappedTx.Tx.(data.GuardedTransactionHandler)
	if !ok {
		return
	}

	guardian := txAsGuardedTransaction.GetGuardianAddr()
	wrappedTx.Guardian.Store(&guardian)
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
