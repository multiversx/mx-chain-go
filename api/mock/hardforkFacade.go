package mock

// HardforkFacade -
type HardforkFacade struct {
	TriggerCalled       func(epoch uint32, withEarlyEndOfEpoch bool) error
	IsSelfTriggerCalled func() bool
}

// Trigger -
func (hf *HardforkFacade) Trigger(epoch uint32, withEarlyEndOfEpoch bool) error {
	if hf.TriggerCalled != nil {
		return hf.TriggerCalled(epoch, withEarlyEndOfEpoch)
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

// IsInterfaceNil -
func (hf *HardforkFacade) IsInterfaceNil() bool {
	return hf == nil
}
