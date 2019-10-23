package dataValidators

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

// nilHeaderValidator represents a header handler validator that doesn't check the validity of provided headerHandler
type nilHeaderValidator struct {
}

// NewNilHeaderValidator creates a new nil header handler validator instance
func NewNilHeaderValidator() (*nilHeaderValidator, error) {
	return &nilHeaderValidator{}, nil
}

// HeaderValidForProcessing is a nil implementation that will return true
func (nhv *nilHeaderValidator) HeaderValidForProcessing(headerHandler process.HdrValidatorHandler) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (nhv *nilHeaderValidator) IsInterfaceNil() bool {
	if nhv == nil {
		return true
	}
	return false
}
