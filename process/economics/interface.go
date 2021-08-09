package economics

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
)

// BuiltInFunctionsCostHandler is able to calculated the cost of a built in function call
type BuiltInFunctionsCostHandler interface {
	ComputeBuiltInCost(tx data.TransactionWithFeeHandler) uint64
	IsBuiltInFuncCall(tx data.TransactionWithFeeHandler) bool
	IsInterfaceNil() bool
}
