package fee

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
)

type economicsDataWithComputeFee interface {
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64
}
