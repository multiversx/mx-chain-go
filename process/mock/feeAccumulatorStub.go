package mock

import "math/big"

// FeeAccumulatorStub is a stub which implements TransactionFeeHandler interface
type FeeAccumulatorStub struct {
	CreateBlockStartedCalled    func()
	GetAccumulatedFeesCalled    func() *big.Int
	ProcessTransactionFeeCalled func(cost *big.Int, hash []byte)
	RevertFeesCalled            func(txHashes [][]byte)
}

// RevertFees -
func (f *FeeAccumulatorStub) RevertFees(txHashes [][]byte) {
	if f.RevertFeesCalled != nil {
		f.RevertFeesCalled(txHashes)
	}
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
func (f *FeeAccumulatorStub) ProcessTransactionFee(cost *big.Int, txHash []byte) {
	if f.ProcessTransactionFeeCalled != nil {
		f.ProcessTransactionFeeCalled(cost, txHash)
	}
}

// IsInterfaceNil -
func (f *FeeAccumulatorStub) IsInterfaceNil() bool {
	return f == nil
}
