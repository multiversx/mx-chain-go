package factory

import "github.com/ElrondNetwork/elrond-go-sandbox/data/state"

// MetaAccountCreator has a method to create a new meta accound
type MetaAccountCreator struct {
}

// NewMetaAccountCreator creates a meta account creator
func NewMetaAccountCreator() state.AccountFactory {
	return &MetaAccountCreator{}
}

// CreateAccount calls the new Account creator and returns the result
func (c *MetaAccountCreator) CreateAccount(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
	account, err := state.NewMetaAccount(address, tracker)
	if err != nil {
		return nil, err
	}

	return account, nil
}
