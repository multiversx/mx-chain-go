package mock

type FeeHandlerMock struct {
	MinGasPriceCalled      func() uint64
	MinGasLimitForTxCalled func() uint64
	MinTxFeeCalled         func() uint64
}

func (fhm *FeeHandlerMock) MinGasPrice() uint64 {
	return fhm.MinGasPriceCalled()
}

func (fhm *FeeHandlerMock) MinGasLimitForTx() uint64 {
	return fhm.MinGasLimitForTxCalled()
}

func (fhm *FeeHandlerMock) MinTxFee() uint64 {
	return fhm.MinTxFeeCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (fhm *FeeHandlerMock) IsInterfaceNil() bool {
	if fhm == nil {
		return true
	}
	return false
}
