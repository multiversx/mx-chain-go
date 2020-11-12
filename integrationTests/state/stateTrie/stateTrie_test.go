package stateTrie

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	transaction2 "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/trie/evictionWaitingList"
	factory2 "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountsDB_RetrieveDataWithSomeValuesShouldWork(t *testing.T) {
	//test simulates creation of a new account, data trie retrieval,
	//adding a (key, value) pair in that data trie, committing changes
	//and then reloading the data trie based on the root hash generated before
	t.Parallel()

	key1 := []byte("ABC")
	val1 := []byte("123")
	key2 := []byte("DEF")
	val2 := []byte("456")
	_, account, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()

	_ = account.DataTrieTracker().SaveKeyValue(key1, val1)
	_ = account.DataTrieTracker().SaveKeyValue(key2, val2)

	err := adb.SaveAccount(account)
	assert.Nil(t, err)

	_, err = adb.Commit()
	assert.Nil(t, err)

	acc, err := adb.LoadAccount(account.AddressBytes())
	assert.Nil(t, err)
	recoveredAccount := acc.(state.UserAccountHandler)

	//verify data
	dataRecovered, err := recoveredAccount.DataTrieTracker().RetrieveValue(key1)
	assert.Nil(t, err)
	assert.Equal(t, val1, dataRecovered)

	dataRecovered, err = recoveredAccount.DataTrieTracker().RetrieveValue(key2)
	assert.Nil(t, err)
	assert.Equal(t, val2, dataRecovered)
}

func TestAccountsDB_PutCodeWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	_, account, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()
	account.SetCode([]byte("Smart contract code"))
	err := adb.SaveAccount(account)
	assert.Nil(t, err)
	assert.NotNil(t, account.GetCodeHash())
	assert.Equal(t, []byte("Smart contract code"), account.GetCode())

	fmt.Printf("SC code is at address: %v\n", account.GetCodeHash())

	acc, err := adb.LoadAccount(account.AddressBytes())
	assert.Nil(t, err)
	recoveredAccount := acc.(state.UserAccountHandler)

	assert.Equal(t, account.GetCode(), recoveredAccount.GetCode())
	assert.Equal(t, account.GetCodeHash(), recoveredAccount.GetCodeHash())
}

func TestAccountsDB_SaveAccountStateWithSomeValues_ShouldWork(t *testing.T) {
	t.Parallel()

	_, account, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()

	err := adb.SaveAccount(account)
	assert.Nil(t, err)
}

func TestAccountsDB_GetJournalizedAccountReturnExistingAccntShouldWork(t *testing.T) {
	t.Parallel()

	balance := big.NewInt(40)
	adr, accountHandler, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()
	account := accountHandler.(state.UserAccountHandler)
	_ = account.AddToBalance(balance)

	err := adb.SaveAccount(account)
	assert.Nil(t, err)

	accountHandlerRecovered, err := adb.LoadAccount(adr)
	assert.Nil(t, err)
	accountRecovered := accountHandlerRecovered.(state.UserAccountHandler)
	assert.NotNil(t, accountRecovered)
	assert.Equal(t, balance, accountRecovered.GetBalance())
}

func TestAccountsDB_GetJournalizedAccountReturnNotFoundAccntShouldWork(t *testing.T) {
	//test when the account does not exists
	t.Parallel()

	adr, _, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()

	//same address of the unsaved account
	accountHandlerRecovered, err := adb.LoadAccount(adr)
	assert.Nil(t, err)
	accountRecovered := accountHandlerRecovered.(state.UserAccountHandler)
	assert.NotNil(t, accountRecovered)
	assert.Equal(t, big.NewInt(0), accountRecovered.GetBalance())
}

func TestAccountsDB_GetExistingAccountConcurrentlyShouldWork(t *testing.T) {
	t.Parallel()

	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	adb, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	wg := sync.WaitGroup{}
	wg.Add(100)

	addresses := make([][]byte, 0)

	//generating 100 different addresses
	for len(addresses) < 100 {
		addr := integrationTests.CreateRandomAddress()

		found := false
		for i := 0; i < len(addresses); i++ {
			if bytes.Equal(addresses[i], addr) {
				found = true
				break
			}
		}

		if !found {
			addresses = append(addresses, addr)
		}
	}

	for i := 0; i < 50; i++ {
		go func(idx int) {
			accnt, err := adb.GetExistingAccount(addresses[idx*2])

			assert.Equal(t, state.ErrAccNotFound, err)
			assert.Nil(t, accnt)

			wg.Done()
		}(i)

		go func(idx int) {
			accnt, err := adb.LoadAccount(addresses[idx*2+1])
			assert.Nil(t, err)
			assert.NotNil(t, accnt)

			err = adb.SaveAccount(accnt)
			assert.Nil(t, err)

			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestAccountsDB_CommitTwoOkAccountsShouldWork(t *testing.T) {
	//test creates 2 accounts (one with a data root)
	//verifies that commit saves the new tries and that can be loaded back
	t.Parallel()

	adr1, _, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()
	adr2 := integrationTests.CreateRandomAddress()

	//first account has the balance of 40
	balance1 := big.NewInt(40)
	state1, err := adb.LoadAccount(adr1)
	assert.Nil(t, err)
	_ = state1.(state.UserAccountHandler).AddToBalance(balance1)

	//second account has the balance of 50 and some data
	balance2 := big.NewInt(50)
	acc, err := adb.LoadAccount(adr2)
	assert.Nil(t, err)

	state2 := acc.(state.UserAccountHandler)
	_ = state2.AddToBalance(balance2)

	key := []byte("ABC")
	val := []byte("123")
	_ = state2.DataTrieTracker().SaveKeyValue(key, val)

	_ = adb.SaveAccount(state1)
	_ = adb.SaveAccount(state2)

	//states are now prepared, committing

	h, err := adb.Commit()
	assert.Nil(t, err)
	fmt.Printf("Result hash: %v\n", base64.StdEncoding.EncodeToString(h))

	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	fmt.Printf("Data committed! Root: %v\n", base64.StdEncoding.EncodeToString(rootHash))

	//reloading a new trie to test if data is inside
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	err = adb.RecreateTrie(rootHash)
	assert.Nil(t, err)

	//checking state1
	newState1, err := adb.LoadAccount(adr1)
	assert.Nil(t, err)
	assert.Equal(t, balance1, newState1.(state.UserAccountHandler).GetBalance())

	//checking state2
	newState2, err := adb.LoadAccount(adr2)
	assert.Nil(t, err)
	assert.Equal(t, balance2, newState2.(state.UserAccountHandler).GetBalance())
	assert.NotNil(t, newState2.(state.UserAccountHandler).GetRootHash())
	valRecovered, err := newState2.(state.UserAccountHandler).DataTrieTracker().RetrieveValue(key)
	assert.Nil(t, err)
	assert.Equal(t, val, valRecovered)
}

func TestTrieDB_RecreateFromStorageShouldWork(t *testing.T) {
	hasher := integrationTests.TestHasher
	store := integrationTests.CreateMemUnit()
	evictionWaitListSize := uint(100)
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(evictionWaitListSize, memorydb.New(), integrationTests.TestMarshalizer)
	trieStorage, _ := trie.NewTrieStorageManager(store, integrationTests.TestMarshalizer, hasher, config.DBConfig{}, ewl, config.TrieStorageManagerConfig{})

	maxTrieLevelInMemory := uint(5)
	tr1, _ := trie.NewTrie(trieStorage, integrationTests.TestMarshalizer, hasher, maxTrieLevelInMemory)

	key := hasher.Compute("key")
	value := hasher.Compute("value")

	_ = tr1.Update(key, value)
	h1, _ := tr1.Root()
	err := tr1.Commit()
	assert.Nil(t, err)

	tr2, err := tr1.Recreate(h1)
	assert.Nil(t, err)

	valRecov, err := tr2.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, value, valRecov)
}

func TestAccountsDB_CommitTwoOkAccountsWithRecreationFromStorageShouldWork(t *testing.T) {
	//test creates 2 accounts (one with a data root)
	//verifies that commit saves the new tries and that can be loaded back
	t.Parallel()

	trieStore, mu := integrationTests.CreateTrieStorageManager()
	adb, _ := integrationTests.CreateAccountsDB(0, trieStore)
	adr1 := integrationTests.CreateRandomAddress()
	adr2 := integrationTests.CreateRandomAddress()

	//first account has the balance of 40
	balance1 := big.NewInt(40)
	state1, err := adb.LoadAccount(adr1)
	assert.Nil(t, err)
	_ = state1.(state.UserAccountHandler).AddToBalance(balance1)

	//second account has the balance of 50 and some data
	balance2 := big.NewInt(50)
	acc, err := adb.LoadAccount(adr2)
	assert.Nil(t, err)

	state2 := acc.(state.UserAccountHandler)
	_ = state2.AddToBalance(balance2)

	key := []byte("ABC")
	val := []byte("123")
	_ = state2.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue(key, val)

	_ = adb.SaveAccount(state1)
	_ = adb.SaveAccount(state2)

	//states are now prepared, committing

	h, err := adb.Commit()
	assert.Nil(t, err)
	fmt.Printf("Result hash: %v\n", base64.StdEncoding.EncodeToString(h))

	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	fmt.Printf("Data committed! Root: %v\n", base64.StdEncoding.EncodeToString(rootHash))

	ewl, _ := evictionWaitingList.NewEvictionWaitingList(100, memorydb.New(), integrationTests.TestMarshalizer)
	trieStorage, _ := trie.NewTrieStorageManager(mu, integrationTests.TestMarshalizer, integrationTests.TestHasher, config.DBConfig{}, ewl, config.TrieStorageManagerConfig{})
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(trieStorage, integrationTests.TestMarshalizer, integrationTests.TestHasher, maxTrieLevelInMemory)
	adb, _ = state.NewAccountsDB(tr, integrationTests.TestHasher, integrationTests.TestMarshalizer, factory.NewAccountCreator())

	//reloading a new trie to test if data is inside
	err = adb.RecreateTrie(h)
	assert.Nil(t, err)

	//checking state1
	newState1, err := adb.LoadAccount(adr1)
	assert.Nil(t, err)
	assert.Equal(t, balance1, newState1.(state.UserAccountHandler).GetBalance())

	//checking state2
	acc2, err := adb.LoadAccount(adr2)
	assert.Nil(t, err)
	newState2 := acc2.(state.UserAccountHandler)
	assert.Equal(t, balance2, newState2.GetBalance())
	assert.NotNil(t, newState2.GetRootHash())
	valRecovered, err := newState2.DataTrieTracker().RetrieveValue(key)
	assert.Nil(t, err)
	assert.Equal(t, val, valRecovered)
}

func TestAccountsDB_CommitAnEmptyStateShouldWork(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "this test should not have panicked")
		}
	}()

	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	adb, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	hash, err := adb.Commit()

	assert.Nil(t, err)
	assert.Equal(t, make([]byte, state.HashLength), hash)
}

