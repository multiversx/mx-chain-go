package mock

// RewardsStakingProviderStub -
type RewardsStakingProviderStub struct {
	CleanCalled                   func()
	ComputeRewardsForBlsKeyCalled func(blsKey []byte) error
}

// ComputeRewardsForBlsKey -
func (rsps *RewardsStakingProviderStub) ComputeRewardsForBlsKey(blsKey []byte) error {
	if rsps.ComputeRewardsForBlsKeyCalled != nil {
		return rsps.ComputeRewardsForBlsKeyCalled(blsKey)
	}

	return nil
}

// Clean -
func (rsps *RewardsStakingProviderStub) Clean() {
	if rsps.CleanCalled != nil {
		rsps.CleanCalled()
	}
}

// IsInterfaceNil -
func (rsps *RewardsStakingProviderStub) IsInterfaceNil() bool {
	return rsps == nil
}
