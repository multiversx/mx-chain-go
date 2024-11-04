package txcache

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

type BunchOfTransactions []*WrappedTransaction

// WrappedTransaction contains a transaction, its hash and extra information
type WrappedTransaction struct {
	Tx              data.TransactionHandler
	TxHash          []byte
	SenderShardID   uint32
	ReceiverShardID uint32
	Size            int64

	PricePerUnit uint64
	HashFnv32    uint32
}

// precomputeFields computes (and caches) the (average) price per gas unit.
func (transaction *WrappedTransaction) precomputeFields(txGasHandler TxGasHandler) {
	fee := txGasHandler.ComputeTxFee(transaction.Tx).Uint64()

	gasLimit := transaction.Tx.GetGasLimit()
	if gasLimit == 0 {
		return
	}

	transaction.PricePerUnit = fee / gasLimit
	transaction.HashFnv32 = fnv32(string(transaction.TxHash))
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
func (transaction *WrappedTransaction) isTransactionMoreDesirableByProtocol(otherTransaction *WrappedTransaction) bool {
	// First, compare by price per unit
	ppu := transaction.PricePerUnit
	ppuOther := otherTransaction.PricePerUnit
	if ppu != ppuOther {
		return ppu > ppuOther
	}

	// In the end, compare by hash number of transaction hash
	return transaction.HashFnv32 > otherTransaction.HashFnv32
}
