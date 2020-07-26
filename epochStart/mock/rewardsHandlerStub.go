package mock

// RewardsHandlerStub -
type RewardsHandlerStub struct {
	LeaderPercentageCalled                 func() float64
	ProtocolSustainabilityPercentageCalled func() float64
	ProtocolSustainabilityAddressCalled    func() string
	MinInflationRateCalled                 func() float64
	MaxInflationRateCalled                 func(year uint32) float64
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

// IsInterfaceNil -
func (r *RewardsHandlerStub) IsInterfaceNil() bool {
	return r == nil
}
