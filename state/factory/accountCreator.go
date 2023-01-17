package factory

import (
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// AccountCreator has method to create a new account
type AccountCreator struct {
}

// NewAccountCreator creates an account creator
func NewAccountCreator() state.AccountFactory {
	return &AccountCreator{}
}

// CreateAccount calls the new Account creator and returns the result
func (ac *AccountCreator) CreateAccount(address []byte) (vmcommon.AccountHandler, error) {
	return state.NewUserAccount(address)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ac *AccountCreator) IsInterfaceNil() bool {
	return ac == nil
}
