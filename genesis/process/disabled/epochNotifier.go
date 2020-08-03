package disabled

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

// EpochNotifier represents a disabled epoch notifier implementation
type EpochNotifier struct {
}

// RegisterNotifyHandler will call the non nil handler with the 0 epoch value
func (en *EpochNotifier) RegisterNotifyHandler(handler core.EpochNotifiedHandler) {
	if check.IfNil(handler) {
		return
	}

	handler.NewEpochConfirmed(0)
}

// CurrentEpoch always returns 0
func (en *EpochNotifier) CurrentEpoch() uint32 {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (en *EpochNotifier) IsInterfaceNil() bool {
	return en == nil
}
