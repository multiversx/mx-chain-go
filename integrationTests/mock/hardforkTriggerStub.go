package mock

// HardforkTriggerStub -
type HardforkTriggerStub struct {
	TriggerCalled                func() error
	IsSelfTriggerCalled          func() bool
	TriggerReceivedCalled        func(payload []byte, pkBytes []byte) error
	RecordedTriggerMessageCalled func() ([]byte, bool)
}

// Trigger -
func (hts *HardforkTriggerStub) Trigger() error {
	if hts.TriggerCalled != nil {
		return hts.TriggerCalled()
	}

	return nil
}

// IsSelfTrigger -
func (hts *HardforkTriggerStub) IsSelfTrigger() bool {
	if hts.IsSelfTriggerCalled != nil {
		return hts.IsSelfTriggerCalled()
	}

	return false
}

// TriggerReceived -
func (hts *HardforkTriggerStub) TriggerReceived(payload []byte, pkBytes []byte) error {
	if hts.TriggerReceivedCalled != nil {
		return hts.TriggerReceivedCalled(payload, pkBytes)
	}

	return nil
}

// RecordedTriggerMessage -
func (hts *HardforkTriggerStub) RecordedTriggerMessage() ([]byte, bool) {
	if hts.RecordedTriggerMessageCalled != nil {
		return hts.RecordedTriggerMessageCalled()
	}

	return nil, false
}

// IsInterfaceNil -
func (hts *HardforkTriggerStub) IsInterfaceNil() bool {
	return hts == nil
}
