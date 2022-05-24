package timemachine

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type disabledEpochNotifier struct {
}

// RegisterNotifyHandler does nothing
func (notifier *disabledEpochNotifier) RegisterNotifyHandler(_ vmcommon.EpochSubscriberHandler) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (notifier *disabledEpochNotifier) IsInterfaceNil() bool {
	return notifier == nil
}
