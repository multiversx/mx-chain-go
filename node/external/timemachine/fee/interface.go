package fee

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
)

// EconomicsDataWithComputeFee defines an economics data with computed fee
type EconomicsDataWithComputeFee interface {
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	ComputeTxFeeInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) *big.Int
	ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeGasUsedAndFeeBasedOnRefundValueInEpoch(tx data.TransactionWithFeeHandler, refundValue *big.Int, epoch uint32) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeTxFeeBasedOnGasUsedInEpoch(tx data.TransactionWithFeeHandler, gasUsed uint64, epoch uint32) *big.Int
	ComputeGasLimitInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) uint64
	IsInterfaceNil() bool
}
