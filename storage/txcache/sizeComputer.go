package txcache

import "github.com/ElrondNetwork/elrond-go/data"

// computeTxSize returns an approximation
func computeTxSize(tx data.TransactionHandler) int64 {
	sizeOfBoundedTxFields := int64(128)
	sizeOfData := int64(len(tx.GetData()))
	return sizeOfBoundedTxFields + sizeOfData
}

// computeTxGas returns an approximation
func computeTxGas(tx data.TransactionHandler) int64 {
	gasPrice := int64(tx.GetGasPrice())
	sizeOfData := int64(len(tx.GetData()))
	return gasPrice * sizeOfData
}
