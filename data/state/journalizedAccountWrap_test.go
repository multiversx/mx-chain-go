package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewJournalizedAccountWrapWithNilAccountAdapterShouldErr(t *testing.T) {
	_, err := state.NewJournalizedAccountWrap(mock.NewTrackableAccountWrapMock(), nil)

	assert.NotNil(t, err)
}

func TestNewJournalizedAccountWrapWithNilModifyingAccountWrapperShouldErr(t *testing.T) {
	_, err := state.NewJournalizedAccountWrap(nil, mock.NewAccountsAdapterMock())

	assert.NotNil(t, err)
}

func TestNewJournalizedAccountWrapSetNonceWithJournal(t *testing.T) {
	var jeAdded state.JournalEntry

	wasCalledSave := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalledSave = true

		return nil
	}

	acntAdapter.AddJournalEntryCalled = func(je state.JournalEntry) {
		jeAdded = je
	}

	jaw, err := state.NewJournalizedAccountWrap(mock.NewTrackableAccountWrapMock(), acntAdapter)
	assert.Nil(t, err)
	err = jaw.SetNonceWithJournal(1)
	assert.Nil(t, err)
	assert.True(t, wasCalledSave)
	assert.NotNil(t, jeAdded)
}

func TestNewJournalizedAccountWrapSetBalanceWithJournal(t *testing.T) {
	var jeAdded state.JournalEntry

	wasCalledSave := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalledSave = true

		return nil
	}

	acntAdapter.AddJournalEntryCalled = func(je state.JournalEntry) {
		jeAdded = je
	}

	jaw, err := state.NewJournalizedAccountWrap(mock.NewTrackableAccountWrapMock(), acntAdapter)
	assert.Nil(t, err)
	err = jaw.SetBalanceWithJournal(*big.NewInt(1))
	assert.Nil(t, err)
	assert.True(t, wasCalledSave)
	assert.NotNil(t, jeAdded)
}

func TestNewJournalizedAccountWrapSetCodeHashWithJournal(t *testing.T) {
	var jeAdded state.JournalEntry

	wasCalledSave := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalledSave = true

		return nil
	}

	acntAdapter.AddJournalEntryCalled = func(je state.JournalEntry) {
		jeAdded = je
	}

	jaw, err := state.NewJournalizedAccountWrap(mock.NewTrackableAccountWrapMock(), acntAdapter)
	assert.Nil(t, err)
	err = jaw.SetCodeHashWithJournal([]byte("aa"))
	assert.Nil(t, err)
	assert.True(t, wasCalledSave)
	assert.NotNil(t, jeAdded)
}

func TestNewJournalizedAccountWrapSetRootHashWithJournal(t *testing.T) {
	var jeAdded state.JournalEntry

	wasCalledSave := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalledSave = true

		return nil
	}

	acntAdapter.AddJournalEntryCalled = func(je state.JournalEntry) {
		jeAdded = je
	}

	jaw, err := state.NewJournalizedAccountWrap(mock.NewTrackableAccountWrapMock(), acntAdapter)
	assert.Nil(t, err)
	err = jaw.SetRootHashWithJournal([]byte("aa"))
	assert.Nil(t, err)
	assert.True(t, wasCalledSave)
	assert.NotNil(t, jeAdded)
}
