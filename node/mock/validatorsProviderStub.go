package mock

import (
	"github.com/multiversx/mx-chain-core-go/data/validator"
)

// ValidatorsProviderStub -
type ValidatorsProviderStub struct {
	GetLatestValidatorsCalled func() map[string]*validator.ValidatorStatistics
}

// GetLatestValidators -
func (vp *ValidatorsProviderStub) GetLatestValidators() map[string]*validator.ValidatorStatistics {
	if vp.GetLatestValidatorsCalled != nil {
		return vp.GetLatestValidatorsCalled()
	}
	return nil
}

// Close -
func (vp *ValidatorsProviderStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (vp *ValidatorsProviderStub) IsInterfaceNil() bool {
	return vp == nil
}
