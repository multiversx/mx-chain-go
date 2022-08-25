package timemachine

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// DisabledEpochNotifier is a no-operation EpochNotifier
type DisabledEpochNotifier struct {
}

// RegisterNotifyHandler does nothing
func (notifier *DisabledEpochNotifier) RegisterNotifyHandler(_ vmcommon.EpochSubscriberHandler) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (notifier *DisabledEpochNotifier) IsInterfaceNil() bool {
	return notifier == nil
}
