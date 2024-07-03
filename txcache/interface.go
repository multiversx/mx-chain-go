package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
)

type scoreComputer interface {
	computeScore(scoreParams senderScoreParams) uint32
}

// TxGasHandler handles a transaction gas and gas cost
type TxGasHandler interface {
	MinGasPrice() uint64
	MaxGasLimitPerTx() uint64
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	IsInterfaceNil() bool
}

// ForEachTransaction is an iterator callback
type ForEachTransaction func(txHash []byte, value *WrappedTransaction)
