package dataValidators

import "github.com/ElrondNetwork/elrond-go/data"

// nilTxHandlerProcessValidator represents a tx handler validator that doesn't check the validity of provided txHandler
type nilTxHandlerProcessValidator struct {
}

// NewNilTxHandlerProcessValidator creates a new nil tx handler validator instance
func NewNilTxHandlerProcessValidator() (*nilTxHandlerProcessValidator, error) {
	return &nilTxHandlerProcessValidator{}, nil
}

// CheckTxHandlerValid is a nil implementation that will return true
func (nthv *nilTxHandlerProcessValidator) CheckTxHandlerValid(txHandler data.TransactionHandler) bool {
	return true
}
