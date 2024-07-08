package economics

import (
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// EpochNotifier raises epoch change events
type EpochNotifier interface {
	RegisterNotifyHandler(handler vmcommon.EpochSubscriberHandler)
	IsInterfaceNil() bool
}
