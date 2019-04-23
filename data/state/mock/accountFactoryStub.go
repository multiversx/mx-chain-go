package mock

import "github.com/ElrondNetwork/elrond-go-sandbox/data/state"

type AccountsFactoryStub struct {
	CreateAccountCalled func(address state.AddressContainer, tracker state.AccountTracker) state.AccountWrapper
}

func (afs *AccountsFactoryStub) CreateAccount(address state.AddressContainer, tracker state.AccountTracker) state.AccountWrapper {
	return afs.CreateAccountCalled(address, tracker)
}
