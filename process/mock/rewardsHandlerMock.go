package mock

import "math/big"

type RewardsHandlerMock struct {
	RewardsValueCalled        func() *big.Int
	CommunityPercentageCalled func() float64
	LeaderPercentageCalled    func() float64
	BurnPercentageCalled      func() float64
}

func (rhm *RewardsHandlerMock) RewardsValue() *big.Int {
	return rhm.RewardsValueCalled()
}

func (rhm *RewardsHandlerMock) CommunityPercentage() float64 {
	return rhm.CommunityPercentageCalled()
}

func (rhm *RewardsHandlerMock) LeaderPercentage() float64 {
	return rhm.LeaderPercentageCalled()
}

func (rhm *RewardsHandlerMock) BurnPercentage() float64 {
	return rhm.BurnPercentageCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rhm *RewardsHandlerMock) IsInterfaceNil() bool {
	if rhm == nil {
		return true
	}
	return false
}
