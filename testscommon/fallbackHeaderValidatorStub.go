package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// FallBackHeaderValidatorStub -
type FallBackHeaderValidatorStub struct {
	ShouldApplyFallbackValidationCalled func(headerHandler data.HeaderHandler) bool
}

// ShouldApplyFallbackValidation -
func (fhvs *FallBackHeaderValidatorStub) ShouldApplyFallbackValidation(headerHandler data.HeaderHandler) bool {
	if fhvs.ShouldApplyFallbackValidationCalled != nil {
		return fhvs.ShouldApplyFallbackValidationCalled(headerHandler)
	}
	return false
}

// IsInterfaceNil -
func (fhvs *FallBackHeaderValidatorStub) IsInterfaceNil() bool {
	return fhvs == nil
}
