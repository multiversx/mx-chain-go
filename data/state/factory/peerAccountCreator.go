package factory

import "github.com/ElrondNetwork/elrond-go/data/state"

// MetaAccountCreator has a method to create a new meta accound
type PeerAccountCreator struct {
}

// NewPeerAccountCreator creates a meta account creator
func NewPeerAccountCreator() state.AccountFactory {
	return &PeerAccountCreator{}
}

// CreateAccount calls the new Account creator and returns the result
func (c *PeerAccountCreator) CreateAccount(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
	account, err := state.NewPeerAccount(address, tracker)
	if err != nil {
		return nil, err
	}

	return account, nil
}
