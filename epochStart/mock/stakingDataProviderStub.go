package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/state"
)

// StakingDataProviderStub -
type StakingDataProviderStub struct {
	CleanCalled                           func()
	PrepareStakingDataCalled              func(keys map[uint32][][]byte) error
	GetTotalStakeEligibleNodesCalled      func() *big.Int
	GetTotalTopUpStakeEligibleNodesCalled func() *big.Int
	GetNodeStakedTopUpCalled              func(blsKey []byte) (*big.Int, error)
	FillValidatorInfoCalled               func(blsKey []byte) error
	ComputeUnQualifiedNodesCalled         func(validatorInfos state.ShardValidatorsInfoMapHandler) ([][]byte, map[string][][]byte, error)
	GetBlsKeyOwnerCalled                  func(blsKey []byte) (string, error)
	GetNumStakedNodesCalled               func(owner []byte) (int64, error)
	GetTotalTopUpCalled                   func(owner []byte) (*big.Int, error)
}

// FillValidatorInfo -
func (sdps *StakingDataProviderStub) FillValidatorInfo(blsKey []byte) error {
	if sdps.FillValidatorInfoCalled != nil {
		return sdps.FillValidatorInfoCalled(blsKey)
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

// GetNumStakedNodes -
func (sdps *StakingDataProviderStub) GetNumStakedNodes(owner []byte) (int64, error) {
	if sdps.GetNumStakedNodesCalled != nil {
		return sdps.GetNumStakedNodesCalled(owner)
	}
	return 0, nil
}

// GetTotalTopUp -
func (sdps *StakingDataProviderStub) GetTotalTopUp(owner []byte) (*big.Int, error) {
	if sdps.GetTotalTopUpCalled != nil {
		return sdps.GetTotalTopUpCalled(owner)
	}
	return big.NewInt(0), nil
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

// GetBlsKeyOwner -
func (sdps *StakingDataProviderStub) GetBlsKeyOwner(blsKey []byte) (string, error) {
	if sdps.GetBlsKeyOwnerCalled != nil {
		return sdps.GetBlsKeyOwnerCalled(blsKey)
	}
	return "", nil
}

// EpochConfirmed -
func (sdps *StakingDataProviderStub) EpochConfirmed(uint32, uint64) {
}

// IsInterfaceNil -
func (sdps *StakingDataProviderStub) IsInterfaceNil() bool {
	return sdps == nil
}
