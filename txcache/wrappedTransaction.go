package txcache

import (
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
	HashFnv32    atomic.Uint32
}

// precomputeFields computes (and caches) the (average) price per gas unit.
func (wrappedTx *WrappedTransaction) precomputeFields(txGasHandler TxGasHandler) {
	fee := txGasHandler.ComputeTxFee(wrappedTx.Tx).Uint64()

	gasLimit := wrappedTx.Tx.GetGasLimit()
	if gasLimit == 0 {
		return
	}

	wrappedTx.PricePerUnit.Store(fee / gasLimit)
	wrappedTx.HashFnv32.Store(fnv32(string(wrappedTx.TxHash)))
}

// fnv32 implements https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function for 32 bits
func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

// Equality is out of scope (not possible in our case).
func (wrappedTx *WrappedTransaction) isTransactionMoreValuableForNetwork(otherTransaction *WrappedTransaction) bool {
	// First, compare by price per unit
	ppu := wrappedTx.PricePerUnit.Load()
	ppuOther := otherTransaction.PricePerUnit.Load()
	if ppu != ppuOther {
		return ppu > ppuOther
	}

	// In the end, compare by hash number of transaction hash
	return wrappedTx.HashFnv32.Load() > otherTransaction.HashFnv32.Load()
}
