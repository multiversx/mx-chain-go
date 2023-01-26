package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// BuiltInCostHandlerStub -
type BuiltInCostHandlerStub struct {
	ComputeBuiltInCostCalled func(tx data.TransactionWithFeeHandler) uint64
	IsBuiltInFuncCallCalled  func(tx data.TransactionWithFeeHandler) bool
}

// ComputeBuiltInCost -
func (stub *BuiltInCostHandlerStub) ComputeBuiltInCost(tx data.TransactionWithFeeHandler) uint64 {
	if stub.ComputeBuiltInCostCalled != nil {
		return stub.ComputeBuiltInCostCalled(tx)
	}

	return 1
}

// IsBuiltInFuncCall -
func (stub *BuiltInCostHandlerStub) IsBuiltInFuncCall(tx data.TransactionWithFeeHandler) bool {
	if stub.IsBuiltInFuncCallCalled != nil {
		return stub.IsBuiltInFuncCallCalled(tx)
	}

	return false
}

// IsInterfaceNil returns true if underlying object is nil
func (stub *BuiltInCostHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
