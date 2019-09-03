package factory

import "github.com/ElrondNetwork/elrond-go/data/state"

// PeerAccountCreator has a method to create a new peer account
type PeerAccountCreator struct {
}

// NewPeerAccountCreator creates a peer account creator
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

// IsInterfaceNil returns true if there is no value under the interface
func (c *PeerAccountCreator) IsInterfaceNil() bool {
	if c == nil {
		return true
	}
	return false
}
