package mock

import "github.com/ElrondNetwork/elrond-go-sandbox/data/state"

type AccountTrackerStub struct {
	SaveAccountCalled func(accountWrapper state.AccountWrapper) error
	JournalizeCalled  func(entry state.JournalEntry)
}

func (ats *AccountTrackerStub) SaveAccount(accountWrapper state.AccountWrapper) error {
	return ats.SaveAccountCalled(accountWrapper)
}

func (ats *AccountTrackerStub) Journalize(entry state.JournalEntry) {
	ats.JournalizeCalled(entry)
}
