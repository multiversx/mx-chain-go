package state_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

//------- BaseJournalEntryCreation

func TestNewBaseJournalEntryCreation_NilUpdaterShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewBaseJournalEntryCreation([]byte("key"), nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilUpdater, err)
}

func TestNewBaseJournalEntryCreation_NilKeyShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewBaseJournalEntryCreation(nil, &mock.UpdaterStub{})

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilOrEmptyKey, err)
}

func TestNewBaseJournalEntryCreation_ShouldWork(t *testing.T) {
	t.Parallel()

	entry, err := state.NewBaseJournalEntryCreation([]byte("key"), &mock.UpdaterStub{})

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestBaseJournalEntryCreation_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	updater := &mock.UpdaterStub{
		UpdateCalled: func(key, value []byte) error {
			wasCalled = true
			return nil
		},
	}

	entry, _ := state.NewBaseJournalEntryCreation([]byte("key"), updater)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

//------- BaseJournalEntryCodeHash

func TestNewBaseJournalEntryCodeHash_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewBaseJournalEntryCodeHash(nil, nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestNewBaseJournalEntryCodeHash_ShouldWork(t *testing.T) {
	t.Parallel()

	entry, err := state.NewBaseJournalEntryCodeHash(mock.NewAccountWrapMock(nil, nil), nil)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestBaseJournalEntryCodeHash_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	testCodeHash := []byte("code hash to revert")
	accnt := mock.NewAccountWrapMock(nil, nil)

	entry, _ := state.NewBaseJournalEntryCodeHash(accnt, testCodeHash)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, testCodeHash, accnt.GetCodeHash())
}

//------- BaseJournalEntryRootHash

func TestNewBaseJournalEntryRootHash_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewBaseJournalEntryRootHash(nil, nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestNewBaseJournalEntryRootHash_ShouldWork(t *testing.T) {
	t.Parallel()

	entry, err := state.NewBaseJournalEntryRootHash(mock.NewAccountWrapMock(nil, nil), nil)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestBaseJournalEntryRootHash_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	testRootHash := []byte("root hash to revert")
	accnt := mock.NewAccountWrapMock(nil, nil)

	entry, _ := state.NewBaseJournalEntryRootHash(accnt, testRootHash)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, testRootHash, accnt.GetRootHash())
}

//------- BaseJournalEntryData

func TestNewBaseJournalEntryData_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewBaseJournalEntryData(nil, nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestNewBaseJournalEntryData_ShouldWork(t *testing.T) {
	t.Parallel()

	entry, err := state.NewBaseJournalEntryData(mock.NewAccountWrapMock(nil, nil), nil)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestBaseJournalEntryData_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false
	tracker := &mock.DataTrieTrackerStub{
		ClearDataCachesCalled: func() {
			wasCalled = true
		},
	}

	accnt := mock.NewAccountWrapMock(nil, nil)
	accnt.SetDataTrieTracker(tracker)
	tr := mock.NewMockTrie()

	entry, _ := state.NewBaseJournalEntryData(accnt, tr)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.True(t, tr == entry.Trie())
	assert.True(t, wasCalled)
}
