package testscommon

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

// FeeComputerStub -
type FeeComputerStub struct {
	ComputeTransactionFeeCalled                  func(tx *transaction.ApiTransactionResult) *big.Int
	ComputeGasUsedAndFeeBasedOnRefundValueCalled func(tx *transaction.ApiTransactionResult, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsedCalled             func(tx *transaction.ApiTransactionResult, gasUsed uint64) *big.Int
	ComputeGasLimitCalled                        func(tx *transaction.ApiTransactionResult) uint64
}

// ComputeTransactionFee -
func (stub *FeeComputerStub) ComputeTransactionFee(tx *transaction.ApiTransactionResult) *big.Int {
	if stub.ComputeTransactionFeeCalled != nil {
		return stub.ComputeTransactionFeeCalled(tx)
	}

	return big.NewInt(0)
}

// ComputeGasUsedAndFeeBasedOnRefundValue -
func (stub *FeeComputerStub) ComputeGasUsedAndFeeBasedOnRefundValue(tx *transaction.ApiTransactionResult, refundValue *big.Int) (uint64, *big.Int) {
	if stub.ComputeGasUsedAndFeeBasedOnRefundValueCalled != nil {
		return stub.ComputeGasUsedAndFeeBasedOnRefundValueCalled(tx, refundValue)
	}
	return 0, big.NewInt(0)
}

// ComputeTxFeeBasedOnGasUsed -
func (stub *FeeComputerStub) ComputeTxFeeBasedOnGasUsed(tx *transaction.ApiTransactionResult, gasUsed uint64) *big.Int {
	if stub.ComputeTxFeeBasedOnGasUsedCalled != nil {
		return stub.ComputeTxFeeBasedOnGasUsedCalled(tx, gasUsed)
	}

	return big.NewInt(0)
}

// ComputeGasLimit -
func (stub *FeeComputerStub) ComputeGasLimit(tx *transaction.ApiTransactionResult) uint64 {
	if stub.ComputeGasLimitCalled != nil {
		return stub.ComputeGasLimitCalled(tx)
	}

	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *FeeComputerStub) IsInterfaceNil() bool {
	return false
}
