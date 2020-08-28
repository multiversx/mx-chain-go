package mock

// ManualEpochStartNotifierStub -
type ManualEpochStartNotifierStub struct {
	NewEpochCalled     func(epoch uint32)
	CurrentEpochCalled func() uint32
}

// NewEpoch -
func (mesns *ManualEpochStartNotifierStub) NewEpoch(epoch uint32) {
	if mesns.NewEpochCalled != nil {
		mesns.NewEpochCalled(epoch)
	}
}

// CurrentEpoch -
func (mesns *ManualEpochStartNotifierStub) CurrentEpoch() uint32 {
	if mesns.CurrentEpochCalled != nil {
		return mesns.CurrentEpochCalled()
	}

	return 0
}

// IsInterfaceNil -
func (mesns *ManualEpochStartNotifierStub) IsInterfaceNil() bool {
	return mesns == nil
}
