package mock

// RewardsHandlerMock -
type RewardsHandlerMock struct {
	MaxInflationRateCalled func() float64
	MinInflationRateCalled func() float64
	LeaderPercentageCalled func() float64
}

// LeaderPercentage -
func (rhm *RewardsHandlerMock) LeaderPercentage() float64 {
	return rhm.LeaderPercentageCalled()
}

// MinInflationRate -
func (rhm *RewardsHandlerMock) MinInflationRate() float64 {
	return rhm.MinInflationRateCalled()
}

// MaxInflationRate -
func (rhm *RewardsHandlerMock) MaxInflationRate() float64 {
	return rhm.MaxInflationRateCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rhm *RewardsHandlerMock) IsInterfaceNil() bool {
	return rhm == nil
}
