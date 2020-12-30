package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

// StakingDataProviderStub -
type StakingDataProviderStub struct {
	CleanCalled                           func()
	PrepareStakingDataCalled              func(keys map[uint32][][]byte) error
	GetTotalStakeEligibleNodesCalled      func() *big.Int
	GetTotalTopUpStakeEligibleNodesCalled func() *big.Int
	GetNodeStakedTopUpCalled              func(blsKey []byte) (*big.Int, error)
	FillValidatorInfoCalled               func(blsKey []byte) error
	ComputeUnQualifiedNodesCalled         func(validatorInfos map[uint32][]*state.ValidatorInfo) ([][]byte, map[string][][]byte, error)
}

// FillValidatorInfo -
func (sdps *StakingDataProviderStub) FillValidatorInfo(blsKey []byte) error {
	if sdps.FillValidatorInfoCalled != nil {
		return sdps.FillValidatorInfoCalled(blsKey)
	}
	return nil
}

// ComputeUnQualifiedNodes -
func (sdps *StakingDataProviderStub) ComputeUnQualifiedNodes(validatorInfos map[uint32][]*state.ValidatorInfo) ([][]byte, map[string][][]byte, error) {
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

// PrepareStakingDataForRewards -
func (sdps *StakingDataProviderStub) PrepareStakingDataForRewards(keys map[uint32][][]byte) error {
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
