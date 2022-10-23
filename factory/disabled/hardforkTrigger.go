package disabled

// hardforkTrigger implements HardforkTrigger interface but does nothing as it is disabled
type hardforkTrigger struct {
}

// HardforkTrigger returns a disabled hardforkTrigger
func HardforkTrigger() *hardforkTrigger {
	return &hardforkTrigger{}
}

// TriggerReceived does nothing as it is disabled
func (h *hardforkTrigger) TriggerReceived(_ []byte, _ []byte, _ []byte) (bool, error) {
	return false, nil
}

// RecordedTriggerMessage does nothing as it is disabled
func (h *hardforkTrigger) RecordedTriggerMessage() ([]byte, bool) {
	return nil, false
}

// NotifyTriggerReceivedV2 does nothing as it is disabled
func (h *hardforkTrigger) NotifyTriggerReceivedV2() <-chan struct{} {
	return nil
}

// CreateData does nothing as it is disabled
func (h *hardforkTrigger) CreateData() []byte {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *hardforkTrigger) IsInterfaceNil() bool {
	return h == nil
}
