package economicsmocks

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
)

// EconomicsHandlerMock -
type EconomicsHandlerMock struct {
	MaxInflationRateCalled                              func(year uint32) float64
	MinInflationRateCalled                              func() float64
	LeaderPercentageCalled                              func() float64
	ProtocolSustainabilityPercentageCalled              func() float64
	ProtocolSustainabilityAddressCalled                 func() string
	SetMaxGasLimitPerBlockCalled                        func(maxGasLimitPerBlock uint64)
	SetMinGasPriceCalled                                func(minGasPrice uint64)
	SetMinGasLimitCalled                                func(minGasLimit uint64)
	MaxGasLimitPerBlockCalled                           func(shardID uint32) uint64
	MaxGasLimitPerMiniBlockCalled                       func(shardID uint32) uint64
	MaxGasLimitPerBlockForSafeCrossShardCalled          func() uint64
	MaxGasLimitPerMiniBlockForSafeCrossShardCalled      func() uint64
	MaxGasLimitPerTxCalled                              func() uint64
	ComputeGasLimitCalled                               func(tx data.TransactionWithFeeHandler) uint64
	ComputeFeeCalled                                    func(tx data.TransactionWithFeeHandler) *big.Int
	CheckValidityTxValuesCalled                         func(tx data.TransactionWithFeeHandler) error
	ComputeMoveBalanceFeeCalled                         func(tx data.TransactionWithFeeHandler) *big.Int
	ComputeMoveBalanceFeeInEpochCalled                  func(tx data.TransactionWithFeeHandler, epoch uint32) *big.Int
	ComputeTxFeeCalled                                  func(tx data.TransactionWithFeeHandler) *big.Int
	DeveloperPercentageCalled                           func() float64
	MinGasPriceCalled                                   func() uint64
	MinGasLimitCalled                                   func() uint64
	GasPerDataByteCalled                                func() uint64
	MaxGasHigherFactorAcceptedCalled                    func() uint64
	RewardsTopUpGradientPointCalled                     func() *big.Int
	RewardsTopUpFactorCalled                            func() float64
	ComputeFeeForProcessingCalled                       func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int
	GasPriceModifierCalled                              func() float64
	SplitTxGasInCategoriesCalled                        func(tx data.TransactionWithFeeHandler) (uint64, uint64)
	GasPriceForProcessingCalled                         func(tx data.TransactionWithFeeHandler) uint64
	GasPriceForMoveCalled                               func(tx data.TransactionWithFeeHandler) uint64
	MinGasPriceForProcessingCalled                      func() uint64
	ComputeGasUsedAndFeeBasedOnRefundValueCalled        func(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsedCalled                    func(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeGasLimitBasedOnBalanceCalled                 func(tx data.TransactionWithFeeHandler, balance *big.Int) (uint64, error)
	SetStatusHandlerCalled                              func(statusHandler core.AppStatusHandler) error
	ComputeTxFeeInEpochCalled                           func(tx data.TransactionWithFeeHandler, epoch uint32) *big.Int
	ComputeGasLimitInEpochCalled                        func(tx data.TransactionWithFeeHandler, epoch uint32) uint64
	ComputeGasUsedAndFeeBasedOnRefundValueInEpochCalled func(tx data.TransactionWithFeeHandler, refundValue *big.Int, epoch uint32) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsedInEpochCalled             func(tx data.TransactionWithFeeHandler, gasUsed uint64, epoch uint32) *big.Int
	GenesisTotalSupplyCalled                            func() *big.Int
	MaxGasPriceSetGuardianCalled                        func() uint64
	LeaderPercentageInEpochCalled                       func(epoch uint32) float64
	DeveloperPercentageInEpochCalled                    func(epoch uint32) float64
	ProtocolSustainabilityPercentageInEpochCalled       func(epoch uint32) float64
	ProtocolSustainabilityAddressInEpochCalled          func(epoch uint32) string
	RewardsTopUpGradientPointInEpochCalled              func(epoch uint32) *big.Int
	RewardsTopUpFactorInEpochCalled                     func(epoch uint32) float64
}

// ComputeGasUnitsFromRefundValue -
func (ehm *EconomicsHandlerMock) ComputeGasUnitsFromRefundValue(_ data.TransactionWithFeeHandler, _ *big.Int, _ uint32) uint64 {
	return 0
}

// LeaderPercentage -
func (ehm *EconomicsHandlerMock) LeaderPercentage() float64 {
	return ehm.LeaderPercentageCalled()
}

// ProtocolSustainabilityPercentage will return the protocol sustainability percentage value
func (ehm *EconomicsHandlerMock) ProtocolSustainabilityPercentage() float64 {
	return ehm.ProtocolSustainabilityPercentageCalled()
}

// ProtocolSustainabilityAddress will return the protocol sustainability address
func (ehm *EconomicsHandlerMock) ProtocolSustainabilityAddress() string {
	return ehm.ProtocolSustainabilityAddressCalled()
}

// MinInflationRate -
func (ehm *EconomicsHandlerMock) MinInflationRate() float64 {
	return ehm.MinInflationRateCalled()
}

// MaxInflationRate -
func (ehm *EconomicsHandlerMock) MaxInflationRate(year uint32) float64 {
	return ehm.MaxInflationRateCalled(year)
}

// MinGasPrice -
func (ehm *EconomicsHandlerMock) MinGasPrice() uint64 {
	if ehm.MinGasPriceCalled != nil {
		return ehm.MinGasPriceCalled()
	}
	return 0
}

// MinGasLimit will return min gas limit
func (ehm *EconomicsHandlerMock) MinGasLimit() uint64 {
	if ehm.MinGasLimitCalled != nil {
		return ehm.MinGasLimitCalled()
	}
	return 0
}

// ExtraGasLimitGuardedTx -
func (ehm *EconomicsHandlerMock) ExtraGasLimitGuardedTx() uint64 {
	return 0
}

// MaxGasPriceSetGuardian -
func (ehm *EconomicsHandlerMock) MaxGasPriceSetGuardian() uint64 {
	if ehm.MaxGasPriceSetGuardianCalled != nil {
		return ehm.MaxGasPriceSetGuardianCalled()
	}
	return 0
}

// GasPerDataByte -
func (ehm *EconomicsHandlerMock) GasPerDataByte() uint64 {
	return 0
}

// DeveloperPercentage -
func (ehm *EconomicsHandlerMock) DeveloperPercentage() float64 {
	if ehm.DeveloperPercentageCalled != nil {
		return ehm.DeveloperPercentageCalled()
	}
	return 0.0
}

// GenesisTotalSupply -
func (ehm *EconomicsHandlerMock) GenesisTotalSupply() *big.Int {
	if ehm.GenesisTotalSupplyCalled != nil {
		return ehm.GenesisTotalSupplyCalled()
	}
	return big.NewInt(0)
}

// SetMaxGasLimitPerBlock -
func (ehm *EconomicsHandlerMock) SetMaxGasLimitPerBlock(maxGasLimitPerBlock uint64) {
	ehm.SetMaxGasLimitPerBlockCalled(maxGasLimitPerBlock)
}

// SetMinGasPrice -
func (ehm *EconomicsHandlerMock) SetMinGasPrice(minGasPrice uint64) {
	ehm.SetMinGasPriceCalled(minGasPrice)
}

// SetMinGasLimit -
func (ehm *EconomicsHandlerMock) SetMinGasLimit(minGasLimit uint64) {
	ehm.SetMinGasLimitCalled(minGasLimit)
}

// MaxGasLimitPerBlock -
func (ehm *EconomicsHandlerMock) MaxGasLimitPerBlock(shardID uint32) uint64 {
	if ehm.MaxGasLimitPerBlockCalled != nil {
		return ehm.MaxGasLimitPerBlockCalled(shardID)
	}
	return 0
}

// MaxGasLimitPerMiniBlock -
func (ehm *EconomicsHandlerMock) MaxGasLimitPerMiniBlock(shardID uint32) uint64 {
	if ehm.MaxGasLimitPerMiniBlockCalled != nil {
		return ehm.MaxGasLimitPerMiniBlockCalled(shardID)
	}
	return 0
}

// MaxGasLimitPerBlockForSafeCrossShard -
func (ehm *EconomicsHandlerMock) MaxGasLimitPerBlockForSafeCrossShard() uint64 {
	if ehm.MaxGasLimitPerBlockForSafeCrossShardCalled != nil {
		return ehm.MaxGasLimitPerBlockForSafeCrossShardCalled()
	}
	return 0
}

// MaxGasLimitPerMiniBlockForSafeCrossShard -
func (ehm *EconomicsHandlerMock) MaxGasLimitPerMiniBlockForSafeCrossShard() uint64 {
	if ehm.MaxGasLimitPerMiniBlockForSafeCrossShardCalled != nil {
		return ehm.MaxGasLimitPerMiniBlockForSafeCrossShardCalled()
	}
	return 0
}

// MaxGasLimitPerTx -
func (ehm *EconomicsHandlerMock) MaxGasLimitPerTx() uint64 {
	if ehm.MaxGasLimitPerTxCalled != nil {
		return ehm.MaxGasLimitPerTxCalled()
	}
	return 0
}

// ComputeGasLimit -
func (ehm *EconomicsHandlerMock) ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64 {
	if ehm.ComputeGasLimitCalled != nil {
		return ehm.ComputeGasLimitCalled(tx)
	}
	return 0
}

// ComputeFee -
func (ehm *EconomicsHandlerMock) ComputeFee(tx data.TransactionWithFeeHandler) *big.Int {
	if ehm.ComputeFeeCalled != nil {
		return ehm.ComputeFeeCalled(tx)
	}
	return big.NewInt(0)
}

// CheckValidityTxValues -
func (ehm *EconomicsHandlerMock) CheckValidityTxValues(tx data.TransactionWithFeeHandler) error {
	if ehm.CheckValidityTxValuesCalled != nil {
		return ehm.CheckValidityTxValuesCalled(tx)
	}
	return nil
}

// ComputeMoveBalanceFee -
func (ehm *EconomicsHandlerMock) ComputeMoveBalanceFee(tx data.TransactionWithFeeHandler) *big.Int {
	if ehm.ComputeMoveBalanceFeeCalled != nil {
		return ehm.ComputeMoveBalanceFeeCalled(tx)
	}
	return big.NewInt(0)
}

// ComputeMoveBalanceFeeInEpoch -
func (ehm *EconomicsHandlerMock) ComputeMoveBalanceFeeInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) *big.Int {
	if ehm.ComputeMoveBalanceFeeInEpochCalled != nil {
		return ehm.ComputeMoveBalanceFeeInEpochCalled(tx, epoch)
	}
	return big.NewInt(0)
}

// ComputeGasLimitBasedOnBalance -
func (ehm *EconomicsHandlerMock) ComputeGasLimitBasedOnBalance(tx data.TransactionWithFeeHandler, balance *big.Int) (uint64, error) {
	if ehm.ComputeGasLimitBasedOnBalanceCalled != nil {
		return ehm.ComputeGasLimitBasedOnBalanceCalled(tx, balance)
	}
	return 0, nil
}

// ComputeTxFee -
func (ehm *EconomicsHandlerMock) ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int {
	if ehm.ComputeTxFeeCalled != nil {
		return ehm.ComputeTxFeeCalled(tx)
	}
	return big.NewInt(0)
}

// RewardsTopUpGradientPoint -
func (ehm *EconomicsHandlerMock) RewardsTopUpGradientPoint() *big.Int {
	if ehm.RewardsTopUpGradientPointCalled != nil {
		return ehm.RewardsTopUpGradientPointCalled()
	}
	return big.NewInt(0)
}

// RewardsTopUpFactor -
func (ehm *EconomicsHandlerMock) RewardsTopUpFactor() float64 {
	if ehm.RewardsTopUpFactorCalled != nil {
		return ehm.RewardsTopUpFactorCalled()
	}
	return 0
}

// ComputeFeeForProcessing -
func (ehm *EconomicsHandlerMock) ComputeFeeForProcessing(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	if ehm.ComputeFeeForProcessingCalled != nil {
		return ehm.ComputeFeeForProcessingCalled(tx, gasToUse)
	}
	return big.NewInt(0)
}

// GasPriceModifier -
func (ehm *EconomicsHandlerMock) GasPriceModifier() float64 {
	if ehm.GasPriceModifierCalled != nil {
		return ehm.GasPriceModifierCalled()
	}
	return 0
}

// SplitTxGasInCategories -
func (ehm *EconomicsHandlerMock) SplitTxGasInCategories(tx data.TransactionWithFeeHandler) (uint64, uint64) {
	if ehm.SplitTxGasInCategoriesCalled != nil {
		return ehm.SplitTxGasInCategoriesCalled(tx)
	}
	return 0, 0
}

// GasPriceForProcessing -
func (ehm *EconomicsHandlerMock) GasPriceForProcessing(tx data.TransactionWithFeeHandler) uint64 {
	if ehm.GasPriceForProcessingCalled != nil {
		return ehm.GasPriceForProcessingCalled(tx)
	}
	return 0
}

// GasPriceForMove -
func (ehm *EconomicsHandlerMock) GasPriceForMove(tx data.TransactionWithFeeHandler) uint64 {
	if ehm.GasPriceForMoveCalled != nil {
		return ehm.GasPriceForMoveCalled(tx)
	}
	return 0
}

// MinGasPriceForProcessing -
func (ehm *EconomicsHandlerMock) MinGasPriceForProcessing() uint64 {
	if ehm.MinGasPriceForProcessingCalled != nil {
		return ehm.MinGasPriceForProcessingCalled()
	}
	return 0
}

// ComputeGasUsedAndFeeBasedOnRefundValue -
func (ehm *EconomicsHandlerMock) ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int) {
	if ehm.ComputeGasUsedAndFeeBasedOnRefundValueCalled != nil {
		return ehm.ComputeGasUsedAndFeeBasedOnRefundValueCalled(tx, refundValue)
	}
	return 0, big.NewInt(0)
}

