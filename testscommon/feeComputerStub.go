package testscommon

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
)

// FeeComputerStub -
type FeeComputerStub struct {
	ComputeTransactionFeeCalled               func(tx data.TransactionWithFeeHandler, epoch int) *big.Int
	ComputeTransactionFeeForMoveBalanceCalled func(tx data.TransactionWithFeeHandler, epoch int) *big.Int
}

// ComputeTransactionFee -
func (stub *FeeComputerStub) ComputeTransactionFee(tx data.TransactionWithFeeHandler, epoch int) *big.Int {
	if stub.ComputeTransactionFeeCalled != nil {
		return stub.ComputeTransactionFeeCalled(tx, epoch)
	}

	return big.NewInt(0)
}

// ComputeTransactionFee -
func (stub *FeeComputerStub) ComputeTransactionFeeForMoveBalance(tx data.TransactionWithFeeHandler, epoch int) *big.Int {
	if stub.ComputeTransactionFeeForMoveBalanceCalled != nil {
		return stub.ComputeTransactionFeeForMoveBalanceCalled(tx, epoch)
	}

	return big.NewInt(0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *FeeComputerStub) IsInterfaceNil() bool {
	return false
}
