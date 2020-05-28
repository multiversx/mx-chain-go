package txcache

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/data"
)

const estimatedSizeOfBoundedTxFields = uint64(128)

// WrappedTransaction contains a transaction, its hash and extra information
type WrappedTransaction struct {
	Tx                     data.TransactionHandler
	TxHash                 []byte
	SenderShardID          uint32
	ReceiverShardID        uint32
	isImmuneToEvictionFlag atomic.Flag
}

// GetKey gets the transaction hash
func (wrappedTx *WrappedTransaction) GetKey() []byte {
	return wrappedTx.TxHash
}

// Payload gets the inner (wrapped) transaction
func (wrappedTx *WrappedTransaction) Payload() interface{} {
	return wrappedTx.Tx
}

// Size gets the size (in bytes) of the transaction
func (wrappedTx *WrappedTransaction) Size() int {
	return int(estimateTxSize(wrappedTx))
}

// IsImmuneToEviction returns whether the transaction is immune to eviction
func (wrappedTx *WrappedTransaction) IsImmuneToEviction() bool {
	return wrappedTx.isImmuneToEvictionFlag.IsSet()
}

// ImmunizeAgainstEviction marks the transaction as immune to eviction
func (wrappedTx *WrappedTransaction) ImmunizeAgainstEviction() {
	wrappedTx.isImmuneToEvictionFlag.Set()
}

func (wrappedTx *WrappedTransaction) sameAs(another *WrappedTransaction) bool {
	return bytes.Equal(wrappedTx.TxHash, another.TxHash)
}

// estimateTxSize returns an approximation
func estimateTxSize(tx *WrappedTransaction) uint64 {
	sizeOfData := uint64(len(tx.Tx.GetData()))
	return estimatedSizeOfBoundedTxFields + sizeOfData
}

// estimateTxGas returns an approximation for the necessary computation units (gas units)
func estimateTxGas(tx *WrappedTransaction) uint64 {
	gasLimit := tx.Tx.GetGasLimit()
	return gasLimit
}

// estimateTxFee returns an approximation for the cost of a transaction, in nano ERD
// TODO: switch to integer operations (as opposed to float operations).
// TODO: do not assume the order of magnitude of minGasPrice.
func estimateTxFee(tx *WrappedTransaction) uint64 {
	// In order to obtain the result as nano ERD (not as "atomic" 10^-18 ERD), we have to divide by 10^9
	// In order to have better precision, we divide the factors by 10^6, and 10^3 respectively
	gasLimit := float32(tx.Tx.GetGasLimit()) / 1000000
	gasPrice := float32(tx.Tx.GetGasPrice()) / 1000
	feeInNanoERD := gasLimit * gasPrice
	return uint64(feeInNanoERD)
}
