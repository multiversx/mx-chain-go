package state_test

import (
	"context"
	"encoding/binary"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/testscommon"
	mockState "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	testTrie "github.com/multiversx/mx-chain-go/testscommon/trie"

	"github.com/multiversx/mx-chain-go/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dummyRootHash = []byte("new root hash")

func createBlockInfoProviderStub(rootHash []byte) *testscommon.BlockInfoProviderStub {
	return &testscommon.BlockInfoProviderStub{
		GetBlockInfoCalled: func() common.BlockInfo {
			return holders.NewBlockInfo(nil, 0, rootHash)
		},
	}
}

func TestNewAccountsDBApi(t *testing.T) {
	t.Parallel()

	t.Run("nil accounts adapter should error", func(t *testing.T) {
		t.Parallel()

		accountsApi, err := state.NewAccountsDBApi(nil, &testscommon.BlockInfoProviderStub{})

		assert.True(t, check.IfNil(accountsApi))
		assert.Equal(t, state.ErrNilAccountsAdapter, err)
	})
	t.Run("nil block info provider should error", func(t *testing.T) {
		t.Parallel()

		accountsApi, err := state.NewAccountsDBApi(&mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
		}, nil)

		assert.True(t, check.IfNil(accountsApi))
		assert.Equal(t, state.ErrNilBlockInfoProvider, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		accountsApi, err := state.NewAccountsDBApi(&mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
		}, &testscommon.BlockInfoProviderStub{})

		assert.False(t, check.IfNil(accountsApi))
		assert.Nil(t, err)
	})
}

func TestAccountsDBAPi_recreateTrieIfNecessary(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")

	t.Run("block info provider returns nil or empty root hash should error", func(t *testing.T) {
		t.Parallel()

		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				require.Fail(t, "should have not called RecreateAllTriesCalled")

				return nil
			},
		}

		blockInfoProvider := &testscommon.BlockInfoProviderStub{
			GetBlockInfoCalled: func() common.BlockInfo {
				return nil
			},
		}
		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockInfoProvider)

		t.Run("nil block info should error", func(t *testing.T) {
			accountsApi.SetCurrentBlockInfo(holders.NewBlockInfo(nil, 0, dummyRootHash))
			assert.True(t, errors.Is(accountsApi.RecreateTrieIfNecessary(), state.ErrNilBlockInfo))

		})
		t.Run("block info contains a nil root hash should error", func(t *testing.T) {
			accountsApi.SetCurrentBlockInfo(holders.NewBlockInfo(nil, 0, dummyRootHash))
			blockInfoProvider.GetBlockInfoCalled = func() common.BlockInfo {
				return holders.NewBlockInfo(nil, 0, nil)
			}
			assert.True(t, errors.Is(accountsApi.RecreateTrieIfNecessary(), state.ErrNilRootHash))
		})
		t.Run("block info contains an empty root hash should error", func(t *testing.T) {
			accountsApi.SetCurrentBlockInfo(holders.NewBlockInfo(nil, 0, dummyRootHash))
			blockInfoProvider.GetBlockInfoCalled = func() common.BlockInfo {
				return holders.NewBlockInfo(nil, 0, make([]byte, 0))
			}
			assert.True(t, errors.Is(accountsApi.RecreateTrieIfNecessary(), state.ErrNilRootHash))
		})
	})
	t.Run("root hash already set, return nil and do not call recreate", func(t *testing.T) {
		t.Parallel()

		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				require.Fail(t, "should have not called RecreateAllTriesCalled")

				return nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		accountsApi.SetCurrentBlockInfo(holders.NewBlockInfo(nil, 0, dummyRootHash))

		assert.Nil(t, accountsApi.RecreateTrieIfNecessary())
	})
	t.Run("different root hash should call recreate", func(t *testing.T) {
		t.Parallel()

		oldRootHash := []byte("old root hash")
		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				assert.Equal(t, rootHash, rootHash)

				return nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		accountsApi.SetCurrentBlockInfo(holders.NewBlockInfo(nil, 0, oldRootHash))

		assert.Nil(t, accountsApi.RecreateTrieIfNecessary())
		lastRootHash, err := accountsApi.RootHash()
		assert.Equal(t, dummyRootHash, lastRootHash)
		assert.Nil(t, err)
	})
	t.Run("recreate fails should return error and set last root hash to nil", func(t *testing.T) {
		t.Parallel()

		oldRootHash := []byte("old root hash")
		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				assert.Equal(t, rootHash, rootHash)

				return expectedErr
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		accountsApi.SetCurrentBlockInfo(holders.NewBlockInfo(nil, 0, oldRootHash))

		assert.Equal(t, expectedErr, accountsApi.RecreateTrieIfNecessary())
		lastRootHash, err := accountsApi.RootHash()
		assert.Nil(t, lastRootHash)
		assert.Equal(t, state.ErrNilRootHash, err)
	})
}

