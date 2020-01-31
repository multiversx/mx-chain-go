package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/stretchr/testify/assert"
)

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
