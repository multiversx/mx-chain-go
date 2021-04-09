package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

// BuiltInCostHandlerStub -
type BuiltInCostHandlerStub struct {
}

// ComputeBuiltInCost -
func (b BuiltInCostHandlerStub) ComputeBuiltInCost(_ process.TransactionWithFeeHandler) uint64 {
	return 1
}

// IsBuiltInFuncCall -
func (b BuiltInCostHandlerStub) IsBuiltInFuncCall(_ process.TransactionWithFeeHandler) bool {
	return false
}

// IsInterfaceNil -
func (b *BuiltInCostHandlerStub) IsInterfaceNil() bool {
	return b == nil
}