func TestAccountsDB_CommitAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	adr1, _, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()

	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrEmpty := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	state1, err := adb.LoadAccount(adr1)
	assert.Nil(t, err)
	_ = adb.SaveAccount(state1)

	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - created account: %v\n", hrCreated)

	_ = state1.(state.UserAccountHandler).AddToBalance(big.NewInt(40))
	_ = adb.SaveAccount(state1)

	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrWithBalance := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - account with balance 40: %v\n", hrWithBalance)

	_, err = adb.Commit()
	assert.Nil(t, err)
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCommit := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - committed: %v\n", hrCommit)

	//commit hash == account with balance
	assert.Equal(t, hrCommit, hrWithBalance)

	_ = state1.(state.UserAccountHandler).SubFromBalance(big.NewInt(40))
	_ = adb.SaveAccount(state1)

	//root hash == hrCreated
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, hrCreated, base64.StdEncoding.EncodeToString(rootHash))
	fmt.Printf("State root - account with balance 0: %v\n", base64.StdEncoding.EncodeToString(rootHash))

	err = adb.RemoveAccount(adr1)
	assert.Nil(t, err)

	//root hash == hrEmpty
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, hrEmpty, base64.StdEncoding.EncodeToString(rootHash))
	fmt.Printf("State root - empty: %v\n", base64.StdEncoding.EncodeToString(rootHash))
}

//------- Revert

func TestAccountsDB_RevertNonceStepByStepAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := integrationTests.CreateRandomAddress()
	adr2 := integrationTests.CreateRandomAddress()

	//Step 1. create accounts objects
	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	adb, _ := integrationTests.CreateAccountsDB(0, trieStorage)
	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrEmpty := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.LoadAccount(adr1)
	assert.Nil(t, err)
	_ = adb.SaveAccount(state1)

	snapshotCreated1 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated1 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.LoadAccount(adr2)
	assert.Nil(t, err)
	_ = adb.SaveAccount(state2)
	snapshotCreated2 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated2 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Save the preset snapshot id
	snapshotPreSet := adb.JournalLen()

	//Step 3. Set Nonces and save data
	state1.(state.UserAccountHandler).IncreaseNonce(40)
	_ = adb.SaveAccount(state1)

	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrWithNonce1 := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - account with nonce 40: %v\n", hrWithNonce1)

	state2.(state.UserAccountHandler).IncreaseNonce(50)
	_ = adb.SaveAccount(state2)

	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrWithNonce2 := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - account with nonce 50: %v\n", hrWithNonce2)

	//Test 3.1. current root hash shall not match created root hash hrCreated2
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	assert.NotEqual(t, hrCreated2, rootHash)

	//Step 4. Revert account nonce and test
	err = adb.RevertToSnapshot(snapshotPreSet)
	assert.Nil(t, err)

	//Test 4.1. current root hash shall match created root hash hrCreated
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrFinal := base64.StdEncoding.EncodeToString(rootHash)
	assert.Equal(t, hrCreated2, hrFinal)
	fmt.Printf("State root - reverted last 2 nonces set: %v\n", hrFinal)
}

func TestAccountsDB_RevertBalanceStepByStepAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := integrationTests.CreateRandomAddress()
	adr2 := integrationTests.CreateRandomAddress()

	//Step 1. create accounts objects
	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	adb, _ := integrationTests.CreateAccountsDB(0, trieStorage)
	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrEmpty := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.LoadAccount(adr1)
	assert.Nil(t, err)
	_ = adb.SaveAccount(state1)

	snapshotCreated1 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated1 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.LoadAccount(adr2)
	assert.Nil(t, err)
	_ = adb.SaveAccount(state2)

	snapshotCreated2 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated2 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Save the preset snapshot id
	snapshotPreSet := adb.JournalLen()

	//Step 3. Set balances and save data
	_ = state1.(state.UserAccountHandler).AddToBalance(big.NewInt(40))
	_ = adb.SaveAccount(state1)

	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrWithBalance1 := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - account with balance 40: %v\n", hrWithBalance1)

	_ = state2.(state.UserAccountHandler).AddToBalance(big.NewInt(50))
	_ = adb.SaveAccount(state2)

	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrWithBalance2 := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - account with balance 50: %v\n", hrWithBalance2)

	//Test 3.1. current root hash shall not match created root hash hrCreated2
	assert.NotEqual(t, hrCreated2, rootHash)

	//Step 4. Revert account balances and test
	err = adb.RevertToSnapshot(snapshotPreSet)
	assert.Nil(t, err)

	//Test 4.1. current root hash shall match created root hash hrCreated
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrFinal := base64.StdEncoding.EncodeToString(rootHash)
	assert.Equal(t, hrCreated2, hrFinal)
	fmt.Printf("State root - reverted last 2 balance set: %v\n", hrFinal)
}

