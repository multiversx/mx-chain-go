package mock

// EpochHandlerStub -
type EpochHandlerStub struct {
	EpochCalled func() uint32
}

// Epoch -
func (ehs *EpochHandlerStub) Epoch() uint32 {
	if ehs.EpochCalled != nil {
		return ehs.EpochCalled()
	}

	return uint32(0)
}

// IsInterfaceNil -
func (ehs *EpochHandlerStub) IsInterfaceNil() bool {
	return ehs == nil
}
