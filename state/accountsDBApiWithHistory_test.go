package state_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common/holders"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	mockState "github.com/ElrondNetwork/elrond-go/testscommon/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccountsDBApiWithHistory(t *testing.T) {
	t.Run("nil accounts adapter should error", func(t *testing.T) {
		accountsApi, err := state.NewAccountsDBApiWithHistory(nil)
		assert.True(t, check.IfNil(accountsApi))
		assert.Equal(t, state.ErrNilAccountsAdapter, err)
	})

	t.Run("should work", func(t *testing.T) {
		accountsApi, err := state.NewAccountsDBApiWithHistory(&mockState.AccountsStub{})
		assert.False(t, check.IfNil(accountsApi))
		assert.Nil(t, err)
	})
}

func TestAccountsDBApiWithHistory_NotPermittedOrNotImplementedOperationsDoNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should have not panicked")
		}
	}()

	accountsApi, _ := state.NewAccountsDBApiWithHistory(&mockState.AccountsStub{})

	account, err := accountsApi.GetExistingAccount([]byte{})
	assert.Nil(t, account)
	assert.Equal(t, state.ErrFunctionalityNotImplemented, err)

	account, err = accountsApi.GetAccountFromBytes([]byte{}, []byte{})
	assert.Nil(t, account)
	assert.Equal(t, state.ErrFunctionalityNotImplemented, err)

	account, err = accountsApi.LoadAccount([]byte{})
	assert.Nil(t, account)
	assert.Equal(t, state.ErrFunctionalityNotImplemented, err)

	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.SaveAccount(nil))
	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.RemoveAccount(nil))

	resultBytes, err := accountsApi.CommitInEpoch(0, 0)
	assert.Nil(t, resultBytes)
	assert.Equal(t, state.ErrOperationNotPermitted, err)

	resultBytes, err = accountsApi.Commit()
	assert.Nil(t, resultBytes)
	assert.Equal(t, state.ErrOperationNotPermitted, err)

	assert.Equal(t, 0, accountsApi.JournalLen())
	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.RevertToSnapshot(0))
	assert.Equal(t, []byte(nil), accountsApi.GetCode(nil))

	resultBytes, err = accountsApi.RootHash()
	assert.Nil(t, resultBytes)
	assert.Equal(t, state.ErrOperationNotPermitted, err)

	accountsApi.PruneTrie(nil, 0, state.NewPruningHandler(state.EnableDataRemoval))
	accountsApi.CancelPrune(nil, 0)
	accountsApi.SnapshotState(nil)
	accountsApi.SetStateCheckpoint(nil)

	assert.Equal(t, false, accountsApi.IsPruningEnabled())
	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.GetAllLeaves(nil, nil, nil))

	resultedMap, err := accountsApi.RecreateAllTries(nil)
	assert.Nil(t, resultedMap)
	assert.Equal(t, state.ErrOperationNotPermitted, err)

	trie, err := accountsApi.GetTrie(nil)
	assert.Nil(t, trie)
	assert.Equal(t, state.ErrFunctionalityNotImplemented, err)

	assert.Equal(t, []byte(nil), accountsApi.GetStackDebugFirstEntry())
	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.RecreateTrie(nil))
}