func TestAccountsDB_RevertCodeStepByStepAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	//adr1 puts code hash + code inside trie. adr2 has the same code hash
	//revert should work

	code := []byte("ABC")
	adr1 := integrationTests.CreateRandomAddress()
	adr2 := integrationTests.CreateRandomAddress()

	//Step 1. create accounts objects
	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	adb, _ := integrationTests.CreateAccountsDB(0, trieStorage)
	rootHash, err := adb.RootHash()
	require.Nil(t, err)
	hrEmpty := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.LoadAccount(adr1)
	require.Nil(t, err)
	state1.(state.UserAccountHandler).SetCode(code)
	_ = adb.SaveAccount(state1)

	snapshotCreated1 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	require.Nil(t, err)
	hrCreated1 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.LoadAccount(adr2)
	require.Nil(t, err)
	state2.(state.UserAccountHandler).SetCode(code)
	_ = adb.SaveAccount(state2)

	snapshotCreated2 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	require.Nil(t, err)
	hrCreated2 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)

	//Test 2.1. test that hashes and snapshots ID are different
	require.NotEqual(t, snapshotCreated2, snapshotCreated1)
	require.NotEqual(t, hrCreated1, hrCreated2)

	//Step 3. Revert second account
	err = adb.RevertToSnapshot(snapshotCreated1)
	require.Nil(t, err)

	//Test 3.1. current root hash shall match created root hash hrCreated1
	rootHash, err = adb.RootHash()
	require.Nil(t, err)
	hrCrt := base64.StdEncoding.EncodeToString(rootHash)
	require.Equal(t, hrCreated1, hrCrt)
	fmt.Printf("State root - reverted last account: %v\n", hrCrt)

	//Step 4. Revert first account
	err = adb.RevertToSnapshot(0)
	require.Nil(t, err)

	//Test 4.1. current root hash shall match empty root hash
	rootHash, err = adb.RootHash()
	require.Nil(t, err)
	hrCrt = base64.StdEncoding.EncodeToString(rootHash)
	require.Equal(t, hrEmpty, hrCrt)
	fmt.Printf("State root - reverted first account: %v\n", hrCrt)
}

func TestAccountsDB_RevertDataStepByStepAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	//adr1 puts data inside trie. adr2 puts the same data
	//revert should work

	key := []byte("ABC")
	val := []byte("123")
	adr1 := integrationTests.CreateRandomAddress()
	adr2 := integrationTests.CreateRandomAddress()

	//Step 1. create accounts objects
	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	adb, _ := integrationTests.CreateAccountsDB(0, trieStorage)
	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrEmpty := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.LoadAccount(adr1)
	assert.Nil(t, err)
	_ = state1.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue(key, val)
	err = adb.SaveAccount(state1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated1 := base64.StdEncoding.EncodeToString(rootHash)
	rootHash, err = state1.(state.UserAccountHandler).DataTrie().Root()
	assert.Nil(t, err)
	hrRoot1 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)
	fmt.Printf("Data root - 1-st account: %v\n", hrRoot1)

	state2, err := adb.LoadAccount(adr2)
	assert.Nil(t, err)
	_ = state2.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue(key, val)
	err = adb.SaveAccount(state2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated2 := base64.StdEncoding.EncodeToString(rootHash)
	rootHash, err = state1.(state.UserAccountHandler).DataTrie().Root()
	assert.Nil(t, err)
	hrRoot2 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)
	fmt.Printf("Data root - 2-nd account: %v\n", hrRoot2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Test 2.2 test whether the datatrie roots match
	assert.Equal(t, hrRoot1, hrRoot2)

	//Step 3. Revert 2-nd account ant test roots
	err = adb.RevertToSnapshot(snapshotCreated1)
	assert.Nil(t, err)
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated2Rev := base64.StdEncoding.EncodeToString(rootHash)

	assert.Equal(t, hrCreated1, hrCreated2Rev)

	//Step 4. Revert 1-st account ant test roots
	err = adb.RevertToSnapshot(0)
	assert.Nil(t, err)
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated1Rev := base64.StdEncoding.EncodeToString(rootHash)

	assert.Equal(t, hrEmpty, hrCreated1Rev)
}

func TestAccountsDB_RevertDataStepByStepWithCommitsAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	//adr1 puts data inside trie. adr2 puts the same data
	//revert should work

	key := []byte("ABC")
	val := []byte("123")
	newVal := []byte("124")
	adr1 := integrationTests.CreateRandomAddress()
	adr2 := integrationTests.CreateRandomAddress()

	//Step 1. create accounts objects
	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	adb, _ := integrationTests.CreateAccountsDB(0, trieStorage)
	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrEmpty := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.LoadAccount(adr1)
	assert.Nil(t, err)
	_ = state1.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue(key, val)
	err = adb.SaveAccount(state1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated1 := base64.StdEncoding.EncodeToString(rootHash)
	rootHash, err = state1.(state.UserAccountHandler).DataTrie().Root()
	assert.Nil(t, err)
	hrRoot1 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)
	fmt.Printf("Data root - 1-st account: %v\n", hrRoot1)

	state2, err := adb.LoadAccount(adr2)
	assert.Nil(t, err)
	_ = state2.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue(key, val)
	err = adb.SaveAccount(state2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated2 := base64.StdEncoding.EncodeToString(rootHash)
	rootHash, err = state2.(state.UserAccountHandler).DataTrie().Root()
	assert.Nil(t, err)
	hrRoot2 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)
	fmt.Printf("Data root - 2-nd account: %v\n", hrRoot2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Test 2.2 test that the datatrie roots are different
	assert.NotEqual(t, hrRoot1, hrRoot2)

	//Step 3. Commit
	rootCommit, _ := adb.Commit()
	hrCommit := base64.StdEncoding.EncodeToString(rootCommit)
	fmt.Printf("State root - committed: %v\n", hrCommit)

	//Step 4. 2-nd account changes its data
	snapshotMod := adb.JournalLen()

	state2, err = adb.LoadAccount(adr2)
	assert.Nil(t, err)
	_ = state2.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue(key, newVal)
	err = adb.SaveAccount(state2)
	assert.Nil(t, err)
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated2p1 := base64.StdEncoding.EncodeToString(rootHash)
	rootHash, err = state2.(state.UserAccountHandler).DataTrie().Root()
	assert.Nil(t, err)
	hrRoot2p1 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - modified 2-nd account: %v\n", hrCreated2p1)
	fmt.Printf("Data root - 2-nd account: %v\n", hrRoot2p1)

	//Test 4.1 test that hashes are different
	assert.NotEqual(t, hrCreated2p1, hrCreated2)

	//Test 4.2 test whether the datatrie roots match/mismatch
	assert.NotEqual(t, hrRoot2, hrRoot2p1)

	//Step 5. Revert 2-nd account modification
	err = adb.RevertToSnapshot(snapshotMod)
	assert.Nil(t, err)
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated2Rev := base64.StdEncoding.EncodeToString(rootHash)

	state2, err = adb.LoadAccount(adr2)
	assert.Nil(t, err)
	rootHash, err = state2.(state.UserAccountHandler).DataTrie().Root()
	assert.Nil(t, err)
	hrRoot2Rev := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - reverted 2-nd account: %v\n", hrCreated2Rev)
	fmt.Printf("Data root - 2-nd account: %v\n", hrRoot2Rev)
	assert.Equal(t, hrCommit, hrCreated2Rev)
	assert.Equal(t, hrRoot2, hrRoot2Rev)
}

func TestAccountsDB_ExecBalanceTxExecution(t *testing.T) {
	t.Parallel()

	adrSrc := integrationTests.CreateRandomAddress()
	adrDest := integrationTests.CreateRandomAddress()

	//Step 1. create accounts objects
	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	adb, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	acntSrc, err := adb.LoadAccount(adrSrc)
	assert.Nil(t, err)
	acntDest, err := adb.LoadAccount(adrDest)
	assert.Nil(t, err)

	//Set a high balance to src's account
	_ = acntSrc.(state.UserAccountHandler).AddToBalance(big.NewInt(1000))
	_ = adb.SaveAccount(acntSrc)

	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrOriginal := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("Original root hash: %s\n", hrOriginal)

	integrationTests.PrintShardAccount(acntSrc.(state.UserAccountHandler), "Source")
	integrationTests.PrintShardAccount(acntDest.(state.UserAccountHandler), "Destination")

	fmt.Println("Executing OK transaction...")
	integrationTests.AdbEmulateBalanceTxSafeExecution(acntSrc.(state.UserAccountHandler), acntDest.(state.UserAccountHandler), adb, big.NewInt(64))

	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrOK := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("After executing an OK tx root hash: %s\n", hrOK)

	integrationTests.PrintShardAccount(acntSrc.(state.UserAccountHandler), "Source")
	integrationTests.PrintShardAccount(acntDest.(state.UserAccountHandler), "Destination")

	fmt.Println("Executing NOK transaction...")
	integrationTests.AdbEmulateBalanceTxSafeExecution(acntSrc.(state.UserAccountHandler), acntDest.(state.UserAccountHandler), adb, big.NewInt(10000))

	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrNok := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("After executing a NOK tx root hash: %s\n", hrNok)

	integrationTests.PrintShardAccount(acntSrc.(state.UserAccountHandler), "Source")
	integrationTests.PrintShardAccount(acntDest.(state.UserAccountHandler), "Destination")

	assert.NotEqual(t, hrOriginal, hrOK)
	assert.Equal(t, hrOK, hrNok)

}

