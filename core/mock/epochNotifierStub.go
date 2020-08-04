package mock

// EpochNotifierStub -
type EpochNotifierStub struct {
	EpochConfirmedCalled func(epoch uint32)
}

// EpochConfirmed -
func (ens *EpochNotifierStub) EpochConfirmed(epoch uint32) {
	if ens.EpochConfirmedCalled != nil {
		ens.EpochConfirmedCalled(epoch)
	}
}

// IsInterfaceNil -
func (ens *EpochNotifierStub) IsInterfaceNil() bool {
	return ens == nil
}
