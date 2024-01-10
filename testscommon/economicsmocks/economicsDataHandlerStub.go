package economicsmocks

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
)

// EconomicsHandlerStub -
type EconomicsHandlerStub struct {
	MaxGasLimitPerBlockCalled                           func(shardID uint32) uint64
	MaxGasLimitPerMiniBlockCalled                       func() uint64
	MaxGasLimitPerBlockForSafeCrossShardCalled          func() uint64
	MaxGasLimitPerMiniBlockForSafeCrossShardCalled      func() uint64
	MaxGasLimitPerTxCalled                              func() uint64
	ComputeGasLimitCalled                               func(tx data.TransactionWithFeeHandler) uint64
	ComputeMoveBalanceFeeCalled                         func(tx data.TransactionWithFeeHandler) *big.Int
	ComputeTxFeeCalled                                  func(tx data.TransactionWithFeeHandler) *big.Int
	CheckValidityTxValuesCalled                         func(tx data.TransactionWithFeeHandler) error
	DeveloperPercentageCalled                           func() float64
	MinGasPriceCalled                                   func() uint64
	GasPriceModifierCalled                              func() float64
	LeaderPercentageCalled                              func() float64
	ProtocolSustainabilityPercentageCalled              func() float64
	ProtocolSustainabilityAddressCalled                 func() string
	MinInflationRateCalled                              func() float64
	MaxInflationRateCalled                              func(year uint32) float64
	GasPerDataByteCalled                                func() uint64
	MinGasLimitCalled                                   func() uint64
	ExtraGasLimitGuardedTxCalled                        func() uint64
	MaxGasPriceSetGuardianCalled                        func() uint64
	GenesisTotalSupplyCalled                            func() *big.Int
	ComputeFeeForProcessingCalled                       func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int
	RewardsTopUpGradientPointCalled                     func() *big.Int
	RewardsTopUpFactorCalled                            func() float64
	SplitTxGasInCategoriesCalled                        func(tx data.TransactionWithFeeHandler) (uint64, uint64)
	GasPriceForProcessingCalled                         func(tx data.TransactionWithFeeHandler) uint64
	GasPriceForMoveCalled                               func(tx data.TransactionWithFeeHandler) uint64
	MinGasPriceProcessingCalled                         func() uint64
	ComputeGasUsedAndFeeBasedOnRefundValueCalled        func(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsedCalled                    func(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeGasLimitBasedOnBalanceCalled                 func(tx data.TransactionWithFeeHandler, balance *big.Int) (uint64, error)
	SetStatusHandlerCalled                              func(statusHandler core.AppStatusHandler) error
	ComputeTxFeeInEpochCalled                           func(tx data.TransactionWithFeeHandler, epoch uint32) *big.Int
	ComputeGasLimitInEpochCalled                        func(tx data.TransactionWithFeeHandler, epoch uint32) uint64
	ComputeGasUsedAndFeeBasedOnRefundValueInEpochCalled func(tx data.TransactionWithFeeHandler, refundValue *big.Int, epoch uint32) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsedInEpochCalled             func(tx data.TransactionWithFeeHandler, gasUsed uint64, epoch uint32) *big.Int
}

// ComputeFeeForProcessing -
func (e *EconomicsHandlerStub) ComputeFeeForProcessing(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	if e.ComputeFeeForProcessingCalled != nil {
		return e.ComputeFeeForProcessingCalled(tx, gasToUse)
	}
	return big.NewInt(0)
}

// ComputeGasLimitBasedOnBalance -
func (e *EconomicsHandlerStub) ComputeGasLimitBasedOnBalance(tx data.TransactionWithFeeHandler, balance *big.Int) (uint64, error) {
	if e.ComputeGasLimitBasedOnBalanceCalled != nil {
		return e.ComputeGasLimitBasedOnBalanceCalled(tx, balance)
	}
	return 0, nil
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

// ExtraGasLimitGuardedTx -
func (e *EconomicsHandlerStub) ExtraGasLimitGuardedTx() uint64 {
	if e.ExtraGasLimitGuardedTxCalled != nil {
		return e.ExtraGasLimitGuardedTxCalled()
	}
	return 0
}

// MaxGasPriceSetGuardian -
func (e *EconomicsHandlerStub) MaxGasPriceSetGuardian() uint64 {
	if e.MaxGasPriceSetGuardianCalled != nil {
		return e.MaxGasPriceSetGuardianCalled()
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
func (e *EconomicsHandlerStub) MaxGasLimitPerBlock(shardID uint32) uint64 {
	if e.MaxGasLimitPerBlockCalled != nil {
		return e.MaxGasLimitPerBlockCalled(shardID)
	}
	return 1000000
}

// MaxGasLimitPerMiniBlock -
func (e *EconomicsHandlerStub) MaxGasLimitPerMiniBlock(uint32) uint64 {
	if e.MaxGasLimitPerMiniBlockCalled != nil {
		return e.MaxGasLimitPerMiniBlockCalled()
	}
	return 1000000
}

// MaxGasLimitPerBlockForSafeCrossShard -
func (e *EconomicsHandlerStub) MaxGasLimitPerBlockForSafeCrossShard() uint64 {
	if e.MaxGasLimitPerBlockForSafeCrossShardCalled != nil {
		return e.MaxGasLimitPerBlockForSafeCrossShardCalled()
	}
	return 1000000
}

// MaxGasLimitPerMiniBlockForSafeCrossShard -
func (e *EconomicsHandlerStub) MaxGasLimitPerMiniBlockForSafeCrossShard() uint64 {
	if e.MaxGasLimitPerMiniBlockForSafeCrossShardCalled != nil {
		return e.MaxGasLimitPerMiniBlockForSafeCrossShardCalled()
	}
	return 1000000
}

// MaxGasLimitPerTx -
func (e *EconomicsHandlerStub) MaxGasLimitPerTx() uint64 {
	if e.MaxGasLimitPerTxCalled != nil {
		return e.MaxGasLimitPerTxCalled()
	}
	return 1000000
}

// ComputeGasLimit -
func (e *EconomicsHandlerStub) ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64 {
	if e.ComputeGasLimitCalled != nil {
		return e.ComputeGasLimitCalled(tx)
	}
	return 0
}

// ComputeMoveBalanceFee -
func (e *EconomicsHandlerStub) ComputeMoveBalanceFee(tx data.TransactionWithFeeHandler) *big.Int {
	if e.ComputeMoveBalanceFeeCalled != nil {
		return e.ComputeMoveBalanceFeeCalled(tx)
	}
	return big.NewInt(0)
}

// ComputeTxFee -
func (e *EconomicsHandlerStub) ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int {
	if e.ComputeTxFeeCalled != nil {
		return e.ComputeTxFeeCalled(tx)
	}
	return big.NewInt(0)
}

// CheckValidityTxValues -
func (e *EconomicsHandlerStub) CheckValidityTxValues(tx data.TransactionWithFeeHandler) error {
	if e.CheckValidityTxValuesCalled != nil {
		return e.CheckValidityTxValuesCalled(tx)
	}
	return nil
}

// RewardsTopUpGradientPoint -
func (e *EconomicsHandlerStub) RewardsTopUpGradientPoint() *big.Int {
	if e.RewardsTopUpGradientPointCalled != nil {
		return e.RewardsTopUpGradientPointCalled()
	}

	return big.NewInt(0)
}

// RewardsTopUpFactor -
func (e *EconomicsHandlerStub) RewardsTopUpFactor() float64 {
	if e.RewardsTopUpFactorCalled != nil {
		return e.RewardsTopUpFactorCalled()
	}

	return 0
}

// SplitTxGasInCategories -
func (e *EconomicsHandlerStub) SplitTxGasInCategories(tx data.TransactionWithFeeHandler) (uint64, uint64) {
	if e.SplitTxGasInCategoriesCalled != nil {
		return e.SplitTxGasInCategoriesCalled(tx)
	}

	processingGas := uint64(0)
	if e.ComputeGasLimit(tx) > e.MinGasLimit() {
		processingGas = e.ComputeGasLimit(tx) - e.MinGasLimit()
	}

	return e.MinGasLimit(), processingGas
}

// GasPriceForProcessing -
func (e *EconomicsHandlerStub) GasPriceForProcessing(tx data.TransactionWithFeeHandler) uint64 {
	if e.GasPriceForProcessingCalled != nil {
		return e.GasPriceForProcessingCalled(tx)
	}
	return 1
}

// GasPriceForMove -
func (e *EconomicsHandlerStub) GasPriceForMove(tx data.TransactionWithFeeHandler) uint64 {
	if e.GasPriceForMoveCalled != nil {
		return e.GasPriceForMoveCalled(tx)
	}
	return 100
}

// MinGasPriceForProcessing -
func (e *EconomicsHandlerStub) MinGasPriceForProcessing() uint64 {
	if e.MinGasPriceProcessingCalled != nil {
		return e.MinGasPriceProcessingCalled()
	}

	return 1
}

// ComputeGasUsedAndFeeBasedOnRefundValue -
func (e *EconomicsHandlerStub) ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int) {
	if e.ComputeGasUsedAndFeeBasedOnRefundValueCalled != nil {
		return e.ComputeGasUsedAndFeeBasedOnRefundValueCalled(tx, refundValue)
	}

	return 0, nil
}

// ComputeTxFeeBasedOnGasUsed -
func (e *EconomicsHandlerStub) ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
	if e.ComputeTxFeeBasedOnGasUsedCalled != nil {
		return e.ComputeTxFeeBasedOnGasUsedCalled(tx, gasUsed)
	}

	return nil
}

// SetStatusHandler -
func (e *EconomicsHandlerStub) SetStatusHandler(statusHandler core.AppStatusHandler) error {
	if e.SetStatusHandlerCalled != nil {
		return e.SetStatusHandlerCalled(statusHandler)
	}
	return nil
}

// ComputeTxFeeInEpoch -
func (e *EconomicsHandlerStub) ComputeTxFeeInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) *big.Int {
	if e.ComputeTxFeeInEpochCalled != nil {
		return e.ComputeTxFeeInEpochCalled(tx, epoch)
	}
	return nil
}

// ComputeGasLimitInEpoch -
func (e *EconomicsHandlerStub) ComputeGasLimitInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) uint64 {
	if e.ComputeGasLimitInEpochCalled != nil {
		return e.ComputeGasLimitInEpochCalled(tx, epoch)
	}
	return 0
}

// ComputeGasUsedAndFeeBasedOnRefundValueInEpoch -
func (e *EconomicsHandlerStub) ComputeGasUsedAndFeeBasedOnRefundValueInEpoch(tx data.TransactionWithFeeHandler, refundValue *big.Int, epoch uint32) (uint64, *big.Int) {
	if e.ComputeGasUsedAndFeeBasedOnRefundValueInEpochCalled != nil {
		return e.ComputeGasUsedAndFeeBasedOnRefundValueInEpochCalled(tx, refundValue, epoch)
	}
	return 0, big.NewInt(0)
}

// ComputeTxFeeBasedOnGasUsedInEpoch -
func (e *EconomicsHandlerStub) ComputeTxFeeBasedOnGasUsedInEpoch(tx data.TransactionWithFeeHandler, gasUsed uint64, epoch uint32) *big.Int {
	if e.ComputeTxFeeBasedOnGasUsedInEpochCalled != nil {
		return e.ComputeTxFeeBasedOnGasUsedInEpochCalled(tx, gasUsed, epoch)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *EconomicsHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
