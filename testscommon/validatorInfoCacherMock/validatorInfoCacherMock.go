package validatorInfoCacherMock

import "github.com/ElrondNetwork/elrond-go/state"

// ValidatorInfoCacherMock -
type ValidatorInfoCacherMock struct {
	CleanCalled            func()
	AddValidatorInfoCalled func(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo)
	GetValidatorInfoCalled func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error)
}

// Clean -
func (vicm *ValidatorInfoCacherMock) Clean() {
	if vicm.CleanCalled != nil {
		vicm.CleanCalled()
	}
}

// GetValidatorInfo -
func (vicm *ValidatorInfoCacherMock) GetValidatorInfo(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
	if vicm.GetValidatorInfoCalled != nil {
		return vicm.GetValidatorInfoCalled(validatorInfoHash)
	}

	return nil, nil
}

// AddValidatorInfo -
func (vicm *ValidatorInfoCacherMock) AddValidatorInfo(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo) {
	if vicm.AddValidatorInfoCalled != nil {
		vicm.AddValidatorInfoCalled(validatorInfoHash, validatorInfo)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (vicm *ValidatorInfoCacherMock) IsInterfaceNil() bool {
	return vicm == nil
}
