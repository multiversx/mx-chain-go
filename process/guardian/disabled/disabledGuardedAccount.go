package disabled

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type disabledGuardedAccount struct{}

// NewDisabledGuardedAccountHandler returns a disabled implementation
func NewDisabledGuardedAccountHandler() *disabledGuardedAccount {
	return &disabledGuardedAccount{}
}

// GetActiveGuardian returns nil, nil as this is a disabled implementation
func (dga *disabledGuardedAccount) GetActiveGuardian(_ vmcommon.UserAccountHandler) ([]byte, error) {
	return nil, nil
}

// SetGuardian returns nil as this is a disabled implementation
func (dga *disabledGuardedAccount) SetGuardian(_ vmcommon.UserAccountHandler, _ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dga *disabledGuardedAccount) IsInterfaceNil() bool {
	return dga == nil
}
