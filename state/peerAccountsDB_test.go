package state_test

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerAccountsDB(t *testing.T) {
	t.Parallel()

	t.Run("nil trie should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.Trie = nil

		adb, err := state.NewPeerAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilTrie, err)
	})
	t.Run("nil hasher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.Hasher = nil

		adb, err := state.NewPeerAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilHasher, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.Marshaller = nil

		adb, err := state.NewPeerAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilMarshalizer, err)
	})
	t.Run("nil account factory should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.AccountFactory = nil

		adb, err := state.NewPeerAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilAccountFactory, err)
	})
	t.Run("nil storage pruning manager should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.StoragePruningManager = nil

		adb, err := state.NewPeerAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilStoragePruningManager, err)
	})
	t.Run("nil process status handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.ProcessStatusHandler = nil

		adb, err := state.NewPeerAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilProcessStatusHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()

		adb, err := state.NewPeerAccountsDB(args)
		assert.False(t, check.IfNil(adb))
		assert.Nil(t, err)
	})
}

func TestNewPeerAccountsDB_SnapshotState(t *testing.T) {
	t.Parallel()

	snapshotCalled := false
	args := createMockAccountsDBArgs()
	args.Trie = &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				TakeSnapshotCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ chan []byte, _ chan error, _ common.SnapshotStatisticsHandler, _ uint32) {
					snapshotCalled = true
				},
			}
		},
	}
	adb, err := state.NewPeerAccountsDB(args)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))

	adb.SnapshotState([]byte("rootHash"))
	assert.True(t, snapshotCalled)
}

func TestNewPeerAccountsDB_SnapshotStateGetLatestStorageEpochErrDoesNotSnapshot(t *testing.T) {
	t.Parallel()

	snapshotCalled := false
	args := createMockAccountsDBArgs()
	args.Trie = &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				GetLatestStorageEpochCalled: func() (uint32, error) {
					return 0, fmt.Errorf("new error")
				},
				TakeSnapshotCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ chan []byte, _ chan error, _ common.SnapshotStatisticsHandler, _ uint32) {
					snapshotCalled = true
				},
			}
		},
	}
	adb, err := state.NewPeerAccountsDB(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))

	adb.SnapshotState([]byte("rootHash"))
	assert.False(t, snapshotCalled)
}

func TestNewPeerAccountsDB_SetStateCheckpoint(t *testing.T) {
	t.Parallel()

	checkpointCalled := false
	args := createMockAccountsDBArgs()
	args.Trie = &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				SetCheckpointCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ chan []byte, _ chan error, _ common.SnapshotStatisticsHandler) {
					checkpointCalled = true
				},
			}
		},
	}
	adb, err := state.NewPeerAccountsDB(args)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))

	adb.SetStateCheckpoint([]byte("rootHash"))
	assert.True(t, checkpointCalled)
}

func TestNewPeerAccountsDB_RecreateAllTries(t *testing.T) {
	t.Parallel()

	recreateCalled := false
	args := createMockAccountsDBArgs()
	args.Trie = &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
		RecreateCalled: func(_ []byte) (common.Trie, error) {
			recreateCalled = true
			return nil, nil
		},
	}
	adb, err := state.NewPeerAccountsDB(args)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))

	tries, err := adb.RecreateAllTries([]byte("rootHash"))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tries))
	assert.True(t, recreateCalled)
}

func TestPeerAccountsDB_SetSyncerAndStartSnapshotIfNeeded(t *testing.T) {
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
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, []byte(common.ActiveDBKey)) {
						return nil, fmt.Errorf("key not found")
					}
					return []byte("rootHash"), nil
				},
				ShouldTakeSnapshotCalled: func() bool {
					return true
				},
				TakeSnapshotCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ chan []byte, _ chan error, _ common.SnapshotStatisticsHandler, _ uint32) {
					mutex.Lock()
					takeSnapshotCalled = true
					mutex.Unlock()
				},
				GetLatestStorageEpochCalled: func() (uint32, error) {
					return 1, nil
				},
			}
		},
	}

	args := createMockAccountsDBArgs()
	args.Trie = trieStub
	adb, err := state.NewPeerAccountsDB(args)
	assert.Nil(t, err)
	assert.NotNil(t, adb)
	adb.SetSyncerAndStartSnapshotIfNeeded(&mock.AccountsDBSyncerStub{})

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
		args := createMockAccountsDBArgs()
		args.Trie = &trieMock.TrieStub{
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
		}
		adb, _ := state.NewPeerAccountsDB(args)

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
		args := createMockAccountsDBArgs()
		args.Trie = &trieMock.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{
					PutInEpochWithoutCacheCalled: func(key []byte, value []byte, epoch uint32) error {
						assert.Equal(t, common.ActiveDBKey, string(key))
						assert.Equal(t, common.ActiveDBVal, string(value))
						putWasCalled = true

						return expectedErr
					},
				}
			},
		}
		adb, _ := state.NewPeerAccountsDB(args)

		adb.MarkSnapshotDone()
		assert.True(t, putWasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		putWasCalled := false
		args := createMockAccountsDBArgs()
		args.Trie = &trieMock.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{
					PutInEpochWithoutCacheCalled: func(key []byte, value []byte, epoch uint32) error {
						assert.Equal(t, common.ActiveDBKey, string(key))
						assert.Equal(t, common.ActiveDBVal, string(value))
						putWasCalled = true

						return nil
					},
				}
			},
		}
		adb, _ := state.NewPeerAccountsDB(args)

		adb.MarkSnapshotDone()
		assert.True(t, putWasCalled)
	})

}

