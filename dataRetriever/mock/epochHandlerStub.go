package mock

type EpochHandlerStub struct {
	EpochCalled func() uint32
}

func (ehs *EpochHandlerStub) Epoch() uint32 {
	if ehs.EpochCalled != nil {
		return ehs.EpochCalled()
	}

	return uint32(0)
}

func (ehs *EpochHandlerStub) IsInterfaceNil() bool {
	return ehs == nil
}
