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

// IsSpecialBuiltIn -
func (b BuiltInCostHandlerStub) IsSpecialBuiltIn(_ process.TransactionWithFeeHandler) bool {
	return false
}
