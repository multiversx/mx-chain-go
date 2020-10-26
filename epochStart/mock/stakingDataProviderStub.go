package mock

// StakingDataProviderStub -
type StakingDataProviderStub struct {
	CleanCalled                func()
	PrepareDataForBlsKeyCalled func(blsKey []byte) error
}

// PrepareDataForBlsKey -
func (sdps *StakingDataProviderStub) PrepareDataForBlsKey(blsKey []byte) error {
	if sdps.PrepareDataForBlsKeyCalled != nil {
		return sdps.PrepareDataForBlsKeyCalled(blsKey)
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
