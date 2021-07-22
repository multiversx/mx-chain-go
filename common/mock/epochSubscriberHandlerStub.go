package mock

// EpochSubscriberHandlerStub -
type EpochSubscriberHandlerStub struct {
	EpochConfirmedCalled func(epoch uint32, timestamp uint64)
}

// EpochConfirmed -
func (eshs *EpochSubscriberHandlerStub) EpochConfirmed(epoch uint32, timestamp uint64) {
	if eshs.EpochConfirmedCalled != nil {
		eshs.EpochConfirmedCalled(epoch, timestamp)
	}
}

// IsInterfaceNil -
func (eshs *EpochSubscriberHandlerStub) IsInterfaceNil() bool {
	return eshs == nil
}
