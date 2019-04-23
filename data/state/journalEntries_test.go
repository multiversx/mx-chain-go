package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

//------- JournalEntryNonce

func TestNewJournalEntryNonce_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryNonce(nil, 0)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountWrapper, err)
}

func TestNewJournalEntryNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryNonce(mock.NewAccountWrapMock(nil, nil), 0)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestJournalEntryNonce_WrongAccountTypeShouldErr(t *testing.T) {
	t.Parallel()

	accnt := mock.NewAccountWrapMock(nil, nil)
	entry, _ := state.NewJournalEntryNonce(accnt, 0)
	_, err := entry.Revert()

	assert.Equal(t, state.ErrWrongTypeAssertion, err)
}

func TestJournalEntryNonce_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	nonce := uint64(445)
	accnt, _ := state.NewAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewJournalEntryNonce(accnt, nonce)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, nonce, accnt.Nonce)
}

//------- JournalEntryBalance

func TestNewJournalEntryBalance_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryBalance(nil, nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountWrapper, err)
}

func TestNewJournalEntryBalance_ShouldWork(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryBalance(mock.NewAccountWrapMock(nil, nil), nil)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestNewJournalEntryBalance_WrongAccountTypeShouldErr(t *testing.T) {
	t.Parallel()

	accnt := mock.NewAccountWrapMock(nil, nil)
	entry, _ := state.NewJournalEntryBalance(accnt, nil)
	_, err := entry.Revert()

	assert.Equal(t, state.ErrWrongTypeAssertion, err)
}

func TestNewJournalEntryBalance_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	balance := big.NewInt(34)
	accnt, _ := state.NewAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewJournalEntryBalance(accnt, balance)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, balance, accnt.Balance)
}
