package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/scheduled"
)

// FeeAccumulatorStub is a stub which implements TransactionFeeHandler interface
type FeeAccumulatorStub struct {
	CreateBlockStartedCalled                 func(gasAndFees scheduled.GasAndFees)
	GetAccumulatedFeesCalled                 func() *big.Int
	GetDeveloperFeesCalled                   func() *big.Int
	ProcessTransactionFeeCalled              func(cost *big.Int, devFee *big.Int, hash []byte)
	ProcessTransactionFeeRelayedUserTxCalled func(cost *big.Int, devFee *big.Int, userTxHash []byte, originalTxHash []byte)
	RevertFeesCalled                         func(txHashes [][]byte)
}

// RevertFees -
func (f *FeeAccumulatorStub) RevertFees(txHashes [][]byte) {
	if f.RevertFeesCalled != nil {
		f.RevertFeesCalled(txHashes)
	}
}

// CreateBlockStarted -
func (f *FeeAccumulatorStub) CreateBlockStarted(gasAndFees scheduled.GasAndFees) {
	if f.CreateBlockStartedCalled != nil {
		f.CreateBlockStartedCalled(gasAndFees)
	}
}

// GetAccumulatedFees -
func (f *FeeAccumulatorStub) GetAccumulatedFees() *big.Int {
	if f.GetAccumulatedFeesCalled != nil {
		return f.GetAccumulatedFeesCalled()
	}
	return big.NewInt(0)
}

// GetDeveloperFees -
func (f *FeeAccumulatorStub) GetDeveloperFees() *big.Int {
	if f.GetDeveloperFeesCalled != nil {
		return f.GetDeveloperFeesCalled()
	}
	return big.NewInt(0)
}

// ProcessTransactionFee -
func (f *FeeAccumulatorStub) ProcessTransactionFee(cost *big.Int, devFee *big.Int, txHash []byte) {
	if f.ProcessTransactionFeeCalled != nil {
		f.ProcessTransactionFeeCalled(cost, devFee, txHash)
	}
}

// ProcessTransactionFeeRelayedUserTx -
func (f *FeeAccumulatorStub) ProcessTransactionFeeRelayedUserTx(cost *big.Int, devFee *big.Int, userTxHash []byte, originalTxHash []byte) {
	if f.ProcessTransactionFeeRelayedUserTxCalled != nil {
		f.ProcessTransactionFeeRelayedUserTxCalled(cost, devFee, userTxHash, originalTxHash)
	}
}

// IsInterfaceNil -
func (f *FeeAccumulatorStub) IsInterfaceNil() bool {
	return f == nil
}
