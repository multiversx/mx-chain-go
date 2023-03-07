package economics

import (
	"github.com/multiversx/mx-chain-core-go/data"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// BuiltInFunctionsCostHandler is able to calculate the cost of a built-in function call
type BuiltInFunctionsCostHandler interface {
	ComputeBuiltInCost(tx data.TransactionWithFeeHandler) uint64
	IsBuiltInFuncCall(tx data.TransactionWithFeeHandler) bool
	IsInterfaceNil() bool
}

// EpochNotifier raises epoch change events
type EpochNotifier interface {
	RegisterNotifyHandler(handler vmcommon.EpochSubscriberHandler)
	IsInterfaceNil() bool
}