func TestAccountsDB_ExecALotOfBalanceTxOK(t *testing.T) {
	t.Parallel()

	adrSrc := integrationTests.CreateRandomAddress()
	adrDest := integrationTests.CreateRandomAddress()

	//Step 1. create accounts objects
	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	adb, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	acntSrc, err := adb.LoadAccount(adrSrc)
	assert.Nil(t, err)
	acntDest, err := adb.LoadAccount(adrDest)
	assert.Nil(t, err)

	//Set a high balance to src's account
	_ = acntSrc.(state.UserAccountHandler).AddToBalance(big.NewInt(10000000))
	_ = adb.SaveAccount(acntSrc)

	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrOriginal := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("Original root hash: %s\n", hrOriginal)

	for i := 1; i <= 1000; i++ {
		err = integrationTests.AdbEmulateBalanceTxExecution(adb, acntSrc.(state.UserAccountHandler), acntDest.(state.UserAccountHandler), big.NewInt(int64(i)))

		assert.Nil(t, err)
	}

	integrationTests.PrintShardAccount(acntSrc.(state.UserAccountHandler), "Source")
	integrationTests.PrintShardAccount(acntDest.(state.UserAccountHandler), "Destination")
}

func TestAccountsDB_ExecALotOfBalanceTxOKorNOK(t *testing.T) {
	t.Parallel()

	adrSrc := integrationTests.CreateRandomAddress()
	adrDest := integrationTests.CreateRandomAddress()

	//Step 1. create accounts objects
	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	adb, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	acntSrc, err := adb.LoadAccount(adrSrc)
	assert.Nil(t, err)
	acntDest, err := adb.LoadAccount(adrDest)
	assert.Nil(t, err)

	//Set a high balance to src's account
	_ = acntSrc.(state.UserAccountHandler).AddToBalance(big.NewInt(10000000))
	_ = adb.SaveAccount(acntSrc)

	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrOriginal := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("Original root hash: %s\n", hrOriginal)

	st := time.Now()
	for i := 1; i <= 1000; i++ {
		err = integrationTests.AdbEmulateBalanceTxExecution(adb, acntSrc.(state.UserAccountHandler), acntDest.(state.UserAccountHandler), big.NewInt(int64(i)))
		assert.Nil(t, err)

		err = integrationTests.AdbEmulateBalanceTxExecution(adb, acntDest.(state.UserAccountHandler), acntSrc.(state.UserAccountHandler), big.NewInt(int64(1000000)))
		assert.NotNil(t, err)
	}

	fmt.Printf("Done in %v\n", time.Since(st))

	integrationTests.PrintShardAccount(acntSrc.(state.UserAccountHandler), "Source")
	integrationTests.PrintShardAccount(acntDest.(state.UserAccountHandler), "Destination")
}

func BenchmarkCreateOneMillionAccountsWithMockDB(b *testing.B) {
	nrOfAccounts := 1000000
	balance := 1500000
	persist := mock.MockDB{}

	adb, _, tr := createAccounts(nrOfAccounts, balance, persist)
	var rtm runtime.MemStats
	runtime.GC()

	runtime.ReadMemStats(&rtm)
	fmt.Printf("Mem before committing %v accounts with mockDB: go mem - %s, sys mem - %s \n",
		nrOfAccounts,
		core.ConvertBytes(rtm.Alloc),
		core.ConvertBytes(rtm.Sys),
	)

	_, _ = adb.Commit()

	runtime.ReadMemStats(&rtm)
	fmt.Printf("Mem after committing %v accounts with mockDB: go mem - %s, sys mem - %s \n",
		nrOfAccounts,
		core.ConvertBytes(rtm.Alloc),
		core.ConvertBytes(rtm.Sys),
	)

	runtime.GC()
	runtime.ReadMemStats(&rtm)
	fmt.Printf("Mem after committing %v accounts with mockDB and garbage collection : go mem - %s, sys mem - %s \n",
		nrOfAccounts,
		core.ConvertBytes(rtm.Alloc),
		core.ConvertBytes(rtm.Sys),
	)

	_ = tr.String()
}

func BenchmarkCreateOneMillionAccounts(b *testing.B) {
	nrOfAccounts := 1000000
	nrTxs := 15000
	txVal := 100
	balance := nrTxs * txVal
	persist := mock.NewCountingDB()

	adb, addr, _ := createAccounts(nrOfAccounts, balance, persist)
	var rtm runtime.MemStats

	_, _ = adb.Commit()
	runtime.GC()

	runtime.ReadMemStats(&rtm)
	goMemCollapsedTrie := rtm.Alloc

	fmt.Printf("Partially collapsed trie - ")
	createAndExecTxs(b, addr, nrTxs, nrOfAccounts, txVal, adb)

	runtime.ReadMemStats(&rtm)
	goMemAfterTxExec := rtm.Alloc

	fmt.Printf("Extra memory after the exec of %v txs: %s \n",
		nrTxs,
		core.ConvertBytes(goMemAfterTxExec-goMemCollapsedTrie),
	)

	fmt.Println("Total nr. of nodes in trie: ", persist.GetCounter())
	persist.Reset()
	_, _ = adb.Commit()
	fmt.Printf("Nr. of modified nodes after %v txs: %v \n", nrTxs, persist.GetCounter())

	rootHash, err := adb.RootHash()
	assert.Nil(b, err)

	_ = adb.RecreateTrie(rootHash)
	fmt.Printf("Completely collapsed trie - ")
	createAndExecTxs(b, addr, nrTxs, nrOfAccounts, txVal, adb)
}

func createAccounts(
	nrOfAccounts int,
	balance int,
	persist storage.Persister,
) (*state.AccountsDB, [][]byte, data.Trie) {
	cache, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 10, Shards: 1, SizeInBytes: 0})
	store, _ := storageUnit.NewStorageUnit(cache, persist)
	evictionWaitListSize := uint(100)

	ewl, _ := evictionWaitingList.NewEvictionWaitingList(evictionWaitListSize, memorydb.New(), integrationTests.TestMarshalizer)
	trieStorage, _ := trie.NewTrieStorageManager(store, integrationTests.TestMarshalizer, integrationTests.TestHasher, config.DBConfig{}, ewl, config.TrieStorageManagerConfig{})
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(trieStorage, integrationTests.TestMarshalizer, integrationTests.TestHasher, maxTrieLevelInMemory)
	adb, _ := state.NewAccountsDB(tr, integrationTests.TestHasher, integrationTests.TestMarshalizer, factory.NewAccountCreator())

	addr := make([][]byte, nrOfAccounts)
	for i := 0; i < nrOfAccounts; i++ {
		addr[i] = integrationTests.CreateAccount(adb, 0, big.NewInt(int64(balance)))
	}

	return adb, addr, tr
}

