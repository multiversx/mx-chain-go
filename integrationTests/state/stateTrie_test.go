package state

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/big"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	transaction2 "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

func TestAccountsDB_RetrieveDataWithSomeValuesShouldWork(t *testing.T) {
	//test simulates creation of a new account, data trie retrieval,
	//adding a (key, value) pair in that data trie, committing changes
	//and then reloading the data trie based on the root hash generated before
	t.Parallel()

	_, account, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()

	account.DataTrieTracker().SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	account.DataTrieTracker().SaveKeyValue([]byte{68, 69, 70}, []byte{35, 36, 37})

	err := adb.SaveDataTrie(account)
	assert.Nil(t, err)

	_, err = adb.Commit()
	assert.Nil(t, err)

	recoveredAccount, err := adb.GetAccountWithJournal(account.AddressContainer())
	assert.Nil(t, err)

	//verify data
	dataRecovered, err := recoveredAccount.DataTrieTracker().RetrieveValue([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{32, 33, 34}, dataRecovered)

	dataRecovered, err = recoveredAccount.DataTrieTracker().RetrieveValue([]byte{68, 69, 70})
	assert.Nil(t, err)
	assert.Equal(t, []byte{35, 36, 37}, dataRecovered)
}

func TestAccountsDB_PutCodeWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	_, account, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()

	err := adb.PutCode(account, []byte("Smart contract code"))
	assert.Nil(t, err)
	assert.NotNil(t, account.GetCodeHash())
	assert.Equal(t, []byte("Smart contract code"), account.GetCode())

	fmt.Printf("SC code is at address: %v\n", account.GetCodeHash())

	recoveredAccount, err := adb.GetAccountWithJournal(account.AddressContainer())
	assert.Nil(t, err)

	assert.Equal(t, account.GetCode(), recoveredAccount.GetCode())
	assert.Equal(t, account.GetCodeHash(), recoveredAccount.GetCodeHash())
}

func TestAccountsDB_SaveDataNoDirtyShouldWork(t *testing.T) {
	t.Parallel()

	_, account, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()

	err := adb.SaveDataTrie(account)
	assert.Nil(t, err)
	assert.Equal(t, 0, adb.JournalLen())
}