func TestAccountsDBAPi_doRecreateTrieWhenReEntranceHappened(t *testing.T) {
	t.Parallel()

	targetRootHash := []byte("root hash")
	numCalled := 0
	accountsAdapter := &mockState.AccountsStub{
		GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
		RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
			numCalled++
			return nil
		},
	}

	accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
	assert.Nil(t, accountsApi.DoRecreateTrieWithBlockInfo(holders.NewBlockInfo(nil, 0, targetRootHash)))
	assert.Nil(t, accountsApi.DoRecreateTrieWithBlockInfo(holders.NewBlockInfo(nil, 0, targetRootHash)))
	assert.Equal(t, 1, numCalled)
}

func TestAccountsDBApi_NotPermittedOperations(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should have not panicked")
		}
	}()

	accountsApi, _ := state.NewAccountsDBApi(&mockState.AccountsStub{
		GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
	}, createBlockInfoProviderStub(dummyRootHash))

	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.SaveAccount(nil))
	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.RemoveAccount(nil))
	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.RevertToSnapshot(0))

	buff, err := accountsApi.CommitInEpoch(0, 0)
	assert.Nil(t, buff)
	assert.Equal(t, state.ErrOperationNotPermitted, err)

	buff, err = accountsApi.Commit()
	assert.Nil(t, buff)
	assert.Equal(t, state.ErrOperationNotPermitted, err)

	resultedMap, err := accountsApi.RecreateAllTries(nil)
	assert.Nil(t, resultedMap)
	assert.Equal(t, state.ErrOperationNotPermitted, err)
}

func TestAccountsDBApi_RecreateTrie(t *testing.T) {
	t.Parallel()

	t.Run("should error if the roothash holder is nil", func(t *testing.T) {
		t.Parallel()
		accountsApi, _ := state.NewAccountsDBApi(&mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(options common.RootHashHolder) error {
				assert.Fail(t, "should have not called accountsApi.RecreateTrie")

				return nil
			},
		}, createBlockInfoProviderStub(dummyRootHash))

		err := accountsApi.RecreateTrie(nil)
		assert.Equal(t, trie.ErrNilRootHashHolder, err)
	})

	t.Run("should error if the inner account RecreateTrie errors", func(t *testing.T) {
		t.Parallel()
		expectedErr := errors.New("root hash err")
		trieStub := &testTrie.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &storageManager.StorageManagerStub{}
			},
		}
		trieStub.RecreateCalled = func(_ common.RootHashHolder) (tree common.Trie, e error) {
			return nil, expectedErr
		}

		adb := generateAccountDBFromTrie(trieStub)
		accountsApi, _ := state.NewAccountsDBApi(adb, createBlockInfoProviderStub(dummyRootHash))

		err := accountsApi.RecreateTrie(holders.NewDefaultRootHashesHolder([]byte("hash")))
		assert.Equal(t, expectedErr, err)
	})

	t.Run("should early exit if block info is the same", func(t *testing.T) {
		t.Parallel()
		currentHash := []byte("hash1")
		accountsApi, _ := state.NewAccountsDBApi(&mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieIfNeededCalled: func(options common.RootHashHolder) error {
				assert.Fail(t, "should have not called accountsApi.RecreateTrieIfNeeded")

				return nil
			},
		}, createBlockInfoProviderStub(dummyRootHash))
		accountsApi.SetCurrentBlockInfo(holders.NewBlockInfo(nil, 0, currentHash))

		err := accountsApi.RecreateTrie(holders.NewDefaultRootHashesHolder(currentHash))
		assert.NoError(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		wasCalled := false
		accountsApi, _ := state.NewAccountsDBApi(&mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				wasCalled = true
				return nil
			},
		}, createBlockInfoProviderStub(dummyRootHash))

		err := accountsApi.RecreateTrie(holders.NewDefaultRootHashesHolder([]byte{}))
		assert.NoError(t, err)
		assert.True(t, wasCalled)
	})
}

