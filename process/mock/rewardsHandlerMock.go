package mock

import "math/big"

// RewardsHandlerMock -
type RewardsHandlerMock struct {
	MaxInflationRateCalled                 func() float64
	MinInflationRateCalled                 func() float64
	LeaderPercentageCalled                 func() float64
	ProtocolSustainabilityPercentageCalled func() float64
	ProtocolSustainabilityAddressCalled    func() string
	RewardsTopUpGradientPointCalled        func() *big.Int
	RewardsTopUpFactorCalled               func() float64
	GetTailInflationActivationEpochCalled  func() uint32
	GetMaximumYearlyInflationCalled        func() float64
	GetDecayPercentageCalled               func() float64
	GetMinimumInflationCalled              func() float64
	EcosystemGrowthPercentageInEpochCalled func(epoch uint32) float64
	EcosystemGrowthAddressInEpochCalled    func(epoch uint32) string
	GrowthDividendPercentageInEpochCalled  func(epoch uint32) float64
	GrowthDividendAddressInEpochCalled     func(epoch uint32) string
}

// LeaderPercentage -
func (rhm *RewardsHandlerMock) LeaderPercentage() float64 {
	return rhm.LeaderPercentageCalled()
}

// LeaderPercentageInEpoch -
func (rhm *RewardsHandlerMock) LeaderPercentageInEpoch(epoch uint32) float64 {
	return rhm.LeaderPercentageCalled()
}

// DeveloperPercentageInEpoch -
func (rhm *RewardsHandlerMock) DeveloperPercentageInEpoch(epoch uint32) float64 {
	return 0
}

// ProtocolSustainabilityPercentage will return the protocol sustainability percentage value
func (rhm *RewardsHandlerMock) ProtocolSustainabilityPercentage() float64 {
	return rhm.ProtocolSustainabilityPercentageCalled()
}

// ProtocolSustainabilityAddress will return the protocol sustainability address
func (rhm *RewardsHandlerMock) ProtocolSustainabilityAddress() string {
	return rhm.ProtocolSustainabilityAddressCalled()
}

// MinInflationRate -
func (rhm *RewardsHandlerMock) MinInflationRate() float64 {
	return rhm.MinInflationRateCalled()
}

// MaxInflationRate -
func (rhm *RewardsHandlerMock) MaxInflationRate(uint32) float64 {
	return rhm.MaxInflationRateCalled()
}

// RewardsTopUpGradientPoint -
func (rhm *RewardsHandlerMock) RewardsTopUpGradientPoint() *big.Int {
	return rhm.RewardsTopUpGradientPointCalled()
}

// RewardsTopUpFactor -
func (rhm *RewardsHandlerMock) RewardsTopUpFactor() float64 {
	return rhm.RewardsTopUpFactorCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rhm *RewardsHandlerMock) IsInterfaceNil() bool {
	return rhm == nil
}

// ProtocolSustainabilityPercentageInEpoch -
func (rhm *RewardsHandlerMock) ProtocolSustainabilityPercentageInEpoch(epoch uint32) float64 {
	return 0
}

// ProtocolSustainabilityAddressInEpoch -
func (rhm *RewardsHandlerMock) ProtocolSustainabilityAddressInEpoch(epoch uint32) string {
	return ""
}

// RewardsTopUpGradientPointInEpoch -
func (rhm *RewardsHandlerMock) RewardsTopUpGradientPointInEpoch(epoch uint32) *big.Int {
	return nil
}

// RewardsTopUpFactorInEpoch -
func (rhm *RewardsHandlerMock) RewardsTopUpFactorInEpoch(epoch uint32) float64 {
	return 0
}

// GetTailInflationActivationEpoch -
func (rhm *RewardsHandlerMock) GetTailInflationActivationEpoch() uint32 {
	if rhm.GetTailInflationActivationEpochCalled != nil {
		return rhm.GetTailInflationActivationEpochCalled()
	}
	return 0
}

// GetMaximumYearlyInflation -
func (rhm *RewardsHandlerMock) GetMaximumYearlyInflation() float64 {
	if rhm.GetMaximumYearlyInflationCalled != nil {
		return rhm.GetMaximumYearlyInflationCalled()
	}
	return 0
}

// GetDecayPercentage -
func (rhm *RewardsHandlerMock) GetDecayPercentage() float64 {
	if rhm.GetDecayPercentageCalled != nil {
		return rhm.GetDecayPercentageCalled()
	}
	return 0
}

// GetMinimumInflation -
func (rhm *RewardsHandlerMock) GetMinimumInflation() float64 {
	if rhm.GetMinimumInflationCalled != nil {
		return rhm.GetMinimumInflationCalled()
	}
	return 0
}

// EcosystemGrowthPercentageInEpoch -
func (rhm *RewardsHandlerMock) EcosystemGrowthPercentageInEpoch(epoch uint32) float64 {
	if rhm.EcosystemGrowthPercentageInEpochCalled != nil {
		return rhm.EcosystemGrowthPercentageInEpochCalled(epoch)
	}
	return 0
}

// EcosystemGrowthAddressInEpoch -
func (rhm *RewardsHandlerMock) EcosystemGrowthAddressInEpoch(epoch uint32) string {
	if rhm.EcosystemGrowthAddressInEpochCalled != nil {
		return rhm.EcosystemGrowthAddressInEpochCalled(epoch)
	}
	return ""
}

// GrowthDividendPercentageInEpoch -
func (rhm *RewardsHandlerMock) GrowthDividendPercentageInEpoch(epoch uint32) float64 {
	if rhm.GrowthDividendPercentageInEpochCalled != nil {
		return rhm.GrowthDividendPercentageInEpochCalled(epoch)
	}
	return 0
}

// GrowthDividendAddressInEpoch -
func (rhm *RewardsHandlerMock) GrowthDividendAddressInEpoch(epoch uint32) string {
	if rhm.GrowthDividendAddressInEpochCalled != nil {
		return rhm.GrowthDividendAddressInEpochCalled(epoch)
	}
	return ""
}