// ComputeTxFeeBasedOnGasUsed -
func (ehm *EconomicsHandlerMock) ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
	if ehm.ComputeTxFeeBasedOnGasUsedCalled != nil {
		return ehm.ComputeTxFeeBasedOnGasUsedCalled(tx, gasUsed)
	}
	return big.NewInt(0)
}

// SetStatusHandler -
func (ehm *EconomicsHandlerMock) SetStatusHandler(statusHandler core.AppStatusHandler) error {
	if ehm.SetStatusHandlerCalled != nil {
		return ehm.SetStatusHandlerCalled(statusHandler)
	}
	return nil
}

// ComputeTxFeeInEpoch -
func (ehm *EconomicsHandlerMock) ComputeTxFeeInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) *big.Int {
	if ehm.ComputeTxFeeInEpochCalled != nil {
		return ehm.ComputeTxFeeInEpochCalled(tx, epoch)
	}
	return nil
}

// ComputeGasLimitInEpoch -
func (ehm *EconomicsHandlerMock) ComputeGasLimitInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) uint64 {
	if ehm.ComputeGasLimitInEpochCalled != nil {
		return ehm.ComputeGasLimitInEpochCalled(tx, epoch)
	}
	return 0
}

// ComputeGasUsedAndFeeBasedOnRefundValueInEpoch -
func (ehm *EconomicsHandlerMock) ComputeGasUsedAndFeeBasedOnRefundValueInEpoch(tx data.TransactionWithFeeHandler, refundValue *big.Int, epoch uint32) (uint64, *big.Int) {
	if ehm.ComputeGasUsedAndFeeBasedOnRefundValueInEpochCalled != nil {
		return ehm.ComputeGasUsedAndFeeBasedOnRefundValueInEpochCalled(tx, refundValue, epoch)
	}
	return 0, big.NewInt(0)
}

