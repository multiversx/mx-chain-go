package state_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	mockState "github.com/ElrondNetwork/elrond-go/testscommon/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/trie"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccountsDBApi(t *testing.T) {
	t.Parallel()

	t.Run("nil accounts adapter should error", func(t *testing.T) {
		t.Parallel()

		accountsApi, err := state.NewAccountsDBApi(nil, &testscommon.ChainHandlerStub{})

		assert.True(t, check.IfNil(accountsApi))
		assert.Equal(t, state.ErrNilAccountsAdapter, err)
	})
	t.Run("nil chain handler should error", func(t *testing.T) {
		t.Parallel()

		accountsApi, err := state.NewAccountsDBApi(&mockState.AccountsStub{}, nil)

		assert.True(t, check.IfNil(accountsApi))
		assert.Equal(t, state.ErrNilChainHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		accountsApi, err := state.NewAccountsDBApi(&mockState.AccountsStub{}, &testscommon.ChainHandlerStub{})

		assert.False(t, check.IfNil(accountsApi))
		assert.Nil(t, err)
	})
}

func TestAccountsDBAPi_recreateTrieIfNecessary(t *testing.T) {
	t.Parallel()

	blockchainRootHash := []byte("root hash")
	blockchain := &testscommon.ChainHandlerStub{
		GetCurrentBlockRootHashCalled: func() []byte {
			return blockchainRootHash
		},
	}
	expectedErr := errors.New("expected error")

	t.Run("blockchain returns nil or empty root hash should error", func(t *testing.T) {
		t.Parallel()

		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieCalled: func(rootHash []byte) error {
				require.Fail(t, "should have not called RecreateAllTriesCalled")

				return nil
			},
		}

		chainHandler := &testscommon.ChainHandlerStub{
			GetCurrentBlockRootHashCalled: func() []byte {
				return nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, chainHandler)
		accountsApi.SetLastRootHash(blockchainRootHash)
		assert.True(t, errors.Is(accountsApi.RecreateTrieIfNecessary(), state.ErrNilRootHash))

		accountsApi.SetLastRootHash(blockchainRootHash)
		chainHandler.GetCurrentBlockRootHashCalled = func() []byte {
			return make([]byte, 0)
		}
		assert.True(t, errors.Is(accountsApi.RecreateTrieIfNecessary(), state.ErrNilRootHash))
	})
	t.Run("root hash already set, return nil and do not call recreate", func(t *testing.T) {
		t.Parallel()

		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieCalled: func(rootHash []byte) error {
				require.Fail(t, "should have not called RecreateAllTriesCalled")

				return nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
		accountsApi.SetLastRootHash(blockchainRootHash)

		assert.Nil(t, accountsApi.RecreateTrieIfNecessary())
	})
	t.Run("different root hash should call recreate", func(t *testing.T) {
		t.Parallel()

		oldRootHash := []byte("old root hash")
		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieCalled: func(rootHash []byte) error {
				assert.Equal(t, blockchainRootHash, rootHash)

				return nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
		accountsApi.SetLastRootHash(oldRootHash)

		assert.Nil(t, accountsApi.RecreateTrieIfNecessary())
		lastRootHash, err := accountsApi.RootHash()
		assert.Equal(t, blockchainRootHash, lastRootHash)
		assert.Nil(t, err)
	})
	t.Run("recreate fails should return error and set last root hash to nil", func(t *testing.T) {
		t.Parallel()

		oldRootHash := []byte("old root hash")
		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieCalled: func(rootHash []byte) error {
				assert.Equal(t, blockchainRootHash, rootHash)

				return expectedErr
			},
		}

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
		accountsApi.SetLastRootHash(oldRootHash)

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

	accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, &testscommon.ChainHandlerStub{})
	assert.Nil(t, accountsApi.DoRecreateTrie(targetRootHash))
	assert.Nil(t, accountsApi.DoRecreateTrie(targetRootHash))
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

	accountsApi, _ := state.NewAccountsDBApi(&mockState.AccountsStub{}, &testscommon.ChainHandlerStub{})

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

	accountsApi, _ := state.NewAccountsDBApi(&mockState.AccountsStub{}, &testscommon.ChainHandlerStub{})

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

	accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, &testscommon.ChainHandlerStub{})

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
	blockchain := &testscommon.ChainHandlerStub{
		GetCurrentBlockRootHashCalled: func() []byte {
			return []byte("new root hash")
		},
	}
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

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
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

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
		account, err := accountsApi.GetExistingAccount([]byte("address"))
		assert.False(t, check.IfNil(account))
		assert.Nil(t, err)
		assert.True(t, recreateTrieCalled)
	})
}

func TestAccountsDBApi_GetAccountFromBytes(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	blockchain := &testscommon.ChainHandlerStub{
		GetCurrentBlockRootHashCalled: func() []byte {
			return []byte("new root hash")
		},
	}
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

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
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

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
		account, err := accountsApi.GetAccountFromBytes([]byte("address"), []byte("bytes"))
		assert.False(t, check.IfNil(account))
		assert.Nil(t, err)
		assert.True(t, recreateTrieCalled)
	})
}

func TestAccountsDBApi_LoadAccount(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	blockchain := &testscommon.ChainHandlerStub{
		GetCurrentBlockRootHashCalled: func() []byte {
			return []byte("new root hash")
		},
	}
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

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
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

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
		account, err := accountsApi.LoadAccount([]byte("address"))
		assert.False(t, check.IfNil(account))
		assert.Nil(t, err)
		assert.True(t, recreateTrieCalled)
	})
}

func TestAccountsDBApi_GetCode(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	blockchain := &testscommon.ChainHandlerStub{
		GetCurrentBlockRootHashCalled: func() []byte {
			return []byte("new root hash")
		},
	}
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

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
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

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
		code := accountsApi.GetCode([]byte("address"))
		assert.Equal(t, providedCode, code)
		assert.True(t, recreateTrieCalled)
	})
}

func TestAccountsDBApi_GetAllLeaves(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	blockchain := &testscommon.ChainHandlerStub{
		GetCurrentBlockRootHashCalled: func() []byte {
			return []byte("new root hash")
		},
	}
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

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
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

		accountsApi, _ := state.NewAccountsDBApi(accountsAdapter, blockchain)
		err := accountsApi.GetAllLeaves(providedChan, context.Background(), []byte("address"))
		assert.Nil(t, err)
		assert.True(t, recreateTrieCalled)
	})
}