func TestAccountsDB_RecreateTrieIfNeeded(t *testing.T) {
	t.Parallel()

	t.Run("should error if the roothash holder is nil", func(t *testing.T) {
		t.Parallel()
		accountsApi, _ := state.NewAccountsDBApi(&mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieIfNeededCalled: func(options common.RootHashHolder) error {
				assert.Fail(t, "should have not called accountsApi.RecreateTrieIfNeeded")

				return nil
			},
		}, createBlockInfoProviderStub(dummyRootHash))

		err := accountsApi.RecreateTrieIfNeeded(nil)
		assert.Equal(t, trie.ErrNilRootHashHolder, err)
	})

	t.Run("should error if the inner account RecreateTrieIfNeeded errors", func(t *testing.T) {
		t.Parallel()
		expectedErr := errors.New("root hash err")
		trieStub := &testTrie.TrieStub{
			RootCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
			GetStorageManagerCalled: func() common.StorageManager {
				return &storageManager.StorageManagerStub{}
			},
		}
		adb := generateAccountDBFromTrie(trieStub)
		accountsApi, _ := state.NewAccountsDBApi(adb, createBlockInfoProviderStub(dummyRootHash))

		err := accountsApi.RecreateTrieIfNeeded(holders.NewDefaultRootHashesHolder([]byte("hash")))
		assert.Equal(t, expectedErr, err)
	})

	t.Run("should early exit if block info is the same", func(t *testing.T) {
		t.Parallel()
		currentHash := []byte("hash1")
		trieStub := &testTrie.TrieStub{
			RootCalled: func() ([]byte, error) {
				return currentHash, nil
			},
			GetStorageManagerCalled: func() common.StorageManager {
				return &storageManager.StorageManagerStub{}
			},
		}
		adb := generateAccountDBFromTrie(trieStub)
		accountsApi, _ := state.NewAccountsDBApi(adb, createBlockInfoProviderStub(dummyRootHash))
		accountsApi.SetCurrentBlockInfo(holders.NewBlockInfo(nil, 0, currentHash))

		err := accountsApi.RecreateTrieIfNeeded(holders.NewDefaultRootHashesHolder(currentHash))
		assert.NoError(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		currentHash := []byte("hash1")
		newHash := []byte("hash2")
		recreatedCalled := false
		recreatedTrie := &testTrie.TrieStub{
			RecreateCalled: func(options common.RootHashHolder) (common.Trie, error) {
				require.Empty(t, options.GetRootHash())
				return &testTrie.TrieStub{}, nil
			},
		}

		// mock trie to check whether RecreateTrie was triggered
		trieStub := &testTrie.TrieStub{
			RootCalled: func() ([]byte, error) {
				return currentHash, nil
			},
			RecreateCalled: func(options common.RootHashHolder) (common.Trie, error) {
				if len(options.GetRootHash()) == 0 {
					return &testTrie.TrieStub{}, nil
				}
				recreatedCalled = true
				return recreatedTrie, nil
			},
		}

		adb := generateAccountDBFromTrie(trieStub)
		accountsApi, _ := state.NewAccountsDBApi(adb, createBlockInfoProviderStub(dummyRootHash))

		// same hash -> no recreation
		optsSame := holders.NewDefaultRootHashesHolder(currentHash)
		err := accountsApi.RecreateTrieIfNeeded(optsSame)
		assert.NoError(t, err)
		assert.False(t, recreatedCalled, "should not recreate if same root hash")

		// different hash -> should recreate
		optsDifferent := holders.NewDefaultRootHashesHolder(newHash)
		err = accountsApi.RecreateTrieIfNeeded(optsDifferent)
		assert.NoError(t, err)
		assert.True(t, recreatedCalled, "should recreate if root hash differs")
	})
}

