package mock

import "github.com/ElrondNetwork/elrond-go/data/state"

// ValidatorsProviderStub -
type ValidatorsProviderStub struct {
	GetLatestValidatorsCalled     func() map[string]*state.ValidatorApiResponse
	GetLatestValidatorInfosCalled func() (map[uint32][]*state.ValidatorInfo, error)
}

// GetLatestValidators -
func (vp *ValidatorsProviderStub) GetLatestValidators() map[string]*state.ValidatorApiResponse {
	if vp.GetLatestValidatorsCalled != nil {
		return vp.GetLatestValidatorsCalled()
	}
	return nil
}

// GetLatestValidatorInfos -
func (vp *ValidatorsProviderStub) GetLatestValidatorInfos() (map[uint32][]*state.ValidatorInfo, error) {
	if vp.GetLatestValidatorInfosCalled != nil {
		return vp.GetLatestValidatorInfosCalled()
	}
	return nil, nil
}

// IsInterfaceNil -
func (vp *ValidatorsProviderStub) IsInterfaceNil() bool {
	return vp == nil
}
