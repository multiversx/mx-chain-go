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
	Fee              *big.Int
	PricePerUnit     uint64
	TransferredValue *big.Int
}

// precomputeFields computes (and caches) the (average) price per gas unit.
func (wrappedTx *WrappedTransaction) precomputeFields(host MempoolHost) {
	wrappedTx.Fee = host.ComputeTxFee(wrappedTx.Tx)

	gasLimit := wrappedTx.Tx.GetGasLimit()
	if gasLimit != 0 {
		wrappedTx.PricePerUnit = wrappedTx.Fee.Uint64() / gasLimit
	}

	wrappedTx.TransferredValue = host.GetTransferredValue(wrappedTx.Tx)
}

// Equality is out of scope (not possible in our case).
func (wrappedTx *WrappedTransaction) isTransactionMoreValuableForNetwork(otherTransaction *WrappedTransaction) bool {
	// First, compare by PPU (higher PPU is better).
	if wrappedTx.PricePerUnit != otherTransaction.PricePerUnit {
		return wrappedTx.PricePerUnit > otherTransaction.PricePerUnit
	}

	// If PPU is the same, compare by gas limit (higher gas limit is better, promoting less "execution fragmentation").
	gasLimit := wrappedTx.Tx.GetGasLimit()
	gasLimitOther := otherTransaction.Tx.GetGasLimit()

	if gasLimit != gasLimitOther {
		return gasLimit > gasLimitOther
	}

	// In the end, compare by transaction hash
	return bytes.Compare(wrappedTx.TxHash, otherTransaction.TxHash) < 0
}
