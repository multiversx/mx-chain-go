package economics

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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
