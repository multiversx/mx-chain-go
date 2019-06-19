package state

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

type accountFactory struct {
}

func (af *accountFactory) CreateAccount(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
	return state.NewAccount(address, tracker)
}

//------- Helper funcs

func adbCreateAccountsDBWithStorage() (*state.AccountsDB, storage.Storer) {
	marsh := &marshal.JsonMarshalizer{}

	store := createMemUnit()
	dbw, _ := trie.NewDBWriteCache(store)
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marsh, &accountFactory{})

	return adb, store
}

func adbCreateAccountsDB() *state.AccountsDB {
	marsh := &marshal.JsonMarshalizer{}

	dbw, _ := trie.NewDBWriteCache(createMemUnit())
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marsh, &accountFactory{})

	return adb
}

func generateAddressJurnalAccountAccountsDB() (state.AddressContainer, state.AccountHandler, *state.AccountsDB) {
	adr := createDummyAddress()
	adb := adbCreateAccountsDB()

	account, _ := state.NewAccount(adr, adb)

	return adr, account, adb
}

func adbEmulateBalanceTxExecution(acntSrc, acntDest *state.Account, value *big.Int) error {

	srcVal := acntSrc.Balance
	destVal := acntDest.Balance

	if srcVal.Cmp(value) < 0 {
		return errors.New("not enough funds")
	}

	//substract value from src
	err := acntSrc.SetBalanceWithJournal(srcVal.Sub(srcVal, value))
	if err != nil {
		return err
	}

	//add value to dest
	err = acntDest.SetBalanceWithJournal(destVal.Add(destVal, value))
	if err != nil {
		return err
	}

	//increment src's nonce
	err = acntSrc.SetNonceWithJournal(acntSrc.Nonce + 1)
	if err != nil {
		return err
	}

	return nil
}

func adbEmulateBalanceTxSafeExecution(acntSrc, acntDest *state.Account,
	accounts state.AccountsAdapter, value *big.Int) {

	snapshot := accounts.JournalLen()

	err := adbEmulateBalanceTxExecution(acntSrc, acntDest, value)

	if err != nil {
		fmt.Printf("!!!! Error executing tx (value: %v), reverting...\n", value)
		err = accounts.RevertToSnapshot(snapshot)

		if err != nil {
			panic(err)
		}
	}
}

func adbPrintAccount(account *state.Account, tag string) {
	bal := account.Balance
	fmt.Printf("%s address: %s\n", tag, base64.StdEncoding.EncodeToString(account.AddressContainer().Bytes()))
	fmt.Printf("     Nonce: %d\n", account.Nonce)
	fmt.Printf("     Balance: %d\n", bal.Uint64())
	fmt.Printf("     Code hash: %v\n", base64.StdEncoding.EncodeToString(account.CodeHash))
	fmt.Printf("     Root hash: %v\n\n", base64.StdEncoding.EncodeToString(account.RootHash))
}

//------- Functionality tests

func TestAccountsDB_RetrieveDataWithSomeValuesShouldWork(t *testing.T) {
	//test simulates creation of a new account, data trie retrieval,
	//adding a (key, value) pair in that data trie, commiting changes
	//and then reloading the data trie based on the root hash generated before
	t.Parallel()

	_, account, adb := generateAddressJurnalAccountAccountsDB()

	account.DataTrieTracker().SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	account.DataTrieTracker().SaveKeyValue([]byte{68, 69, 70}, []byte{35, 36, 37})

	err := adb.SaveDataTrie(account)
	assert.Nil(t, err)

	_, err = adb.Commit()
	assert.Nil(t, err)

	recoveredAccount, err := adb.GetAccountWithJournal(account.AddressContainer())
	assert.Nil(t, err)

	//verify data
	data, err := recoveredAccount.DataTrieTracker().RetrieveValue([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{32, 33, 34}, data)

	data, err = recoveredAccount.DataTrieTracker().RetrieveValue([]byte{68, 69, 70})
	assert.Nil(t, err)
	assert.Equal(t, []byte{35, 36, 37}, data)
}

func TestAccountsDB_PutCodeWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressJurnalAccountAccountsDB()

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

	_, account, adb := generateAddressJurnalAccountAccountsDB()

	err := adb.SaveDataTrie(account)
	assert.Nil(t, err)
	assert.Equal(t, 0, adb.JournalLen())
}