func TestAccountsDB_HasAccountNotFoundShouldRetFalse(t *testing.T) {
	t.Parallel()

	adr, _, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()

	//should return false
	val, err := adb.HasAccount(adr)
	assert.Nil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccountFoundShouldRetTrue(t *testing.T) {
	t.Parallel()

	adr, _, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()
	_, err := adb.GetAccountWithJournal(adr)
	assert.Nil(t, err)

	//should return true
	val, err := adb.HasAccount(adr)
	assert.Nil(t, err)
	assert.True(t, val)
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
	account := accountHandler.(*state.Account)
	err := account.SetBalanceWithJournal(balance)
	assert.Nil(t, err)

	err = adb.SaveAccount(account)
	assert.Nil(t, err)

	accountHandlerRecovered, err := adb.GetAccountWithJournal(adr)
	assert.Nil(t, err)
	accountRecovered := accountHandlerRecovered.(*state.Account)
	assert.NotNil(t, accountRecovered)
	assert.Equal(t, accountRecovered.Balance, balance)
}

func TestAccountsDB_GetJournalizedAccountReturnNotFoundAccntShouldWork(t *testing.T) {
	//test when the account does not exists
	t.Parallel()

	adr, _, adb := integrationTests.GenerateAddressJournalAccountAccountsDB()

	//same address of the unsaved account
	accountHandlerRecovered, err := adb.GetAccountWithJournal(adr)
	assert.Nil(t, err)
	accountRecovered := accountHandlerRecovered.(*state.Account)
	assert.NotNil(t, accountRecovered)
	assert.Equal(t, accountRecovered.Balance, big.NewInt(0))
}

func TestAccountsDB_GetExistingAccountConcurrentlyShouldWork(t *testing.T) {
	t.Parallel()

	adb, _, _ := integrationTests.CreateAccountsDB(nil)

	wg := sync.WaitGroup{}
	wg.Add(2000)

	addresses := make([]state.AddressContainer, 0)

	//generating 2000 different addresses
	for len(addresses) < 2000 {
		addr := integrationTests.CreateDummyAddress()

		found := false
		for i := 0; i < len(addresses); i++ {
			if bytes.Equal(addresses[i].Bytes(), addr.Bytes()) {
				found = true
				break
			}
		}

		if !found {
			addresses = append(addresses, addr)
		}
	}

	for i := 0; i < 1000; i++ {
		go func(idx int) {
			accnt, err := adb.GetExistingAccount(addresses[idx*2])

			assert.Equal(t, state.ErrAccNotFound, err)
			assert.Nil(t, accnt)

			wg.Done()
		}(i)

		go func(idx int) {
			accnt, err := adb.GetAccountWithJournal(addresses[idx*2+1])

			assert.Nil(t, err)
			assert.NotNil(t, accnt)

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
	adr2 := integrationTests.CreateDummyAddress()

	//first account has the balance of 40
	balance1 := big.NewInt(40)
	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	err = state1.(*state.Account).SetBalanceWithJournal(balance1)
	assert.Nil(t, err)

	//second account has the balance of 50 and some data
	balance2 := big.NewInt(50)
	state2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)

	err = state2.(*state.Account).SetBalanceWithJournal(balance2)
	assert.Nil(t, err)
	key := []byte{65, 66, 67}
	val := []byte{32, 33, 34}
	state2.DataTrieTracker().SaveKeyValue(key, val)
	err = adb.SaveDataTrie(state2)

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
	newState1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	assert.Equal(t, newState1.(*state.Account).Balance, balance1)

	//checking state2
	newState2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)
	assert.Equal(t, newState2.(*state.Account).Balance, balance2)
	assert.NotNil(t, newState2.(*state.Account).RootHash)
	valRecovered, err := newState2.DataTrieTracker().RetrieveValue(key)
	assert.Nil(t, err)
	assert.Equal(t, val, valRecovered)
}

func TestTrieDB_RecreateFromStorageShouldWork(t *testing.T) {
	hasher := integrationTests.TestHasher
	store := integrationTests.CreateMemUnit()
	tr1, _ := trie.NewTrie(store, integrationTests.TestMarshalizer, hasher)

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

	adb, _, mu := integrationTests.CreateAccountsDB(nil)
	adr1 := integrationTests.CreateDummyAddress()
	adr2 := integrationTests.CreateDummyAddress()

	//first account has the balance of 40
	balance1 := big.NewInt(40)
	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	err = state1.(*state.Account).SetBalanceWithJournal(balance1)
	assert.Nil(t, err)

	//second account has the balance of 50 and some data
	balance2 := big.NewInt(50)
	state2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)

	err = state2.(*state.Account).SetBalanceWithJournal(balance2)
	assert.Nil(t, err)
	key := []byte{65, 66, 67}
	val := []byte{32, 33, 34}
	state2.DataTrieTracker().SaveKeyValue(key, val)
	err = adb.SaveDataTrie(state2)

	//states are now prepared, committing

	h, err := adb.Commit()
	assert.Nil(t, err)
	fmt.Printf("Result hash: %v\n", base64.StdEncoding.EncodeToString(h))

	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	fmt.Printf("Data committed! Root: %v\n", base64.StdEncoding.EncodeToString(rootHash))

	tr, _ := trie.NewTrie(mu, integrationTests.TestMarshalizer, integrationTests.TestHasher)
	adb, _ = state.NewAccountsDB(tr, integrationTests.TestHasher, integrationTests.TestMarshalizer, factory.NewAccountCreator())

	//reloading a new trie to test if data is inside
	err = adb.RecreateTrie(h)
	assert.Nil(t, err)

	//checking state1
	newState1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	assert.Equal(t, newState1.(*state.Account).Balance, balance1)

	//checking state2
	newState2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)
	assert.Equal(t, newState2.(*state.Account).Balance, balance2)
	assert.NotNil(t, newState2.(*state.Account).RootHash)
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

	adb, _, _ := integrationTests.CreateAccountsDB(nil)

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

	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - created account: %v\n", hrCreated)

	err = state1.(*state.Account).SetBalanceWithJournal(big.NewInt(40))
	assert.Nil(t, err)
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

	err = state1.(*state.Account).SetBalanceWithJournal(big.NewInt(0))
	assert.Nil(t, err)

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

	adr1 := integrationTests.CreateDummyAddress()
	adr2 := integrationTests.CreateDummyAddress()

	//Step 1. create accounts objects
	adb, _, _ := integrationTests.CreateAccountsDB(nil)
	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrEmpty := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated1 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)
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
	err = state1.(*state.Account).SetNonceWithJournal(40)
	assert.Nil(t, err)
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrWithNonce1 := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - account with nonce 40: %v\n", hrWithNonce1)

	err = state2.(*state.Account).SetNonceWithJournal(50)
	assert.Nil(t, err)
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

	adr1 := integrationTests.CreateDummyAddress()
	adr2 := integrationTests.CreateDummyAddress()

	//Step 1. create accounts objects
	adb, _, _ := integrationTests.CreateAccountsDB(nil)
	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrEmpty := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated1 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)
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
	err = state1.(*state.Account).SetBalanceWithJournal(big.NewInt(40))
	assert.Nil(t, err)
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrWithBalance1 := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - account with balance 40: %v\n", hrWithBalance1)

	err = state2.(*state.Account).SetBalanceWithJournal(big.NewInt(50))
	assert.Nil(t, err)
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

	adr1 := integrationTests.CreateDummyAddress()
	adr2 := integrationTests.CreateDummyAddress()

	//Step 1. create accounts objects
	adb, _, _ := integrationTests.CreateAccountsDB(nil)
	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrEmpty := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	err = adb.PutCode(state1, []byte{65, 66, 67})
	assert.Nil(t, err)
	snapshotCreated1 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated1 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)
	err = adb.PutCode(state2, []byte{65, 66, 67})
	assert.Nil(t, err)
	snapshotCreated2 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated2 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Step 3. Revert second account
	err = adb.RevertToSnapshot(snapshotCreated1)
	assert.Nil(t, err)

	//Test 3.1. current root hash shall match created root hash hrCreated1
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCrt := base64.StdEncoding.EncodeToString(rootHash)
	assert.Equal(t, hrCreated1, hrCrt)
	fmt.Printf("State root - reverted last account: %v\n", hrCrt)

	//Step 4. Revert first account
	err = adb.RevertToSnapshot(0)
	assert.Nil(t, err)

	//Test 4.1. current root hash shall match empty root hash
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCrt = base64.StdEncoding.EncodeToString(rootHash)
	assert.Equal(t, hrEmpty, hrCrt)
	fmt.Printf("State root - reverted first account: %v\n", hrCrt)
}