func TestAccountsDBApi_RecreateTrieFromEpoch(t *testing.T) {
	t.Parallel()

	t.Run("should error if the roothash holder is nil", func(t *testing.T) {
		accountsApi, _ := state.NewAccountsDBApi(&mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(options common.RootHashHolder) error {
				assert.Fail(t, "should have not called accountsApi.RecreateTrieFromEpochCalled")

				return nil
			},
		}, createBlockInfoProviderStub(dummyRootHash))

		err := accountsApi.RecreateTrie(nil)
		assert.Equal(t, trie.ErrNilRootHashHolder, err)
	})
	t.Run("should work", func(t *testing.T) {
		wasCalled := false
		rootHash := []byte("root hash")
		epoch := core.OptionalUint32{Value: 37, HasValue: true}
		accountsApi, _ := state.NewAccountsDBApi(&mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(options common.RootHashHolder) error {
				wasCalled = true
				assert.Equal(t, rootHash, options.GetRootHash())
				assert.Equal(t, epoch, options.GetEpoch())
				return nil
			},
		}, createBlockInfoProviderStub(dummyRootHash))

		holder := holders.NewRootHashHolder(rootHash, epoch)
		err := accountsApi.RecreateTrie(holder)
		assert.NoError(t, err)
		assert.True(t, wasCalled)
	})
}

func TestAccountsDBApi_EmptyMethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should have not panicked")
		}
	}()

	accountsApi, _ := state.NewAccountsDBApi(&mockState.AccountsStub{
		GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
	}, createBlockInfoProviderStub(dummyRootHash))

	accountsApi.PruneTrie(nil, 0, state.NewPruningHandler(state.EnableDataRemoval))
	accountsApi.CancelPrune(nil, 0)
	accountsApi.SnapshotState(nil, 0)

	assert.Equal(t, 0, accountsApi.JournalLen())
}

func TestAccountsDBApi_SimpleProxyMethodsShouldWork(t *testing.T) {
	t.Parallel()

	isPruningEnabledCalled := false
	getStackDebugFirstEntryCalled := false
	closeCalled := false
	getTrieCalled := false
	accountsAdapter := &mockState.AccountsStub{
		RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
			require.Fail(t, "should have not called RecreateTrieCalled")

			return nil
		},
		IsPruningEnabledCalled: func() bool {
			isPruningEnabledCalled = true
			return false
		},
		GetStackDebugFirstEntryCalled: func() []byte {
			getStackDebugFirstEntryCalled = true
			return nil
		},
		CloseCalled: func() error {
			closeCalled = true
			return nil
		},
		GetTrieCalled: func(i []byte) (common.Trie, error) {
			getTrieCalled = true
			return &testTrie.TrieStub{}, nil
		},
	}

	accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))

	assert.False(t, accountsApi.IsPruningEnabled())
	assert.Nil(t, accountsApi.GetStackDebugFirstEntry())
	assert.Nil(t, accountsApi.Close())

	tr, err := accountsApi.GetTrie(nil)
	assert.False(t, check.IfNil(tr))
	assert.Nil(t, err)

	assert.True(t, isPruningEnabledCalled)
	assert.True(t, getStackDebugFirstEntryCalled)
	assert.True(t, closeCalled)
	assert.True(t, getTrieCalled)
}

