package state_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/holders"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	mockState "github.com/ElrondNetwork/elrond-go/testscommon/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/trie"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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

		accountsApi, err := state.NewAccountsDBApi(&mockState.AccountsStub{}, nil)

		assert.True(t, check.IfNil(accountsApi))
		assert.Equal(t, state.ErrNilBlockInfoProvider, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		accountsApi, err := state.NewAccountsDBApi(&mockState.AccountsStub{}, &testscommon.BlockInfoProviderStub{})

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
			RecreateTrieCalled: func(rootHash []byte) error {
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
			RecreateTrieCalled: func(rootHash []byte) error {
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
			RecreateTrieCalled: func(rootHash []byte) error {
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
			RecreateTrieCalled: func(rootHash []byte) error {
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
		RecreateTrieCalled: func(rootHash []byte) error {
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

	accountsApi, _ := state.NewAccountsDBApi(&mockState.AccountsStub{}, createBlockInfoProviderStub(dummyRootHash))

	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.SaveAccount(nil))
	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.RemoveAccount(nil))
	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.RevertToSnapshot(0))
	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.RecreateTrie(nil))

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

func TestAccountsDBApi_EmptyMethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should have not panicked")
		}
	}()

	accountsApi, _ := state.NewAccountsDBApi(&mockState.AccountsStub{}, createBlockInfoProviderStub(dummyRootHash))

	accountsApi.PruneTrie(nil, 0, state.NewPruningHandler(state.EnableDataRemoval))
	accountsApi.CancelPrune(nil, 0)
	accountsApi.SnapshotState(nil)
	accountsApi.SetStateCheckpoint(nil)

	assert.Equal(t, 0, accountsApi.JournalLen())
}

func TestAccountsDBApi_SimpleProxyMethodsShouldWork(t *testing.T) {
	t.Parallel()

	isPruningEnabledCalled := false
	getStackDebugFirstEntryCalled := false
	closeCalled := false
	getTrieCalled := false
	accountsAdapter := &mockState.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
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
			return &trie.TrieStub{}, nil
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
			RecreateTrieCalled: func(rootHash []byte) error {
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
			RecreateTrieCalled: func(rootHash []byte) error {
				recreateTrieCalled = true
				return nil
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return state.NewUserAccount(addressContainer)
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
			RecreateTrieCalled: func(rootHash []byte) error {
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
			RecreateTrieCalled: func(rootHash []byte) error {
				recreateTrieCalled = true
				return nil
			},
			GetAccountFromBytesCalled: func(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error) {
				return state.NewUserAccount(address)
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
			RecreateTrieCalled: func(rootHash []byte) error {
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
			RecreateTrieCalled: func(rootHash []byte) error {
				recreateTrieCalled = true
				return nil
			},
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return state.NewUserAccount(address)
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
			RecreateTrieCalled: func(rootHash []byte) error {
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
			RecreateTrieCalled: func(rootHash []byte) error {
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
			RecreateTrieCalled: func(rootHash []byte) error {
				return expectedErr
			},
			GetAllLeavesCalled: func(_ chan core.KeyValueHolder, _ context.Context, _ []byte) error {
				require.Fail(t, "should have not called inner method")
				return nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		err := accountsApi.GetAllLeaves(nil, nil, []byte{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("recreate trie works, should call inner method", func(t *testing.T) {
		t.Parallel()

		providedChan := make(chan core.KeyValueHolder)
		recreateTrieCalled := false
		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieCalled: func(rootHash []byte) error {
				recreateTrieCalled = true
				return nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, createBlockInfoProviderStub(dummyRootHash))
		err := accountsApi.GetAllLeaves(providedChan, context.Background(), []byte("address"))
		assert.Nil(t, err)
		assert.True(t, recreateTrieCalled)
	})
}
