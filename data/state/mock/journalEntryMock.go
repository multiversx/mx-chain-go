package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type JournalEntryMock struct {
	RevertCalled func() (state.AccountWrapper, error)
}

func NewJournalEntryMock() *JournalEntryMock {
	return &JournalEntryMock{}
}

func (jem *JournalEntryMock) Revert() (state.AccountWrapper, error) {
	return jem.RevertCalled()
}
