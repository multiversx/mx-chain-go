package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/process"
)

// FeeHandlerStub -
type FeeHandlerStub struct {
	MaxGasLimitPerBlockCalled   func() uint64
	SetMinGasPriceCalled        func(minasPrice uint64)
	SetMinGasLimitCalled        func(minGasLimit uint64)
	ComputeGasLimitCalled       func(tx process.TransactionWithFeeHandler) uint64
	ComputeMoveBalanceFeeCalled func(tx process.TransactionWithFeeHandler) *big.Int
	CheckValidityTxValuesCalled func(tx process.TransactionWithFeeHandler) error
	DeveloperPercentageCalled   func() float64
	MinGasPriceCalled           func() uint64
}

// MinGasPrice -
func (fhs *FeeHandlerStub) MinGasPrice() uint64 {
	if fhs.MinGasPriceCalled != nil {
		return fhs.MinGasPriceCalled()
	}
	return 0
}

// DeveloperPercentage -
func (fhs *FeeHandlerStub) DeveloperPercentage() float64 {
	return fhs.DeveloperPercentageCalled()
}

// MaxGasLimitPerBlock -
func (fhs *FeeHandlerStub) MaxGasLimitPerBlock(uint32) uint64 {
	return fhs.MaxGasLimitPerBlockCalled()
}

// ComputeGasLimit -
func (fhs *FeeHandlerStub) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	return fhs.ComputeGasLimitCalled(tx)
}

// ComputeMoveBalanceFee -
func (fhs *FeeHandlerStub) ComputeMoveBalanceFee(tx process.TransactionWithFeeHandler) *big.Int {
	return fhs.ComputeMoveBalanceFeeCalled(tx)
}

// CheckValidityTxValues -
func (fhs *FeeHandlerStub) CheckValidityTxValues(tx process.TransactionWithFeeHandler) error {
	return fhs.CheckValidityTxValuesCalled(tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (fhs *FeeHandlerStub) IsInterfaceNil() bool {
	return fhs == nil
}
