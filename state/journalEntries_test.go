package state_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewJournalEntryCode_NilUpdaterShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryCode(&state.CodeEntry{}, []byte("code hash"), []byte("code hash"), nil, &testscommon.MarshalizerMock{})
	assert.True(t, check.IfNil(entry))
	assert.Equal(t, state.ErrNilUpdater, err)
}

func TestNewJournalEntryCode_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryCode(&state.CodeEntry{}, []byte("code hash"), []byte("code hash"), &trieMock.TrieStub{}, nil)
	assert.True(t, check.IfNil(entry))
	assert.Equal(t, state.ErrNilMarshalizer, err)
}

func TestNewJournalEntryCode_OkParams(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryCode(&state.CodeEntry{}, []byte("code hash"), []byte("code hash"), &trieMock.TrieStub{}, &testscommon.MarshalizerMock{})
	assert.Nil(t, err)
	assert.False(t, check.IfNil(entry))
}

func TestJournalEntryCode_OldHashAndNewHashAreNil(t *testing.T) {
	t.Parallel()

	trieStub := &trieMock.TrieStub{}
	entry, _ := state.NewJournalEntryCode(&state.CodeEntry{}, nil, nil, trieStub, &testscommon.MarshalizerMock{})

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
	marshalizer := &testscommon.MarshalizerMock{}

	updateCalled := false
	trieStub := &trieMock.TrieStub{
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			serializedCodeEntry, err := marshalizer.Marshal(codeEntry)
			return serializedCodeEntry, 0, err
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

	entry, err := state.NewJournalEntryAccount(&stateMock.AccountWrapMock{})
	assert.Nil(t, err)
	assert.False(t, check.IfNil(entry))
}

func TestJournalEntryAccount_Revert(t *testing.T) {
	t.Parallel()

	expectedAcc := &stateMock.AccountWrapMock{}
	entry, _ := state.NewJournalEntryAccount(expectedAcc)

	acc, err := entry.Revert()
	assert.Nil(t, err)
	assert.Equal(t, expectedAcc, acc)
}

func TestNewJournalEntryAccountCreation_InvalidAddressShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewJournalEntryAccountCreation([]byte{}, &trieMock.TrieStub{})
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

	entry, err := state.NewJournalEntryAccountCreation([]byte("address"), &trieMock.TrieStub{})
	assert.Nil(t, err)
	assert.False(t, check.IfNil(entry))
}

func TestJournalEntryAccountCreation_RevertErr(t *testing.T) {
	t.Parallel()

	updateErr := errors.New("update error")
	address := []byte("address")
	ts := &trieMock.TrieStub{
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
	ts := &trieMock.TrieStub{
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
	accnt := stateMock.NewAccountWrapMock(nil)

	tr := &trieMock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return expectedErr
		},
	}

	accnt.SetDataTrie(tr)
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
	accnt := stateMock.NewAccountWrapMock(nil)

	tr := &trieMock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		RootCalled: func() ([]byte, error) {
			return nil, expectedErr
		},
	}

	accnt.SetDataTrie(tr)
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
	accnt := stateMock.NewAccountWrapMock(nil)

	tr := &trieMock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			updateWasCalled = true
			return nil
		},
		RootCalled: func() ([]byte, error) {
			rootWasCalled = true
			return []byte{}, nil
		},
	}

	accnt.SetDataTrie(tr)
	entry, _ := state.NewJournalEntryDataTrieUpdates(trieUpdates, accnt)

	acc, err := entry.Revert()
	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.True(t, updateWasCalled)
	assert.True(t, rootWasCalled)
}