func TestAccountsDB_HasAccountNotFoundShouldRetFalse(t *testing.T) {
	t.Parallel()

	adr, _, adb := generateAddressJurnalAccountAccountsDB()

	//should return false
	val, err := adb.HasAccount(adr)
	assert.Nil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccountFoundShouldRetTrue(t *testing.T) {
	t.Parallel()

	adr, _, adb := generateAddressJurnalAccountAccountsDB()
	_, err := adb.GetAccountWithJournal(adr)
	assert.Nil(t, err)

	//should return true
	val, err := adb.HasAccount(adr)
	assert.Nil(t, err)
	assert.True(t, val)
}

func TestAccountsDB_SaveAccountStateWithSomeValues_ShouldWork(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressJurnalAccountAccountsDB()

	err := adb.SaveAccount(account)
	assert.Nil(t, err)
}

func TestAccountsDB_GetJournalizedAccountReturnExistingAccntShouldWork(t *testing.T) {
	t.Parallel()

	balance := big.NewInt(40)
	adr, accountHandler, adb := generateAddressJurnalAccountAccountsDB()
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

	adr, _, adb := generateAddressJurnalAccountAccountsDB()

	//same address of the unsaved account
	accountHandlerRecovered, err := adb.GetAccountWithJournal(adr)
	assert.Nil(t, err)
	accountRecovered := accountHandlerRecovered.(*state.Account)
	assert.NotNil(t, accountRecovered)
	assert.Equal(t, accountRecovered.Balance, big.NewInt(0))
}

func TestAccountsDB_CommitTwoOkAccountsShouldWork(t *testing.T) {
	//test creates 2 accounts (one with a data root)
	//verifies that commit saves the new tries and that can be loaded back
	t.Parallel()

	adr1, _, adb := generateAddressJurnalAccountAccountsDB()
	buff := make([]byte, sha256.Sha256{}.Size())
	rand.Read(buff)
	adr2 := createDummyAddress()

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

	fmt.Printf("Data committed! Root: %v\n", base64.StdEncoding.EncodeToString(adb.RootHash()))

	//reloading a new trie to test if data is inside
	err = adb.RecreateTrie(adb.RootHash())
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
	//get data
	err = adb.LoadDataTrie(newState2)
	assert.Nil(t, err)
	valRecovered, err := newState2.DataTrieTracker().RetrieveValue(key)
	assert.Nil(t, err)
	assert.Equal(t, val, valRecovered)
}

func TestTrieDB_RecreateFromStorageShouldWork(t *testing.T) {
	hasher := sha256.Sha256{}

	store := createMemUnit()
	dbw, _ := trie.NewDBWriteCache(store)
	tr1, _ := trie.NewTrie(make([]byte, 32), dbw, hasher)

	key := hasher.Compute("key")
	value := hasher.Compute("value")

	tr1.Update(key, value)
	h1, err := tr1.Commit(nil)
	assert.Nil(t, err)

	dbw, _ = trie.NewDBWriteCache(store)
	tr2, err := trie.NewTrie(h1, dbw, hasher)
	assert.Nil(t, err)

	valRecov, err := tr2.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, value, valRecov)
}

func TestAccountsDB_CommitTwoOkAccountsWithRecreationFromStorageShouldWork(t *testing.T) {
	//test creates 2 accounts (one with a data root)
	//verifies that commit saves the new tries and that can be loaded back
	t.Parallel()

	adb, mu := adbCreateAccountsDBWithStorage()
	adr1 := createDummyAddress()
	buff := make([]byte, sha256.Sha256{}.Size())
	rand.Read(buff)
	adr2 := createDummyAddress()

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

	fmt.Printf("Data committed! Root: %v\n", base64.StdEncoding.EncodeToString(adb.RootHash()))

	dbw, _ := trie.NewDBWriteCache(mu)
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ = state.NewAccountsDB(tr, sha256.Sha256{}, &marshal.JsonMarshalizer{}, &accountFactory{})

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
	//get data
	err = adb.LoadDataTrie(newState2)
	assert.Nil(t, err)
	valRecovered, err := newState2.DataTrieTracker().RetrieveValue(key)
	assert.Nil(t, err)
	assert.Equal(t, val, valRecovered)
}

func TestAccountsDB_CommitAnEmptyStateShouldWork(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "this test should not have paniced")
		}
	}()

	adb, _ := adbCreateAccountsDBWithStorage()

	hash, err := adb.Commit()

	assert.Nil(t, err)
	assert.Equal(t, make([]byte, state.HashLength), hash)
}

