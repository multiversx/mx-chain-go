package state_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/mock"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/disabled"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerAccountsDB_WithNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		nil,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilTrie, err)
}

func TestNewPeerAccountsDB_WithNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&testscommon.TrieStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilHasher, err)
}

func TestNewPeerAccountsDB_WithNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&testscommon.TrieStub{},
		&mock.HasherMock{},
		nil,
		&mock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilMarshalizer, err)
}

func TestNewPeerAccountsDB_WithNilAddressFactoryShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&testscommon.TrieStub{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilAccountFactory, err)
}

func TestNewPeerAccountsDB_WithNilStoragePruningManagerShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&testscommon.TrieStub{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
		nil,
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilStoragePruningManager, err)
}

func TestNewPeerAccountsDB_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&testscommon.TrieStub{
			GetStorageManagerCalled: func() temporary.StorageManager {
				return &testscommon.StorageManagerStub{
					DatabaseCalled: func() temporary.DBWriteCacher {
						return mock.NewMemDbMock()
					},
				}
			},
		},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))
}

func TestNewPeerAccountsDB_SnapshotState(t *testing.T) {
	t.Parallel()

	snapshotCalled := false
	adb, err := state.NewPeerAccountsDB(
		&testscommon.TrieStub{
			GetStorageManagerCalled: func() temporary.StorageManager {
				return &testscommon.StorageManagerStub{
					TakeSnapshotCalled: func(_ []byte, _ bool, _ chan core.KeyValueHolder) {
						snapshotCalled = true
					},
					DatabaseCalled: func() temporary.DBWriteCacher {
						return mock.NewMemDbMock()
					},
				}
			},
		},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))

	adb.SnapshotState([]byte("rootHash"))
	assert.True(t, snapshotCalled)
}

func TestNewPeerAccountsDB_SetStateCheckpoint(t *testing.T) {
	t.Parallel()

	checkpointCalled := false
	adb, err := state.NewPeerAccountsDB(
		&testscommon.TrieStub{
			GetStorageManagerCalled: func() temporary.StorageManager {
				return &testscommon.StorageManagerStub{
					SetCheckpointCalled: func(_ []byte, _ chan core.KeyValueHolder) {
						checkpointCalled = true
					},
					DatabaseCalled: func() temporary.DBWriteCacher {
						return mock.NewMemDbMock()
					},
				}
			},
		},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))

	adb.SetStateCheckpoint([]byte("rootHash"))
	assert.True(t, checkpointCalled)
}

func TestNewPeerAccountsDB_RecreateAllTries(t *testing.T) {
	t.Parallel()

	recreateCalled := false
	adb, err := state.NewPeerAccountsDB(
		&testscommon.TrieStub{
			GetStorageManagerCalled: func() temporary.StorageManager {
				return &testscommon.StorageManagerStub{
					DatabaseCalled: func() temporary.DBWriteCacher {
						return mock.NewMemDbMock()
					},
				}
			},
			RecreateCalled: func(_ []byte) (temporary.Trie, error) {
				recreateCalled = true
				return nil, nil
			},
		},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))

	tries, err := adb.RecreateAllTries([]byte("rootHash"))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tries))
	assert.True(t, recreateCalled)
}
