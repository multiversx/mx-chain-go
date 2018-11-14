package state_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/stretchr/testify/assert"
)

type jurnalEntryMock struct {
	RevertCalled int
	FailRevert   bool

	Addr *state.Address
}

func (tje *jurnalEntryMock) Revert(accounts state.AccountsHandler) error {
	if tje.FailRevert {
		return errors.New("testJurnalEntry general failure")
	}

	tje.RevertCalled++
	return nil
}

func (tje *jurnalEntryMock) DirtyAddress() *state.Address {
	return tje.Addr
}

func TestJurnal_AddEntry_ValidValue_ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJurnal(nil)

	jem := jurnalEntryMock{Addr: state.HexToAddress("1234")}

	j.AddEntry(&jem)
	assert.Equal(t, 1, len(j.Entries()))
	assert.Equal(t, uint32(1), j.DirtyAddresses()[jem.DirtyAddress()])
	assert.Equal(t, uint32(1), j.Len())
}

func TestJurnal_RevertFromSnapshot_OutOfBound_ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJurnal(nil)

	jem := jurnalEntryMock{Addr: state.HexToAddress("1234")}

	j.AddEntry(&jem)

	err := j.RevertFromSnapshot(uint32(3))
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), j.Len())
	assert.Equal(t, 0, jem.RevertCalled)

	j.RevertFromSnapshot(uint32(2))
	assert.Equal(t, uint32(1), j.Len())
	assert.Equal(t, 0, jem.RevertCalled)

	j.RevertFromSnapshot(uint32(0))
	assert.Equal(t, uint32(1), j.Len())
	assert.Equal(t, 0, jem.RevertCalled)
}

func TestJurnal_RevertFromSnapshot_SingleEntry_ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJurnal(nil)

	jem := jurnalEntryMock{Addr: state.HexToAddress("1234")}

	j.AddEntry(&jem)

	err := j.RevertFromSnapshot(uint32(1))
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), j.Len())
	assert.Equal(t, 1, jem.RevertCalled)
}

func TestJurnal_RevertFromSnapshot_5EntriesIdx3_ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJurnal(nil)

	jem := jurnalEntryMock{Addr: state.HexToAddress("1234")}

	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)

	err := j.RevertFromSnapshot(uint32(4))
	assert.Nil(t, err)
	assert.Equal(t, uint32(3), j.Len())
	assert.Equal(t, 2, jem.RevertCalled)
}

func TestJurnal_RevertFromSnapshot_5EntriesIdx0_ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJurnal(nil)

	jem := jurnalEntryMock{Addr: state.HexToAddress("1234")}

	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)

	err := j.RevertFromSnapshot(uint32(1))
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), j.Len())
	assert.Equal(t, 5, jem.RevertCalled)
}

func TestJurnal_RevertFromSnapshot_5EntriesIdx4_ShouldWork(t *testing.T) {
	t.Parallel()

	j := state.NewJurnal(nil)

	jem := jurnalEntryMock{Addr: state.HexToAddress("1234")}

	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)
	j.AddEntry(&jem)

	err := j.RevertFromSnapshot(uint32(5))
	assert.Nil(t, err)
	assert.Equal(t, uint32(4), j.Len())
	assert.Equal(t, 1, jem.RevertCalled)
}

func TestJurnal_RevertFromSnapshot_SingleEntryErrors_ShouldRetErr(t *testing.T) {
	t.Parallel()

	j := state.NewJurnal(nil)

	jem := jurnalEntryMock{Addr: state.HexToAddress("1234")}
	jem.FailRevert = true

	j.AddEntry(&jem)

	err := j.RevertFromSnapshot(uint32(1))
	assert.NotNil(t, err)
	assert.Equal(t, uint32(1), j.Len())
	assert.Equal(t, 0, jem.RevertCalled)
}
