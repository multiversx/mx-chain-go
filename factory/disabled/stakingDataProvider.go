package disabled

import (
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
)

type stakingDataProvider struct {
}

// NewDisabledStakingDataProvider returns a new instance of stakingDataProvider
func NewDisabledStakingDataProvider() *stakingDataProvider {
	return &stakingDataProvider{}
}

// FillValidatorInfo returns a nil error
func (s *stakingDataProvider) FillValidatorInfo(state.ValidatorInfoHandler) error {
	return nil
}

// ComputeUnQualifiedNodes returns nil values
func (s *stakingDataProvider) ComputeUnQualifiedNodes(_ state.ShardValidatorsInfoMapHandler) ([][]byte, map[string][][]byte, error) {
	return nil, nil, nil
}

// GetOwnersData returns nil
func (s *stakingDataProvider) GetOwnersData() map[string]*epochStart.OwnerData {
	return nil
}

// Clean does nothing
func (s *stakingDataProvider) Clean() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *stakingDataProvider) IsInterfaceNil() bool {
	return s == nil
}
