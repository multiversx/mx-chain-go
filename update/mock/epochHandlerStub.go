package mock

// EpochHandlerStub -
type EpochHandlerStub struct {
	MetaEpochCalled       func() uint32
	ForceEpochStartCalled func()
}

// MetaEpoch -
func (ehs *EpochHandlerStub) MetaEpoch() uint32 {
	if ehs.MetaEpochCalled != nil {
		return ehs.MetaEpochCalled()
	}

	return uint32(0)
}

// ForceEpochStart -
func (ehs *EpochHandlerStub) ForceEpochStart() {
	if ehs.ForceEpochStartCalled != nil {
		ehs.ForceEpochStartCalled()
	}
}

// IsInterfaceNil -
func (ehs *EpochHandlerStub) IsInterfaceNil() bool {
	return ehs == nil
}
