package mock

type EconomicsStub struct {
	minGasPrice uint64
}

func NewEconomicsStub(minGasPrice uint64) *EconomicsStub {
	return &EconomicsStub{
		minGasPrice: minGasPrice,
	}
}

func (stub *EconomicsStub) MinGasPrice() uint64 {
	return stub.minGasPrice
}
