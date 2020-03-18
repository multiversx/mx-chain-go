package factory

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// AccountCreator has method to create a new account
type AccountCreator struct {
}

// NewAccountCreator creates an account creator
func NewAccountCreator() state.AccountFactory {
	return &AccountCreator{}
}

// CreateAccount calls the new Account creator and returns the result
func (ac *AccountCreator) CreateAccount(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
	account, err := state.NewAccount(address, tracker)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// GetType returns the account factory type
func (ac *AccountCreator) GetType() state.Type {
	return state.UserAccount
}

// IsInterfaceNil returns true if there is no value under the interface
func (ac *AccountCreator) IsInterfaceNil() bool {
	return ac == nil
}
