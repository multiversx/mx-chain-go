package state_test

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/iteratorChannelsProvider"
	"github.com/multiversx/mx-chain-go/state/lastSnapshotMarker"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	testState "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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
			return &storageManager.StorageManagerStub{
				TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, leavesChan *common.TrieIteratorChannels, _ chan []byte, stats common.SnapshotStatisticsHandler, _ uint32) {
					snapshotCalled = true
					stats.SnapshotFinished()
				},
			}
		},
	}

	snapshotsManager, _ := state.NewSnapshotsManager(state.ArgsNewSnapshotsManager{
		ProcessingMode:       common.Normal,
		Marshaller:           &marshallerMock.MarshalizerMock{},
		AddressConverter:     &testscommon.PubkeyConverterMock{},
		ProcessStatusHandler: &testscommon.ProcessStatusHandlerStub{},
		StateMetrics:         &testState.StateMetricsStub{},
		AccountFactory:       args.AccountFactory,
		ChannelsProvider:     iteratorChannelsProvider.NewPeerStateIteratorChannelsProvider(),
		LastSnapshotMarker:   lastSnapshotMarker.NewLastSnapshotMarker(),
		StateStatsHandler:    statistics.NewStateStatistics(),
	})
	args.SnapshotsManager = snapshotsManager

	adb, err := state.NewPeerAccountsDB(args)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))

	adb.SnapshotState([]byte("rootHash"), 0)
	for adb.IsSnapshotInProgress() {
		time.Sleep(10 * time.Millisecond)
	}
	assert.True(t, snapshotCalled)
}

func TestNewPeerAccountsDB_SnapshotStateGetLatestStorageEpochErrDoesNotSnapshot(t *testing.T) {
	t.Parallel()

	snapshotCalled := false
	args := createMockAccountsDBArgs()
	args.Trie = &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{
				GetLatestStorageEpochCalled: func() (uint32, error) {
					return 0, fmt.Errorf("new error")
				},
				TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, _ *common.TrieIteratorChannels, _ chan []byte, _ common.SnapshotStatisticsHandler, _ uint32) {
					snapshotCalled = true
				},
			}
		},
	}
	adb, err := state.NewPeerAccountsDB(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))

	adb.SnapshotState([]byte("rootHash"), 0)
	assert.False(t, snapshotCalled)
}

