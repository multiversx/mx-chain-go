package integrationTests

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	mock2 "github.com/ElrondNetwork/elrond-go-sandbox/data/trie/mock"
	"github.com/stretchr/testify/assert"
)

var testHasher = mock.HasherMock{}
var testMarshalizer = mock.MarshalizerMock{}

func createAccountsDB(t *testing.T) *state.AccountsDB {
	dbw := trie.NewDBWriteCache(mock2.NewDatabaseMock())
	tr, err := trie.NewTrie(make([]byte, 32), dbw, &testHasher)
	assert.Nil(t, err)

	adb := state.NewAccountsDB(tr, &testHasher, &testMarshalizer)

	return adb
}

func createAddress(t *testing.T, buff []byte) *state.Address {
	adr, err := state.FromPubKeyBytes(buff)
	assert.Nil(t, err)

	return adr
}

func createAccountState(address state.Address) *state.AccountState {
	return state.NewAccountState(address, state.Account{}, testHasher)
}

func TestAccountsDB_RetrieveCode_Found_ShouldWork(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")

	adr := createAddress(t, testHash)
	st := createAccountState(*adr)
	st.CodeHash = testHasher.Compute("12345")

	adb := createAccountsDB(t)
	//manually put it in the trie
	adb.MainTrie.Update(st.CodeHash, []byte{68, 69, 70})

	//should return nil
	err := adb.RetrieveCode(st)
	assert.Nil(t, err)
	assert.Equal(t, st.Code, []byte{68, 69, 70})
}

func TestAccountsDB_RetrieveData_WithSomeValues_ShouldWork(t *testing.T) {
	//test simulates creation of a new account, data trie retrieval,
	//adding a (key, value) pair in that data trie, commiting changes
	//and then reloading the data trie based on the root hash generated before
	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")

	adr := createAddress(t, testHash)
	st := createAccountState(*adr)
	st.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)

	//should not return error
	err := adb.RetrieveData(st)
	assert.Nil(t, err)
	assert.NotNil(t, st.Data)

	fmt.Printf("Root: %v\n", st.Root)

	//add something to the data trie
	st.Data.Update([]byte{65, 66, 67}, []byte{68, 69, 70})
	st.Root = st.Data.Root()
	fmt.Printf("state.Root is %v\n", st.Root)
	hashResult, err := st.Data.Commit(nil)

	assert.Equal(t, hashResult, st.Root)

	//make data trie null and then retrieve the trie and search the data
	st.Data = nil
	err = adb.RetrieveData(st)
	assert.Nil(t, err)
	assert.NotNil(t, st.Data)

	val, err := st.Data.Get([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{68, 69, 70}, val)
}

func TestAccountsDB_PutCode_EmptyCodeHash_ShouldRetNil(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")

	adr := createAddress(t, testHash)
	st := createAccountState(*adr)
	st.Root = st.Addr.Hash(testHasher)

	adb := createAccountsDB(t)

	err := adb.PutCode(st, make([]byte, 0))
	assert.Nil(t, err)
	assert.Nil(t, st.CodeHash)
	assert.Nil(t, st.Code)
}

func TestAccountsDB_PutCode_WithSomeValues_ShouldWork(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")

	adr := createAddress(t, testHash)
	st := createAccountState(*adr)
	st.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)

	snapshotRoot := adb.MainTrie.Root()

	err := adb.PutCode(st, []byte("Smart contract code"))
	assert.Nil(t, err)
	assert.NotNil(t, st.CodeHash)
	assert.Equal(t, []byte("Smart contract code"), st.Code)

	fmt.Printf("SC code is at address: %v\n", st.CodeHash)

	//retrieve directly from the trie
	data, err := adb.MainTrie.Get(st.CodeHash)
	assert.Nil(t, err)
	assert.Equal(t, data, st.Code)

	fmt.Printf("SC code is: %v\n", string(data))

	//main root trie should have been modified
	assert.NotEqual(t, snapshotRoot, adb.MainTrie.Root())
}

