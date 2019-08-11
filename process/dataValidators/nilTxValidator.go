package dataValidators

import "github.com/ElrondNetwork/elrond-go/data"

// nilTxValidator represents a tx handler validator that doesn't check the validity of provided txHandler
type nilTxValidator struct {
}

// NewNilTxValidator creates a new nil tx handler validator instance
func NewNilTxValidator() (*nilTxValidator, error) {
	return &nilTxValidator{}, nil
}

// IsTxValidForProcessing is a nil implementation that will return true
func (ntv *nilTxValidator) IsTxValidForProcessing(txHandler data.TransactionHandler) bool {
	return true
}
