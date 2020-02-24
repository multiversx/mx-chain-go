package mock

// EpochHandlerMock -
type EpochHandlerMock struct {
	EpochValue uint32
}

// Epoch -
func (ehm *EpochHandlerMock) Epoch() uint32 {
	return ehm.EpochValue
}

// IsInterfaceNil -
func (ehm *EpochHandlerMock) IsInterfaceNil() bool {
	return ehm == nil
}
