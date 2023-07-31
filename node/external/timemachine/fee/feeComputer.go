package fee

import (
	"errors"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

var errNilEconomicsData = errors.New("nil economics data")

type feeComputer struct {
	economicsInstance EconomicsDataWithComputeFee
}

// NewFeeComputer creates a fee computer which handles historical transactions, as well
func NewFeeComputer(economicsInstance EconomicsDataWithComputeFee) (*feeComputer, error) {
	if check.IfNil(economicsInstance) {
		return nil, errNilEconomicsData
	}

	computer := &feeComputer{
		economicsInstance: economicsInstance,
	}

	// TODO: Handle fees for guarded transactions, when enabled.

	return computer, nil
}

// ComputeGasUsedAndFeeBasedOnRefundValue computes gas used and fee based on the refund value, at a given epoch
func (computer *feeComputer) ComputeGasUsedAndFeeBasedOnRefundValue(tx *transaction.ApiTransactionResult, refundValue *big.Int) (uint64, *big.Int) {
	return computer.economicsInstance.ComputeGasUsedAndFeeBasedOnRefundValueInEpoch(tx.Tx, refundValue, tx.Epoch)
}

// ComputeTxFeeBasedOnGasUsed computes fee based on gas used, at a given epoch
func (computer *feeComputer) ComputeTxFeeBasedOnGasUsed(tx *transaction.ApiTransactionResult, gasUsed uint64) *big.Int {
	return computer.economicsInstance.ComputeTxFeeBasedOnGasUsedInEpoch(tx.Tx, gasUsed, tx.Epoch)
}

// ComputeGasLimit computes a transaction gas limit, at a given epoch
func (computer *feeComputer) ComputeGasLimit(tx *transaction.ApiTransactionResult) uint64 {
	return computer.economicsInstance.ComputeGasLimitInEpoch(tx.Tx, tx.Epoch)
}

// ComputeTransactionFee computes a transaction fee, at a given epoch
func (computer *feeComputer) ComputeTransactionFee(tx *transaction.ApiTransactionResult) *big.Int {
	return computer.economicsInstance.ComputeTxFeeInEpoch(tx.Tx, tx.Epoch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (computer *feeComputer) IsInterfaceNil() bool {
	return computer == nil
}
