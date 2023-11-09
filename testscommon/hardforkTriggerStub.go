package testscommon

import "github.com/multiversx/mx-chain-go/update"

// HardforkTriggerStub -
type HardforkTriggerStub struct {
	SetExportFactoryHandlerCalled func(exportFactoryHandler update.ExportFactoryHandler) error
	TriggerCalled                 func(epoch uint32, withEarlyEndOfEpoch bool) error
	IsSelfTriggerCalled           func() bool
	TriggerReceivedCalled         func(payload []byte, data []byte, pkBytes []byte) (bool, error)
	RecordedTriggerMessageCalled  func() ([]byte, bool)
	CreateDataCalled              func() []byte
	AddCloserCalled               func(closer update.Closer) error
	NotifyTriggerReceivedV2Called func() <-chan struct{}
}

// SetExportFactoryHandler -
func (hts *HardforkTriggerStub) SetExportFactoryHandler(exportFactoryHandler update.ExportFactoryHandler) error {
	if hts.SetExportFactoryHandlerCalled != nil {
		return hts.SetExportFactoryHandlerCalled(exportFactoryHandler)
	}

	return nil
}

// Trigger -
func (hts *HardforkTriggerStub) Trigger(epoch uint32, withEarlyEndOfEpoch bool) error {
	if hts.TriggerCalled != nil {
		return hts.TriggerCalled(epoch, withEarlyEndOfEpoch)
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

// NotifyTriggerReceivedV2 -
func (hts *HardforkTriggerStub) NotifyTriggerReceivedV2() <-chan struct{} {
	if hts.NotifyTriggerReceivedV2Called != nil {
		return hts.NotifyTriggerReceivedV2Called()
	}

	return make(chan struct{})
}

// IsInterfaceNil -
func (hts *HardforkTriggerStub) IsInterfaceNil() bool {
	return hts == nil
}
