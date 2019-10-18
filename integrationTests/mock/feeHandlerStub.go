package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/process"
)

type FeeHandlerStub struct {
	SetMinGasPriceCalled        func(minasPrice uint64)
	SetMinGasLimitCalled        func(minGasLimit uint64)
	ComputeGasLimitCalled       func(tx process.TransactionWithFeeHandler) uint64
	ComputeFeeCalled            func(tx process.TransactionWithFeeHandler) *big.Int
	CheckValidityTxValuesCalled func(tx process.TransactionWithFeeHandler) error
}

func (fhs *FeeHandlerStub) SetMinGasPrice(minGasPrice uint64) {
	fhs.SetMinGasPriceCalled(minGasPrice)
}

func (fhs *FeeHandlerStub) SetMinGasLimit(minGasLimit uint64) {
	fhs.SetMinGasLimitCalled(minGasLimit)
}

func (fhs *FeeHandlerStub) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	return fhs.ComputeGasLimitCalled(tx)
}

func (fhs *FeeHandlerStub) ComputeFee(tx process.TransactionWithFeeHandler) *big.Int {
	return fhs.ComputeFeeCalled(tx)
}

func (fhs *FeeHandlerStub) CheckValidityTxValues(tx process.TransactionWithFeeHandler) error {
	return fhs.CheckValidityTxValuesCalled(tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (fhs *FeeHandlerStub) IsInterfaceNil() bool {
	if fhs == nil {
		return true
	}
	return false
}
