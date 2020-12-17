package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/process"
)

// EconomicsHandlerStub -
type EconomicsHandlerStub struct {
	MaxGasLimitPerBlockCalled              func() uint64
	ComputeGasLimitCalled                  func(tx process.TransactionWithFeeHandler) uint64
	ComputeMoveBalanceFeeCalled            func(tx process.TransactionWithFeeHandler) *big.Int
	ComputeTxFeeCalled                     func(tx process.TransactionWithFeeHandler) *big.Int
	CheckValidityTxValuesCalled            func(tx process.TransactionWithFeeHandler) error
	DeveloperPercentageCalled              func() float64
	MinGasPriceCalled                      func() uint64
	GasPriceModifierCalled                 func() float64
	LeaderPercentageCalled                 func() float64
	ProtocolSustainabilityPercentageCalled func() float64
	ProtocolSustainabilityAddressCalled    func() string
	MinInflationRateCalled                 func() float64
	MaxInflationRateCalled                 func(year uint32) float64
	GasPerDataByteCalled                   func() uint64
	MinGasLimitCalled                      func() uint64
	GenesisTotalSupplyCalled               func() *big.Int
	ComputeFeeForProcessingCalled          func(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int
	SplitTxGasInCategoriesCalled           func(tx process.TransactionWithFeeHandler) (uint64, uint64)
	GasPriceForProcessingCalled            func(tx process.TransactionWithFeeHandler) uint64
	GasPriceForMoveCalled                  func(tx process.TransactionWithFeeHandler) uint64
	MinGasPriceProcessingCalled            func() uint64
}

// ComputeFeeForProcessing -
func (e *EconomicsHandlerStub) ComputeFeeForProcessing(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	if e.ComputeFeeForProcessingCalled != nil {
		return e.ComputeFeeForProcessingCalled(tx, gasToUse)
	}
	return big.NewInt(0)
}

// LeaderPercentage -
func (e *EconomicsHandlerStub) LeaderPercentage() float64 {
	if e.LeaderPercentageCalled != nil {
		return e.LeaderPercentageCalled()
	}
	return 0.0
}

// ProtocolSustainabilityPercentage -
func (e *EconomicsHandlerStub) ProtocolSustainabilityPercentage() float64 {
	if e.ProtocolSustainabilityAddressCalled != nil {
		return e.ProtocolSustainabilityPercentageCalled()
	}
	return 0.0
}

// ProtocolSustainabilityAddress -
func (e *EconomicsHandlerStub) ProtocolSustainabilityAddress() string {
	if e.ProtocolSustainabilityAddressCalled != nil {
		return e.ProtocolSustainabilityAddressCalled()
	}
	return ""
}

// MinInflationRate -
func (e *EconomicsHandlerStub) MinInflationRate() float64 {
	if e.MinInflationRateCalled != nil {
		return e.MinInflationRateCalled()
	}
	return 0.0
}

// MaxInflationRate -
func (e *EconomicsHandlerStub) MaxInflationRate(year uint32) float64 {
	if e.MaxInflationRateCalled != nil {
		return e.MaxInflationRateCalled(year)
	}
	return 0.0
}

// GasPerDataByte -
func (e *EconomicsHandlerStub) GasPerDataByte() uint64 {
	if e.GasPerDataByteCalled != nil {
		return e.GasPerDataByteCalled()
	}
	return 0
}

// MinGasLimit -
func (e *EconomicsHandlerStub) MinGasLimit() uint64 {
	if e.MinGasLimitCalled != nil {
		return e.MinGasLimitCalled()
	}
	return 0
}

// GenesisTotalSupply -
func (e *EconomicsHandlerStub) GenesisTotalSupply() *big.Int {
	if e.GenesisTotalSupplyCalled != nil {
		return e.GenesisTotalSupplyCalled()
	}
	return big.NewInt(100000000)
}

// GasPriceModifier -
func (e *EconomicsHandlerStub) GasPriceModifier() float64 {
	if e.GasPriceModifierCalled != nil {
		return e.GasPriceModifierCalled()
	}
	return 1.0
}

// MinGasPrice -
func (e *EconomicsHandlerStub) MinGasPrice() uint64 {
	if e.MinGasPriceCalled != nil {
		return e.MinGasPriceCalled()
	}
	return 0
}

// DeveloperPercentage -
func (e *EconomicsHandlerStub) DeveloperPercentage() float64 {
	if e.DeveloperPercentageCalled != nil {
		return e.DeveloperPercentageCalled()
	}

	return 0.0
}

// MaxGasLimitPerBlock -
func (e *EconomicsHandlerStub) MaxGasLimitPerBlock(uint32) uint64 {
	if e.MaxGasLimitPerBlockCalled != nil {
		return e.MaxGasLimitPerBlockCalled()
	}
	return 1000000
}

// ComputeGasLimit -
func (e *EconomicsHandlerStub) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	if e.ComputeGasLimitCalled != nil {
		return e.ComputeGasLimitCalled(tx)
	}
	return 0
}

// ComputeMoveBalanceFee -
func (e *EconomicsHandlerStub) ComputeMoveBalanceFee(tx process.TransactionWithFeeHandler) *big.Int {
	if e.ComputeMoveBalanceFeeCalled != nil {
		return e.ComputeMoveBalanceFeeCalled(tx)
	}
	return big.NewInt(0)
}

// ComputeTxFee -
func (e *EconomicsHandlerStub) ComputeTxFee(tx process.TransactionWithFeeHandler) *big.Int {
	if e.ComputeTxFeeCalled != nil {
		return e.ComputeTxFeeCalled(tx)
	}
	return big.NewInt(0)
}

// CheckValidityTxValues -
func (e *EconomicsHandlerStub) CheckValidityTxValues(tx process.TransactionWithFeeHandler) error {
	if e.CheckValidityTxValuesCalled != nil {
		return e.CheckValidityTxValuesCalled(tx)
	}
	return nil
}

// SplitTxGasInCategories -
func (e *EconomicsHandlerStub) SplitTxGasInCategories(tx process.TransactionWithFeeHandler) (uint64, uint64) {
	if e.SplitTxGasInCategoriesCalled != nil {
		return e.SplitTxGasInCategoriesCalled(tx)
	}

	var processingGas uint64
	if e.ComputeGasLimit(tx) > e.MinGasLimit() {
		processingGas = e.ComputeGasLimit(tx) - e.MinGasLimit()
	}
	processingGas = 0

	return e.MinGasLimit(), processingGas
}

// GasPriceForProcessing -
func (e *EconomicsHandlerStub) GasPriceForProcessing(tx process.TransactionWithFeeHandler) uint64 {
	if e.GasPriceForProcessingCalled != nil {
		return e.GasPriceForProcessingCalled(tx)
	}
	return 1
}

// GasPriceForMove -
func (e *EconomicsHandlerStub) GasPriceForMove(tx process.TransactionWithFeeHandler) uint64 {
	if e.GasPriceForMoveCalled != nil {
		return e.GasPriceForMoveCalled(tx)
	}
	return 100
}

// MinGasPriceProcessing -
func (e *EconomicsHandlerStub) MinGasPriceProcessing() uint64 {
	if e.MinGasPriceProcessingCalled != nil {
		return e.MinGasPriceProcessingCalled()
	}

	return 1
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *EconomicsHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
