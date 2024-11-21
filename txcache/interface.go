package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
)

// TxGasHandler handles a transaction gas and gas cost
type TxGasHandler interface {
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	IsInterfaceNil() bool
}

// ForEachTransaction is an iterator callback
type ForEachTransaction func(txHash []byte, value *WrappedTransaction)
