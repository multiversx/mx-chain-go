package state_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/disabled"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerAccountsDB_WithNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		nil,
		&hashingMocks.HasherMock{},
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
		&hashingMocks.HasherMock{},
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
		&hashingMocks.HasherMock{},
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
		&hashingMocks.HasherMock{},
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
		&hashingMocks.HasherMock{},
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
					TakeSnapshotCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ common.SnapshotStatisticsHandler, _ uint32) {
						snapshotCalled = true
					},
				}
			},
		},
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		&stateMock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))

	adb.SnapshotState([]byte("rootHash"))
	assert.True(t, snapshotCalled)
}

func TestNewPeerAccountsDB_SnapshotStateGetLatestStorageEpochErrDoesNotSnapshot(t *testing.T) {
	t.Parallel()

	snapshotCalled := false
	adb, err := state.NewPeerAccountsDB(
		&trieMock.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{
					GetLatestStorageEpochCalled: func() (uint32, error) {
						return 0, fmt.Errorf("new error")
					},
					TakeSnapshotCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ common.SnapshotStatisticsHandler, _ uint32) {
						snapshotCalled = true
					},
				}
			},
		},
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		&stateMock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))

	adb.SnapshotState([]byte("rootHash"))
	assert.False(t, snapshotCalled)
}

func TestNewPeerAccountsDB_SetStateCheckpoint(t *testing.T) {
	t.Parallel()

	checkpointCalled := false
	adb, err := state.NewPeerAccountsDB(
		&trieMock.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{
					SetCheckpointCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ common.SnapshotStatisticsHandler) {
						checkpointCalled = true
					},
				}
			},
		},
		&hashingMocks.HasherMock{},
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
		&hashingMocks.HasherMock{},
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

func TestPeerAccountsDB_NewAccountsDbStartsSnapshotAfterRestart(t *testing.T) {
	t.Parallel()

	rootHash := []byte("rootHash")
	mutex := sync.RWMutex{}
	takeSnapshotCalled := false
	trieStub := &trieMock.TrieStub{
		RootCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				GetCalled: func(_ []byte) ([]byte, error) {
					return nil, fmt.Errorf("key not found")
				},
				ShouldTakeSnapshotCalled: func() bool {
					return true
				},
				TakeSnapshotCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ common.SnapshotStatisticsHandler, _ uint32) {
					mutex.Lock()
					takeSnapshotCalled = true
					mutex.Unlock()
				},
			}
		},
	}

	adb, err := state.NewPeerAccountsDB(
		trieStub,
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		&stateMock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)
	assert.Nil(t, err)
	assert.NotNil(t, adb)

	time.Sleep(time.Second)
	mutex.RLock()
	assert.True(t, takeSnapshotCalled)
	mutex.RUnlock()
}

func TestPeerAccountsDB_MarkSnapshotDone(t *testing.T) {
	t.Parallel()

	t.Run("get latest epoch fails", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not failed %v", r))
			}
		}()

		expectedErr := errors.New("expected error")
		adb, _ := state.NewPeerAccountsDB(
			&trieMock.TrieStub{
				GetStorageManagerCalled: func() common.StorageManager {
					return &testscommon.StorageManagerStub{
						PutInEpochCalled: func(bytes []byte, bytes2 []byte, u uint32) error {
							assert.Fail(t, "should have not called put in epoch")
							return nil
						},
						GetLatestStorageEpochCalled: func() (uint32, error) {
							return 0, expectedErr
						},
					}
				},
			},
			&hashingMocks.HasherMock{},
			&testscommon.MarshalizerMock{},
			&stateMock.AccountsFactoryStub{},
			disabled.NewDisabledStoragePruningManager(),
		)

		adb.MarkSnapshotDone()
	})
	t.Run("put fails should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not failed %v", r))
			}
		}()

		expectedErr := errors.New("expected error")
		putWasCalled := false
		adb, _ := state.NewPeerAccountsDB(
			&trieMock.TrieStub{
				GetStorageManagerCalled: func() common.StorageManager {
					return &testscommon.StorageManagerStub{
						PutInEpochCalled: func(key []byte, value []byte, epoch uint32) error {
							assert.Equal(t, common.ActiveDBKey, string(key))
							assert.Equal(t, common.ActiveDBVal, string(value))
							putWasCalled = true

							return expectedErr
						},
					}
				},
			},
			&hashingMocks.HasherMock{},
			&testscommon.MarshalizerMock{},
			&stateMock.AccountsFactoryStub{},
			disabled.NewDisabledStoragePruningManager(),
		)

		adb.MarkSnapshotDone()
		assert.True(t, putWasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		putWasCalled := false
		adb, _ := state.NewPeerAccountsDB(
			&trieMock.TrieStub{
				GetStorageManagerCalled: func() common.StorageManager {
					return &testscommon.StorageManagerStub{
						PutInEpochCalled: func(key []byte, value []byte, epoch uint32) error {
							assert.Equal(t, common.ActiveDBKey, string(key))
							assert.Equal(t, common.ActiveDBVal, string(value))
							putWasCalled = true

							return nil
						},
					}
				},
			},
			&hashingMocks.HasherMock{},
			&testscommon.MarshalizerMock{},
			&stateMock.AccountsFactoryStub{},
			disabled.NewDisabledStoragePruningManager(),
		)

		adb.MarkSnapshotDone()
		assert.True(t, putWasCalled)
	})

}
