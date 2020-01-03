package txcache

import "github.com/ElrondNetwork/elrond-go/data"

func computeTxSize(tx data.TransactionHandler) int64 {
	sizeOfBoundedTxFields := int64(128)
	sizeOfData := int64(len(tx.GetData()))
	return sizeOfBoundedTxFields + sizeOfData
}