func TestAccountsDBApi_GetExistingAccount(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("recreate trie fails", func(t *testing.T) {
		t.Parallel()

		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				return expectedErr
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Fail(t, "should have not called inner method")
				return nil, nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		account, err := accountsApi.GetExistingAccount(nil)
		assert.Nil(t, account)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("recreate trie works, should call inner method", func(t *testing.T) {
		t.Parallel()

		recreateTrieCalled := false
		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				recreateTrieCalled = true
				return nil
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return createUserAcc(addressContainer), nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		account, err := accountsApi.GetExistingAccount([]byte("address"))
		assert.False(t, check.IfNil(account))
		assert.Nil(t, err)
		assert.True(t, recreateTrieCalled)
	})
}

func TestAccountsDBApi_GetAccountFromBytes(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("recreate trie fails", func(t *testing.T) {
		t.Parallel()

		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				return expectedErr
			},
			GetAccountFromBytesCalled: func(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error) {
				require.Fail(t, "should have not called inner method")
				return nil, nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		account, err := accountsApi.GetAccountFromBytes(nil, nil)
		assert.Nil(t, account)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("recreate trie works, should call inner method", func(t *testing.T) {
		t.Parallel()

		recreateTrieCalled := false
		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				recreateTrieCalled = true
				return nil
			},
			GetAccountFromBytesCalled: func(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error) {
				return createUserAcc(address), nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		account, err := accountsApi.GetAccountFromBytes([]byte("address"), []byte("bytes"))
		assert.False(t, check.IfNil(account))
		assert.Nil(t, err)
		assert.True(t, recreateTrieCalled)
	})
}

func TestAccountsDBApi_LoadAccount(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("recreate trie fails", func(t *testing.T) {
		t.Parallel()

		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				return expectedErr
			},
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				require.Fail(t, "should have not called inner method")
				return nil, nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		account, err := accountsApi.LoadAccount(nil)
		assert.Nil(t, account)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("recreate trie works, should call inner method", func(t *testing.T) {
		t.Parallel()

		recreateTrieCalled := false
		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				recreateTrieCalled = true
				return nil
			},
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return createUserAcc(address), nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		account, err := accountsApi.LoadAccount([]byte("address"))
		assert.False(t, check.IfNil(account))
		assert.Nil(t, err)
		assert.True(t, recreateTrieCalled)
	})
}

func TestAccountsDBApi_GetCode(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("recreate trie fails", func(t *testing.T) {
		t.Parallel()

		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				return expectedErr
			},
			GetCodeCalled: func(codeHash []byte) []byte {
				require.Fail(t, "should have not called inner method")
				return nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		code := accountsApi.GetCode(nil)
		assert.Nil(t, code)
	})
	t.Run("recreate trie works, should call inner method", func(t *testing.T) {
		t.Parallel()

		providedCode := []byte("code")
		recreateTrieCalled := false
		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				recreateTrieCalled = true
				return nil
			},
			GetCodeCalled: func(codeHash []byte) []byte {
				return providedCode
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		code := accountsApi.GetCode([]byte("address"))
		assert.Equal(t, providedCode, code)
		assert.True(t, recreateTrieCalled)
	})
}

func TestAccountsDBApi_GetAllLeaves(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("recreate trie fails", func(t *testing.T) {
		t.Parallel()

		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				return expectedErr
			},
			GetAllLeavesCalled: func(_ *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.TrieLeafParser) error {
				require.Fail(t, "should have not called inner method")
				return nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		err := accountsApi.GetAllLeaves(&common.TrieIteratorChannels{}, nil, []byte{}, parsers.NewMainTrieLeafParser())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("recreate trie works, should call inner method", func(t *testing.T) {
		t.Parallel()

		providedChan := &common.TrieIteratorChannels{LeavesChan: make(chan core.KeyValueHolder)}
		recreateTrieCalled := false
		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				recreateTrieCalled = true
				return nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		err := accountsApi.GetAllLeaves(providedChan, context.Background(), []byte("address"), parsers.NewMainTrieLeafParser())
		assert.Nil(t, err)
		assert.True(t, recreateTrieCalled)
	})
}

func TestAccountsDBApi_GetAccountWithBlockInfoWhenHighConcurrency(t *testing.T) {
	numTestRuns := 500
	numChangesOfCurrentBlockInfo := 100
	numConcurrentRoutinesReadingAccount := 100

	for run := 0; run < numTestRuns; run++ {
		dummyAccount := createDummyAccountWithBalanceBytes([]byte{42, 0, 0, 0, 0, 0, 0, 0})
		currentBlockInfo := createDummyBlockInfoWithNonce(42)
		var dummyAccountMutex sync.RWMutex
		var currentBlockInfoMutex sync.RWMutex

		accountsAdapter := &mockState.AccountsStub{
			GetTrieCalled: func(_ []byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
			RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
				dummyAccountMutex.Lock()
				defer dummyAccountMutex.Unlock()

				// When a trie is recreated, we "add" to it a single account,
				// having the balance correlated with the trie rootHash (for the sake of the test, for easier assertions).
				dummyAccount = createDummyAccountWithBalanceBytes(rootHash.GetRootHash())
				return nil
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				dummyAccountMutex.RLock()
				defer dummyAccountMutex.RUnlock()
				return dummyAccount, nil
			},
		}

		blockInfoProvider := &testscommon.BlockInfoProviderStub{
			GetBlockInfoCalled: func() common.BlockInfo {
				currentBlockInfoMutex.RLock()
				defer currentBlockInfoMutex.RUnlock()
				return currentBlockInfo
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockInfoProvider)

		var wg sync.WaitGroup

		// Simulate the processing goroutine
		wg.Add(1)
		go func() {
			for i := 1; i <= numChangesOfCurrentBlockInfo; i++ {
				currentBlockInfoMutex.Lock()
				currentBlockInfo = createDummyBlockInfoWithNonce(uint64(i))
				currentBlockInfoMutex.Unlock()
			}
			wg.Done()
		}()

		// Simulate many concurrent goroutines calling GetAccountWithBlockInfo()
		for i := 0; i < numConcurrentRoutinesReadingAccount; i++ {
			wg.Add(1)

			go func() {
				account, blockInfo, _ := accountsApi.GetAccountWithBlockInfo([]byte("address"), holders.NewRootHashHolderAsEmpty())
				userAccount := account.(state.UserAccountHandler)
				expectedDummyBalanceBytes := blockInfo.GetRootHash()
				actualBalanceBytes := userAccount.GetBalance().Bytes()

				assert.Equal(t, expectedDummyBalanceBytes, actualBalanceBytes)
				wg.Done()
			}()
		}

		wg.Wait()
	}
}

func createDummyAccountWithBalanceBytes(balanceBytes []byte) state.UserAccountHandler {
	return &mockState.UserAccountStub{Balance: new(big.Int).SetBytes(balanceBytes)}
}

func createDummyBlockInfoWithNonce(nonce uint64) common.BlockInfo {
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, nonce)
	dummyBlockInfo := holders.NewBlockInfo(nonceBytes, nonce, nonceBytes)

	return dummyBlockInfo
}

func TestAccountsDBApi_PreparesDataTrieCreatorOncePerSelection(t *testing.T) {
	for _, method := range []string{"existing", "load", "bytes"} {
		t.Run(method, func(t *testing.T) {
			creator := &testTrie.TrieStub{}
			preparations, bindings := 0, 0
			inner := &mockState.AccountsStub{
				RecreateTrieCalled: func(common.RootHashHolder) error { return nil },
				GetTrieCalled: func(root []byte) (common.Trie, error) {
					require.Empty(t, root)
					preparations++
					return creator, nil
				},
			}
			load := func([]byte) (vmcommon.AccountHandler, error) {
				return &mockState.UserAccountStub{
					SetDataTrieCreatorCalled: func(actual common.DataTrieCreator) {
						require.Same(t, creator, actual)
						bindings++
					},
				}, nil
			}
			inner.GetExistingAccountCalled, inner.LoadAccountCalled = load, load
			inner.GetAccountFromBytesCalled = func(address, _ []byte) (vmcommon.AccountHandler, error) { return load(address) }
			adapter, err := state.NewAccountsDBApi(inner, createBlockInfoProviderStub(dummyRootHash))
			require.NoError(t, err)
			for i := 0; i < 5; i++ {
				_, err = creatorTestAccount(adapter, method)
				require.NoError(t, err)
			}
			require.Equal(t, 1, preparations)
			require.Equal(t, 5, bindings)
		})
	}
}

func creatorTestAccount(adapter state.AccountsAdapterAPI, method string) (vmcommon.AccountHandler, error) {
	switch method {
	case "load":
		return adapter.LoadAccount([]byte("address"))
	case "bytes":
		return adapter.GetAccountFromBytes([]byte("address"), []byte("account bytes"))
	default:
		return adapter.GetExistingAccount([]byte("address"))
	}
}

func TestAccountsDBApi_FailedCreatorPreparationInvalidatesSelection(t *testing.T) {
	for _, method := range []string{"recreate", "if-needed", "block-info"} {
		for _, failure := range []string{"recreate error", "creator error", "nil creator"} {
			t.Run(method+"/"+failure, func(t *testing.T) {
				creator := &testTrie.TrieStub{}
				expected := errors.New("injected selection error")
				fail := false
				recreations := 0
				recreate := func(common.RootHashHolder) error {
					recreations++
					if fail && failure == "recreate error" {
						return expected
					}
					return nil
				}
				inner := &mockState.AccountsStub{
					RecreateTrieCalled: recreate, RecreateTrieIfNeededCalled: recreate,
					GetTrieCalled: func([]byte) (common.Trie, error) {
						if fail && failure == "creator error" {
							return nil, expected
						}
						if fail && failure == "nil creator" {
							return nil, nil
						}
						return creator, nil
					},
				}
				adapter, err := state.NewAccountsDBApi(inner, createBlockInfoProviderStub(dummyRootHash))
				require.NoError(t, err)
				selectRoot := func(root []byte) error {
					switch method {
					case "if-needed":
						return adapter.RecreateTrieIfNeeded(holders.NewDefaultRootHashesHolder(root))
					case "block-info":
						return adapter.DoRecreateTrieWithBlockInfo(holders.NewBlockInfo(nil, 0, root))
					default:
						return adapter.RecreateTrie(holders.NewDefaultRootHashesHolder(root))
					}
				}
				require.NoError(t, selectRoot([]byte("root-A")))
				fail = true
				err = selectRoot([]byte("root-B"))
				if failure == "nil creator" {
					require.ErrorIs(t, err, state.ErrNilTrie)
				} else {
					require.ErrorIs(t, err, expected)
				}
				root, err := adapter.RootHash()
				require.ErrorIs(t, err, state.ErrNilRootHash)
				require.Nil(t, root)
				fail = false
				require.NoError(t, selectRoot([]byte("root-B")))
				root, err = adapter.RootHash()
				require.NoError(t, err)
				require.Equal(t, []byte("root-B"), root)
				require.Equal(t, 3, recreations, "failed preparation must not make a same-root retry skip recreation")
			})
		}
	}
}

func TestAccountsDBApi_CreatorBindingIsProtectedFromConcurrentRecreation(t *testing.T) {
	for _, method := range []string{"existing", "load", "bytes"} {
		t.Run(method, func(t *testing.T) {
			creators := []common.Trie{&testTrie.TrieStub{}, &testTrie.TrieStub{}}
			selected := 0
			readerStarted, releaseReader := make(chan struct{}), make(chan struct{})
			bound := make(chan common.DataTrieCreator, 1)
			inner := &mockState.AccountsStub{
				RecreateTrieCalled: func(options common.RootHashHolder) error {
					if string(options.GetRootHash()) == "root-B" {
						selected = 1
					}
					return nil
				},
				GetTrieCalled: func([]byte) (common.Trie, error) { return creators[selected], nil },
			}
			load := func([]byte) (vmcommon.AccountHandler, error) {
				close(readerStarted)
				<-releaseReader
				return &mockState.UserAccountStub{SetDataTrieCreatorCalled: func(creator common.DataTrieCreator) { bound <- creator }}, nil
			}
			inner.GetExistingAccountCalled, inner.LoadAccountCalled = load, load
			inner.GetAccountFromBytesCalled = func(address, _ []byte) (vmcommon.AccountHandler, error) { return load(address) }
			adapter, err := state.NewAccountsDBApi(inner, createBlockInfoProviderStub([]byte("root-A")))
			require.NoError(t, err)
			readDone := make(chan error, 1)
			go func() { _, readErr := creatorTestAccount(adapter, method); readDone <- readErr }()
			<-readerStarted
			writeDone, writerStarted := make(chan error, 1), make(chan struct{})
			go func() {
				close(writerStarted)
				writeDone <- adapter.RecreateTrie(holders.NewDefaultRootHashesHolder([]byte("root-B")))
			}()
			<-writerStarted
			crossed := false
			var writeErr error
			select {
			case writeErr = <-writeDone:
				crossed = true
			case <-time.After(25 * time.Millisecond):
			}
			close(releaseReader)
			require.NoError(t, <-readDone)
			if !crossed {
				require.NoError(t, <-writeDone)
			}
			require.NoError(t, writeErr)
			require.False(t, crossed, "recreation crossed account retrieval")
			select {
			case creator := <-bound:
				require.Same(t, creators[0], creator)
			default:
				t.Fatalf("%s did not bind its selected creator", method)
			}
		})
	}
}

type creatorSelectionBlockInfo struct {
	common.BlockInfo
	equal func(common.BlockInfo) bool
}

func (info *creatorSelectionBlockInfo) Equal(other common.BlockInfo) bool {
	return info.equal(other)
}

func TestAccountsDBApi_RejectsSelectionThatFailedBeforeAccountLock(t *testing.T) {
	for _, method := range []string{"existing", "load", "bytes"} {
		t.Run(method, func(t *testing.T) {
			expected := errors.New("other recreation failed")
			accountCalls := 0
			load := func([]byte) (vmcommon.AccountHandler, error) {
				accountCalls++
				return &mockState.UserAccountStub{}, nil
			}
			inner := &mockState.AccountsStub{
				RecreateTrieCalled: func(options common.RootHashHolder) error {
					if string(options.GetRootHash()) == "root-B" {
						return expected
					}
					return nil
				},
				GetTrieCalled:            func([]byte) (common.Trie, error) { return &testTrie.TrieStub{}, nil },
				GetExistingAccountCalled: load, LoadAccountCalled: load,
				GetAccountFromBytesCalled: func(address, _ []byte) (vmcommon.AccountHandler, error) { return load(address) },
			}
			info := &creatorSelectionBlockInfo{BlockInfo: holders.NewBlockInfo(nil, 0, []byte("root-A"))}
			adapter, err := state.NewAccountsDBApi(inner, &testscommon.BlockInfoProviderStub{
				GetBlockInfoCalled: func() common.BlockInfo { return info },
			})
			require.NoError(t, err)
			require.NoError(t, adapter.RecreateTrie(holders.NewDefaultRootHashesHolder([]byte("root-A"))))
			info.equal = func(common.BlockInfo) bool {
				// Equality runs after the initial selection read lock is released.
				// Reproduce another caller failing before the account lock is taken.
				require.ErrorIs(t, adapter.RecreateTrie(holders.NewDefaultRootHashesHolder([]byte("root-B"))), expected)
				return true
			}
			account, err := creatorTestAccount(adapter, method)
			require.ErrorIs(t, err, state.ErrNilRootHash)
			require.Nil(t, account)
			require.Zero(t, accountCalls)
		})
	}
}

func TestAccountsDBApi_AccountEntryPointsBindDataTrieCreator(t *testing.T) {
	t.Parallel()

	for _, method := range []string{"GetAccountFromBytes", "LoadAccount", "GetExistingAccount"} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			for _, scenario := range []string{"user account", "peer account", "nil account", "inner error"} {
				t.Run(scenario, func(t *testing.T) {
					t.Parallel()

					address, accountBytes := []byte("address"), []byte("serialized account")
					bindings, innerCalls := 0, 0
					creator := &testTrie.TrieStub{}
					var expected vmcommon.AccountHandler
					var innerErr error
					switch scenario {
					case "user account":
						expected = &mockState.UserAccountStub{
							SetDataTrieCreatorCalled: func(gotCreator common.DataTrieCreator) {
								require.Same(t, creator, gotCreator)
								bindings++
							},
							EnableDataTrieRootValidationCalled: func() {
								t.Fatal("binding the creator already enables root validation")
							},
							RetrieveValueCalled: func([]byte) ([]byte, uint32, error) {
								t.Fatal("returning an account must not read its storage")
								return nil, 0, nil
							},
						}
					case "peer account":
						expected = &mockState.PeerAccountHandlerMock{}
					case "inner error":
						innerErr = errors.New("account load failed")
					}

					inner := &mockState.AccountsStub{
						GetTrieCalled: func(_ []byte) (common.Trie, error) { return creator, nil },
						RecreateTrieCalled: func(options common.RootHashHolder) error {
							require.Equal(t, dummyRootHash, options.GetRootHash())
							return nil
						},
						GetAccountFromBytesCalled: func(gotAddress, gotBytes []byte) (vmcommon.AccountHandler, error) {
							require.Equal(t, "GetAccountFromBytes", method)
							require.Equal(t, address, gotAddress)
							require.Equal(t, accountBytes, gotBytes)
							innerCalls++
							return expected, innerErr
						},
						GetExistingAccountCalled: func(gotAddress []byte) (vmcommon.AccountHandler, error) {
							require.Equal(t, "GetExistingAccount", method)
							require.Equal(t, address, gotAddress)
							innerCalls++
							return expected, innerErr
						},
						LoadAccountCalled: func(gotAddress []byte) (vmcommon.AccountHandler, error) {
							require.Equal(t, "LoadAccount", method)
							require.Equal(t, address, gotAddress)
							innerCalls++
							return expected, innerErr
						},
					}
					adapter, err := state.NewAccountsDBApi(inner, createBlockInfoProviderStub(dummyRootHash))
					require.NoError(t, err)

					var account vmcommon.AccountHandler
					switch method {
					case "GetAccountFromBytes":
						account, err = adapter.GetAccountFromBytes(address, accountBytes)
					case "LoadAccount":
						account, err = adapter.LoadAccount(address)
					case "GetExistingAccount":
						account, err = adapter.GetExistingAccount(address)
					}
					require.Equal(t, 1, innerCalls)
					if innerErr != nil {
						require.ErrorIs(t, err, innerErr)
						require.Nil(t, account)
					} else {
						require.NoError(t, err)
						require.Equal(t, expected, account)
					}
					expectedBindings := 0
					if scenario == "user account" {
						expectedBindings = 1
					}
					require.Equal(t, expectedBindings, bindings)
				})
			}
		})
	}
}
