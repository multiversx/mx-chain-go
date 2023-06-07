package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data/guardians"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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

// HasActiveGuardian returns false as this is a disabled implementation
func (dga *disabledGuardedAccount) HasActiveGuardian(_ common.UserAccountHandler) bool {
	return false
}

// HasPendingGuardian returns false as this is a disabled implementation
func (dga *disabledGuardedAccount) HasPendingGuardian(_ common.UserAccountHandler) bool {
	return false
}

// SetGuardian returns nil as this is a disabled implementation
func (dga *disabledGuardedAccount) SetGuardian(_ vmcommon.UserAccountHandler, _ []byte, _ []byte, _ []byte) error {
	return nil
}

// CleanOtherThanActive does nothing as this is a disabled implementation
func (dga *disabledGuardedAccount) CleanOtherThanActive(_ vmcommon.UserAccountHandler) {}

// GetConfiguredGuardians returns nil, nil, nil as this is a disabled component
func (dga *disabledGuardedAccount) GetConfiguredGuardians(_ common.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
	return nil, nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dga *disabledGuardedAccount) IsInterfaceNil() bool {
	return dga == nil
}
