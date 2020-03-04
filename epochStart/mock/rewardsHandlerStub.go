package mock

// RewardsHandlerStub -
type RewardsHandlerStub struct {
	LeaderPercentageCalled func() float64
	MinInflationRateCalled func() float64
	MaxInflationRateCalled func() float64
}

// LeaderPercentage -
func (r *RewardsHandlerStub) LeaderPercentage() float64 {
	if r.LeaderPercentageCalled != nil {
		return r.LeaderPercentageCalled()
	}

	return 1
}

// MinInflationRate -
func (r *RewardsHandlerStub) MinInflationRate() float64 {
	if r.MinInflationRateCalled != nil {
		return r.MinInflationRateCalled()
	}

	return 1
}

// MaxInflationRate -
func (r *RewardsHandlerStub) MaxInflationRate() float64 {
	if r.MaxInflationRateCalled != nil {
		return r.MaxInflationRateCalled()
	}

	return 1000000
}

// IsInterfaceNil -
func (r *RewardsHandlerStub) IsInterfaceNil() bool {
	return r == nil
}
