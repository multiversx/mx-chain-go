package state_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewJournalEntryCode_NilUpdaterShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryCode(&state.CodeEntry{}, []byte("code hash"), []byte("code hash"), nil, &mock.MarshalizerMock{})
	assert.True(t, check.IfNil(entry))
	assert.Equal(t, state.ErrNilUpdater, err)
}

func TestNewJournalEntryCode_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryCode(&state.CodeEntry{}, []byte("code hash"), []byte("code hash"), &testscommon.TrieStub{}, nil)
	assert.True(t, check.IfNil(entry))
	assert.Equal(t, state.ErrNilMarshalizer, err)
}

func TestNewJournalEntryCode_OkParams(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryCode(&state.CodeEntry{}, []byte("code hash"), []byte("code hash"), &testscommon.TrieStub{}, &mock.MarshalizerMock{})
	assert.Nil(t, err)
	assert.False(t, check.IfNil(entry))
}

func TestJournalEntryCode_OldHashAndNewHashAreNil(t *testing.T) {
	t.Parallel()

	trieStub := &testscommon.TrieStub{}
	entry, _ := state.NewJournalEntryCode(&state.CodeEntry{}, nil, nil, trieStub, &mock.MarshalizerMock{})

	acc, err := entry.Revert()
	assert.Nil(t, err)
	assert.Nil(t, acc)
}

func TestJournalEntryCode_OldHashIsNilAndNewHashIsNotNil(t *testing.T) {
	t.Parallel()

	codeEntry := &state.CodeEntry{
		Code:          []byte("newCode"),
		NumReferences: 1,
	}
	marshalizer := &mock.MarshalizerMock{}

	updateCalled := false
	trieStub := &testscommon.TrieStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return marshalizer.Marshal(codeEntry)
		},
		UpdateCalled: func(key, value []byte) error {
			updateCalled = true
			return nil
		},
	}
	entry, _ := state.NewJournalEntryCode(
		&state.CodeEntry{},
		nil,
		[]byte("newHash"),
		trieStub,
		marshalizer,
	)

	acc, err := entry.Revert()
	assert.Nil(t, err)
	assert.Nil(t, acc)
	assert.True(t, updateCalled)
}

func TestNewJournalEntryAccount_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryAccount(nil)
	assert.True(t, check.IfNil(entry))
	assert.True(t, errors.Is(err, state.ErrNilAccountHandler))
}

func TestNewJournalEntryAccount_OkParams(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryAccount(&mock.AccountWrapMock{})
	assert.Nil(t, err)
	assert.False(t, check.IfNil(entry))
}

func TestJournalEntryAccount_Revert(t *testing.T) {
	t.Parallel()

	expectedAcc := &mock.AccountWrapMock{}
	entry, _ := state.NewJournalEntryAccount(expectedAcc)

	acc, err := entry.Revert()
	assert.Nil(t, err)
	assert.Equal(t, expectedAcc, acc)
}

func TestNewJournalEntryAccountCreation_InvalidAddressShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryAccountCreation([]byte{}, &testscommon.TrieStub{})
	assert.True(t, check.IfNil(entry))
	assert.Equal(t, state.ErrInvalidAddressLength, err)
}

func TestNewJournalEntryAccountCreation_NilUpdaterShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryAccountCreation([]byte("address"), nil)
	assert.True(t, check.IfNil(entry))
	assert.Equal(t, state.ErrNilUpdater, err)
}

func TestNewJournalEntryAccountCreation_OkParams(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryAccountCreation([]byte("address"), &testscommon.TrieStub{})
	assert.Nil(t, err)
	assert.False(t, check.IfNil(entry))
}

func TestJournalEntryAccountCreation_RevertErr(t *testing.T) {
	t.Parallel()

	updateErr := errors.New("update error")
	address := []byte("address")
	ts := &testscommon.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return updateErr
		},
	}
	entry, _ := state.NewJournalEntryAccountCreation(address, ts)

	acc, err := entry.Revert()
	assert.Equal(t, updateErr, err)
	assert.Nil(t, acc)
}

func TestJournalEntryAccountCreation_RevertUpdatesTheTrie(t *testing.T) {
	t.Parallel()

	updateCalled := false
	address := []byte("address")
	ts := &testscommon.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			assert.Equal(t, address, key)
			assert.Nil(t, value)
			updateCalled = true
			return nil
		},
	}
	entry, _ := state.NewJournalEntryAccountCreation(address, ts)

	acc, err := entry.Revert()
	assert.Nil(t, err)
	assert.Nil(t, acc)
	assert.True(t, updateCalled)
}

func TestNewJournalEntryDataTrieUpdates_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	trieUpdates := make(map[string][]byte)
	trieUpdates["a"] = []byte("b")
	entry, err := state.NewJournalEntryDataTrieUpdates(trieUpdates, nil)

	assert.True(t, check.IfNil(entry))
	assert.True(t, errors.Is(err, state.ErrNilAccountHandler))
}

func TestNewJournalEntryDataTrieUpdates_EmptyTrieUpdatesShouldErr(t *testing.T) {
	t.Parallel()

	trieUpdates := make(map[string][]byte)
	accnt, _ := state.NewUserAccount(make([]byte, 32))
	entry, err := state.NewJournalEntryDataTrieUpdates(trieUpdates, accnt)

	assert.True(t, check.IfNil(entry))
	assert.Equal(t, state.ErrNilOrEmptyDataTrieUpdates, err)
}

func TestNewJournalEntryDataTrieUpdates_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	trieUpdates := make(map[string][]byte)
	trieUpdates["a"] = []byte("b")
	accnt, _ := state.NewUserAccount(make([]byte, 32))
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

	trie := &testscommon.TrieStub{
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

	trie := &testscommon.TrieStub{
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

	trie := &testscommon.TrieStub{
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