// ComputeTxFeeBasedOnGasUsedInEpoch -
func (ehm *EconomicsHandlerMock) ComputeTxFeeBasedOnGasUsedInEpoch(tx data.TransactionWithFeeHandler, gasUsed uint64, epoch uint32) *big.Int {
	if ehm.ComputeTxFeeBasedOnGasUsedInEpochCalled != nil {
		return ehm.ComputeTxFeeBasedOnGasUsedInEpochCalled(tx, gasUsed, epoch)
	}
	return nil
}

// LeaderPercentageInEpoch -
func (ehm *EconomicsHandlerMock) LeaderPercentageInEpoch(epoch uint32) float64 {
	if ehm.LeaderPercentageInEpochCalled != nil {
		return ehm.LeaderPercentageInEpochCalled(epoch)
	}
	return 0
}

// DeveloperPercentageInEpoch -
func (ehm *EconomicsHandlerMock) DeveloperPercentageInEpoch(epoch uint32) float64 {
	if ehm.DeveloperPercentageInEpochCalled != nil {
		return ehm.DeveloperPercentageInEpochCalled(epoch)
	}
	return 0
}

// ProtocolSustainabilityPercentageInEpoch -
func (ehm *EconomicsHandlerMock) ProtocolSustainabilityPercentageInEpoch(epoch uint32) float64 {
	if ehm.ProtocolSustainabilityPercentageInEpochCalled != nil {
		return ehm.ProtocolSustainabilityPercentageInEpochCalled(epoch)
	}
	return 0
}