func createAndExecTxs(
	b *testing.B,
	addr [][]byte,
	nrTxs int,
	nrOfAccounts int,
	txVal int,
	adb *state.AccountsDB,
) {

	txProcessor := integrationTests.CreateSimpleTxProcessor(adb)
	var totalTime int64 = 0

	for i := 0; i < nrTxs; i++ {
		sender := rand.Intn(nrOfAccounts)
		receiver := rand.Intn(nrOfAccounts)
		for sender == receiver {
			receiver = rand.Intn(nrOfAccounts)
		}

		tx := &transaction2.Transaction{
			Nonce:   1,
			Value:   big.NewInt(int64(txVal)),
			SndAddr: addr[sender],
			RcvAddr: addr[receiver],
		}

		startTime := time.Now()
		_, err := txProcessor.ProcessTransaction(tx)
		duration := time.Since(startTime)
		totalTime += int64(duration)
		assert.Nil(b, err)
	}
	fmt.Printf("Time needed for executing %v transactions: %v \n", nrTxs, time.Duration(totalTime))
}

func BenchmarkTxExecution(b *testing.B) {
	adrSrc := integrationTests.CreateRandomAddress()
	adrDest := integrationTests.CreateRandomAddress()

	//Step 1. create accounts objects
	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	adb, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	acntSrc, err := adb.LoadAccount(adrSrc)
	assert.Nil(b, err)
	acntDest, err := adb.LoadAccount(adrDest)
	assert.Nil(b, err)

	//Set a high balance to src's account
	_ = acntSrc.(state.UserAccountHandler).AddToBalance(big.NewInt(10000000))
	_ = adb.SaveAccount(acntSrc)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		integrationTests.AdbEmulateBalanceTxSafeExecution(acntSrc.(state.UserAccountHandler), acntDest.(state.UserAccountHandler), adb, big.NewInt(1))
	}
}

func TestTrieDbPruning_GetAccountAfterPruning(t *testing.T) {
	t.Parallel()

	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}
	evictionWaitListSize := uint(100)
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(evictionWaitListSize, memorydb.New(), integrationTests.TestMarshalizer)
	trieStorage, _ := trie.NewTrieStorageManager(memorydb.New(), integrationTests.TestMarshalizer, integrationTests.TestHasher, config.DBConfig{}, ewl, generalCfg)
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(trieStorage, integrationTests.TestMarshalizer, integrationTests.TestHasher, maxTrieLevelInMemory)
	adb, _ := state.NewAccountsDB(tr, integrationTests.TestHasher, integrationTests.TestMarshalizer, factory.NewAccountCreator())

	hexPubkeyConverter, _ := pubkeyConverter.NewHexPubkeyConverter(32)
	address1, _ := hexPubkeyConverter.Decode("0000000000000000000000000000000000000000000000000000000000000000")
	address2, _ := hexPubkeyConverter.Decode("0000000000000000000000000000000000000000000000000000000000000001")
	address3, _ := hexPubkeyConverter.Decode("0000000000000000000000000000000000000000000000000000000000000002")

	newDefaultAccount(adb, address1)
	newDefaultAccount(adb, address2)
	account := newDefaultAccount(adb, address3)

	rootHash1, _ := adb.Commit()
	_ = account.(state.UserAccountHandler).AddToBalance(big.NewInt(1))
	_ = adb.SaveAccount(account)
	rootHash2, _ := adb.Commit()
	tr.Prune(rootHash1, data.OldRoot)

	err := adb.RecreateTrie(rootHash2)
	assert.Nil(t, err)
	acc, err := adb.GetExistingAccount(address1)
	assert.NotNil(t, acc)
	assert.Nil(t, err)
}

func newDefaultAccount(adb *state.AccountsDB, address []byte) state.AccountHandler {
	account, _ := adb.LoadAccount(address)
	_ = adb.SaveAccount(account)

	return account
}

func TestAccountsDB_RecreateTrieInvalidatesDataTriesCache(t *testing.T) {
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}
	evictionWaitListSize := uint(100)
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(evictionWaitListSize, memorydb.New(), integrationTests.TestMarshalizer)
	trieStorage, _ := trie.NewTrieStorageManager(memorydb.New(), integrationTests.TestMarshalizer, integrationTests.TestHasher, config.DBConfig{}, ewl, generalCfg)
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(trieStorage, integrationTests.TestMarshalizer, integrationTests.TestHasher, maxTrieLevelInMemory)
	adb, _ := state.NewAccountsDB(tr, integrationTests.TestHasher, integrationTests.TestMarshalizer, factory.NewAccountCreator())

	hexAddressPubkeyConverter, _ := pubkeyConverter.NewHexPubkeyConverter(32)
	address1, _ := hexAddressPubkeyConverter.Decode("0000000000000000000000000000000000000000000000000000000000000000")

	key1 := []byte("ABC")
	key2 := []byte("ABD")
	value1 := []byte("dog")
	value2 := []byte("puppy")

	acc1, _ := adb.LoadAccount(address1)
	state1 := acc1.(state.UserAccountHandler)
	_ = state1.DataTrieTracker().SaveKeyValue(key1, value1)
	_ = state1.DataTrieTracker().SaveKeyValue(key2, value1)
	_ = adb.SaveAccount(state1)
	rootHash, err := adb.Commit()
	require.Nil(t, err)

	acc1, _ = adb.LoadAccount(address1)
	state1 = acc1.(state.UserAccountHandler)
	_ = state1.DataTrieTracker().SaveKeyValue(key1, value2)
	_ = adb.SaveAccount(state1)
	_, err = adb.Commit()
	require.Nil(t, err)

	acc1, _ = adb.LoadAccount(address1)
	state1 = acc1.(state.UserAccountHandler)
	_ = state1.DataTrieTracker().SaveKeyValue(key2, value2)
	_ = adb.SaveAccount(state1)
	err = adb.RevertToSnapshot(0)
	require.Nil(t, err)

	err = adb.RecreateTrie(rootHash)
	require.Nil(t, err)
	acc1, _ = adb.LoadAccount(address1)
	state1 = acc1.(state.UserAccountHandler)

	retrievedVal, _ := state1.DataTrieTracker().RetrieveValue(key1)
	assert.Equal(t, value1, retrievedVal)
}

func TestTrieDbPruning_GetDataTrieTrackerAfterPruning(t *testing.T) {
	t.Parallel()

	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}
	evictionWaitListSize := uint(100)
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(evictionWaitListSize, memorydb.New(), integrationTests.TestMarshalizer)
	trieStorage, _ := trie.NewTrieStorageManager(memorydb.New(), integrationTests.TestMarshalizer, integrationTests.TestHasher, config.DBConfig{}, ewl, generalCfg)
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(trieStorage, integrationTests.TestMarshalizer, integrationTests.TestHasher, maxTrieLevelInMemory)
	adb, _ := state.NewAccountsDB(tr, integrationTests.TestHasher, integrationTests.TestMarshalizer, factory.NewAccountCreator())

	hexAddressPubkeyConverter, _ := pubkeyConverter.NewHexPubkeyConverter(32)
	address1, _ := hexAddressPubkeyConverter.Decode("0000000000000000000000000000000000000000000000000000000000000000")
	address2, _ := hexAddressPubkeyConverter.Decode("0000000000000000000000000000000000000000000000000000000000000001")

	key1 := []byte("ABC")
	key2 := []byte("ABD")
	value1 := []byte("dog")
	value2 := []byte("puppy")

	acc1, _ := adb.LoadAccount(address1)
	state1 := acc1.(state.UserAccountHandler)
	_ = state1.DataTrieTracker().SaveKeyValue(key1, value1)
	_ = state1.DataTrieTracker().SaveKeyValue(key2, value1)
	_ = adb.SaveAccount(state1)

	acc2, _ := adb.LoadAccount(address2)
	state2 := acc2.(state.UserAccountHandler)
	_ = state2.DataTrieTracker().SaveKeyValue(key1, value1)
	_ = state2.DataTrieTracker().SaveKeyValue(key2, value1)
	_ = adb.SaveAccount(state2)

	oldRootHash, _ := adb.Commit()

	acc2, _ = adb.LoadAccount(address2)
	state2 = acc2.(state.UserAccountHandler)
	_ = state2.DataTrieTracker().SaveKeyValue(key1, value2)
	_ = adb.SaveAccount(state2)

	newRootHash, _ := adb.Commit()
	tr.Prune(oldRootHash, data.OldRoot)

	err := adb.RecreateTrie(newRootHash)
	assert.Nil(t, err)
	acc, err := adb.GetExistingAccount(address1)
	assert.NotNil(t, acc)
	assert.Nil(t, err)

	collapseTrie(state1, t)
	collapseTrie(state2, t)

	val, err := state1.DataTrieTracker().RetrieveValue(key1)
	assert.Nil(t, err)
	assert.Equal(t, value1, val)

	val, err = state2.DataTrieTracker().RetrieveValue(key2)
	assert.Nil(t, err)
	assert.Equal(t, value1, val)
}

