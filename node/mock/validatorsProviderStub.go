package mock

import "github.com/ElrondNetwork/elrond-go/data/state"

// ValidatorsProviderStub -
type ValidatorsProviderStub struct {
	GetLatestValidatorsCalled func() map[string]*state.ValidatorApiResponse
}

// GetLatestValidators -
func (vp *ValidatorsProviderStub) GetLatestValidators() map[string]*state.ValidatorApiResponse {
	if vp.GetLatestValidatorsCalled != nil {
		return vp.GetLatestValidatorsCalled()
	}
	return nil
}

// IsInterfaceNil -
func (vp *ValidatorsProviderStub) IsInterfaceNil() bool {
	return vp == nil
}
