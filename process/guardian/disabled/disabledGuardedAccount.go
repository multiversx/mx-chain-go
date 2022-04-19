package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
)

type disabledGuardedAccount struct{}

// NewDisabledGuardedAccountHandler returns a disabled implementation
func NewDisabledGuardedAccountHandler() *disabledGuardedAccount {
	return &disabledGuardedAccount{}
}

// GetActiveGuardian returns nil, nil as this is a disabled implementation
func (dga *disabledGuardedAccount) GetActiveGuardian(uah data.UserAccountHandler) ([]byte, error) {
	return nil, nil
}

// SetGuardian returns nil as this is a disabled implementation
func (dga *disabledGuardedAccount) SetGuardian(_ data.UserAccountHandler, _ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dga *disabledGuardedAccount) IsInterfaceNil() bool {
	return dga == nil
}
