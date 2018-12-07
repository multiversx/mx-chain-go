package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type JournalEntryMock struct {
	RevertCalled         func(accountsAdapter state.AccountsAdapter) error
	DirtiedAddressCalled func() state.AddressContainer
}

func NewJournalEntryMock() *JournalEntryMock {
	return &JournalEntryMock{}
}

func (jem *JournalEntryMock) Revert(accountsAdapter state.AccountsAdapter) error {
	return jem.RevertCalled(accountsAdapter)
}

func (jem *JournalEntryMock) DirtiedAddress() state.AddressContainer {
	return jem.DirtiedAddressCalled()
}
