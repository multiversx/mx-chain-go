package dataValidators

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

// nilTxValidator represents a tx handler validator that doesn't check the validity of provided txHandler
type nilTxValidator struct {
}

// NewNilTxValidator creates a new nil tx handler validator instance
func NewNilTxValidator() (*nilTxValidator, error) {
	return &nilTxValidator{}, nil
}

// IsTxValidForProcessing is a nil implementation that will return true
func (ntv *nilTxValidator) IsTxValidForProcessing(interceptedTx process.TxValidatorHandler) bool {
	return true
}

// NumRejectedTxs will return number of rejected transaction
func (ntv *nilTxValidator) NumRejectedTxs() uint64 {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (ntv *nilTxValidator) IsInterfaceNil() bool {
	if ntv == nil {
		return true
	}
	return false
}
