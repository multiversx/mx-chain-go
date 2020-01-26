package mock

type EconomicsStubForTxPool struct {
	minGasPrice uint64
}

func NewEconomicsStubForTxPool(minGasPrice uint64) *EconomicsStubForTxPool {
	return &EconomicsStubForTxPool{
		minGasPrice: minGasPrice,
	}
}

func (stub *EconomicsStubForTxPool) MinGasPrice() uint64 {
	return stub.minGasPrice
}

func (stub *EconomicsStubForTxPool) IsInterfaceNil() bool {
	return stub == nil
}
