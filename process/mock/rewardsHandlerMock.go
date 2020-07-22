package mock

// RewardsHandlerMock -
type RewardsHandlerMock struct {
	MaxInflationRateCalled    func() float64
	MinInflationRateCalled    func() float64
	LeaderPercentageCalled    func() float64
	CommunityPercentageCalled func() float64
	CommunityAddressCalled    func() string
}

// LeaderPercentage -
func (rhm *RewardsHandlerMock) LeaderPercentage() float64 {
	return rhm.LeaderPercentageCalled()
}

// CommunityPercentage will return the community percentage value
func (rhm *RewardsHandlerMock) CommunityPercentage() float64 {
	return rhm.CommunityPercentageCalled()
}

// CommunityAddress will return the community address
func (rhm *RewardsHandlerMock) CommunityAddress() string {
	return rhm.CommunityAddressCalled()
}

// MinInflationRate -
func (rhm *RewardsHandlerMock) MinInflationRate() float64 {
	return rhm.MinInflationRateCalled()
}

// MaxInflationRate -
func (rhm *RewardsHandlerMock) MaxInflationRate(uint32) float64 {
	return rhm.MaxInflationRateCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rhm *RewardsHandlerMock) IsInterfaceNil() bool {
	return rhm == nil
}
