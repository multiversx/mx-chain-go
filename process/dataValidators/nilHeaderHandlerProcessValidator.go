package dataValidators

import "github.com/ElrondNetwork/elrond-go/data"

// nilHeaderHandlerProcessValidator represents a header handler validator that doesn't check the validity of provided headerHandler
type nilHeaderHandlerProcessValidator struct {
}

// NewNilHeaderHandlerProcessValidator creates a new nil header handler validator instance
func NewNilHeaderHandlerProcessValidator() (*nilHeaderHandlerProcessValidator, error) {
	return &nilHeaderHandlerProcessValidator{}, nil
}

// CheckHeaderHandlerValid is a nil implementation that will return true
func (nhhv *nilHeaderHandlerProcessValidator) CheckHeaderHandlerValid(headerHandler data.HeaderHandler) bool {
	return true
}
