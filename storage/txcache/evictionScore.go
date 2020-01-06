package txcache

import "github.com/ElrondNetwork/elrond-go/data"

// estimateTxSize returns an approximation
func estimateTxSize(tx data.TransactionHandler) int64 {
	sizeOfBoundedTxFields := int64(128)
	sizeOfData := int64(len(tx.GetData()))
	return sizeOfBoundedTxFields + sizeOfData
}

// estimateTxGas returns an approximation
func estimateTxGas(tx data.TransactionHandler) int64 {
	gasPrice := int64(tx.GetGasPrice())
	sizeOfData := int64(len(tx.GetData()))
	return gasPrice * sizeOfData
}
