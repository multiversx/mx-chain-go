package testscommon

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
)

// FeeComputerStub -
type FeeComputerStub struct {
	ComputeTransactionFeeCalled func(tx data.TransactionWithFeeHandler, epoch int) (*big.Int, error)
}

// ComputeTransactionFee -
func (stub *FeeComputerStub) ComputeTransactionFee(tx data.TransactionWithFeeHandler, epoch int) (*big.Int, error) {
	if stub.ComputeTransactionFeeCalled != nil {
		return stub.ComputeTransactionFeeCalled(tx, epoch)
	}

	return big.NewInt(0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *FeeComputerStub) IsInterfaceNil() bool {
	return false
}
