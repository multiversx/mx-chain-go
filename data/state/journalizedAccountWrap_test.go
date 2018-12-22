package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewJournalizedAccountWrap_WithNilAccountAdapterShouldErr(t *testing.T) {
	_, err := state.NewJournalizedAccountWrap(mock.NewTrackableAccountWrapMock(), nil)

	assert.NotNil(t, err)
}

func TestNewJournalizedAccountWrap_WithNilModifyingAccountWrapperShouldErr(t *testing.T) {
	_, err := state.NewJournalizedAccountWrap(nil, mock.NewAccountsAdapterMock())

	assert.NotNil(t, err)
}

func TestNewJournalizedAccountWrapFromAccountContainer_WithNilAddressContainerShouldErr(t *testing.T) {
	_, err := state.NewJournalizedAccountWrapFromAccountContainer(
		nil,
		state.NewAccount(),
		mock.NewAccountsAdapterMock(),
	)

	assert.Equal(t, state.ErrNilAddressContainer, err)
}

func TestNewJournalizedAccountWrapFromAccountContainer_WithNilAccountShouldErr(t *testing.T) {
	_, err := state.NewJournalizedAccountWrapFromAccountContainer(
		mock.NewAddressMock(),
		nil,
		mock.NewAccountsAdapterMock(),
	)

	assert.Equal(t, state.ErrNilAccount, err)
}

func TestNewJournalizedAccountWrapFromAccountContainer_WithNilAccountsAdapterShouldErr(t *testing.T) {
	_, err := state.NewJournalizedAccountWrapFromAccountContainer(
		mock.NewAddressMock(),
		state.NewAccount(),
		nil,
	)

	assert.Equal(t, state.ErrNilAccountsAdapter, err)
}

func TestJournalizedAccountWrap_SetNonceWithJournal(t *testing.T) {
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

func TestJournalizedAccountWrap_SetBalanceWithJournal(t *testing.T) {
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

func TestJournalizedAccountWrap_SetCodeHashWithJournal(t *testing.T) {
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

func TestJournalizedAccountWrap_SetRootHashWithJournal(t *testing.T) {
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

func TestJournalizedAccountWrap_AppendDataRegistrationWithJournal(t *testing.T) {
	var jeAdded state.JournalEntry

	wasCalledSave := false
	wasCalledAppend := false

	acntAdapter := mock.NewAccountsAdapterMock()
	acntAdapter.SaveAccountStateCalled = func(acountWrapper state.JournalizedAccountWrapper) error {
		wasCalledSave = true

		return nil
	}

	acntAdapter.AddJournalEntryCalled = func(je state.JournalEntry) {
		jeAdded = je
	}

	trackableAccount := mock.NewTrackableAccountWrapMock()
	trackableAccount.AppendRegistrationDataCalled = func(data *state.RegistrationData) error {
		wasCalledAppend = true
		return nil
	}

	jaw, err := state.NewJournalizedAccountWrap(trackableAccount, acntAdapter)
	assert.Nil(t, err)
	err = jaw.AppendDataRegistrationWithJournal(&state.RegistrationData{})
	assert.Nil(t, err)
	assert.True(t, wasCalledSave)
	assert.True(t, wasCalledAppend)
	assert.NotNil(t, jeAdded)
}