func TestAccountsDB_HasAccount_NotFound_ShouldRetFalse(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")

	adr := createAddress(t, testHash)
	st := createAccountState(*adr)
	st.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)

	//should return false
	val, err := adb.HasAccount(*adr)
	assert.Nil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccount_Found_ShouldRetTrue(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")

	adr := createAddress(t, testHash)
	st := createAccountState(*adr)
	st.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)
	adb.MainTrie.Update(testHasher.Compute(string(adr.Bytes())), []byte{65})

	//should return true
	val, err := adb.HasAccount(*adr)
	assert.Nil(t, err)
	assert.True(t, val)
}

func TestAccountsDB_SaveAccountState_WithSomeValues_ShouldWork(t *testing.T) {
	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")

	adr := createAddress(t, testHash)
	st := createAccountState(*adr)
	st.Root = testHasher.Compute("12345")

	adb := createAccountsDB(t)

	//should return error
	err := adb.SaveAccountState(st)
	assert.Nil(t, err)
}

func TestAccountsDB_GetOrCreateAccount_ReturnExistingAccnt_ShouldWork(t *testing.T) {
	//test when the account exists
	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")

	adr := createAddress(t, testHash)
	st := createAccountState(*adr)
	st.Balance = big.NewInt(40)

	adb := createAccountsDB(t)
	adb.SaveAccountState(st)

	acnt, err := adb.GetOrCreateAccount(*adr)
	assert.Nil(t, err)
	assert.NotNil(t, acnt)
	assert.Equal(t, acnt.Balance, big.NewInt(40))
}

func TestAccountsDB_GetOrCreateAccount_ReturnNotFoundAccnt_ShouldWork(t *testing.T) {
	//test when the account does not exists
	testHash := testHasher.Compute("ABCDEFGHIJKLMNOP")

	adr := createAddress(t, testHash)
	st := createAccountState(*adr)
	st.Balance = big.NewInt(40)

	adb := createAccountsDB(t)

	//same address of the unsaved account
	acnt, err := adb.GetOrCreateAccount(*adr)
	assert.Nil(t, err)
	assert.NotNil(t, acnt)
	assert.Equal(t, acnt.Balance, big.NewInt(0))
}

func TestAccountsDB_Commit_2okAccounts_ShouldWork(t *testing.T) {
	//test creates 2 accounts (one with a data root)
	//verifies that commit saves the new tries and that can be loaded back

	testHash1 := testHasher.Compute("ABCDEFGHIJKLMNOP")
	adr1 := createAddress(t, testHash1)

	testHash2 := testHasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr2 := createAddress(t, testHash2)

	adb := createAccountsDB(t)

	//first account has the balance of 40
	state1, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	state1.Balance = big.NewInt(40)

	//second account has the balance of 50 and something in data root
	state2, err := adb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)

	state2.Balance = big.NewInt(50)
	state2.Root = make([]byte, encoding.HashLength)
	err = adb.RetrieveData(state2)
	assert.Nil(t, err)

	state2.Data.Update([]byte{65, 66, 67}, []byte{68, 69, 70})

	//states are now prepared, committing

	h, err := adb.Commit()
	assert.Nil(t, err)
	fmt.Printf("Result hash: %v\n", base64.StdEncoding.EncodeToString(h))

	fmt.Printf("Data committed! Root: %v\n",
		base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))

	//reloading a new trie to test if data is inside
	newTrie, err := adb.MainTrie.Recreate(adb.MainTrie.Root(), adb.MainTrie.DBW())
	assert.Nil(t, err)
	newAdb := state.NewAccountsDB(newTrie, testHasher, &testMarshalizer)

	//checking state1
	newState1, err := newAdb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	assert.Equal(t, newState1.Balance, big.NewInt(40))

	//checking state2
	newState2, err := newAdb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)
	assert.Equal(t, newState2.Balance, big.NewInt(50))
	assert.NotNil(t, newState2.Root)
	//get data
	err = adb.RetrieveData(newState2)
	assert.Nil(t, err)
	assert.NotNil(t, newState2.Data)
	val, err := newState2.Data.Get([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, val, []byte{68, 69, 70})
}

