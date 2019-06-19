package mock

import "github.com/ElrondNetwork/elrond-go/data/state"

type AccountTrackerStub struct {
	SaveAccountCalled func(accountHandler state.AccountHandler) error
	JournalizeCalled  func(entry state.JournalEntry)
}

func (ats *AccountTrackerStub) SaveAccount(accountHandler state.AccountHandler) error {
	return ats.SaveAccountCalled(accountHandler)
}

func (ats *AccountTrackerStub) Journalize(entry state.JournalEntry) {
	ats.JournalizeCalled(entry)
}
