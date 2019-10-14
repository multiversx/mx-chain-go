package mock

type FeeHandlerMock struct {
}

func NewFeeHandlerMock() *FeeHandlerMock {
	return &FeeHandlerMock{}
}

func (fhm *FeeHandlerMock) MinGasPrice() uint64 {
	return uint64(0)
}

func (fhm *FeeHandlerMock) MinGasLimit() uint64 {
	return uint64(5)
}

func (fhm *FeeHandlerMock) IsInterfaceNil() bool {
	return fhm == nil
}
