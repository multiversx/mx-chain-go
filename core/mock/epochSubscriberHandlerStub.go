package mock

// EpochSubscriberHandlerStub -
type EpochSubscriberHandlerStub struct {
	EpochConfirmedCalled func(epoch uint32)
}

// EpochConfirmed -
func (eshs *EpochSubscriberHandlerStub) EpochConfirmed(epoch uint32) {
	if eshs.EpochConfirmedCalled != nil {
		eshs.EpochConfirmedCalled(epoch)
	}
}

// IsInterfaceNil -
func (eshs *EpochSubscriberHandlerStub) IsInterfaceNil() bool {
	return eshs == nil
}
