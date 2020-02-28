package mock

import "github.com/ElrondNetwork/elrond-go/data/state"

// AccountTrackerStub -
type AccountTrackerStub struct {
	SaveAccountCalled func(accountHandler state.AccountHandler) error
	JournalizeCalled  func(entry state.JournalEntry)
}

// SaveAccount -
func (ats *AccountTrackerStub) SaveAccount(accountHandler state.AccountHandler) error {
	if ats.SaveAccountCalled != nil {
		return ats.SaveAccountCalled(accountHandler)
	}
	return nil
}

// Journalize -
func (ats *AccountTrackerStub) Journalize(entry state.JournalEntry) {
	if ats.JournalizeCalled != nil {
		ats.JournalizeCalled(entry)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ats *AccountTrackerStub) IsInterfaceNil() bool {
	return ats == nil
}