func TestAccountsDB_RevertDataStepByStepAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	//adr1 puts data inside trie. adr2 puts the same data
	//revert should work

	adr1 := integrationTests.CreateDummyAddress()
	adr2 := integrationTests.CreateDummyAddress()

	//Step 1. create accounts objects
	adb, _, _ := integrationTests.CreateAccountsDB(nil)
	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrEmpty := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	state1.DataTrieTracker().SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	err = adb.SaveDataTrie(state1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated1 := base64.StdEncoding.EncodeToString(rootHash)
	rootHash, err = state1.DataTrie().Root()
	assert.Nil(t, err)
	hrRoot1 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)
	fmt.Printf("Data root - 1-st account: %v\n", hrRoot1)

	state2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)
	state2.DataTrieTracker().SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	err = adb.SaveDataTrie(state2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated2 := base64.StdEncoding.EncodeToString(rootHash)
	rootHash, err = state1.DataTrie().Root()
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

	adr1 := integrationTests.CreateDummyAddress()
	adr2 := integrationTests.CreateDummyAddress()

	//Step 1. create accounts objects
	adb, _, _ := integrationTests.CreateAccountsDB(nil)
	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrEmpty := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	state1.DataTrieTracker().SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	err = adb.SaveDataTrie(state1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated1 := base64.StdEncoding.EncodeToString(rootHash)
	rootHash, err = state1.DataTrie().Root()
	assert.Nil(t, err)
	hrRoot1 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)
	fmt.Printf("Data root - 1-st account: %v\n", hrRoot1)

	state2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)
	state2.DataTrieTracker().SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	err = adb.SaveDataTrie(state2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.JournalLen()
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated2 := base64.StdEncoding.EncodeToString(rootHash)
	rootHash, err = state1.DataTrie().Root()
	assert.Nil(t, err)
	hrRoot2 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)
	fmt.Printf("Data root - 2-nd account: %v\n", hrRoot2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Test 2.2 test whether the datatrie roots match
	assert.Equal(t, hrRoot1, hrRoot2)

	//Step 3. Commit
	rootCommit, err := adb.Commit()
	hrCommit := base64.StdEncoding.EncodeToString(rootCommit)
	fmt.Printf("State root - committed: %v\n", hrCommit)

	//Step 4. 2-nd account changes its data
	snapshotMod := adb.JournalLen()
	state2.DataTrieTracker().SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 35})
	err = adb.SaveDataTrie(state2)
	assert.Nil(t, err)
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated2p1 := base64.StdEncoding.EncodeToString(rootHash)
	rootHash, err = state2.DataTrie().Root()
	assert.Nil(t, err)
	hrRoot2p1 := base64.StdEncoding.EncodeToString(rootHash)

	fmt.Printf("State root - modified 2-nd account: %v\n", hrCreated2p1)
	fmt.Printf("Data root - 2-nd account: %v\n", hrRoot2p1)

	//Test 4.1 test that hashes are different
	assert.NotEqual(t, hrCreated2p1, hrCreated2)

	//Test 4.2 test whether the datatrie roots match/mismatch
	assert.Equal(t, hrRoot1, hrRoot2)
	assert.NotEqual(t, hrRoot2, hrRoot2p1)

	//Step 5. Revert 2-nd account modification
	err = adb.RevertToSnapshot(snapshotMod)
	assert.Nil(t, err)
	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrCreated2Rev := base64.StdEncoding.EncodeToString(rootHash)
	rootHash, err = state2.DataTrie().Root()
	assert.Nil(t, err)
	hrRoot2Rev := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("State root - reverted 2-nd account: %v\n", hrCreated2Rev)
	fmt.Printf("Data root - 2-nd account: %v\n", hrRoot2Rev)
	assert.Equal(t, hrCommit, hrCreated2Rev)
	assert.Equal(t, hrRoot2, hrRoot2Rev)
}

