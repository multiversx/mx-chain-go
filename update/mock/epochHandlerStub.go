package mock

// EpochHandlerStub -
type EpochHandlerStub struct {
	MetaEpochCalled func() uint32
}

// MetaEpoch -
func (ehs *EpochHandlerStub) MetaEpoch() uint32 {
	if ehs.MetaEpochCalled != nil {
		return ehs.MetaEpochCalled()
	}

	return uint32(0)
}

// IsInterfaceNil -
func (ehs *EpochHandlerStub) IsInterfaceNil() bool {
	return ehs == nil
}
