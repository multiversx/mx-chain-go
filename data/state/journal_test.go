package state_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

func TestJournalAddEntry(t *testing.T) {
	j := state.NewJournal()

	jem := mock.NewJournalEntryMock()
	j.AddEntry(jem)
	assert.Equal(t, 1, len(j.Entries()))
}

func TestJournalAddNilEntryShouldDiscard(t *testing.T) {
	j := state.NewJournal()

	j.AddEntry(nil)
	assert.Equal(t, 0, len(j.Entries()))
}

func TestJournalLen(t *testing.T) {
	j := state.NewJournal()

	jem := mock.NewJournalEntryMock()
	j.AddEntry(jem)
	assert.Equal(t, 1, j.Len())
}

func TestJournalClear(t *testing.T) {
	j := state.NewJournal()

	jem := mock.NewJournalEntryMock()
	j.AddEntry(jem)
	assert.Equal(t, 1, j.Len())

	j.Clear()
	assert.Equal(t, 0, j.Len())
}

func TestJournalRevertWith2EntriesShouldWork(t *testing.T) {
	j := state.NewJournal()

	wasCalled1 := false
	wasCalled2 := false

	jem1 := mock.NewJournalEntryMock()
	jem1.RevertCalled = func(accountsAdapter state.AccountsAdapter) error {
		wasCalled1 = true
		return nil
	}

	jem2 := mock.NewJournalEntryMock()
	jem2.RevertCalled = func(accountsAdapter state.AccountsAdapter) error {
		wasCalled2 = true
		return nil
	}

	j.AddEntry(jem1)
	j.AddEntry(jem2)

	err := j.RevertToSnapshot(0, nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, j.Len())
	assert.True(t, wasCalled1)
	assert.True(t, wasCalled2)
}

func TestJournalRevertWithOutsideBoundsShouldIgnore(t *testing.T) {
	j := state.NewJournal()

	wasCalled1 := false
	wasCalled2 := false

	jem1 := mock.NewJournalEntryMock()
	jem1.RevertCalled = func(accountsAdapter state.AccountsAdapter) error {
		wasCalled1 = true
		return nil
	}

	jem2 := mock.NewJournalEntryMock()
	jem2.RevertCalled = func(accountsAdapter state.AccountsAdapter) error {
		wasCalled2 = true
		return nil
	}

	j.AddEntry(jem1)
	j.AddEntry(jem2)

	err := j.RevertToSnapshot(-1, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, j.Len())
	assert.False(t, wasCalled1)
	assert.False(t, wasCalled2)

	err = j.RevertToSnapshot(3, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, j.Len())
	assert.False(t, wasCalled1)
	assert.False(t, wasCalled2)
}
