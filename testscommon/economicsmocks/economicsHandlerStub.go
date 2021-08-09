package economicsmocks

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
)

// EconomicsHandlerStub -
type EconomicsHandlerStub struct {
	MaxInflationRateCalled                       func(year uint32) float64
	MinInflationRateCalled                       func() float64
	LeaderPercentageCalled                       func() float64
	ProtocolSustainabilityPercentageCalled       func() float64
	ProtocolSustainabilityAddressCalled          func() string
	SetMaxGasLimitPerBlockCalled                 func(maxGasLimitPerBlock uint64)
	SetMinGasPriceCalled                         func(minGasPrice uint64)
	SetMinGasLimitCalled                         func(minGasLimit uint64)
	ComputeGasLimitCalled                        func(tx data.TransactionWithFeeHandler) uint64
	ComputeMoveBalanceFeeCalled                  func(tx data.TransactionWithFeeHandler) *big.Int
	ComputeTxFeeCalled                           func(tx data.TransactionWithFeeHandler) *big.Int
	CheckValidityTxValuesCalled                  func(tx data.TransactionWithFeeHandler) error
	DeveloperPercentageCalled                    func() float64
	MinGasPriceCalled                            func() uint64
	LeaderPercentageCalled                       func() float64
	ProtocolSustainabilityPercentageCalled       func() float64
	ProtocolSustainabilityAddressCalled          func() string
	MinInflationRateCalled                       func() float64
	MaxInflationRateCalled                       func(year uint32) float64
	GasPriceModifierCalled                       func() float64
	ComputeFeeForProcessingCalled                func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int
	GenesisTotalSupplyCalled                     func() *big.Int
	RewardsTopUpGradientPointCalled              func() *big.Int
	RewardsTopUpFactorCalled                     func() float64
	SplitTxGasInCategoriesCalled                 func(tx data.TransactionWithFeeHandler) (uint64, uint64)
	GasPriceForProcessingCalled                  func(tx data.TransactionWithFeeHandler) uint64
	GasPriceForMoveCalled                        func(tx data.TransactionWithFeeHandler) uint64
	MinGasPriceForProcessingCalled               func() uint64
	ComputeGasUsedAndFeeBasedOnRefundValueCalled func(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsedCalled             func(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeGasLimitBasedOnBalanceCalled          func(tx data.TransactionWithFeeHandler, balance *big.Int) (uint64, error)
}

// ComputeGasLimitBasedOnBalance -
func (ehs *EconomicsHandlerStub) ComputeGasLimitBasedOnBalance(tx data.TransactionWithFeeHandler, balance *big.Int) (uint64, error) {
	if ehs.ComputeGasLimitBasedOnBalanceCalled != nil {
		return ehs.ComputeGasLimitBasedOnBalanceCalled(tx, balance)
	}
	return 0, nil
}

// ComputeFeeForProcessing -
func (ehs *EconomicsHandlerStub) ComputeFeeForProcessing(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	if ehs.ComputeFeeForProcessingCalled != nil {
		return ehs.ComputeFeeForProcessingCalled(tx, gasToUse)
	}
	return big.NewInt(0)
}

// GasPriceModifier -
func (ehs *EconomicsHandlerStub) GasPriceModifier() float64 {
	if ehs.GasPriceModifierCalled != nil {
		return ehs.GasPriceModifierCalled()
	}
	return 1.0
}

// MinGasPrice -
func (ehs *EconomicsHandlerStub) MinGasPrice() uint64 {
	if ehs.MinGasPriceCalled != nil {
		return ehs.MinGasPriceCalled()
	}
	return 0
}

// MinGasLimit will return min gas limit
func (ehs *EconomicsHandlerStub) MinGasLimit() uint64 {
	return 0
}

// GasPerDataByte -
func (ehs *EconomicsHandlerStub) GasPerDataByte() uint64 {
	return 0
}

// DeveloperPercentage -
func (ehs *EconomicsHandlerStub) DeveloperPercentage() float64 {
	return ehs.DeveloperPercentageCalled()
}

// MaxGasLimitPerBlock -
func (ehs *EconomicsHandlerStub) MaxGasLimitPerBlock(uint32) uint64 {
	return ehs.MaxGasLimitPerBlockCalled()
}

// ComputeGasLimit -
func (ehs *EconomicsHandlerStub) ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64 {
	return ehs.ComputeGasLimitCalled(tx)
}

// ComputeMoveBalanceFee -
func (ehs *EconomicsHandlerStub) ComputeMoveBalanceFee(tx data.TransactionWithFeeHandler) *big.Int {
	return ehs.ComputeMoveBalanceFeeCalled(tx)
}

// ComputeTxFee -
func (ehs *EconomicsHandlerStub) ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int {
	return ehs.ComputeTxFeeCalled(tx)
}

// CheckValidityTxValues -
func (ehs *EconomicsHandlerStub) CheckValidityTxValues(tx data.TransactionWithFeeHandler) error {
	if ehs.CheckValidityTxValuesCalled != nil {
		return ehs.CheckValidityTxValuesCalled(tx)
	}

	return nil
}

// LeaderPercentage -
func (ehs *EconomicsHandlerStub) LeaderPercentage() float64 {
	if ehs.LeaderPercentageCalled != nil {
		return ehs.LeaderPercentageCalled()
	}
	return 0.1
}

// ProtocolSustainabilityPercentage will return the protocol sustainability percentage value
func (ehs *EconomicsHandlerStub) ProtocolSustainabilityPercentage() float64 {
	if ehs.ProtocolSustainabilityPercentageCalled != nil {
		return ehs.ProtocolSustainabilityPercentageCalled()
	}
	return 0.1
}

// ProtocolSustainabilityAddress will return the protocol sustainability address
func (ehs *EconomicsHandlerStub) ProtocolSustainabilityAddress() string {
	if ehs.ProtocolSustainabilityAddressCalled != nil {
		return ehs.ProtocolSustainabilityAddressCalled()
	}
	return "erd14uqxan5rgucsf6537ll4vpwyc96z7us5586xhc5euv8w96rsw95sfl6a49"
}

// MinInflationRate -
func (ehs *EconomicsHandlerStub) MinInflationRate() float64 {
	if ehs.MinInflationRateCalled != nil {
		return ehs.MinInflationRateCalled()
	}
	return 0.01
}

// MaxInflationRate -
func (ehs *EconomicsHandlerStub) MaxInflationRate(year uint32) float64 {
	if ehs.MaxInflationRateCalled != nil {
		return ehs.MaxInflationRateCalled(year)
	}
	return 0.1
}

// MinGasPrice -
func (ehs *EconomicsHandlerStub) MinGasPrice() uint64 {
	if ehs.MinGasPriceCalled != nil {
		return ehs.MinGasPriceCalled()
	}
	return 0
}

// MinGasLimit will return min gas limit
func (ehs *EconomicsHandlerStub) MinGasLimit() uint64 {
	return 0
}

// GasPerDataByte -
func (ehs *EconomicsHandlerStub) GasPerDataByte() uint64 {
	return 0
}

// DeveloperPercentage -
func (ehs *EconomicsHandlerStub) DeveloperPercentage() float64 {
	if ehs.DeveloperPercentageCalled != nil {
		return ehs.DeveloperPercentageCalled()
	}
	return 0
}

// GenesisTotalSupply -
func (ehs *EconomicsHandlerStub) GenesisTotalSupply() *big.Int {
	if ehs.GenesisTotalSupplyCalled != nil {
		return ehs.GenesisTotalSupplyCalled()
	}
	return big.NewInt(0)
}

// SetMaxGasLimitPerBlock -
func (ehs *EconomicsHandlerStub) SetMaxGasLimitPerBlock(maxGasLimitPerBlock uint64) {
	if ehs.SetMaxGasLimitPerBlockCalled != nil {
		ehs.SetMaxGasLimitPerBlockCalled(maxGasLimitPerBlock)
	}
}

// SetMinGasPrice -
func (ehs *EconomicsHandlerStub) SetMinGasPrice(minGasPrice uint64) {
	if ehs.SetMinGasPriceCalled != nil {
		ehs.SetMinGasPriceCalled(minGasPrice)
	}
}

// SetMinGasLimit -
func (ehs *EconomicsHandlerStub) SetMinGasLimit(minGasLimit uint64) {
	if ehs.SetMinGasLimitCalled != nil {
		ehs.SetMinGasLimitCalled(minGasLimit)
	}
}

// MaxGasLimitPerBlock -
func (ehs *EconomicsHandlerStub) MaxGasLimitPerBlock(shard uint32) uint64 {
	if ehs.MaxGasLimitPerBlockCalled != nil {
		return ehs.MaxGasLimitPerBlockCalled(shard)
	}
	return 1500000000
}

// ComputeGasLimit -
func (ehs *EconomicsHandlerStub) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	if ehs.ComputeGasLimitCalled != nil {
		return ehs.ComputeGasLimitCalled(tx)
	}
	return 0
}

// ComputeFee -
func (ehs *EconomicsHandlerStub) ComputeFee(tx process.TransactionWithFeeHandler) *big.Int {
	if ehs.ComputeFeeCalled != nil {
		return ehs.ComputeFeeCalled(tx)
	}
	return big.NewInt(0)
}

// CheckValidityTxValues -
func (ehs *EconomicsHandlerStub) CheckValidityTxValues(tx process.TransactionWithFeeHandler) error {
	if ehs.CheckValidityTxValuesCalled != nil {
		return ehs.CheckValidityTxValuesCalled(tx)
	}
	return nil
}

// ComputeMoveBalanceFee -
func (ehs *EconomicsHandlerStub) ComputeMoveBalanceFee(tx process.TransactionWithFeeHandler) *big.Int {
	if ehs.ComputeMoveBalanceFeeCalled != nil {
		return ehs.ComputeMoveBalanceFeeCalled(tx)
	}
	return big.NewInt(0)

}

// ComputeGasLimitBasedOnBalance -
func (ehs *EconomicsHandlerStub) ComputeGasLimitBasedOnBalance(tx process.TransactionWithFeeHandler, balance *big.Int) (uint64, error) {
	if ehs.ComputeGasLimitBasedOnBalanceCalled != nil {
		return ehs.ComputeGasLimitBasedOnBalanceCalled(tx, balance)
	}
	return 0, nil
}

// ComputeTxFee -
func (ehs *EconomicsHandlerStub) ComputeTxFee(tx process.TransactionWithFeeHandler) *big.Int {
	if ehs.ComputeTxFeeCalled != nil {
		return ehs.ComputeTxFeeCalled(tx)
	}
	return big.NewInt(0)
}

// RewardsTopUpGradientPoint -
func (ehs *EconomicsHandlerStub) RewardsTopUpGradientPoint() *big.Int {
	if ehs.RewardsTopUpGradientPointCalled != nil {
		return ehs.RewardsTopUpGradientPointCalled()
	}
	return big.NewInt(0)
}

// RewardsTopUpFactor -
func (ehs *EconomicsHandlerStub) RewardsTopUpFactor() float64 {
	if ehs.RewardsTopUpFactorCalled != nil {
		return ehs.RewardsTopUpFactorCalled()
	}
	return 0
}

// ComputeFeeForProcessing -
func (ehs *EconomicsHandlerStub) ComputeFeeForProcessing(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	if ehs.ComputeFeeForProcessingCalled != nil {
		return ehs.ComputeFeeForProcessingCalled(tx, gasToUse)
	}
	return big.NewInt(0)
}

// GasPriceModifier -
func (ehs *EconomicsHandlerStub) GasPriceModifier() float64 {
	if ehs.GasPriceModifierCalled != nil {
		return ehs.GasPriceModifierCalled()
	}
	return 0
}

// SplitTxGasInCategories -
func (ehs *EconomicsHandlerStub) SplitTxGasInCategories(tx data.TransactionWithFeeHandler) (uint64, uint64) {
	if ehs.SplitTxGasInCategoriesCalled != nil {
		return ehs.SplitTxGasInCategoriesCalled(tx)
	}
	return 0, 0
}

// GasPriceForProcessing -
func (ehs *EconomicsHandlerStub) GasPriceForProcessing(tx data.TransactionWithFeeHandler) uint64 {
	if ehs.GasPriceForProcessingCalled != nil {
		return ehs.GasPriceForProcessingCalled(tx)
	}
	return 0
}

// GasPriceForMove -
func (ehs *EconomicsHandlerStub) GasPriceForMove(tx data.TransactionWithFeeHandler) uint64 {
	if ehs.GasPriceForMoveCalled != nil {
		return ehs.GasPriceForMoveCalled(tx)
	}
	return 0
}

// MinGasPriceForProcessing -
func (ehs *EconomicsHandlerStub) MinGasPriceForProcessing() uint64 {
	if ehs.MinGasPriceForProcessingCalled != nil {
		return ehs.MinGasPriceForProcessingCalled()
	}
	return 0
}

// ComputeGasUsedAndFeeBasedOnRefundValue -
func (ehs *EconomicsHandlerStub) ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int) {
	if ehs.ComputeGasUsedAndFeeBasedOnRefundValueCalled != nil {
		return ehs.ComputeGasUsedAndFeeBasedOnRefundValueCalled(tx, refundValue)
	}
	return 0, big.NewInt(0)
}

// ComputeTxFeeBasedOnGasUsed -
func (ehs *EconomicsHandlerStub) ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
	if ehs.ComputeTxFeeBasedOnGasUsedCalled != nil {
		return ehs.ComputeTxFeeBasedOnGasUsedCalled(tx, gasUsed)
	}
	return big.NewInt(0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ehs *EconomicsHandlerStub) IsInterfaceNil() bool {
	return ehs == nil
}