func TestAccountsDBApiWithHistory_GetAccountWithBlockInfo(t *testing.T) {
	rootHash := []byte("rootHash")
	options := holders.NewGetAccountStateOptions(rootHash)
	arbitraryError := errors.New("arbitrary error")

	t.Run("recreate trie fails", func(t *testing.T) {
		expectedErr := errors.New("expected error")

		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieCalled: func(rootHash []byte) error {
				return expectedErr
			},
		}

		accountsApi, _ := state.NewAccountsDBApiWithHistory(accountsAdapter)
		account, blockInfo, err := accountsApi.GetAccountWithBlockInfo(testscommon.TestPubKeyAlice, options)
		assert.Nil(t, account)
		assert.Nil(t, blockInfo)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("recreate trie works, should call inner.GetExistingAccount()", func(t *testing.T) {
		var recreatedRootHash []byte

		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieCalled: func(rootHash []byte) error {
				recreatedRootHash = rootHash
				return nil
			},
			GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				if bytes.Equal(address, testscommon.TestPubKeyAlice) {
					return state.NewUserAccount(address)
				}

				return nil, errors.New("not found")
			},
		}

		accountsApi, _ := state.NewAccountsDBApiWithHistory(accountsAdapter)
		account, blockInfo, err := accountsApi.GetAccountWithBlockInfo(testscommon.TestPubKeyAlice, options)
		assert.Nil(t, err)
		assert.Equal(t, blockInfo.GetRootHash(), rootHash)
		assert.Equal(t, testscommon.TestPubKeyAlice, account.AddressBytes())
		assert.Equal(t, rootHash, recreatedRootHash)
	})

	t.Run("account is missing, should return error with block info", func(t *testing.T) {
		var recreatedRootHash []byte

		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieCalled: func(rootHash []byte) error {
				recreatedRootHash = rootHash
				return nil
			},
			GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return nil, state.ErrAccNotFound
			},
		}

		accountsApi, _ := state.NewAccountsDBApiWithHistory(accountsAdapter)
		account, blockInfo, err := accountsApi.GetAccountWithBlockInfo(testscommon.TestPubKeyAlice, options)
		assert.Nil(t, account)
		assert.Nil(t, blockInfo)
		assert.Equal(t, state.NewErrAccountNotFoundAtBlock(holders.NewBlockInfo(nil, 0, rootHash)), err)
		assert.Equal(t, rootHash, recreatedRootHash)
	})

	t.Run("arbitrary error on GetExistingAccount(), should be returned (forwarded)", func(t *testing.T) {
		var recreatedRootHash []byte

		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieCalled: func(rootHash []byte) error {
				recreatedRootHash = rootHash
				return nil
			},
			GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return nil, arbitraryError
			},
		}

		accountsApi, _ := state.NewAccountsDBApiWithHistory(accountsAdapter)
		account, blockInfo, err := accountsApi.GetAccountWithBlockInfo(testscommon.TestPubKeyAlice, options)
		assert.Nil(t, account)
		assert.Nil(t, blockInfo)
		assert.Equal(t, arbitraryError, err)
		assert.Equal(t, rootHash, recreatedRootHash)
	})
}

func TestAccountsDBApiWithHistory_GetCodeWithBlockInfo(t *testing.T) {
	contractCodeHash := []byte("codeHash")
	rootHash := []byte("rootHash")
	options := holders.NewGetAccountStateOptions(rootHash)

	t.Run("recreate trie fails", func(t *testing.T) {
		expectedErr := errors.New("expected error")

		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieCalled: func(rootHash []byte) error {
				return expectedErr
			},
		}

		accountsApi, _ := state.NewAccountsDBApiWithHistory(accountsAdapter)
		account, blockInfo, err := accountsApi.GetCodeWithBlockInfo(contractCodeHash, options)
		assert.Nil(t, account)
		assert.Nil(t, blockInfo)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("recreate trie works, should call inner.GetExistingAccount()", func(t *testing.T) {
		var recreatedRootHash []byte

		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieCalled: func(rootHash []byte) error {
				recreatedRootHash = rootHash
				return nil
			},
			GetCodeCalled: func(codeHash []byte) []byte {
				if bytes.Equal(codeHash, contractCodeHash) {
					return []byte("code")
				}

				return nil
			},
		}

		accountsApi, _ := state.NewAccountsDBApiWithHistory(accountsAdapter)
		code, blockInfo, err := accountsApi.GetCodeWithBlockInfo(contractCodeHash, options)
		assert.Nil(t, err)
		assert.Equal(t, blockInfo.GetRootHash(), rootHash)
		assert.Equal(t, []byte("code"), code)
		assert.Equal(t, rootHash, recreatedRootHash)
	})
}
