package mock

// EpochNotifierStub -
type EpochNotifierStub struct {
	NewEpochConfirmedCalled func(epoch uint32)
}

// NewEpochConfirmed -
func (ens *EpochNotifierStub) NewEpochConfirmed(epoch uint32) {
	if ens.NewEpochConfirmedCalled != nil {
		ens.NewEpochConfirmedCalled(epoch)
	}
}

// IsInterfaceNil -
func (ens *EpochNotifierStub) IsInterfaceNil() bool {
	return ens == nil
}
