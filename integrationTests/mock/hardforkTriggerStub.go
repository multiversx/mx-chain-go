package mock

import "github.com/ElrondNetwork/elrond-go/update"

// HardforkTriggerStub -
type HardforkTriggerStub struct {
	TriggerCalled                func(epoch uint32, forced bool) error
	IsSelfTriggerCalled          func() bool
	TriggerReceivedCalled        func(payload []byte, data []byte, pkBytes []byte) (bool, error)
	RecordedTriggerMessageCalled func() ([]byte, bool)
	CreateDataCalled             func() []byte
	AddCloserCalled              func(closer update.Closer) error
	NotifyTriggerReceivedCalled  func() <-chan struct{}
}

// Trigger -
func (hts *HardforkTriggerStub) Trigger(epoch uint32, forced bool) error {
	if hts.TriggerCalled != nil {
		return hts.TriggerCalled(epoch, forced)
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
func (hts *HardforkTriggerStub) TriggerReceived(payload []byte, data []byte, pkBytes []byte) (bool, error) {
	if hts.TriggerReceivedCalled != nil {
		return hts.TriggerReceivedCalled(payload, data, pkBytes)
	}

	return false, nil
}

// RecordedTriggerMessage -
func (hts *HardforkTriggerStub) RecordedTriggerMessage() ([]byte, bool) {
	if hts.RecordedTriggerMessageCalled != nil {
		return hts.RecordedTriggerMessageCalled()
	}

	return nil, false
}

// CreateData -
func (hts *HardforkTriggerStub) CreateData() []byte {
	if hts.CreateDataCalled != nil {
		return hts.CreateDataCalled()
	}

	return make([]byte, 0)
}

// AddCloser -
func (hts *HardforkTriggerStub) AddCloser(closer update.Closer) error {
	if hts.AddCloserCalled != nil {
		return hts.AddCloserCalled(closer)
	}

	return nil
}

// NotifyTriggerReceived -
func (hts *HardforkTriggerStub) NotifyTriggerReceived() <-chan struct{} {
	if hts.NotifyTriggerReceivedCalled != nil {
		return hts.NotifyTriggerReceivedCalled()
	}

	return make(chan struct{})
}

// IsInterfaceNil -
func (hts *HardforkTriggerStub) IsInterfaceNil() bool {
	return hts == nil
}
