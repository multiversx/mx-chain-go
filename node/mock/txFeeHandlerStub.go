package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/process"
)

// EconomicsHandlerStub -
type EconomicsHandlerStub struct {
	MaxGasLimitPerBlockCalled              func() uint64
	SetMinGasPriceCalled                   func(minasPrice uint64)
	SetMinGasLimitCalled                   func(minGasLimit uint64)
	ComputeGasLimitCalled                  func(tx process.TransactionWithFeeHandler) uint64
	ComputeMoveBalanceFeeCalled            func(tx process.TransactionWithFeeHandler) *big.Int
	ComputeTxFeeCalled          func(tx process.TransactionWithFeeHandler) *big.Int
	CheckValidityTxValuesCalled            func(tx process.TransactionWithFeeHandler) error
	DeveloperPercentageCalled              func() float64
	MinGasPriceCalled                      func() uint64
	LeaderPercentageCalled                 func() float64
	ProtocolSustainabilityPercentageCalled func() float64
	ProtocolSustainabilityAddressCalled    func() string
	MinInflationRateCalled                 func() float64
	MaxInflationRateCalled                 func(year uint32) float64
}

// MinGasPrice -
func (fhs *EconomicsHandlerStub) MinGasPrice() uint64 {
	if fhs.MinGasPriceCalled != nil {
		return fhs.MinGasPriceCalled()
	}
	return 0
}

// MinGasLimit will return min gas limit
func (fhs *EconomicsHandlerStub) MinGasLimit() uint64 {
	return 0
}

// GasPerDataByte -
func (fhs *EconomicsHandlerStub) GasPerDataByte() uint64 {
	return 0
}

// DeveloperPercentage -
func (fhs *EconomicsHandlerStub) DeveloperPercentage() float64 {
	return fhs.DeveloperPercentageCalled()
}

// GenesisTotalSupply -
func (fhs *EconomicsHandlerStub) GenesisTotalSupply() *big.Int {
	return big.NewInt(0)
}

// MaxGasLimitPerBlock -
func (fhs *EconomicsHandlerStub) MaxGasLimitPerBlock(uint32) uint64 {
	return fhs.MaxGasLimitPerBlockCalled()
}

// ComputeGasLimit -
func (fhs *EconomicsHandlerStub) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	return fhs.ComputeGasLimitCalled(tx)
}

// ComputeMoveBalanceFee -
func (fhs *EconomicsHandlerStub) ComputeMoveBalanceFee(tx process.TransactionWithFeeHandler) *big.Int {
	return fhs.ComputeMoveBalanceFeeCalled(tx)
}

// ComputeTxFee -
func (fhs *EconomicsHandlerStub) ComputeTxFee(tx process.TransactionWithFeeHandler) *big.Int {
	return fhs.ComputeTxFeeCalled(tx)
}

// CheckValidityTxValues -
func (fhs *EconomicsHandlerStub) CheckValidityTxValues(tx process.TransactionWithFeeHandler) error {
	return fhs.CheckValidityTxValuesCalled(tx)
}

// LeaderPercentage -
func (fhs *EconomicsHandlerStub) LeaderPercentage() float64 {
	if fhs.LeaderPercentageCalled != nil {
		return fhs.LeaderPercentageCalled()
	}

	return 1
}

// ProtocolSustainabilityPercentage will return the protocol sustainability percentage value
func (fhs *EconomicsHandlerStub) ProtocolSustainabilityPercentage() float64 {
	if fhs.ProtocolSustainabilityPercentageCalled != nil {
		return fhs.ProtocolSustainabilityPercentageCalled()
	}

	return 0.1
}

// ProtocolSustainabilityAddress will return the protocol sustainability address
func (fhs *EconomicsHandlerStub) ProtocolSustainabilityAddress() string {
	if fhs.ProtocolSustainabilityAddressCalled != nil {
		return fhs.ProtocolSustainabilityAddressCalled()
	}

	return "1111"
}

// MinInflationRate -
func (fhs *EconomicsHandlerStub) MinInflationRate() float64 {
	if fhs.MinInflationRateCalled != nil {
		return fhs.MinInflationRateCalled()
	}

	return 1
}

// MaxInflationRate -
func (fhs *EconomicsHandlerStub) MaxInflationRate(year uint32) float64 {
	if fhs.MaxInflationRateCalled != nil {
		return fhs.MaxInflationRateCalled(year)
	}

	return 1000000
}

// IsInterfaceNil returns true if there is no value under the interface
func (fhs *EconomicsHandlerStub) IsInterfaceNil() bool {
	return fhs == nil
}
