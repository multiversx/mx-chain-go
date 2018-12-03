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

func (tje *journalEntryMock) DirtyAddress() *state.Address {
	return tje.Addr
}

func jCreateRandomAddress() *state.Address {
	buff := make([]byte, state.AdrLen)
	rand.Read(buff)

	addr, err := state.NewAddress(buff)
	if err != nil {
		panic(err)
	}

	return addr
}

func TestJournal_AddEntry_ValidValue_ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJournal(nil)

	jem := journalEntryMock{Addr: jCreateRandomAddress()}

	j.AddEntry(&jem)
	assert.Equal(t, 1, len(j.Entries()))
	assert.Equal(t, 1, j.DirtyAddresses()[jem.DirtyAddress()])
	assert.Equal(t, 1, j.Len())
}

func TestJournal_RevertFromSnapshot_OutOfBound_ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJournal(nil)

	jem := journalEntryMock{Addr: jCreateRandomAddress()}

	j.AddEntry(&jem)

	err := j.RevertFromSnapshot(3)
	assert.Nil(t, err)
	assert.Equal(t, 1, j.Len())
	assert.Equal(t, 0, jem.RevertCalled)

	j.RevertFromSnapshot(2)
	assert.Equal(t, 1, j.Len())
	assert.Equal(t, 0, jem.RevertCalled)

	j.RevertFromSnapshot(0)
	assert.Equal(t, 0, j.Len())
	assert.Equal(t, 1, jem.RevertCalled)
}

func TestJournal_RevertFromSnapshot_SingleEntry_ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJournal(nil)

	jem := journalEntryMock{Addr: jCreateRandomAddress()}

	j.AddEntry(&jem)

	err := j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, 0, j.Len())
	assert.Equal(t, 1, jem.RevertCalled)
}

func TestJournal_RevertFromSnapshot_5EntriesIdx3_ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJournal(nil)

	jem := journalEntryMock{Addr: jCreateRandomAddress()}

	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)

	err := j.RevertFromSnapshot(3)
	assert.Nil(t, err)
	assert.Equal(t, 3, j.Len())
	assert.Equal(t, 2, jem.RevertCalled)
}

func TestJournal_RevertFromSnapshot_5EntriesIdx0_ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJournal(nil)

	jem := journalEntryMock{Addr: jCreateRandomAddress()}

	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)

	err := j.RevertFromSnapshot(0)
	assert.Nil(t, err)
	assert.Equal(t, 0, j.Len())
	assert.Equal(t, 5, jem.RevertCalled)
}

func TestJournal_RevertFromSnapshot_5EntriesIdx4_ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJournal(nil)

	jem := journalEntryMock{Addr: jCreateRandomAddress()}

	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)

	err := j.RevertFromSnapshot(4)
	assert.Nil(t, err)
	assert.Equal(t, 4, j.Len())
	assert.Equal(t, 1, jem.RevertCalled)
}

func TestJournal_RevertFromSnapshot_SingleEntryErrors_ShouldRetErr(t *testing.T) {
	t.Parallel()

	j := state.NewJournal(nil)

	jem := journalEntryMock{Addr: jCreateRandomAddress()}
	jem.FailRevert = true

	j.AddEntry(&jem)

	err := j.RevertFromSnapshot(0)
	assert.NotNil(t, err)
	assert.Equal(t, 1, j.Len())
	assert.Equal(t, 0, jem.RevertCalled)
}

func TestJournal_Clear(t *testing.T) {
	t.Parallel()

	j := state.NewJournal(nil)

	jem := journalEntryMock{Addr: jCreateRandomAddress()}

	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)

	assert.Equal(t, 5, j.Len())

	j.Clear()

	assert.Equal(t, 0, j.Len())
}
