package disabled

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/state"
)

var zeroBI = big.NewInt(0)

type stakingDataProvider struct {
}

// NewDisabledStakingDataProvider returns a new instance of stakingDataProvider
func NewDisabledStakingDataProvider() *stakingDataProvider {
	return &stakingDataProvider{}
}

// GetTotalStakeEligibleNodes returns an empty big integer
func (s *stakingDataProvider) GetTotalStakeEligibleNodes() *big.Int {
	return zeroBI
}

// GetTotalTopUpStakeEligibleNodes returns an empty big integer
func (s *stakingDataProvider) GetTotalTopUpStakeEligibleNodes() *big.Int {
	return zeroBI
}

// GetNodeStakedTopUp returns an empty big integer and a nil error
func (s *stakingDataProvider) GetNodeStakedTopUp(_ []byte) (*big.Int, error) {
	return zeroBI, nil
}

// PrepareStakingData returns a nil error
func (s *stakingDataProvider) PrepareStakingData(state.ShardValidatorsInfoMapHandler) error {
	return nil
}

// FillValidatorInfo returns a nil error
func (s *stakingDataProvider) FillValidatorInfo(state.ValidatorInfoHandler) error {
	return nil
}

// ComputeUnQualifiedNodes returns nil values
func (s *stakingDataProvider) ComputeUnQualifiedNodes(_ state.ShardValidatorsInfoMapHandler) ([][]byte, map[string][][]byte, error) {
	return nil, nil, nil
}

// GetBlsKeyOwner returns an empty key and a nil error
func (s *stakingDataProvider) GetBlsKeyOwner(_ []byte) (string, error) {
	return "", nil
}

// GetNumOfValidatorsInCurrentEpoch returns 0
func (s *stakingDataProvider) GetNumOfValidatorsInCurrentEpoch() uint32 {
	return 0
}

// GetOwnersData returns nil
func (s *stakingDataProvider) GetOwnersData() map[string]*epochStart.OwnerData {
	return nil
}

// Clean does nothing
func (s *stakingDataProvider) Clean() {
}

// EpochConfirmed does nothing
func (s *stakingDataProvider) EpochConfirmed(_ uint32, _ uint64) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *stakingDataProvider) IsInterfaceNil() bool {
	return s == nil
}
