package txcache

import "github.com/ElrondNetwork/elrond-go/data"

const estimatedSizeOfBoundedTxFields = uint32(128)

// estimateTxSize returns an approximation
func estimateTxSize(tx data.TransactionHandler) int64 {
	sizeOfData := int64(len(tx.GetData()))
	return int64(estimatedSizeOfBoundedTxFields) + sizeOfData
}

// estimateTxGas returns an approximation
func estimateTxGas(tx data.TransactionHandler) int64 {
	gasPrice := int64(tx.GetGasPrice())
	gasLimit := int64(tx.GetGasLimit())
	return gasPrice * gasLimit
}
