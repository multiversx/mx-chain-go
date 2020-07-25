package mock

// RewardsHandlerMock -
type RewardsHandlerMock struct {
	MaxInflationRateCalled                 func() float64
	MinInflationRateCalled                 func() float64
	LeaderPercentageCalled                 func() float64
	ProtocolSustainabilityPercentageCalled func() float64
	ProtocolSustainabilityAddressCalled    func() string
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
