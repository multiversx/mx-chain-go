package factory

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/accounts"
)

// AccountCreator has method to create a new account
type AccountCreator struct {
}

// NewAccountCreator creates an account creator
func NewAccountCreator() state.AccountFactory {
	return &AccountCreator{}
}

// CreateAccount calls the new Account creator and returns the result
func (ac *AccountCreator) CreateAccount(address state.AddressContainer) (state.AccountHandler, error) {
	return accounts.NewUserAccount(address)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ac *AccountCreator) IsInterfaceNil() bool {
	return ac == nil
}
