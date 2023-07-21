package fee

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
)

type economicsDataWithComputeFee interface {
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	ComputeTxFeeInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) *big.Int
	ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeGasUsedAndFeeBasedOnRefundValueInEpoch(tx data.TransactionWithFeeHandler, refundValue *big.Int, epoch uint32) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeTxFeeBasedOnGasUsedInEpoch(tx data.TransactionWithFeeHandler, gasUsed uint64, epoch uint32) *big.Int
	ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64
}