func TestAccountsDB_CommitAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	adr1, _, adb := generateAddressJurnalAccountAccountsDB()

	hrEmpty := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	hrCreated := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - created account: %v\n", hrCreated)

	err = state1.(*state.Account).SetBalanceWithJournal(big.NewInt(40))
	assert.Nil(t, err)
	hrWithBalance := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - account with balance 40: %v\n", hrWithBalance)

	_, err = adb.Commit()
	assert.Nil(t, err)
	hrCommit := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - committed: %v\n", hrCommit)

	//commit hash == account with balance
	assert.Equal(t, hrCommit, hrWithBalance)

	err = state1.(*state.Account).SetBalanceWithJournal(big.NewInt(0))
	assert.Nil(t, err)

	//root hash == hrCreated
	assert.Equal(t, hrCreated, base64.StdEncoding.EncodeToString(adb.RootHash()))
	fmt.Printf("State root - account with balance 0: %v\n", base64.StdEncoding.EncodeToString(adb.RootHash()))

	err = adb.RemoveAccount(adr1)
	assert.Nil(t, err)

	//root hash == hrEmpty
	assert.Equal(t, hrEmpty, base64.StdEncoding.EncodeToString(adb.RootHash()))
	fmt.Printf("State root - empty: %v\n", base64.StdEncoding.EncodeToString(adb.RootHash()))
}

//------- Revert

func TestAccountsDB_RevertNonceStepByStepAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := createDummyAddress()
	adr2 := createDummyAddress()

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.JournalLen()
	hrCreated1 := base64.StdEncoding.EncodeToString(adb.RootHash())

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.JournalLen()
	hrCreated2 := base64.StdEncoding.EncodeToString(adb.RootHash())

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Save the preset snapshot id
	snapshotPreSet := adb.JournalLen()

	//Step 3. Set Nonces and save data
	err = state1.(*state.Account).SetNonceWithJournal(40)
	assert.Nil(t, err)
	hrWithNonce1 := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - account with nonce 40: %v\n", hrWithNonce1)

	err = state2.(*state.Account).SetNonceWithJournal(50)
	assert.Nil(t, err)
	hrWithNonce2 := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - account with nonce 50: %v\n", hrWithNonce2)

	//Test 3.1. current root hash shall not match created root hash hrCreated2
	assert.NotEqual(t, hrCreated2, adb.RootHash())

	//Step 4. Revert account nonce and test
	err = adb.RevertToSnapshot(snapshotPreSet)
	assert.Nil(t, err)

	//Test 4.1. current root hash shall match created root hash hrCreated
	hrFinal := base64.StdEncoding.EncodeToString(adb.RootHash())
	assert.Equal(t, hrCreated2, hrFinal)
	fmt.Printf("State root - reverted last 2 nonces set: %v\n", hrFinal)
}

func TestAccountsDB_RevertBalanceStepByStepAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := createDummyAddress()
	adr2 := createDummyAddress()

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.JournalLen()
	hrCreated1 := base64.StdEncoding.EncodeToString(adb.RootHash())

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.JournalLen()
	hrCreated2 := base64.StdEncoding.EncodeToString(adb.RootHash())

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Save the preset snapshot id
	snapshotPreSet := adb.JournalLen()

	//Step 3. Set balances and save data
	err = state1.(*state.Account).SetBalanceWithJournal(big.NewInt(40))
	assert.Nil(t, err)
	hrWithBalance1 := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - account with balance 40: %v\n", hrWithBalance1)

	err = state2.(*state.Account).SetBalanceWithJournal(big.NewInt(50))
	assert.Nil(t, err)
	hrWithBalance2 := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - account with balance 50: %v\n", hrWithBalance2)

	//Test 3.1. current root hash shall not match created root hash hrCreated2
	assert.NotEqual(t, hrCreated2, adb.RootHash())

	//Step 4. Revert account balances and test
	err = adb.RevertToSnapshot(snapshotPreSet)
	assert.Nil(t, err)

	//Test 4.1. current root hash shall match created root hash hrCreated
	hrFinal := base64.StdEncoding.EncodeToString(adb.RootHash())
	assert.Equal(t, hrCreated2, hrFinal)
	fmt.Printf("State root - reverted last 2 balance set: %v\n", hrFinal)
}

