package mock

import (
	"github.com/multiversx/mx-chain-go/state"
)

// ValidatorInfoForCurrentEpochStub -
type ValidatorInfoForCurrentEpochStub struct {
	CleanCalled            func()
	GetValidatorInfoCalled func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error)
	AddValidatorInfoCalled func(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo)
}

// Clean -
func (v *ValidatorInfoForCurrentEpochStub) Clean() {
	if v.CleanCalled != nil {
		v.CleanCalled()
	}
}

// GetValidatorInfo -
func (v *ValidatorInfoForCurrentEpochStub) GetValidatorInfo(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
	if v.GetValidatorInfoCalled != nil {
		return v.GetValidatorInfoCalled(validatorInfoHash)
	}
	return nil, nil
}

// AddValidatorInfo -
func (v *ValidatorInfoForCurrentEpochStub) AddValidatorInfo(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo) {
	if v.AddValidatorInfoCalled != nil {
		v.AddValidatorInfoCalled(validatorInfoHash, validatorInfo)
	}
}

// IsInterfaceNil -
func (v *ValidatorInfoForCurrentEpochStub) IsInterfaceNil() bool {
	return v == nil
}
