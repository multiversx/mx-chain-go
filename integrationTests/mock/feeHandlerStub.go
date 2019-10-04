package mock

type FeeHandlerStub struct {
	MinGasPriceCalled      func() uint64
	MinGasLimitForTxCalled func() uint64
	MinTxFeeCalled         func() uint64
}

func (fhs *FeeHandlerStub) MinGasPrice() uint64 {
	return fhs.MinGasPriceCalled()
}

func (fhs *FeeHandlerStub) MinGasLimitForTx() uint64 {
	return fhs.MinGasLimitForTxCalled()
}

func (fhs *FeeHandlerStub) MinTxFee() uint64 {
	return fhs.MinTxFeeCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (fhs *FeeHandlerStub) IsInterfaceNil() bool {
	if fhs == nil {
		return true
	}
	return false
}