func TestAccountsDB_ExecBalanceTxExecution(t *testing.T) {
	t.Parallel()

	adrSrc := integrationTests.CreateDummyAddress()
	adrDest := integrationTests.CreateDummyAddress()

	//Step 1. create accounts objects
	adb, _, _ := integrationTests.CreateAccountsDB(nil)

	acntSrc, err := adb.GetAccountWithJournal(adrSrc)
	assert.Nil(t, err)
	acntDest, err := adb.GetAccountWithJournal(adrDest)
	assert.Nil(t, err)

	//Set a high balance to src's account
	err = acntSrc.(*state.Account).SetBalanceWithJournal(big.NewInt(1000))
	assert.Nil(t, err)

	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrOriginal := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("Original root hash: %s\n", hrOriginal)

	integrationTests.PrintShardAccount(acntSrc.(*state.Account), "Source")
	integrationTests.PrintShardAccount(acntDest.(*state.Account), "Destination")

	fmt.Println("Executing OK transaction...")
	integrationTests.AdbEmulateBalanceTxSafeExecution(acntSrc.(*state.Account), acntDest.(*state.Account), adb, big.NewInt(64))

	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrOK := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("After executing an OK tx root hash: %s\n", hrOK)

	integrationTests.PrintShardAccount(acntSrc.(*state.Account), "Source")
	integrationTests.PrintShardAccount(acntDest.(*state.Account), "Destination")

	fmt.Println("Executing NOK transaction...")
	integrationTests.AdbEmulateBalanceTxSafeExecution(acntSrc.(*state.Account), acntDest.(*state.Account), adb, big.NewInt(10000))

	rootHash, err = adb.RootHash()
	assert.Nil(t, err)
	hrNok := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("After executing a NOK tx root hash: %s\n", hrNok)

	integrationTests.PrintShardAccount(acntSrc.(*state.Account), "Source")
	integrationTests.PrintShardAccount(acntDest.(*state.Account), "Destination")

	assert.NotEqual(t, hrOriginal, hrOK)
	assert.Equal(t, hrOK, hrNok)

}

func TestAccountsDB_ExecALotOfBalanceTxOK(t *testing.T) {
	t.Parallel()

	adrSrc := integrationTests.CreateDummyAddress()
	adrDest := integrationTests.CreateDummyAddress()

	//Step 1. create accounts objects
	adb, _, _ := integrationTests.CreateAccountsDB(nil)

	acntSrc, err := adb.GetAccountWithJournal(adrSrc)
	assert.Nil(t, err)
	acntDest, err := adb.GetAccountWithJournal(adrDest)
	assert.Nil(t, err)

	//Set a high balance to src's account
	err = acntSrc.(*state.Account).SetBalanceWithJournal(big.NewInt(10000000))
	assert.Nil(t, err)

	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrOriginal := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("Original root hash: %s\n", hrOriginal)

	for i := 1; i <= 1000; i++ {
		err := integrationTests.AdbEmulateBalanceTxExecution(acntSrc.(*state.Account), acntDest.(*state.Account), big.NewInt(int64(i)))

		assert.Nil(t, err)
	}

	integrationTests.PrintShardAccount(acntSrc.(*state.Account), "Source")
	integrationTests.PrintShardAccount(acntDest.(*state.Account), "Destination")
}

