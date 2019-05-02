package factory

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

// AccountCreator has method to create a new account
type AccountCreator struct {
}

// NewAccountCreator creates an account creator
func NewAccountCreator() state.AccountFactory {
	return &AccountCreator{}
}

// CreateAccount calls the new Account creator and returns the result
func (c *AccountCreator) CreateAccount(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
	account, err := state.NewAccount(address, tracker)
	if err != nil {
		return nil, err
	}

	return account, nil
}