func TestNewPeerAccountsDB_RecreateAllTries(t *testing.T) {
	t.Parallel()

	recreateCalled := false
	args := createMockAccountsDBArgs()
	args.Trie = &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
		},
		RecreateCalled: func(_ common.RootHashHolder) (common.Trie, error) {
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
			return &storageManager.StorageManagerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, []byte(common.ActiveDBKey)) {
						return nil, fmt.Errorf("key not found")
					}
					return []byte("rootHash"), nil
				},
				ShouldTakeSnapshotCalled: func() bool {
					return true
				},
				TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, _ *common.TrieIteratorChannels, _ chan []byte, _ common.SnapshotStatisticsHandler, _ uint32) {
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
	err = adb.SetSyncer(&mock.AccountsDBSyncerStub{})
	assert.Nil(t, err)
	err = adb.StartSnapshotIfNeeded()
	assert.Nil(t, err)

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
				return &storageManager.StorageManagerStub{
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
				return &storageManager.StorageManagerStub{
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
				return &storageManager.StorageManagerStub{
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
	t.Run("epoch 0, GetLatestStorageEpoch errors should not put", func(t *testing.T) {
		trieStub := &trieMock.TrieStub{
			RootCalled: func() ([]byte, error) {
				return rootHash, nil
			},
			GetStorageManagerCalled: func() common.StorageManager {
				return &storageManager.StorageManagerStub{
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
		err := adb.SetSyncer(&mock.AccountsDBSyncerStub{})
		assert.Nil(t, err)
		err = adb.StartSnapshotIfNeeded()
		assert.Nil(t, err)
	})
	t.Run("in import DB mode", func(t *testing.T) {
		trieStub := &trieMock.TrieStub{
			RootCalled: func() ([]byte, error) {
				return rootHash, nil
			},
			GetStorageManagerCalled: func() common.StorageManager {
				return &storageManager.StorageManagerStub{
					ShouldTakeSnapshotCalled: func() bool {
						return true
					},
					GetLatestStorageEpochCalled: func() (uint32, error) {
						return 1, nil
					},
					GetFromCurrentEpochCalled: func(i []byte) ([]byte, error) {
						return nil, expectedErr
					},
				}
			},
		}

		args := createMockAccountsDBArgs()
		args.SnapshotsManager, _ = state.NewSnapshotsManager(state.ArgsNewSnapshotsManager{
			ProcessingMode:       common.ImportDb,
			Marshaller:           &marshallerMock.MarshalizerMock{},
			AddressConverter:     &testscommon.PubkeyConverterMock{},
			ProcessStatusHandler: &testscommon.ProcessStatusHandlerStub{},
			StateMetrics:         &testState.StateMetricsStub{},
			AccountFactory:       args.AccountFactory,
			ChannelsProvider:     iteratorChannelsProvider.NewUserStateIteratorChannelsProvider(),
			LastSnapshotMarker:   lastSnapshotMarker.NewLastSnapshotMarker(),
			StateStatsHandler:    statistics.NewStateStatistics(),
		})
		args.Trie = trieStub
		adb, _ := state.NewPeerAccountsDB(args)
		err := adb.SetSyncer(&mock.AccountsDBSyncerStub{})
		assert.Nil(t, err)
		err = adb.StartSnapshotIfNeeded()
		assert.Nil(t, err)
	})
}

func TestPeerAccountsDB_SnapshotStateOnAClosedStorageManagerShouldNotMarkActiveDB(t *testing.T) {
	t.Parallel()

	mut := sync.RWMutex{}
	lastSnapshotStartedWasPut := false
	activeDBWasPut := false
	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{
				ShouldTakeSnapshotCalled: func() bool {
					return true
				},
				TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, _ *common.TrieIteratorChannels, _ chan []byte, stats common.SnapshotStatisticsHandler, _ uint32) {
					stats.SnapshotFinished()
				},
				IsClosedCalled: func() bool {
					return true
				},
				PutInEpochCalled: func(key []byte, val []byte, epoch uint32) error {
					mut.Lock()
					defer mut.Unlock()

					if string(key) == common.ActiveDBKey {
						activeDBWasPut = true
					}

					if string(key) == lastSnapshotMarker.LastSnapshot {
						lastSnapshotStartedWasPut = true
					}

					return nil
				},
			}
		},
	}
	args := createMockAccountsDBArgs()
	args.Trie = trieStub

	adb, _ := state.NewPeerAccountsDB(args)
	adb.SnapshotState([]byte("roothash"), 0)
	time.Sleep(time.Second)

	mut.RLock()
	defer mut.RUnlock()
	assert.False(t, lastSnapshotStartedWasPut)
	assert.False(t, activeDBWasPut)
}

func TestGetPeerAccountAndReturnIfNew(t *testing.T) {
	t.Parallel()

	t.Run("should return error if failed account creation", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		adb := &testState.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return nil, errors.New("account does not exist")
			},
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				return nil, expectedErr
			},
		}

		acc, isNew, err := state.GetPeerAccountAndReturnIfNew(adb, []byte("address"))
		assert.Nil(t, acc)
		assert.False(t, isNew)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("should return error if wrong type of existent account", func(t *testing.T) {
		t.Parallel()

		adb := &testState.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return &testState.UserAccountStub{}, nil
			},
		}

		acc, isNew, err := state.GetPeerAccountAndReturnIfNew(adb, []byte("address"))
		assert.Nil(t, acc)
		assert.False(t, isNew)
		assert.Equal(t, state.ErrWrongTypeAssertion, err)
	})

	t.Run("should return error if wrong type of created accounts", func(t *testing.T) {
		t.Parallel()

		adb := &testState.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return nil, errors.New("account does not exist")
			},
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				return &testState.UserAccountStub{}, nil
			},
		}

		acc, isNew, err := state.GetPeerAccountAndReturnIfNew(adb, []byte("address"))
		assert.Nil(t, acc)
		assert.False(t, isNew)
		assert.Equal(t, state.ErrWrongTypeAssertion, err)
	})

	t.Run("should work if account exists", func(t *testing.T) {
		t.Parallel()

		adb := &testState.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return &testState.PeerAccountHandlerMock{}, nil
			},
		}

		acc, isNew, err := state.GetPeerAccountAndReturnIfNew(adb, []byte("address"))
		assert.NotNil(t, acc)
		assert.False(t, isNew)
		assert.Nil(t, err)
	})

	t.Run("should create account if missing", func(t *testing.T) {
		t.Parallel()

		adb := &testState.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return nil, errors.New("account does not exist")
			},
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				return &testState.PeerAccountHandlerMock{}, nil
			},
		}

		acc, isNew, err := state.GetPeerAccountAndReturnIfNew(adb, []byte("address"))
		assert.NotNil(t, acc)
		assert.True(t, isNew)
		assert.Nil(t, err)
	})
}