// ProtocolSustainabilityAddressInEpoch -
func (ehm *EconomicsHandlerMock) ProtocolSustainabilityAddressInEpoch(epoch uint32) string {
	if ehm.ProtocolSustainabilityAddressInEpochCalled != nil {
		return ehm.ProtocolSustainabilityAddressInEpochCalled(epoch)
	}
	return ""
}

// RewardsTopUpGradientPointInEpoch -
func (ehm *EconomicsHandlerMock) RewardsTopUpGradientPointInEpoch(epoch uint32) *big.Int {
	if ehm.RewardsTopUpGradientPointInEpochCalled != nil {
		return ehm.RewardsTopUpGradientPointInEpochCalled(epoch)
	}
	return big.NewInt(0)
}

// RewardsTopUpFactorInEpoch -
func (ehm *EconomicsHandlerMock) RewardsTopUpFactorInEpoch(epoch uint32) float64 {
	if ehm.RewardsTopUpFactorInEpochCalled != nil {
		return ehm.RewardsTopUpFactorInEpochCalled(epoch)
	}
	return 0
}

// MaxGasHigherFactorAccepted -
func (ehm *EconomicsHandlerMock) MaxGasHigherFactorAccepted() uint64 {
	if ehm.MaxGasHigherFactorAcceptedCalled != nil {
		return ehm.MaxGasHigherFactorAcceptedCalled()
	}
	return 10
}

// IsInterfaceNil returns true if there is no value under the interface
func (ehm *EconomicsHandlerMock) IsInterfaceNil() bool {
	return ehm == nil
}
