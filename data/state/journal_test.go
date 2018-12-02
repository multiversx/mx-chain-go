package state_test

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/stretchr/testify/assert"
)

type journalEntryMock struct {
	RevertCalled int
	FailRevert   bool

	Addr *state.Address
}

func (tje *journalEntryMock) Revert(accounts state.AccountsHandler) error {
	if tje.FailRevert {
		return errors.New("testJournalEntry general failure")
	}

	tje.RevertCalled++
	return nil
}

func (tje *journalEntryMock) DirtiedAddress() state.AddressHandler {
	return tje.Addr
}

func jurnalCreateRandomAddress() *state.Address {
	buff := make([]byte, state.AdrLen)
	rand.Read(buff)

	addr, err := state.NewAddress(buff)
	if err != nil {
		panic(err)
	}

	return addr
}

func TestJournalAddEntryValidValueShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJournal()

	jem := journalEntryMock{Addr: jurnalCreateRandomAddress()}

	j.AddEntry(&jem)
	assert.Equal(t, 1, len(j.Entries()))
	assert.Equal(t, 1, j.Len())
}

func TestJournalRevertToSnapshotOutOfBoundShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJournal()

	jem := journalEntryMock{Addr: jurnalCreateRandomAddress()}

	j.AddEntry(&jem)

	err := j.RevertToSnapshot(3, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, j.Len())
	assert.Equal(t, 0, jem.RevertCalled)

	err = j.RevertToSnapshot(2, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, j.Len())
	assert.Equal(t, 0, jem.RevertCalled)

	err = j.RevertToSnapshot(0, nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, j.Len())
	assert.Equal(t, 1, jem.RevertCalled)
}

func TestJournalRevertToSnapshotSingleEntryShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJournal()

	jem := journalEntryMock{Addr: jurnalCreateRandomAddress()}

	j.AddEntry(&jem)

	err := j.RevertToSnapshot(0, nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, j.Len())
	assert.Equal(t, 1, jem.RevertCalled)
}

func TestJournalRevertToSnapshot5EntriesIdx3ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJournal()

	jem := journalEntryMock{Addr: jurnalCreateRandomAddress()}

	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)

	err := j.RevertToSnapshot(3, nil)
	assert.Nil(t, err)
	assert.Equal(t, 3, j.Len())
	assert.Equal(t, 2, jem.RevertCalled)
}

func TestJournalRevertToSnapshot5EntriesIdx0ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJournal()

	jem := journalEntryMock{Addr: jurnalCreateRandomAddress()}

	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)

	err := j.RevertToSnapshot(0, nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, j.Len())
	assert.Equal(t, 5, jem.RevertCalled)
}

func TestJournalRevertToSnapshot5EntriesIdx4ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJournal()

	jem := journalEntryMock{Addr: jurnalCreateRandomAddress()}

	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)

	err := j.RevertToSnapshot(4, nil)
	assert.Nil(t, err)
	assert.Equal(t, 4, j.Len())
	assert.Equal(t, 1, jem.RevertCalled)
}

func TestJournalRevertToSnapshotSingleEntryErrorsShouldRetErr(t *testing.T) {
	t.Parallel()

	j := state.NewJournal()

	jem := journalEntryMock{Addr: jurnalCreateRandomAddress()}
	jem.FailRevert = true

	j.AddEntry(&jem)

	err := j.RevertToSnapshot(0, nil)
	assert.NotNil(t, err)
	assert.Equal(t, 1, j.Len())
	assert.Equal(t, 0, jem.RevertCalled)
}

func TestJournalClear(t *testing.T) {
	t.Parallel()

	j := state.NewJournal()

	jem := journalEntryMock{Addr: jurnalCreateRandomAddress()}

	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)

	assert.Equal(t, 5, j.Len())

	j.Clear()

	assert.Equal(t, 0, j.Len())
}
