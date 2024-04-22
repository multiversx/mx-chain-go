package stakingcommon

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
)

// StakingDataProviderStub -
type StakingDataProviderStub struct {
	CleanCalled                           func()
	PrepareStakingDataCalled              func(validatorsMap state.ShardValidatorsInfoMapHandler) error
	GetTotalStakeEligibleNodesCalled      func() *big.Int
	GetTotalTopUpStakeEligibleNodesCalled func() *big.Int
	GetNodeStakedTopUpCalled              func(blsKey []byte) (*big.Int, error)
	FillValidatorInfoCalled               func(validator state.ValidatorInfoHandler) error
	ComputeUnQualifiedNodesCalled         func(validatorInfos state.ShardValidatorsInfoMapHandler) ([][]byte, map[string][][]byte, error)
	GetBlsKeyOwnerCalled                  func(blsKey []byte) (string, error)
	GetOwnersDataCalled                   func() map[string]*epochStart.OwnerData
}

// FillValidatorInfo -
func (sdps *StakingDataProviderStub) FillValidatorInfo(validator state.ValidatorInfoHandler) error {
	if sdps.FillValidatorInfoCalled != nil {
		return sdps.FillValidatorInfoCalled(validator)
	}
	return nil
}

// ComputeUnQualifiedNodes -
func (sdps *StakingDataProviderStub) ComputeUnQualifiedNodes(validatorInfos state.ShardValidatorsInfoMapHandler) ([][]byte, map[string][][]byte, error) {
	if sdps.ComputeUnQualifiedNodesCalled != nil {
		return sdps.ComputeUnQualifiedNodesCalled(validatorInfos)
	}
	return nil, nil, nil
}

// GetTotalStakeEligibleNodes -
func (sdps *StakingDataProviderStub) GetTotalStakeEligibleNodes() *big.Int {
	if sdps.GetTotalStakeEligibleNodesCalled != nil {
		return sdps.GetTotalStakeEligibleNodesCalled()
	}
	return big.NewInt(0)
}

// GetTotalTopUpStakeEligibleNodes -
func (sdps *StakingDataProviderStub) GetTotalTopUpStakeEligibleNodes() *big.Int {
	if sdps.GetTotalTopUpStakeEligibleNodesCalled != nil {
		return sdps.GetTotalTopUpStakeEligibleNodesCalled()
	}
	return big.NewInt(0)
}

// GetNodeStakedTopUp -
func (sdps *StakingDataProviderStub) GetNodeStakedTopUp(blsKey []byte) (*big.Int, error) {
	if sdps.GetNodeStakedTopUpCalled != nil {
		return sdps.GetNodeStakedTopUpCalled(blsKey)
	}
	return big.NewInt(0), nil
}

// PrepareStakingData -
func (sdps *StakingDataProviderStub) PrepareStakingData(validatorsMap state.ShardValidatorsInfoMapHandler) error {
	if sdps.PrepareStakingDataCalled != nil {
		return sdps.PrepareStakingDataCalled(validatorsMap)
	}
	return nil
}

// Clean -
func (sdps *StakingDataProviderStub) Clean() {
	if sdps.CleanCalled != nil {
		sdps.CleanCalled()
	}
}

// GetBlsKeyOwner -
func (sdps *StakingDataProviderStub) GetBlsKeyOwner(blsKey []byte) (string, error) {
	if sdps.GetBlsKeyOwnerCalled != nil {
		return sdps.GetBlsKeyOwnerCalled(blsKey)
	}
	return "", nil
}

// GetNumOfValidatorsInCurrentEpoch -
func (sdps *StakingDataProviderStub) GetNumOfValidatorsInCurrentEpoch() uint32 {
	return 0
}

// GetOwnersData -
func (sdps *StakingDataProviderStub) GetOwnersData() map[string]*epochStart.OwnerData {
	if sdps.GetOwnersDataCalled != nil {
		return sdps.GetOwnersDataCalled()
	}
	return nil
}

// EpochConfirmed -
func (sdps *StakingDataProviderStub) EpochConfirmed(uint32, uint64) {
}

// IsInterfaceNil -
func (sdps *StakingDataProviderStub) IsInterfaceNil() bool {
	return sdps == nil
}
