package mock

// ManualEpochStartNotifierStub -
type ManualEpochStartNotifierStub struct {
	NewEpochCalled func(epoch uint32)
}

// NewEpoch -
func (mesns *ManualEpochStartNotifierStub) NewEpoch(epoch uint32) {
	if mesns.NewEpochCalled != nil {
		mesns.NewEpochCalled(epoch)
	}
}

// IsInterfaceNil -
func (mesns *ManualEpochStartNotifierStub) IsInterfaceNil() bool {
	return mesns == nil
}
