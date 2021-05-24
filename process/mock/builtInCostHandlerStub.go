package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

// BuiltInCostHandlerStub -
type BuiltInCostHandlerStub struct {
	ComputeBuiltInCostCalled func(tx process.TransactionWithFeeHandler) uint64
}

// ComputeBuiltInCost -
func (b *BuiltInCostHandlerStub) ComputeBuiltInCost(tx process.TransactionWithFeeHandler) uint64 {
	if b.ComputeBuiltInCostCalled != nil {
		return b.ComputeBuiltInCostCalled(tx)
	}
	return 1
}

// IsBuiltInFuncCall -
func (b *BuiltInCostHandlerStub) IsBuiltInFuncCall(_ process.TransactionWithFeeHandler) bool {
	return false
}

// IsInterfaceNil -
func (b *BuiltInCostHandlerStub) IsInterfaceNil() bool {
	return b == nil
}
