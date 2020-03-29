package trigger

func (t *trigger) RegisteredHandlers() []func() {
	t.mutTriggerHandlers.RLock()
	defer t.mutTriggerHandlers.RUnlock()

	return t.triggerHandlers
}