func TestAccountsDB_ExecALotOfBalanceTxOKorNOK(t *testing.T) {
	t.Parallel()

	adrSrc := integrationTests.CreateDummyAddress()
	adrDest := integrationTests.CreateDummyAddress()

	//Step 1. create accounts objects
	adb, _, _ := integrationTests.CreateAccountsDB(nil)

	acntSrc, err := adb.GetAccountWithJournal(adrSrc)
	assert.Nil(t, err)
	acntDest, err := adb.GetAccountWithJournal(adrDest)
	assert.Nil(t, err)

	//Set a high balance to src's account
	err = acntSrc.(*state.Account).SetBalanceWithJournal(big.NewInt(10000000))
	assert.Nil(t, err)

	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	hrOriginal := base64.StdEncoding.EncodeToString(rootHash)
	fmt.Printf("Original root hash: %s\n", hrOriginal)

	st := time.Now()
	for i := 1; i <= 1000; i++ {
		err := integrationTests.AdbEmulateBalanceTxExecution(acntSrc.(*state.Account), acntDest.(*state.Account), big.NewInt(int64(i)))
		assert.Nil(t, err)

		err = integrationTests.AdbEmulateBalanceTxExecution(acntDest.(*state.Account), acntSrc.(*state.Account), big.NewInt(int64(1000000)))
		assert.NotNil(t, err)
	}

	fmt.Printf("Done in %v\n", time.Now().Sub(st))

	integrationTests.PrintShardAccount(acntSrc.(*state.Account), "Source")
	integrationTests.PrintShardAccount(acntDest.(*state.Account), "Destination")
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
) (*state.AccountsDB, []state.AddressContainer, data.Trie) {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
	store, _ := storageUnit.NewStorageUnit(cache, persist)
	tr, _ := trie.NewTrie(store, integrationTests.TestMarshalizer, integrationTests.TestHasher)
	adb, _ := state.NewAccountsDB(tr, integrationTests.TestHasher, integrationTests.TestMarshalizer, factory.NewAccountCreator())

	addr := make([]state.AddressContainer, nrOfAccounts)
	for i := 0; i < nrOfAccounts; i++ {
		addr[i] = integrationTests.CreateAccount(adb, 0, big.NewInt(int64(balance)))
	}

	return adb, addr, tr
}

func createAndExecTxs(
	b *testing.B,
	addr []state.AddressContainer,
	nrTxs int,
	nrOfAccounts int,
	txVal int,
	adb *state.AccountsDB,
) {

	txProcessor := integrationTests.CreateTxProcessor(adb)
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
			SndAddr: addr[sender].Bytes(),
			RcvAddr: addr[receiver].Bytes(),
		}

		startTime := time.Now()
		err := txProcessor.ProcessTransaction(tx, 0)
		duration := time.Now().Sub(startTime)
		totalTime += int64(duration)
		assert.Nil(b, err)
	}
	fmt.Printf("Time needed for executing %v transactions: %v \n", nrTxs, time.Duration(totalTime))
}

func BenchmarkTxExecution(b *testing.B) {
	adrSrc := integrationTests.CreateDummyAddress()
	adrDest := integrationTests.CreateDummyAddress()

	//Step 1. create accounts objects
	adb, _, _ := integrationTests.CreateAccountsDB(nil)

	acntSrc, err := adb.GetAccountWithJournal(adrSrc)
	assert.Nil(b, err)
	acntDest, err := adb.GetAccountWithJournal(adrDest)
	assert.Nil(b, err)

	//Set a high balance to src's account
	err = acntSrc.(*state.Account).SetBalanceWithJournal(big.NewInt(10000000))
	assert.Nil(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		integrationTests.AdbEmulateBalanceTxSafeExecution(acntSrc.(*state.Account), acntDest.(*state.Account), adb, big.NewInt(1))
	}
}
