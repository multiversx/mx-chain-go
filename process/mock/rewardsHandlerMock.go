package mock

import "math/big"

// RewardsHandlerMock -
type RewardsHandlerMock struct {
	MaxInflationRateCalled                 func() float64
	LeaderPercentageCalled                 func() float64
	ProtocolSustainabilityPercentageCalled func() float64
	ProtocolSustainabilityAddressCalled    func() string
	RewardsTopUpGradientPointCalled        func() *big.Int
	RewardsTopUpFactorCalled               func() float64
	EcosystemGrowthPercentageInEpochCalled func(epoch uint32) float64
	EcosystemGrowthAddressInEpochCalled    func(epoch uint32) string
	GrowthDividendPercentageInEpochCalled  func(epoch uint32) float64
	GrowthDividendAddressInEpochCalled     func(epoch uint32) string
}

// LeaderPercentage -
func (rhm *RewardsHandlerMock) LeaderPercentage() float64 {
	return rhm.LeaderPercentageCalled()
}

// ProtocolSustainabilityPercentage will return the protocol sustainability percentage value
func (rhm *RewardsHandlerMock) ProtocolSustainabilityPercentage() float64 {
	return rhm.ProtocolSustainabilityPercentageCalled()
}

// ProtocolSustainabilityAddress will return the protocol sustainability address
func (rhm *RewardsHandlerMock) ProtocolSustainabilityAddress() string {
	return rhm.ProtocolSustainabilityAddressCalled()
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

// IsInterfaceNil returns true if there is no value under the interface
func (rhm *RewardsHandlerMock) IsInterfaceNil() bool {
	return rhm == nil
}
