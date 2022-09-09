package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
)

// EpochStartNotifier -
type EpochStartNotifier struct {
}

// NewEpochStartNotifier returns the disabled epoch start notifier
func NewEpochStartNotifier() *EpochStartNotifier {
	return &EpochStartNotifier{}
}

// RegisterHandler -
func (desn *EpochStartNotifier) RegisterHandler(_ core.EpochStartActionHandler) {
}

// UnregisterHandler -
func (desn *EpochStartNotifier) UnregisterHandler(_ core.EpochStartActionHandler) {
}

// NotifyAllPrepare -
func (desn *EpochStartNotifier) NotifyAllPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
}

// NotifyAll -
func (desn *EpochStartNotifier) NotifyAll(_ data.HeaderHandler) {
}

// IsInterfaceNil -
func (desn *EpochStartNotifier) IsInterfaceNil() bool {
	return desn == nil
}
