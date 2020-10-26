package mock

// StakingDataProviderStub -
type StakingDataProviderStub struct {
	CleanCalled                   func()
	GetStakingDataForBlsKeyCalled func(blsKey []byte) error
}

// GetStakingDataForBlsKey -
func (sdps *StakingDataProviderStub) GetStakingDataForBlsKey(blsKey []byte) error {
	if sdps.GetStakingDataForBlsKeyCalled != nil {
		return sdps.GetStakingDataForBlsKeyCalled(blsKey)
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
