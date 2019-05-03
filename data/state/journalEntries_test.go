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
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestNewJournalEntryNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	accnt, _ := state.NewAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewJournalEntryNonce(accnt, 0)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
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
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestNewJournalEntryBalance_ShouldWork(t *testing.T) {
	t.Parallel()

	accnt, _ := state.NewAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewJournalEntryBalance(accnt, nil)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
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