func collapseTrie(state state.UserAccountHandler, t *testing.T) {
	stateRootHash := state.GetRootHash()
	stateTrie := state.DataTrieTracker().DataTrie()
	stateNewTrie, _ := stateTrie.Recreate(stateRootHash)
	assert.NotNil(t, stateNewTrie)

	state.DataTrieTracker().SetDataTrie(stateNewTrie)
}

func TestRollbackBlockAndCheckThatPruningIsCancelledOnAccountsTrie(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodesPerShard := 1
	numNodesMeta := 1

	nodes, advertiser, idxProposers := integrationTests.SetupSyncNodesOneShardAndMeta(numNodesPerShard, numNodesMeta)
	defer integrationTests.CloseProcessorNodes(nodes, advertiser)

	integrationTests.StartP2PBootstrapOnProcessorNodes(nodes)
	integrationTests.StartSyncingBlocks(nodes)

	round := uint64(0)
	nonce := uint64(0)

	valMinting := big.NewInt(1000000000)
	valToTransferPerTx := big.NewInt(2)

	fmt.Println("Generating private keys for senders and receivers...")
	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(1), 0)
	nrTxs := 20

	//sender shard keys, receivers  keys
	sendersPrivateKeys := make([]crypto.PrivateKey, nrTxs)
	receiversPublicKeys := make(map[uint32][]crypto.PublicKey)
	for i := 0; i < nrTxs; i++ {
		sendersPrivateKeys[i], _, _ = integrationTests.GenerateSkAndPkInShard(generateCoordinator, 0)
		_, pk, _ := integrationTests.GenerateSkAndPkInShard(generateCoordinator, 0)
		receiversPublicKeys[0] = append(receiversPublicKeys[0], pk)
	}

	fmt.Println("Minting sender addresses...")
	integrationTests.CreateMintingForSenders(nodes, 0, sendersPrivateKeys, valMinting)

	shardNode := nodes[0]

	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)

	rootHashOfFirstBlock, _ := shardNode.AccntState.RootHash()

	assert.Equal(t, uint64(1), nodes[0].BlockChain.GetCurrentBlockHeader().GetNonce())
	assert.Equal(t, uint64(1), nodes[1].BlockChain.GetCurrentBlockHeader().GetNonce())

	fmt.Println("Generating transactions...")
	integrationTests.GenerateAndDisseminateTxs(
		shardNode,
		sendersPrivateKeys,
		receiversPublicKeys,
		valToTransferPerTx,
		1000,
		1000,
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)
	fmt.Println("Delaying for disseminating transactions...")
	time.Sleep(time.Second * 5)

	round, _ = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
	time.Sleep(time.Second * 5)

	rootHashOfRollbackedBlock, _ := shardNode.AccntState.RootHash()

	assert.Equal(t, uint64(2), nodes[0].BlockChain.GetCurrentBlockHeader().GetNonce())
	assert.Equal(t, uint64(2), nodes[1].BlockChain.GetCurrentBlockHeader().GetNonce())

	shardIdToRollbackLastBlock := uint32(0)
	integrationTests.ForkChoiceOneBlock(nodes, shardIdToRollbackLastBlock)
	integrationTests.ResetHighestProbableNonce(nodes, shardIdToRollbackLastBlock, 1)
	integrationTests.EmptyDataPools(nodes, shardIdToRollbackLastBlock)

	assert.Equal(t, uint64(1), nodes[0].BlockChain.GetCurrentBlockHeader().GetNonce())
	assert.Equal(t, uint64(2), nodes[1].BlockChain.GetCurrentBlockHeader().GetNonce())

	rootHash, err := shardNode.AccntState.RootHash()
	assert.Nil(t, err)

	if !bytes.Equal(rootHash, rootHashOfRollbackedBlock) {
		time.Sleep(time.Second * 3)
		err = shardNode.AccntState.RecreateTrie(rootHashOfRollbackedBlock)
		assert.True(t, errors.Is(err, trie.ErrHashNotFound))
	}

	nonces := []*uint64{new(uint64), new(uint64)}
	atomic.AddUint64(nonces[0], 2)
	atomic.AddUint64(nonces[1], 3)

	numOfRounds := 2
	integrationTests.ProposeBlocks(
		nodes,
		&round,
		idxProposers,
		nonces,
		numOfRounds,
	)
	time.Sleep(time.Second * 5)

	err = shardNode.AccntState.RecreateTrie(rootHashOfFirstBlock)
	assert.Nil(t, err)
	assert.Equal(t, uint64(3), nodes[0].BlockChain.GetCurrentBlockHeader().GetNonce())
	assert.Equal(t, uint64(4), nodes[1].BlockChain.GetCurrentBlockHeader().GetNonce())
}

func TestRollbackBlockWithSameRootHashAsPreviousAndCheckThatPruningIsNotDone(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodesPerShard := 1
	numNodesMeta := 1

	nodes, advertiser, idxProposers := integrationTests.SetupSyncNodesOneShardAndMeta(numNodesPerShard, numNodesMeta)
	defer integrationTests.CloseProcessorNodes(nodes, advertiser)

	integrationTests.StartP2PBootstrapOnProcessorNodes(nodes)
	integrationTests.StartSyncingBlocks(nodes)

	round := uint64(0)
	nonce := uint64(0)

	valMinting := big.NewInt(1000000000)

	fmt.Println("Generating private keys for senders and receivers...")
	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(1), 0)
	nrTxs := 20

	//sender shard keys, receivers  keys
	sendersPrivateKeys := make([]crypto.PrivateKey, nrTxs)
	for i := 0; i < nrTxs; i++ {
		sendersPrivateKeys[i], _, _ = integrationTests.GenerateSkAndPkInShard(generateCoordinator, 0)
	}

	fmt.Println("Minting sender addresses...")
	integrationTests.CreateMintingForSenders(nodes, 0, sendersPrivateKeys, valMinting)

	shardNode := nodes[0]

	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)

	rootHashOfFirstBlock, _ := shardNode.AccntState.RootHash()

	assert.Equal(t, uint64(1), nodes[0].BlockChain.GetCurrentBlockHeader().GetNonce())
	assert.Equal(t, uint64(1), nodes[1].BlockChain.GetCurrentBlockHeader().GetNonce())

	_, _ = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
	time.Sleep(time.Second * 5)

	assert.Equal(t, uint64(2), nodes[0].BlockChain.GetCurrentBlockHeader().GetNonce())
	assert.Equal(t, uint64(2), nodes[1].BlockChain.GetCurrentBlockHeader().GetNonce())

	shardIdToRollbackLastBlock := uint32(0)
	integrationTests.ForkChoiceOneBlock(nodes, shardIdToRollbackLastBlock)
	integrationTests.ResetHighestProbableNonce(nodes, shardIdToRollbackLastBlock, 1)
	integrationTests.EmptyDataPools(nodes, shardIdToRollbackLastBlock)

	assert.Equal(t, uint64(1), nodes[0].BlockChain.GetCurrentBlockHeader().GetNonce())
	assert.Equal(t, uint64(2), nodes[1].BlockChain.GetCurrentBlockHeader().GetNonce())

	err := shardNode.AccntState.RecreateTrie(rootHashOfFirstBlock)
	assert.Nil(t, err)
}

