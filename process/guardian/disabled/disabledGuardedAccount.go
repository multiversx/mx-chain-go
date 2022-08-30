package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/data/guardians"
	"github.com/ElrondNetwork/elrond-go/state"
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

// HasActiveGuardian returns false as this is a disabled implementation
func (dga *disabledGuardedAccount) HasActiveGuardian(_ state.UserAccountHandler) bool {
	return false
}

// HasPendingGuardian returns false as this is a disabled implementation
func (dga *disabledGuardedAccount) HasPendingGuardian(_ state.UserAccountHandler) bool {
	return false
}

// SetGuardian returns nil as this is a disabled implementation
func (dga *disabledGuardedAccount) SetGuardian(_ vmcommon.UserAccountHandler, _ []byte, _ []byte) error {
	return nil
}

// CleanOtherThanActive does nothing as this is a disabled implementation
func (dga *disabledGuardedAccount) CleanOtherThanActive(_ vmcommon.UserAccountHandler) {}

// GetConfiguredGuardians returns nil, nil, nil as this is a disabled component
func (dga *disabledGuardedAccount) GetConfiguredGuardians(_ state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
	return nil, nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dga *disabledGuardedAccount) IsInterfaceNil() bool {
	return dga == nil
}
