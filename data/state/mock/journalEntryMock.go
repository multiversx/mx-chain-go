package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type JournalEntryMock struct {
	RevertCalled func() (state.AccountHandler, error)
}

func NewJournalEntryMock() *JournalEntryMock {
	return &JournalEntryMock{}
}

func (jem *JournalEntryMock) Revert() (state.AccountHandler, error) {
	return jem.RevertCalled()
}