func TestAccountsDB_RevertCodeStepByStepAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	//adr1 puts code hash + code inside trie. adr2 has the same code hash
	//revert should work

	adr1 := createDummyAddress()
	adr2 := createDummyAddress()

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	err = adb.PutCode(state1, []byte{65, 66, 67})
	assert.Nil(t, err)
	snapshotCreated1 := adb.JournalLen()
	hrCreated1 := base64.StdEncoding.EncodeToString(adb.RootHash())

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)
	err = adb.PutCode(state2, []byte{65, 66, 67})
	assert.Nil(t, err)
	snapshotCreated2 := adb.JournalLen()
	hrCreated2 := base64.StdEncoding.EncodeToString(adb.RootHash())

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Step 3. Revert second account
	err = adb.RevertToSnapshot(snapshotCreated1)
	assert.Nil(t, err)

	//Test 3.1. current root hash shall match created root hash hrCreated1
	hrCrt := base64.StdEncoding.EncodeToString(adb.RootHash())
	assert.Equal(t, hrCreated1, hrCrt)
	fmt.Printf("State root - reverted last account: %v\n", hrCrt)

	//Step 4. Revert first account
	err = adb.RevertToSnapshot(0)
	assert.Nil(t, err)

	//Test 4.1. current root hash shall match empty root hash
	hrCrt = base64.StdEncoding.EncodeToString(adb.RootHash())
	assert.Equal(t, hrEmpty, hrCrt)
	fmt.Printf("State root - reverted first account: %v\n", hrCrt)
}

func TestAccountsDB_RevertDataStepByStepAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	//adr1 puts data inside trie. adr2 puts the same data
	//revert should work

	adr1 := createDummyAddress()
	adr2 := createDummyAddress()

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	state1.DataTrieTracker().SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	err = adb.SaveDataTrie(state1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.JournalLen()
	hrCreated1 := base64.StdEncoding.EncodeToString(adb.RootHash())
	hrRoot1 := base64.StdEncoding.EncodeToString(state1.DataTrie().Root())

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)
	fmt.Printf("Data root - 1-st account: %v\n", hrRoot1)

	state2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)
	state2.DataTrieTracker().SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	err = adb.SaveDataTrie(state2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.JournalLen()
	hrCreated2 := base64.StdEncoding.EncodeToString(adb.RootHash())
	hrRoot2 := base64.StdEncoding.EncodeToString(state1.DataTrie().Root())

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
	hrCreated2Rev := base64.StdEncoding.EncodeToString(adb.RootHash())

	assert.Equal(t, hrCreated1, hrCreated2Rev)

	//Step 4. Revert 1-st account ant test roots
	err = adb.RevertToSnapshot(0)
	assert.Nil(t, err)
	hrCreated1Rev := base64.StdEncoding.EncodeToString(adb.RootHash())

	assert.Equal(t, hrEmpty, hrCreated1Rev)
}

func TestAccountsDB_RevertDataStepByStepWithCommitsAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	//adr1 puts data inside trie. adr2 puts the same data
	//revert should work

	adr1 := createDummyAddress()
	adr2 := createDummyAddress()

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetAccountWithJournal(adr1)
	assert.Nil(t, err)
	state1.DataTrieTracker().SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	err = adb.SaveDataTrie(state1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.JournalLen()
	hrCreated1 := base64.StdEncoding.EncodeToString(adb.RootHash())
	hrRoot1 := base64.StdEncoding.EncodeToString(state1.DataTrie().Root())

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)
	fmt.Printf("Data root - 1-st account: %v\n", hrRoot1)

	state2, err := adb.GetAccountWithJournal(adr2)
	assert.Nil(t, err)
	state2.DataTrieTracker().SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	err = adb.SaveDataTrie(state2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.JournalLen()
	hrCreated2 := base64.StdEncoding.EncodeToString(adb.RootHash())
	hrRoot2 := base64.StdEncoding.EncodeToString(state1.DataTrie().Root())

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
	hrCreated2p1 := base64.StdEncoding.EncodeToString(adb.RootHash())
	hrRoot2p1 := base64.StdEncoding.EncodeToString(state2.DataTrie().Root())

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
	hrCreated2Rev := base64.StdEncoding.EncodeToString(adb.RootHash())
	hrRoot2Rev := base64.StdEncoding.EncodeToString(state2.DataTrie().Root())
	fmt.Printf("State root - reverted 2-nd account: %v\n", hrCreated2Rev)
	fmt.Printf("Data root - 2-nd account: %v\n", hrRoot2Rev)
	assert.Equal(t, hrCommit, hrCreated2Rev)
	assert.Equal(t, hrRoot2, hrRoot2Rev)
}

