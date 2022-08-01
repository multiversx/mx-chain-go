package transactionsfee

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
)

// FeesProcessorHandler defines the interface for the transaction fees processor
type FeesProcessorHandler interface {
	ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64
	IsInterfaceNil() bool
}
