package trigger

import "github.com/ElrondNetwork/elrond-go/update"

const HardforkTriggerString = hardforkTriggerString
const PayloadSeparator = dataSeparator
const HardforkGracePeriod = hardforkGracePeriod

func (t *trigger) SetTimeHandler(handler func() int64) {
	t.getTimestampHandler = handler
}

func (t *trigger) Closers() []update.Closer {
	t.mutClosers.RLock()
	defer t.mutClosers.RUnlock()

	return t.closers
}
