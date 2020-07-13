package mock

// RewardsHandlerStub -
type RewardsHandlerStub struct {
	LeaderPercentageCalled    func() float64
	CommunityPercentageCalled func() float64
	CommunityAddressCalled    func() string
	MinInflationRateCalled    func() float64
	MaxInflationRateCalled    func(year uint32) float64
}

// LeaderPercentage -
func (r *RewardsHandlerStub) LeaderPercentage() float64 {
	if r.LeaderPercentageCalled != nil {
		return r.LeaderPercentageCalled()
	}

	return 1
}

// CommunityPercentage will return the community percentage value
func (r *RewardsHandlerStub) CommunityPercentage() float64 {
	if r.CommunityPercentageCalled != nil {
		return r.CommunityPercentageCalled()
	}

	return 0.1
}

// CommunityAddress will return the community address
func (r *RewardsHandlerStub) CommunityAddress() string {
	if r.CommunityAddressCalled != nil {
		return r.CommunityAddressCalled()
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
