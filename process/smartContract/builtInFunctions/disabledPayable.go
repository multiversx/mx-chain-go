package builtInFunctions

import "github.com/ElrondNetwork/elrond-go/process"

// disabledPayableHandler is a disabled payable handler implements PayableHandler interface but it is disabled
type disabledPayableHandler struct {
}

// IsPayable returns false and error as this is a disabled payable handler
func (d *disabledPayableHandler) IsPayable(_ []byte) (bool, error) {
	return false, process.ErrAccountNotPayable
}

// IsInterfaceNil returns true if underlying object is nil
func (d *disabledPayableHandler) IsInterfaceNil() bool {
	return d == nil
}