func TestTriePruningWhenBlockIsFinal(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	fmt.Println("Setup nodes...")
	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	senderShard := uint32(0)
	round := uint64(0)
	nonce := uint64(0)

	valMinting := big.NewInt(1000000000)
	valToTransferPerTx := big.NewInt(2)

	nodes, advertiser, idxProposers := integrationTests.SetupSyncNodesOneShardAndMeta(nodesPerShard, numMetachainNodes)
	integrationTests.DisplayAndStartNodes(nodes)

	defer integrationTests.CloseProcessorNodes(nodes, advertiser)

	fmt.Println("Generating private keys for senders and receivers...")
	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 0)
	nrTxs := 20

	//sender shard keys, receivers  keys
	sendersPrivateKeys := make([]crypto.PrivateKey, nrTxs)
	receiversPublicKeys := make(map[uint32][]crypto.PublicKey)
	for i := 0; i < nrTxs; i++ {
		sendersPrivateKeys[i], _, _ = integrationTests.GenerateSkAndPkInShard(generateCoordinator, senderShard)
		_, pk, _ := integrationTests.GenerateSkAndPkInShard(generateCoordinator, senderShard)
		receiversPublicKeys[senderShard] = append(receiversPublicKeys[senderShard], pk)
	}

	fmt.Println("Minting sender addresses...")
	integrationTests.CreateMintingForSenders(nodes, senderShard, sendersPrivateKeys, valMinting)

	shardNode := nodes[0]

	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)

	assert.Equal(t, uint64(1), nodes[0].BlockChain.GetCurrentBlockHeader().GetNonce())
	assert.Equal(t, uint64(1), nodes[1].BlockChain.GetCurrentBlockHeader().GetNonce())

	rootHashOfFirstBlock, _ := shardNode.AccntState.RootHash()

	fmt.Println("Generating transactions...")
	integrationTests.GenerateAndDisseminateTxs(
		shardNode,
		sendersPrivateKeys,
		receiversPublicKeys,
		valToTransferPerTx,
		1000,
		1000,
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)
	fmt.Println("Delaying for disseminating transactions...")
	time.Sleep(time.Second * 5)

	roundsToWait := 6
	for i := 0; i < roundsToWait; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
	}

	assert.Equal(t, uint64(7), nodes[0].BlockChain.GetCurrentBlockHeader().GetNonce())
	assert.Equal(t, uint64(7), nodes[1].BlockChain.GetCurrentBlockHeader().GetNonce())

	err := shardNode.AccntState.RecreateTrie(rootHashOfFirstBlock)
	assert.True(t, errors.Is(err, trie.ErrHashNotFound))
}

func TestSnapshotOnEpochChange(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1
	stateCheckpointModulus := uint(3)

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap(0)

	nodes := integrationTests.CreateNodesWithCustomStateCheckpointModulus(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
		stateCheckpointModulus,
	)

	roundsPerEpoch := uint64(5)
	for _, node := range nodes {
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	sendValue := big.NewInt(5)
	receiverAddress := []byte("12345678901234567890123456789012")
	initialVal := big.NewInt(10000000000)

	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	time.Sleep(integrationTests.StepDelay)

	checkpointsRootHashes := make(map[int][][]byte)
	snapshotsRootHashes := make(map[uint32][][]byte)
	prunedRootHashes := make(map[int][][]byte)

	numShardNodes := numOfShards * nodesPerShard
	numRounds := uint32(9)
	for i := uint64(0); i < uint64(numRounds); i++ {

		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)

		for _, node := range nodes {
			integrationTests.CreateAndSendTransaction(node, nodes, sendValue, receiverAddress, "", integrationTests.AdditionalGasLimit)
		}
		time.Sleep(integrationTests.StepDelay)

		collectSnapshotAndCheckpointHashes(
			nodes,
			numShardNodes,
			checkpointsRootHashes,
			snapshotsRootHashes,
			prunedRootHashes,
			uint64(stateCheckpointModulus),
		)
	}

	numDelayRounds := uint32(6)
	for i := uint64(0); i < uint64(numDelayRounds); i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		time.Sleep(integrationTests.StepDelay)
	}

	for i := 0; i < numOfShards*nodesPerShard; i++ {
		shId := nodes[i].ShardCoordinator.SelfId()
		testNodeStateCheckpointSnapshotAndPruning(t, nodes[i], checkpointsRootHashes[i], snapshotsRootHashes[shId], prunedRootHashes[i])
	}
}

func collectSnapshotAndCheckpointHashes(
	nodes []*integrationTests.TestProcessorNode,
	numShardNodes int,
	checkpointsRootHashes map[int][][]byte,
	snapshotsRootHashes map[uint32][][]byte,
	prunedRootHashes map[int][][]byte,
	stateCheckpointModulus uint64,
) {
	for j := 0; j < numShardNodes; j++ {
		currentBlockHeader := nodes[j].BlockChain.GetCurrentBlockHeader()
		if currentBlockHeader.IsStartOfEpochBlock() {
			continue
		}

		checkpointRound := currentBlockHeader.GetNonce()%stateCheckpointModulus == 0
		if checkpointRound {
			checkpointsRootHashes[j] = append(checkpointsRootHashes[j], currentBlockHeader.GetRootHash())
			continue
		}

		prunedRootHashes[j] = append(prunedRootHashes[j], currentBlockHeader.GetRootHash())
	}

	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != core.MetachainShardId {
			continue
		}

		currentBlockHeader := node.BlockChain.GetCurrentBlockHeader()
		if !currentBlockHeader.IsStartOfEpochBlock() {
			continue
		}

		metaHdr := currentBlockHeader.(*block.MetaBlock)
		snapshotsRootHashes[core.MetachainShardId] = append(snapshotsRootHashes[core.MetachainShardId], metaHdr.GetRootHash())
		for _, epochStartData := range metaHdr.EpochStart.LastFinalizedHeaders {
			snapshotsRootHashes[epochStartData.ShardID] = append(snapshotsRootHashes[epochStartData.ShardID], epochStartData.RootHash)
		}
	}
}

func testNodeStateCheckpointSnapshotAndPruning(
	t *testing.T,
	node *integrationTests.TestProcessorNode,
	checkpointsRootHashes [][]byte,
	snapshotsRootHashes [][]byte,
	prunedRootHashes [][]byte,
) {

	stateTrie := node.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	assert.Equal(t, 3, len(checkpointsRootHashes))
	for i := range checkpointsRootHashes {
		tr, err := stateTrie.Recreate(checkpointsRootHashes[i])
		assert.Nil(t, err)
		assert.NotNil(t, tr)
	}

	assert.Equal(t, 1, len(snapshotsRootHashes))
	for i := range snapshotsRootHashes {
		tr, err := stateTrie.Recreate(snapshotsRootHashes[i])
		assert.Nil(t, err)
		assert.NotNil(t, tr)
	}

	assert.Equal(t, 5, len(prunedRootHashes))
	for i := range prunedRootHashes {
		tr, err := stateTrie.Recreate(prunedRootHashes[i])
		if err == nil {
			fmt.Println(hex.EncodeToString(prunedRootHashes[i]))
		}
		assert.Nil(t, tr)
		assert.NotNil(t, err)
	}
}

func TestContinuouslyAccountCodeChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1
	senderShard := uint32(0)
	round := uint64(0)
	nonce := uint64(0)
	valMinting := big.NewInt(1000000000)

	nodes, advertiser, idxProposers := integrationTests.SetupSyncNodesOneShardAndMeta(nodesPerShard, numMetachainNodes)
	integrationTests.DisplayAndStartNodes(nodes)

	defer integrationTests.CloseProcessorNodes(nodes, advertiser)

	fmt.Println("Generating private keys for senders...")
	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 0)
	nrAccounts := 20

	sendersPrivateKeys := make([]crypto.PrivateKey, nrAccounts)
	for i := 0; i < nrAccounts; i++ {
		sendersPrivateKeys[i], _, _ = integrationTests.GenerateSkAndPkInShard(generateCoordinator, senderShard)
	}

	accounts := make([][]byte, len(sendersPrivateKeys))
	for i := range sendersPrivateKeys {
		accounts[i], _ = sendersPrivateKeys[i].GeneratePublic().ToByteArray()
	}

	fmt.Println("Minting sender addresses...")
	integrationTests.CreateMintingForSenders(nodes, senderShard, sendersPrivateKeys, valMinting)

	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)

	numCodes := 10
	codeMap := getCodeMap(numCodes)
	codeArray := make([][]byte, 0)
	for code := range codeMap {
		codeArray = append(codeArray, []byte(code))
	}

	maxCodeUpdates := 10
	maxCodeDeletes := 10

	shardNode := nodes[0]
	roundsToWait := 50
	for i := 0; i < roundsToWait; i++ {
		numCodeUpdates := rand.Intn(maxCodeUpdates)
		numCodeDeletes := rand.Intn(maxCodeDeletes)

		for j := 0; j < numCodeUpdates; j++ {
			accountIndex := rand.Intn(nrAccounts)
			account, _ := shardNode.AccntState.LoadAccount(accounts[accountIndex])

			updateCode(t, shardNode.AccntState, codeArray, codeMap, account, numCodes)
		}

		for j := 0; j < numCodeDeletes; j++ {
			accountIndex := rand.Intn(nrAccounts)
			account, _ := shardNode.AccntState.LoadAccount(accounts[accountIndex])

			removeCode(t, shardNode.AccntState, codeMap, account)
		}
		_, _ = shardNode.AccntState.Commit()

		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)

		for j := range codeArray {
			fmt.Printf("%v - %v \n", codeArray[j], codeMap[string(codeArray[j])])
		}

		checkCodeConsistency(t, shardNode, codeMap)
	}
}

