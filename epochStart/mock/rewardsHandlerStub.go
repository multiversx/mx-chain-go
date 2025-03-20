package mock

import "math/big"

// RewardsHandlerStub -
type RewardsHandlerStub struct {
	LeaderPercentageCalled                        func() float64
	ProtocolSustainabilityPercentageCalled        func() float64
	ProtocolSustainabilityAddressCalled           func() string
	MinInflationRateCalled                        func() float64
	MaxInflationRateCalled                        func(year uint32) float64
	RewardsTopUpGradientPointCalled               func() *big.Int
	RewardsTopUpFactorCalled                      func() float64
	LeaderPercentageInEpochCalled                 func(epoch uint32) float64
	DeveloperPercentageInEpochCalled              func(epoch uint32) float64
	ProtocolSustainabilityPercentageInEpochCalled func(epoch uint32) float64
	ProtocolSustainabilityAddressInEpochCalled    func(epoch uint32) string
	RewardsTopUpGradientPointInEpochCalled        func(epoch uint32) *big.Int
	RewardsTopUpFactorInEpochCalled               func(epoch uint32) float64
}

// LeaderPercentage -
func (r *RewardsHandlerStub) LeaderPercentage() float64 {
	if r.LeaderPercentageCalled != nil {
		return r.LeaderPercentageCalled()
	}

	return 1
}

// ProtocolSustainabilityPercentage will return the protocol sustainability percentage value
func (r *RewardsHandlerStub) ProtocolSustainabilityPercentage() float64 {
	if r.ProtocolSustainabilityPercentageCalled != nil {
		return r.ProtocolSustainabilityPercentageCalled()
	}

	return 0.1
}

// ProtocolSustainabilityAddress will return the protocol sustainability address
func (r *RewardsHandlerStub) ProtocolSustainabilityAddress() string {
	if r.ProtocolSustainabilityAddressCalled != nil {
		return r.ProtocolSustainabilityAddressCalled()
	}

	return "1111"
}

// MinInflationRate -
func (r *RewardsHandlerStub) MinInflationRate() float64 {
	if r.MinInflationRateCalled != nil {
		return r.MinInflationRateCalled()
	}

	return 1
}

// MaxInflationRate -
func (r *RewardsHandlerStub) MaxInflationRate(year uint32) float64 {
	if r.MaxInflationRateCalled != nil {
		return r.MaxInflationRateCalled(year)
	}

	return 1000000
}

// RewardsTopUpGradientPoint -
func (r *RewardsHandlerStub) RewardsTopUpGradientPoint() *big.Int {
	return r.RewardsTopUpGradientPointCalled()
}

// RewardsTopUpFactor -
func (r *RewardsHandlerStub) RewardsTopUpFactor() float64 {
	return r.RewardsTopUpFactorCalled()
}

// LeaderPercentageInEpoch -
func (r *RewardsHandlerStub) LeaderPercentageInEpoch(epoch uint32) float64 {
	if r.LeaderPercentageInEpochCalled != nil {
		return r.LeaderPercentageInEpochCalled(epoch)
	}
	return 1
}

// DeveloperPercentageInEpoch -
func (r *RewardsHandlerStub) DeveloperPercentageInEpoch(epoch uint32) float64 {
	if r.DeveloperPercentageInEpochCalled != nil {
		return r.DeveloperPercentageInEpochCalled(epoch)
	}
	return 0
}

// ProtocolSustainabilityPercentageInEpoch -
func (r *RewardsHandlerStub) ProtocolSustainabilityPercentageInEpoch(epoch uint32) float64 {
	if r.ProtocolSustainabilityPercentageInEpochCalled != nil {
		return r.ProtocolSustainabilityPercentageInEpochCalled(epoch)
	}
	return 0
}

// ProtocolSustainabilityAddressInEpoch -
func (r *RewardsHandlerStub) ProtocolSustainabilityAddressInEpoch(epoch uint32) string {
	if r.ProtocolSustainabilityAddressInEpochCalled != nil {
		return r.ProtocolSustainabilityAddressInEpochCalled(epoch)
	}
	return "1111"
}

// RewardsTopUpGradientPointInEpoch -
func (r *RewardsHandlerStub) RewardsTopUpGradientPointInEpoch(epoch uint32) *big.Int {
	if r.RewardsTopUpGradientPointInEpochCalled != nil {
		return r.RewardsTopUpGradientPointInEpochCalled(epoch)
	}
	return big.NewInt(0)
}

// RewardsTopUpFactorInEpoch -
func (r *RewardsHandlerStub) RewardsTopUpFactorInEpoch(epoch uint32) float64 {
	if r.RewardsTopUpFactorInEpochCalled != nil {
		return r.RewardsTopUpFactorInEpochCalled(epoch)
	}
	return 0
}

// IsInterfaceNil -
func (r *RewardsHandlerStub) IsInterfaceNil() bool {
	return r == nil
}
