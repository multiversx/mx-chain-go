package dataValidators

import "github.com/ElrondNetwork/elrond-go/data"

// nilHeaderValidator represents a header handler validator that doesn't check the validity of provided headerHandler
type nilHeaderValidator struct {
}

// NewNilHeaderValidator creates a new nil header handler validator instance
func NewNilHeaderValidator() (*nilHeaderValidator, error) {
	return &nilHeaderValidator{}, nil
}

// IsHeaderValidForProcessing is a nil implementation that will return true
func (nhv *nilHeaderValidator) IsHeaderValidForProcessing(headerHandler data.HeaderHandler) bool {
	return true
}