func TestAccountsDB_Commits_2okAccounts3times_ShouldWork(t *testing.T) {
	//test creates 2 accounts
	//test commits 3 times different values
	//test whether the retrieved data for each root snapshot corresponds
	//practically, we test the immutability of the trie structure

	testHash1 := testHasher.Compute("ABCDEFGHIJKLMNOP")
	adr1 := createAddress(t, testHash1)

	testHash2 := testHasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr2 := createAddress(t, testHash2)

	adb := createAccountsDB(t)

	//first account has the balance of 40
	state1, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	state1.Balance = big.NewInt(40)

	//second account has the balance of 50
	state2, err := adb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)
	state2.Balance = big.NewInt(50)

	//commit 1 and snapshot
	snapshot1, err := adb.Commit()
	assert.Nil(t, err)

	//-------------------------------------------------------

	//modify the first account with balance 41
	state1, err = adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	state1.Balance = big.NewInt(41)

	//modify the first account with balance 51
	state2, err = adb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)
	state2.Balance = big.NewInt(51)

	//commit 2 and snapshot
	snapshot2, err := adb.Commit()
	assert.Nil(t, err)

	//-------------------------------------------------------

	//modify the first account with balance 42
	state1, err = adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	state1.Balance = big.NewInt(42)

	//modify the first account with balance 52
	state2, err = adb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)
	state2.Balance = big.NewInt(52)

	//commit 3 and snapshot
	snapshot3, err := adb.Commit()
	assert.Nil(t, err)

	//=======================================================

	//revert the trie to snapshot 1
	adb.MainTrie, err = trie.NewTrie(snapshot1, adb.MainTrie.DBW(), testHasher)
	assert.Nil(t, err)

	//test state 1
	stTest, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	assert.Equal(t, stTest.Balance, big.NewInt(40))
	//test state 2
	stTest, err = adb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)
	assert.Equal(t, stTest.Balance, big.NewInt(50))

	//-------------------------------------------------------

	//revert the trie to snapshot 2
	adb.MainTrie, err = trie.NewTrie(snapshot2, adb.MainTrie.DBW(), testHasher)
	assert.Nil(t, err)

	//test state 1
	stTest, err = adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	assert.Equal(t, stTest.Balance, big.NewInt(41))
	//test state 2
	stTest, err = adb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)
	assert.Equal(t, stTest.Balance, big.NewInt(51))

	//-------------------------------------------------------

	//revert the trie to snapshot 3
	adb.MainTrie, err = trie.NewTrie(snapshot3, adb.MainTrie.DBW(), testHasher)
	assert.Nil(t, err)

	//test state 1
	stTest, err = adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	assert.Equal(t, stTest.Balance, big.NewInt(42))
	//test state 2
	stTest, err = adb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)
	assert.Equal(t, stTest.Balance, big.NewInt(52))
}

func TestAccountsDB_Commit_AccountData_ShouldWork(t *testing.T) {
	testHash1 := testHasher.Compute("ABCDEFGHIJKLMNOP")
	adr1 := createAddress(t, testHash1)

	adb := createAccountsDB(t)
	hrEmpty := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	state1, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)

	state1.Balance = big.NewInt(40)
	hrCreated := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - created account: %v\n", hrCreated)

	adb.SaveAccountState(state1)
	hrWithBalance := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - account with balance 40: %v\n", hrWithBalance)

	adb.Commit()
	hrCommit := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - commited: %v\n", hrCommit)

	//commit hash == account with balance
	assert.Equal(t, hrCommit, hrWithBalance)

	state1.Balance = big.NewInt(0)
	adb.SaveAccountState(state1)

	//root hash == hrCreated
	assert.Equal(t, hrCreated, base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
	fmt.Printf("State root - account with balance 0: %v\n", base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))

	adb.RemoveAccount(*adr1)
	//root hash == hrEmpty
	assert.Equal(t, hrEmpty, base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
	fmt.Printf("State root - empty: %v\n", base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
}
