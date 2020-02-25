package state_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state/accounts"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/stretchr/testify/assert"
)

func TestNewJournalEntryDataTrieUpdates_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	trieUpdates := make(map[string][]byte)
	trieUpdates["a"] = []byte("b")
	entry, err := state.NewJournalEntryDataTrieUpdates(trieUpdates, nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestNewJournalEntryDataTrieUpdates_EmptyTrieUpdatesShouldErr(t *testing.T) {
	t.Parallel()

	trieUpdates := make(map[string][]byte)
	accnt, _ := accounts.NewUserAccount(mock.NewAddressMock())
	entry, err := state.NewJournalEntryDataTrieUpdates(trieUpdates, accnt)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilOrEmptyDataTrieUpdates, err)
}

func TestNewJournalEntryDataTrieUpdates_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	trieUpdates := make(map[string][]byte)
	trieUpdates["a"] = []byte("b")
	accnt, _ := accounts.NewUserAccount(mock.NewAddressMock())
	entry, err := state.NewJournalEntryDataTrieUpdates(trieUpdates, accnt)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(entry))
}

func TestJournalEntryDataTrieUpdates_RevertFailsWhenUpdateFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error")

	trieUpdates := make(map[string][]byte)
	trieUpdates["a"] = []byte("b")
	accnt := mock.NewAccountWrapMock(nil)

	trie := &mock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return expectedErr
		},
	}

	accnt.SetDataTrie(trie)
	//accnt, _ := state.NewAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewJournalEntryDataTrieUpdates(trieUpdates, accnt)

	acc, err := entry.Revert()
	assert.Nil(t, acc)
	assert.Equal(t, expectedErr, err)
}

func TestJournalEntryDataTrieUpdates_RevertFailsWhenAccountRootFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error")

	trieUpdates := make(map[string][]byte)
	trieUpdates["a"] = []byte("b")
	accnt := mock.NewAccountWrapMock(nil)

	trie := &mock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		RootCalled: func() ([]byte, error) {
			return nil, expectedErr
		},
	}

	accnt.SetDataTrie(trie)
	entry, _ := state.NewJournalEntryDataTrieUpdates(trieUpdates, accnt)

	acc, err := entry.Revert()
	assert.Nil(t, acc)
	assert.Equal(t, expectedErr, err)
}

func TestJournalEntryDataTrieUpdates_RevertShouldWork(t *testing.T) {
	t.Parallel()

	updateWasCalled := false
	rootWasCalled := false

	trieUpdates := make(map[string][]byte)
	trieUpdates["a"] = []byte("b")
	accnt := mock.NewAccountWrapMock(nil)

	trie := &mock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			updateWasCalled = true
			return nil
		},
		RootCalled: func() ([]byte, error) {
			rootWasCalled = true
			return []byte{}, nil
		},
	}

	accnt.SetDataTrie(trie)
	entry, _ := state.NewJournalEntryDataTrieUpdates(trieUpdates, accnt)

	acc, err := entry.Revert()
	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.True(t, updateWasCalled)
	assert.True(t, rootWasCalled)
}
