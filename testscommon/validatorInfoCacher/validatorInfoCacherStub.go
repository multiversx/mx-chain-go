package validatorInfoCacherStub

import "github.com/multiversx/mx-chain-go/state"

// ValidatorInfoCacherStub -
type ValidatorInfoCacherStub struct {
	CleanCalled            func()
	AddValidatorInfoCalled func(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo)
	GetValidatorInfoCalled func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error)
}

// Clean -
func (vics *ValidatorInfoCacherStub) Clean() {
	if vics.CleanCalled != nil {
		vics.CleanCalled()
	}
}

// GetValidatorInfo -
func (vics *ValidatorInfoCacherStub) GetValidatorInfo(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
	if vics.GetValidatorInfoCalled != nil {
		return vics.GetValidatorInfoCalled(validatorInfoHash)
	}

	return nil, nil
}

// AddValidatorInfo -
func (vics *ValidatorInfoCacherStub) AddValidatorInfo(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo) {
	if vics.AddValidatorInfoCalled != nil {
		vics.AddValidatorInfoCalled(validatorInfoHash, validatorInfo)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (vics *ValidatorInfoCacherStub) IsInterfaceNil() bool {
	return vics == nil
}
