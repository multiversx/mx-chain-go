package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
)

// BuiltInCostHandlerStub -
type BuiltInCostHandlerStub struct {
}

// ComputeBuiltInCost -
func (b *BuiltInCostHandlerStub) ComputeBuiltInCost(_ data.TransactionWithFeeHandler) uint64 {
	return 1
}

// IsBuiltInFuncCall -
func (b *BuiltInCostHandlerStub) IsBuiltInFuncCall(_ data.TransactionWithFeeHandler) bool {
	return false
}

// IsInterfaceNil -
func (b *BuiltInCostHandlerStub) IsInterfaceNil() bool {
	return b == nil
}
