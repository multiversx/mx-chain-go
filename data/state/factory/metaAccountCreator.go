package factory

import "github.com/ElrondNetwork/elrond-go-sandbox/data/state"

type MetaAccountCreator struct {
}

// NewAccountCreator creates an account creator
func NewMetaAccountCreator() (*AccountCreator, error) {
	return &AccountCreator{}, nil
}

// CreateAccount calls the new Account creator and returns the result
func (c *MetaAccountCreator) CreateAccount(address state.AddressContainer, tracker state.AccountTracker) (state.AccountWrapper, error) {
	account, err := state.NewMetaAccount(address, tracker)
	if err != nil {
		return nil, err
	}

	return account, nil
}
