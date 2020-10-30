package mock

import "math/big"

// StakingDataProviderStub -
type StakingDataProviderStub struct {
	CleanCalled                      func()
	PrepareDataForBlsKeyCalled       func(blsKey []byte) error
	PrepareStakingDataCalled         func(keys map[uint32][][]byte) error
	GetTotalStakeEligibleNodesCalled func() *big.Int
}

// GetTotalStakeEligibleNodes -
func (sdps *StakingDataProviderStub) GetTotalStakeEligibleNodes() *big.Int {
	if sdps.GetTotalStakeEligibleNodesCalled != nil {
		return sdps.GetTotalStakeEligibleNodesCalled()
	}
	return big.NewInt(0)
}

// PrepareStakingData -
func (sdps *StakingDataProviderStub) PrepareStakingData(keys map[uint32][][]byte) error {
	if sdps.PrepareStakingDataCalled != nil {
		return sdps.PrepareStakingDataCalled(keys)
	}
	return nil
}

// Clean -
func (sdps *StakingDataProviderStub) Clean() {
	if sdps.CleanCalled != nil {
		sdps.CleanCalled()
	}
}

// IsInterfaceNil -
func (sdps *StakingDataProviderStub) IsInterfaceNil() bool {
	return sdps == nil
}
