package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

const epochStartNotifierName = "disabledEpochStartNotifier"

// EpochStartNotifier -
type EpochStartNotifier struct {
}

// NewEpochStartNotifier returns the disabled epoch start notifier
func NewEpochStartNotifier() *EpochStartNotifier {
	return &EpochStartNotifier{}
}

// RegisterHandler -
func (desn *EpochStartNotifier) RegisterHandler(_ epochStart.ActionHandler) {
}

// UnregisterHandler -
func (desn *EpochStartNotifier) UnregisterHandler(_ epochStart.ActionHandler) {
}

// NotifyAllPrepare -
func (desn *EpochStartNotifier) NotifyAllPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
}

// NotifyAll -
func (desn *EpochStartNotifier) NotifyAll(_ data.HeaderHandler) {
}

// GetName -
func (desn *EpochStartNotifier) GetName() string {
	return epochStartNotifierName
}

// IsInterfaceNil -
func (desn *EpochStartNotifier) IsInterfaceNil() bool {
	return desn == nil
}
