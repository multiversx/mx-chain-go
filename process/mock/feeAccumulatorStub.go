package mock

import "math/big"

type FeeAccumulatorStub struct {
	CreateBlockStartedCalled    func()
	GetAccumulatedFeesCalled    func() *big.Int
	ProcessTransactionFeeCalled func(cost *big.Int)
}

// CreateBlockStarted -
func (f *FeeAccumulatorStub) CreateBlockStarted() {
	if f.CreateBlockStartedCalled != nil {
		f.CreateBlockStartedCalled()
	}
}

// GetAccumulatedFees -
func (f *FeeAccumulatorStub) GetAccumulatedFees() *big.Int {
	if f.GetAccumulatedFeesCalled != nil {
		return f.GetAccumulatedFeesCalled()
	}
	return big.NewInt(0)
}

// ProcessTransactionFee -
func (f *FeeAccumulatorStub) ProcessTransactionFee(cost *big.Int) {
	if f.ProcessTransactionFeeCalled != nil {
		f.ProcessTransactionFeeCalled(cost)
	}
}

// IsInterfaceNil -
func (f *FeeAccumulatorStub) IsInterfaceNil() bool {
	return f == nil
}
