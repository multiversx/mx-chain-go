package mock

// HardforkFacade -
type HardforkFacade struct {
	TriggerCalled       func() error
	IsSelfTriggerCalled func() bool
}

// Trigger -
func (hf *HardforkFacade) Trigger() error {
	if hf.TriggerCalled != nil {
		return hf.TriggerCalled()
	}

	return nil
}

// IsSelfTrigger -
func (hf *HardforkFacade) IsSelfTrigger() bool {
	if hf.IsSelfTriggerCalled != nil {
		return hf.IsSelfTriggerCalled()
	}

	return false
}
