package state_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	mockState "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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
	assert.Equal(t, state.ErrOperationNotPermitted, accountsApi.GetAllLeaves(&common.TrieIteratorChannels{}, nil, nil, nil))

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
	options := holders.NewRootHashHolder(rootHash, core.OptionalUint32{})
	arbitraryError := errors.New("arbitrary error")

	t.Run("recreate trie fails", func(t *testing.T) {
		expectedErr := errors.New("expected error")

		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieFromEpochCalled: func(_ common.RootHashHolder) error {
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
			RecreateTrieFromEpochCalled: func(options common.RootHashHolder) error {
				recreatedRootHash = options.GetRootHash()
				return nil
			},
			GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				if bytes.Equal(address, testscommon.TestPubKeyAlice) {
					return createUserAcc(address), nil
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
			RecreateTrieFromEpochCalled: func(options common.RootHashHolder) error {
				recreatedRootHash = options.GetRootHash()
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
			RecreateTrieFromEpochCalled: func(options common.RootHashHolder) error {
				recreatedRootHash = options.GetRootHash()
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
	options := holders.NewRootHashHolder(rootHash, core.OptionalUint32{})

	t.Run("recreate trie fails", func(t *testing.T) {
		expectedErr := errors.New("expected error")

		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieFromEpochCalled: func(_ common.RootHashHolder) error {
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
			RecreateTrieFromEpochCalled: func(options common.RootHashHolder) error {
				recreatedRootHash = options.GetRootHash()
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

func TestAccountsDBApiWithHistory_GetAccountWithBlockInfoWhenHighConcurrency(t *testing.T) {
	numTestRuns := 16
	numRootHashes := 128
	numCallsPerRootHash := 128
	recreationsCounterByRootHash := make(map[string]*atomic.Counter)

	for run := 0; run < numTestRuns; run++ {
		for i := 0; i < numRootHashes; i++ {
			rootHashAsString := fmt.Sprintf("%d", i)
			recreationsCounterByRootHash[rootHashAsString] = &atomic.Counter{}
		}

		dummyAccount := createDummyAccountWithBalanceBytes([]byte{42, 0, 0, 0, 0, 0, 0, 0})
		var dummyAccountMutex sync.RWMutex

		accountsAdapter := &mockState.AccountsStub{
			RecreateTrieFromEpochCalled: func(options common.RootHashHolder) error {
				rootHash := options.GetRootHash()

				dummyAccountMutex.Lock()
				defer dummyAccountMutex.Unlock()

				// When a trie is recreated, we "add" to it a single account,
				// having the balance correlated with the trie rootHash (for the sake of the test, for easier assertions).
				dummyAccount = createDummyAccountWithBalanceString(string(rootHash))

				// We also count the re-creation
				counter := recreationsCounterByRootHash[string(rootHash)]
				counter.Increment()
				return nil
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				dummyAccountMutex.RLock()
				defer dummyAccountMutex.RUnlock()
				return dummyAccount, nil
			},
		}

		accountsApiWithHistory, _ := state.NewAccountsDBApiWithHistory(accountsAdapter)

		var wg sync.WaitGroup

		// Simulate many concurrent goroutines calling GetAccountWithBlockInfo()
		for i := 0; i < numRootHashes; i++ {
			for j := 0; j < numCallsPerRootHash; j++ {
				wg.Add(1)

				go func(rootHashAsInt int) {
					rootHashAsString := fmt.Sprintf("%d", rootHashAsInt)
					rootHash := []byte(rootHashAsString)
					options := holders.NewRootHashHolder(rootHash, core.OptionalUint32{})
					account, blockInfo, _ := accountsApiWithHistory.GetAccountWithBlockInfo([]byte("address"), options)
					userAccount := account.(common.UserAccountHandler)

					assert.Equal(t, rootHash, blockInfo.GetRootHash())
					assert.Equal(t, uint64(rootHashAsInt), userAccount.GetBalance().Uint64())
					wg.Done()
				}(i)
			}
		}

		wg.Wait()
	}

	for i := 0; i < numRootHashes; i++ {
		rootHashAsString := fmt.Sprintf("%d", i)
		numRecreations := recreationsCounterByRootHash[rootHashAsString]

		assert.True(t, numRecreations.Get() > 0 && numRecreations.Get() <= int64(numCallsPerRootHash))
	}
}

func createDummyAccountWithBalanceString(balanceString string) common.UserAccountHandler {
	dummyAccount := &mockState.AccountWrapMock{
		Balance: big.NewInt(0),
	}
	dummyBalance, _ := big.NewInt(0).SetString(balanceString, 10)
	_ = dummyAccount.AddToBalance(dummyBalance)

	return dummyAccount
}
