package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// EpochNotifier implements the EpochNotifier interface but does nothing as it is disabled
type EpochNotifier struct {
}

// CheckEpoch does nothing as it is disabled
func (en *EpochNotifier) CheckEpoch(_ data.HeaderHandler) {
}

// RegisterNotifyHandler does nothing as it is disabled
func (en *EpochNotifier) RegisterNotifyHandler(_ vmcommon.EpochSubscriberHandler) {
}

// CurrentEpoch returns 0 as it is disabled
func (en *EpochNotifier) CurrentEpoch() uint32 {
	return 0
}

// IsInterfaceNil returns true if underlying object is nil
func (en *EpochNotifier) IsInterfaceNil() bool {
	return en == nil
}
