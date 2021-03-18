package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/process"
)

// EconomicsHandlerStub -
type EconomicsHandlerStub struct {
	MaxGasLimitPerBlockCalled                    func() uint64
	SetMinGasPriceCalled                         func(minasPrice uint64)
	SetMinGasLimitCalled                         func(minGasLimit uint64)
	ComputeGasLimitCalled                        func(tx process.TransactionWithFeeHandler) uint64
	ComputeMoveBalanceFeeCalled                  func(tx process.TransactionWithFeeHandler) *big.Int
	ComputeTxFeeCalled                           func(tx process.TransactionWithFeeHandler) *big.Int
	CheckValidityTxValuesCalled                  func(tx process.TransactionWithFeeHandler) error
	DeveloperPercentageCalled                    func() float64
	MinGasPriceCalled                            func() uint64
	LeaderPercentageCalled                       func() float64
	ProtocolSustainabilityPercentageCalled       func() float64
	ProtocolSustainabilityAddressCalled          func() string
	MinInflationRateCalled                       func() float64
	MaxInflationRateCalled                       func(year uint32) float64
	ComputeFeeForProcessingCalled                func(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int
	GasPriceModifierCalled                       func() float64
	SplitTxGasInCategoriesCalled                 func(tx process.TransactionWithFeeHandler) (uint64, uint64)
	GasPriceForProcessingCalled                  func(tx process.TransactionWithFeeHandler) uint64
	GasPriceForMoveCalled                        func(tx process.TransactionWithFeeHandler) uint64
	MinGasPriceForProcessingCalled               func() uint64
	ComputeGasUsedAndFeeBasedOnRefundValueCalled func(tx process.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsedCalled             func(tx process.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	RewardsTopUpGradientPointCalled              func() *big.Int
	RewardsTopUpFactorCalled                     func() float64
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

// ComputeFeeForProcessing -
func (fhs *EconomicsHandlerStub) ComputeFeeForProcessing(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	if fhs.ComputeFeeForProcessingCalled != nil {
		return fhs.ComputeFeeForProcessingCalled(tx, gasToUse)
	}
	return big.NewInt(0)
}

// GasPriceModifier -
func (fhs *EconomicsHandlerStub) GasPriceModifier() float64 {
	if fhs.GasPriceModifierCalled != nil {
		return fhs.GasPriceModifierCalled()
	}
	return 0
}

// SplitTxGasInCategories -
func (fhs *EconomicsHandlerStub) SplitTxGasInCategories(tx process.TransactionWithFeeHandler) (uint64, uint64) {
	if fhs.SplitTxGasInCategoriesCalled != nil {
		return fhs.SplitTxGasInCategoriesCalled(tx)
	}
	return 0, 0
}

// GasPriceForProcessing -
func (fhs *EconomicsHandlerStub) GasPriceForProcessing(tx process.TransactionWithFeeHandler) uint64 {
	if fhs.GasPriceForProcessingCalled != nil {
		return fhs.GasPriceForProcessingCalled(tx)
	}
	return 0
}

// GasPriceForMove -
func (fhs *EconomicsHandlerStub) GasPriceForMove(tx process.TransactionWithFeeHandler) uint64 {
	if fhs.GasPriceForMoveCalled != nil {
		return fhs.GasPriceForMoveCalled(tx)
	}
	return 0
}

// MinGasPriceForProcessing -
func (fhs *EconomicsHandlerStub) MinGasPriceForProcessing() uint64 {
	if fhs.MinGasPriceForProcessingCalled != nil {
		return fhs.MinGasPriceForProcessingCalled()
	}
	return 0
}

// ComputeGasUsedAndFeeBasedOnRefundValue -
func (fhs *EconomicsHandlerStub) ComputeGasUsedAndFeeBasedOnRefundValue(tx process.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int) {
	if fhs.ComputeGasUsedAndFeeBasedOnRefundValueCalled != nil {
		return fhs.ComputeGasUsedAndFeeBasedOnRefundValueCalled(tx, refundValue)
	}
	return 0, big.NewInt(0)
}

// ComputeTxFeeBasedOnGasUsed -
func (fhs *EconomicsHandlerStub) ComputeTxFeeBasedOnGasUsed(tx process.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
	if fhs.ComputeTxFeeBasedOnGasUsedCalled != nil {
		return fhs.ComputeTxFeeBasedOnGasUsedCalled(tx, gasUsed)
	}
	return big.NewInt(0)
}

// RewardsTopUpGradientPoint -
func (fhs *EconomicsHandlerStub) RewardsTopUpGradientPoint() *big.Int {
	if fhs.RewardsTopUpGradientPointCalled != nil {
		return fhs.RewardsTopUpGradientPointCalled()
	}
	return big.NewInt(0)
}

// RewardsTopUpFactor -
func (fhs *EconomicsHandlerStub) RewardsTopUpFactor() float64 {
	if fhs.RewardsTopUpFactorCalled != nil {
		return fhs.RewardsTopUpFactorCalled()
	}
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (fhs *EconomicsHandlerStub) IsInterfaceNil() bool {
	return fhs == nil
}
