package txcache

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

type scoreComputer interface {
	computeScore(scoreParams senderScoreParams) uint32
}

// TxGasHandler handles a transaction gas and gas cost
type TxGasHandler interface {
	SplitTxGasInCategories(tx process.TransactionWithFeeHandler) (uint64, uint64)
	GasPriceForProcessing(tx process.TransactionWithFeeHandler) uint64
	GasPriceForMove(tx process.TransactionWithFeeHandler) uint64
	MinGasPrice() uint64
	MinGasLimit() uint64
	MinGasPriceForProcessing() uint64
	IsInterfaceNil() bool
}

// ForEachTransaction is an iterator callback
type ForEachTransaction func(txHash []byte, value *WrappedTransaction)
