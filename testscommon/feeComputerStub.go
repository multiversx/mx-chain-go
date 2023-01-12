package testscommon

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

// FeeComputerStub -
type FeeComputerStub struct {
	ComputeTransactionFeeCalled func(tx *transaction.ApiTransactionResult) *big.Int
}

// ComputeTransactionFee -
func (stub *FeeComputerStub) ComputeTransactionFee(tx *transaction.ApiTransactionResult) *big.Int {
	if stub.ComputeTransactionFeeCalled != nil {
		return stub.ComputeTransactionFeeCalled(tx)
	}

	return big.NewInt(0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *FeeComputerStub) IsInterfaceNil() bool {
	return false
}
