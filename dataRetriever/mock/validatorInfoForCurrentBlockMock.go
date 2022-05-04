package mock

import (
	"github.com/ElrondNetwork/elrond-go/state"
)

// ValidatorInfoForCurrentBlockStub -
type ValidatorInfoForCurrentBlockStub struct {
	CleanCalled            func()
	GetValidatorInfoCalled func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error)
	AddValidatorInfoCalled func(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo)
}

// Clean -
func (t *ValidatorInfoForCurrentBlockStub) Clean() {
	if t.CleanCalled != nil {
		t.CleanCalled()
	}
}

// GetValidatorInfo -
func (v *ValidatorInfoForCurrentBlockStub) GetValidatorInfo(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
	if v.GetValidatorInfoCalled != nil {
		return v.GetValidatorInfoCalled(validatorInfoHash)
	}
	return nil, nil
}

// AddValidatorInfo -
func (v *ValidatorInfoForCurrentBlockStub) AddValidatorInfo(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo) {
	if v.AddValidatorInfoCalled != nil {
		v.AddValidatorInfoCalled(validatorInfoHash, validatorInfo)
	}
}

// IsInterfaceNil -
func (v *ValidatorInfoForCurrentBlockStub) IsInterfaceNil() bool {
	return v == nil
}