func TestPeerAccountsDB_SetSyncerAndStartSnapshotIfNeededMarksActiveDB(t *testing.T) {
	t.Parallel()

	rootHash := []byte("rootHash")
	expectedErr := errors.New("expected error")
	t.Run("epoch 0", func(t *testing.T) {
		putCalled := false
		trieStub := &trieMock.TrieStub{
			RootCalled: func() ([]byte, error) {
				return rootHash, nil
			},
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{
					ShouldTakeSnapshotCalled: func() bool {
						return true
					},
					GetLatestStorageEpochCalled: func() (uint32, error) {
						return 0, nil
					},
					PutCalled: func(key []byte, val []byte) error {
						assert.Equal(t, []byte(common.ActiveDBKey), key)
						assert.Equal(t, []byte(common.ActiveDBVal), val)

						putCalled = true

						return nil
					},
				}
			},
		}

		args := createMockAccountsDBArgs()
		args.Trie = trieStub
		adb, _ := state.NewPeerAccountsDB(args)
		adb.SetSyncerAndStartSnapshotIfNeeded(&mock.AccountsDBSyncerStub{})

		assert.True(t, putCalled)
	})
	t.Run("epoch 0, GetLatestStorageEpoch errors should not put", func(t *testing.T) {
		trieStub := &trieMock.TrieStub{
			RootCalled: func() ([]byte, error) {
				return rootHash, nil
			},
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{
					ShouldTakeSnapshotCalled: func() bool {
						return true
					},
					GetLatestStorageEpochCalled: func() (uint32, error) {
						return 0, expectedErr
					},
					PutCalled: func(key []byte, val []byte) error {
						assert.Fail(t, "should have not called put")

						return nil
					},
				}
			},
		}

		args := createMockAccountsDBArgs()
		args.Trie = trieStub
		adb, _ := state.NewPeerAccountsDB(args)
		adb.SetSyncerAndStartSnapshotIfNeeded(&mock.AccountsDBSyncerStub{})
	})
	t.Run("in import DB mode", func(t *testing.T) {
		putCalled := false
		trieStub := &trieMock.TrieStub{
			RootCalled: func() ([]byte, error) {
				return rootHash, nil
			},
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{
					ShouldTakeSnapshotCalled: func() bool {
						return true
					},
					GetLatestStorageEpochCalled: func() (uint32, error) {
						return 1, nil
					},
					PutCalled: func(key []byte, val []byte) error {
						assert.Equal(t, []byte(common.ActiveDBKey), key)
						assert.Equal(t, []byte(common.ActiveDBVal), val)

						putCalled = true

						return nil
					},
				}
			},
		}

		args := createMockAccountsDBArgs()
		args.ProcessingMode = common.ImportDb
		args.Trie = trieStub
		adb, _ := state.NewPeerAccountsDB(args)
		adb.SetSyncerAndStartSnapshotIfNeeded(&mock.AccountsDBSyncerStub{})

		assert.True(t, putCalled)
	})
}

func TestPeerAccountsDB_SnapshotStateOnAClosedStorageManagerShouldNotMarkActiveDB(t *testing.T) {
	t.Parallel()

	mut := sync.RWMutex{}
	lastSnapshotStartedWasPut := false
	activeDBWasPut := false
	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				ShouldTakeSnapshotCalled: func() bool {
					return true
				},
				TakeSnapshotCalled: func(_ []byte, _ []byte, ch chan core.KeyValueHolder, _ chan []byte, _ chan error, stats common.SnapshotStatisticsHandler, _ uint32) {
					stats.SnapshotFinished()
				},
				IsClosedCalled: func() bool {
					return true
				},
				PutCalled: func(key []byte, val []byte) error {
					mut.Lock()
					defer mut.Unlock()

					if string(key) == state.LastSnapshotStarted {
						lastSnapshotStartedWasPut = true
					}

					return nil
				},
				PutInEpochCalled: func(key []byte, val []byte, epoch uint32) error {
					mut.Lock()
					defer mut.Unlock()

					if string(key) == common.ActiveDBKey {
						activeDBWasPut = true
					}

					return nil
				},
			}
		},
	}
	args := createMockAccountsDBArgs()
	args.Trie = trieStub

	adb, _ := state.NewPeerAccountsDB(args)
	adb.SnapshotState([]byte("roothash"))
	time.Sleep(time.Second)

	mut.RLock()
	defer mut.RUnlock()
	assert.True(t, lastSnapshotStartedWasPut)
	assert.False(t, activeDBWasPut)
}
