package disabled

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

type disabledEpochStartEventNotifier struct {
}

func NewEpochStartEventNotifier() *disabledEpochStartEventNotifier {
	return &disabledEpochStartEventNotifier{}
}

func (d *disabledEpochStartEventNotifier) RegisterHandler(handler epochStart.ActionHandler) {
}

func (d *disabledEpochStartEventNotifier) UnregisterHandler(handler epochStart.ActionHandler) {
}

func (d *disabledEpochStartEventNotifier) NotifyAll(hdr data.HeaderHandler) {
}

func (d *disabledEpochStartEventNotifier) IsInterfaceNil() bool {
	return d == nil
}
