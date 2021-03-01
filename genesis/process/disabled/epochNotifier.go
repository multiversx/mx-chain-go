package disabled

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// EpochNotifier -
type EpochNotifier struct {
}

// CheckEpoch -
func (en *EpochNotifier) CheckEpoch(_ uint32) {
}

// RegisterNotifyHandler -
func (en *EpochNotifier) RegisterNotifyHandler(_ core.EpochSubscriberHandler) {
}

// CurrentEpoch -
func (en *EpochNotifier) CurrentEpoch() uint32 {
	return 0
}

// IsInterfaceNil -
func (en *EpochNotifier) IsInterfaceNil() bool {
	return en == nil
}
