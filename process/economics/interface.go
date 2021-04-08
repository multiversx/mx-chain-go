package economics

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

// BuiltInFunctionsCostHandler is able to calculated the cost of a built in function call
type BuiltInFunctionsCostHandler interface {
	ComputeBuiltInCost(tx process.TransactionWithFeeHandler) uint64
	IsSpecialBuiltIn(tx process.TransactionWithFeeHandler) bool
}
