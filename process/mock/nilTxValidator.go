package mock

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

// CheckTxValidity is a nil implementation that will return nil
func (ntv *nilTxValidator) CheckTxValidity(_ process.TxValidatorHandler) error {
	return nil
}

// CheckTxWhiteList is a nil implementation that will return nil
func (ntv *nilTxValidator) CheckTxWhiteList(_ process.InterceptedData) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ntv *nilTxValidator) IsInterfaceNil() bool {
	return ntv == nil
}