func TestAccountsDB_ExecBalanceTxExecution(t *testing.T) {
	t.Parallel()

	adrSrc := createDummyAddress()
	adrDest := createDummyAddress()

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()

	acntSrc, err := adb.GetAccountWithJournal(adrSrc)
	assert.Nil(t, err)
	acntDest, err := adb.GetAccountWithJournal(adrDest)
	assert.Nil(t, err)

	//Set a high balance to src's account
	err = acntSrc.(*state.Account).SetBalanceWithJournal(big.NewInt(1000))
	assert.Nil(t, err)

	hrOriginal := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("Original root hash: %s\n", hrOriginal)

	adbPrintAccount(acntSrc.(*state.Account), "Source")
	adbPrintAccount(acntDest.(*state.Account), "Destination")

	fmt.Println("Executing OK transaction...")
	adbEmulateBalanceTxSafeExecution(acntSrc.(*state.Account), acntDest.(*state.Account), adb, big.NewInt(64))

	hrOK := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("After executing an OK tx root hash: %s\n", hrOK)

	adbPrintAccount(acntSrc.(*state.Account), "Source")
	adbPrintAccount(acntDest.(*state.Account), "Destination")

	fmt.Println("Executing NOK transaction...")
	adbEmulateBalanceTxSafeExecution(acntSrc.(*state.Account), acntDest.(*state.Account), adb, big.NewInt(10000))

	hrNok := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("After executing a NOK tx root hash: %s\n", hrNok)

	adbPrintAccount(acntSrc.(*state.Account), "Source")
	adbPrintAccount(acntDest.(*state.Account), "Destination")

	assert.NotEqual(t, hrOriginal, hrOK)
	assert.Equal(t, hrOK, hrNok)

}

func TestAccountsDB_ExecALotOfBalanceTxOK(t *testing.T) {
	t.Parallel()

	adrSrc := createDummyAddress()
	adrDest := createDummyAddress()

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()

	acntSrc, err := adb.GetAccountWithJournal(adrSrc)
	assert.Nil(t, err)
	acntDest, err := adb.GetAccountWithJournal(adrDest)
	assert.Nil(t, err)

	//Set a high balance to src's account
	err = acntSrc.(*state.Account).SetBalanceWithJournal(big.NewInt(10000000))
	assert.Nil(t, err)

	hrOriginal := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("Original root hash: %s\n", hrOriginal)

	for i := 1; i <= 1000; i++ {
		err := adbEmulateBalanceTxExecution(acntSrc.(*state.Account), acntDest.(*state.Account), big.NewInt(int64(i)))

		assert.Nil(t, err)
	}

	adbPrintAccount(acntSrc.(*state.Account), "Source")
	adbPrintAccount(acntDest.(*state.Account), "Destination")
}

func TestAccountsDB_ExecALotOfBalanceTxOKorNOK(t *testing.T) {
	t.Parallel()

	adrSrc := createDummyAddress()
	adrDest := createDummyAddress()

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()

	acntSrc, err := adb.GetAccountWithJournal(adrSrc)
	assert.Nil(t, err)
	acntDest, err := adb.GetAccountWithJournal(adrDest)
	assert.Nil(t, err)

	//Set a high balance to src's account
	err = acntSrc.(*state.Account).SetBalanceWithJournal(big.NewInt(10000000))
	assert.Nil(t, err)

	hrOriginal := base64.StdEncoding.EncodeToString(adb.RootHash())
	fmt.Printf("Original root hash: %s\n", hrOriginal)

	st := time.Now()
	for i := 1; i <= 1000; i++ {
		err := adbEmulateBalanceTxExecution(acntSrc.(*state.Account), acntDest.(*state.Account), big.NewInt(int64(i)))

		assert.Nil(t, err)

		err = adbEmulateBalanceTxExecution(acntDest.(*state.Account), acntSrc.(*state.Account), big.NewInt(int64(1000000)))

		assert.NotNil(t, err)
	}

	fmt.Printf("Done in %v\n", time.Now().Sub(st))

	adbPrintAccount(acntSrc.(*state.Account), "Source")
	adbPrintAccount(acntDest.(*state.Account), "Destination")
}

func BenchmarkTxExecution(b *testing.B) {
	adrSrc := createDummyAddress()
	adrDest := createDummyAddress()

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()

	acntSrc, err := adb.GetAccountWithJournal(adrSrc)
	assert.Nil(b, err)
	acntDest, err := adb.GetAccountWithJournal(adrDest)
	assert.Nil(b, err)

	//Set a high balance to src's account
	err = acntSrc.(*state.Account).SetBalanceWithJournal(big.NewInt(10000000))
	assert.Nil(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		adbEmulateBalanceTxSafeExecution(acntSrc.(*state.Account), acntDest.(*state.Account), adb, big.NewInt(1))
	}
}
