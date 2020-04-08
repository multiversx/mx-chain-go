package trigger

const HardforkTriggerString = hardforkTriggerString
const PayloadSeparator = dataSeparator
const HardforkGracePeriod = hardforkGracePeriod

func (t *trigger) RegisteredHandlers() []func() {
	t.mutTriggerHandlers.RLock()
	defer t.mutTriggerHandlers.RUnlock()

	return t.triggerHandlers
}

func (t *trigger) SetTimeHandler(handler func() int64) {
	t.getTimestampHandler = handler
}