func shouldRevert() bool {
	return (rand.Intn(10) % 3) == 0
}

func getCodeMap(numCodes int) map[string]int {
	codeMap := make(map[string]int)
	for i := 0; i < numCodes; i++ {
		code := "code" + strconv.Itoa(i)
		codeMap[code] = 0
	}

	return codeMap
}

func updateCode(
	t *testing.T,
	AccntState state.AccountsAdapter,
	codeArray [][]byte,
	codeMap map[string]int,
	account state.AccountHandler,
	numCodes int,
) {
	snapshot := AccntState.JournalLen()
	codeIndex := rand.Intn(numCodes)
	code := codeArray[codeIndex]

	oldCode := account.(state.UserAccountHandler).GetCode()
	account.(state.UserAccountHandler).SetCode(code)
	_ = AccntState.SaveAccount(account)

	if shouldRevert() && snapshot != 0 {
		err := AccntState.RevertToSnapshot(snapshot)
		assert.Nil(t, err)
		fmt.Printf("updated code %v to account %v and reverted\n", code, hex.EncodeToString(account.AddressBytes()))
		return
	}

	codeMap[string(code)]++
	if len(oldCode) != 0 {
		codeMap[string(oldCode)]--
	}

	fmt.Printf("updated code %v to account %v \n", code, hex.EncodeToString(account.AddressBytes()))
}

func removeCode(
	t *testing.T,
	AccntState state.AccountsAdapter,
	codeMap map[string]int,
	account state.AccountHandler,
) {
	snapshot := AccntState.JournalLen()
	code := account.(state.UserAccountHandler).GetCode()
	account.(state.UserAccountHandler).SetCode(nil)
	_ = AccntState.SaveAccount(account)

	if shouldRevert() && snapshot != 0 {
		err := AccntState.RevertToSnapshot(snapshot)
		assert.Nil(t, err)
		fmt.Printf("removed old code %v from account %v and reverted\n", code, hex.EncodeToString(account.AddressBytes()))
		return
	}

	if len(code) != 0 {
		codeMap[string(code)]--
	}

	fmt.Printf("removed old code %v from account %v \n", code, hex.EncodeToString(account.AddressBytes()))
}
func checkCodeConsistency(
	t *testing.T,
	shardNode *integrationTests.TestProcessorNode,
	codeMap map[string]int,
) {
	for code := range codeMap {
		codeHash := integrationTests.TestHasher.Compute(code)
		tr := shardNode.TrieContainer.Get([]byte(factory2.UserAccountTrie))

		if codeMap[code] != 0 {
			val, err := tr.Get(codeHash)
			assert.Nil(t, err)
			assert.NotNil(t, val)

			var codeEntry state.CodeEntry
			err = integrationTests.TestMarshalizer.Unmarshal(&codeEntry, val)
			assert.Nil(t, err)

			assert.Equal(t, uint32(codeMap[code]), codeEntry.NumReferences)
		}
	}
}

func TestAccountRemoval(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1
	senderShard := uint32(0)
	round := uint64(0)
	nonce := uint64(0)
	valMinting := big.NewInt(1000000000)

	nodes, advertiser, idxProposers := integrationTests.SetupSyncNodesOneShardAndMeta(nodesPerShard, numMetachainNodes)
	integrationTests.DisplayAndStartNodes(nodes)

	defer integrationTests.CloseProcessorNodes(nodes, advertiser)

	fmt.Println("Generating private keys for senders...")
	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 0)
	nrAccounts := 10000

	sendersPrivateKeys := make([]crypto.PrivateKey, nrAccounts)
	for i := 0; i < nrAccounts; i++ {
		sendersPrivateKeys[i], _, _ = integrationTests.GenerateSkAndPkInShard(generateCoordinator, senderShard)
	}

	accounts := make([][]byte, len(sendersPrivateKeys))
	for i := range sendersPrivateKeys {
		accounts[i], _ = sendersPrivateKeys[i].GeneratePublic().ToByteArray()
	}

	fmt.Println("Minting sender addresses...")
	integrationTests.CreateMintingForSenders(nodes, senderShard, sendersPrivateKeys, valMinting)

	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)

	shardNode := nodes[0]

	dataTriesRootHashes, codeMap := generateAccounts(shardNode, accounts)

	_, _ = shardNode.AccntState.Commit()
	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)

	numAccountsToRemove := 2
	roundsToWait := 50

	removedAccounts := make(map[int]struct{})
	for i := 0; i < roundsToWait; i++ {
		for j := 0; j < numAccountsToRemove; j++ {
			accountIndex := rand.Intn(nrAccounts)
			removedAccounts[accountIndex] = struct{}{}
			account, err := shardNode.AccntState.GetExistingAccount(accounts[accountIndex])
			if err != nil {
				continue
			}
			code := account.(state.UserAccountHandler).GetCode()

			_ = shardNode.AccntState.RemoveAccount(account.AddressBytes())

			codeMap[string(code)]--
		}

		_, _ = shardNode.AccntState.Commit()
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		checkCodeConsistency(t, shardNode, codeMap)
	}

	delayRounds := 5
	for i := 0; i < delayRounds; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
	}

	checkDataTrieConsistency(t, shardNode.AccntState, removedAccounts, dataTriesRootHashes)
}

func generateAccounts(
	shardNode *integrationTests.TestProcessorNode,
	accounts [][]byte,
) ([][]byte, map[string]int) {
	numCodes := 100
	codeMap := getCodeMap(numCodes)
	codeArray := make([][]byte, 0)
	for code := range codeMap {
		codeArray = append(codeArray, []byte(code))
	}

	dataTriesRootHashes := make([][]byte, 0)
	dataTrieSize := 5
	for i := 0; i < len(accounts); i++ {
		account, _ := shardNode.AccntState.LoadAccount(accounts[i])

		code := codeArray[rand.Intn(numCodes)]
		account.(state.UserAccountHandler).SetCode(code)
		codeMap[string(code)]++

		for j := 0; j < dataTrieSize; j++ {
			_ = account.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue(getDataTrieEntry())
		}

		_ = shardNode.AccntState.SaveAccount(account)

		rootHash := account.(state.UserAccountHandler).GetRootHash()
		dataTriesRootHashes = append(dataTriesRootHashes, rootHash)
	}

	return dataTriesRootHashes, codeMap
}

func getDataTrieEntry() ([]byte, []byte) {
	index := strconv.Itoa(rand.Intn(math.MaxInt32))
	key := []byte("key" + index)
	value := []byte("value" + index)

	return key, value
}

func checkDataTrieConsistency(
	t *testing.T,
	adb state.AccountsAdapter,
	removedAccounts map[int]struct{},
	dataTriesRootHashes [][]byte,
) {
	for i, rootHash := range dataTriesRootHashes {
		_, ok := removedAccounts[i]
		if ok {
			err := adb.RecreateTrie(rootHash)
			assert.NotNil(t, err)
		} else {
			err := adb.RecreateTrie(rootHash)
			assert.Nil(t, err)
		}
	}
}
