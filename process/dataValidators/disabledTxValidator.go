package dataValidators

import (
	"github.com/multiversx/mx-chain-go/process"
)

// disabledTxValidator represents a tx handler validator that doesn't check the validity of provided txHandler
type disabledTxValidator struct {
}

// NewDisabledTxValidator creates a new disabled tx handler validator instance
func NewDisabledTxValidator() *disabledTxValidator {
	return &disabledTxValidator{}
}

// CheckTxValidity is a disabled implementation that will return nil
func (dtv *disabledTxValidator) CheckTxValidity(_ process.InterceptedTransactionHandler) error {
	return nil
}

// CheckTxWhiteList is a disabled implementation that will return nil
func (dtv *disabledTxValidator) CheckTxWhiteList(_ process.InterceptedData) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dtv *disabledTxValidator) IsInterfaceNil() bool {
	return dtv == nil
}
