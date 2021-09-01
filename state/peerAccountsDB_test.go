package state_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/disabled"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerAccountsDB_WithNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		nil,
		&testscommon.HasherMock{},
		&testscommon.MarshalizerMock{},
		&stateMock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilTrie, err)
}

func TestNewPeerAccountsDB_WithNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&trieMock.TrieStub{},
		nil,
		&testscommon.MarshalizerMock{},
		&stateMock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilHasher, err)
}

func TestNewPeerAccountsDB_WithNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&trieMock.TrieStub{},
		&testscommon.HasherMock{},
		nil,
		&stateMock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilMarshalizer, err)
}

func TestNewPeerAccountsDB_WithNilAddressFactoryShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&trieMock.TrieStub{},
		&testscommon.HasherMock{},
		&testscommon.MarshalizerMock{},
		nil,
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilAccountFactory, err)
}

func TestNewPeerAccountsDB_WithNilStoragePruningManagerShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&trieMock.TrieStub{},
		&testscommon.HasherMock{},
		&testscommon.MarshalizerMock{},
		&stateMock.AccountsFactoryStub{},
		nil,
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilStoragePruningManager, err)
}

func TestNewPeerAccountsDB_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&trieMock.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{}
			},
		},
		&testscommon.HasherMock{},
		&testscommon.MarshalizerMock{},
		&stateMock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))
}

func TestNewPeerAccountsDB_SnapshotState(t *testing.T) {
	t.Parallel()

	snapshotCalled := false
	adb, err := state.NewPeerAccountsDB(
		&trieMock.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{
					TakeSnapshotCalled: func(_ []byte, _ bool, _ chan core.KeyValueHolder) {
						snapshotCalled = true
					},
				}
			},
		},
		&testscommon.HasherMock{},
		&testscommon.MarshalizerMock{},
		&stateMock.AccountsFactoryStub{},
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
		&trieMock.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{
					SetCheckpointCalled: func(_ []byte, _ chan core.KeyValueHolder) {
						checkpointCalled = true
					},
				}
			},
		},
		&testscommon.HasherMock{},
		&testscommon.MarshalizerMock{},
		&stateMock.AccountsFactoryStub{},
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
		&trieMock.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{}
			},
			RecreateCalled: func(_ []byte) (common.Trie, error) {
				recreateCalled = true
				return nil, nil
			},
		},
		&testscommon.HasherMock{},
		&testscommon.MarshalizerMock{},
		&stateMock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))

	tries, err := adb.RecreateAllTries([]byte("rootHash"))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tries))
	assert.True(t, recreateCalled)
}
