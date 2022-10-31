package timemachine

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// DisabledEpochNotifier is a no-operation EpochNotifier
type DisabledEpochNotifier struct {
}

// CurrentEpoch returns 0
func (notifier *DisabledEpochNotifier) CurrentEpoch() uint32 {
	return 0
}

// CheckEpoch does nothing
func (notifier *DisabledEpochNotifier) CheckEpoch(_ data.HeaderHandler) {
}

// RegisterNotifyHandler does nothing
func (notifier *DisabledEpochNotifier) RegisterNotifyHandler(_ vmcommon.EpochSubscriberHandler) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (notifier *DisabledEpochNotifier) IsInterfaceNil() bool {
	return notifier == nil
}
