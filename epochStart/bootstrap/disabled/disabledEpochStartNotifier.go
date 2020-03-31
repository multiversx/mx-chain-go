package disabled

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// EpochStartNotifier -
type EpochStartNotifier struct {
}

// RegisterHandler -
func (desn *EpochStartNotifier) RegisterHandler(handler epochStart.ActionHandler) {
}

// UnregisterHandler -
func (desn *EpochStartNotifier) UnregisterHandler(handler epochStart.ActionHandler) {
}

// NotifyAllPrepare -
func (desn *EpochStartNotifier) NotifyAllPrepare(metaHeader data.HeaderHandler) {
}

// NotifyAll -
func (desn *EpochStartNotifier) NotifyAll(hdr data.HeaderHandler) {
}

// IsInterfaceNil -
func (desn *EpochStartNotifier) IsInterfaceNil() bool {
	return desn == nil
}
