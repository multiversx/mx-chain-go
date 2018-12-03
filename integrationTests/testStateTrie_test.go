package integrationTests

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	mock2 "github.com/ElrondNetwork/elrond-go-sandbox/data/trie/mock"
	"github.com/stretchr/testify/assert"
)

func adbCreateAccountsDB() *state.AccountsDB {
	marsh := mock.MarshalizerMock{}
	dbw := trie.NewDBWriteCache(mock2.NewDatabaseMock())
	tr, err := trie.NewTrie(make([]byte, 32), dbw, mock.HasherMock{})
	if err != nil {
		panic(err)
	}

	adb := state.NewAccountsDB(tr, mock.HasherMock{}, &marsh)

	return adb
}

func adbCreateAddress(t testing.TB, buff []byte) *state.Address {
	adr, err := state.FromPubKeyBytes(buff, mock.HasherMock{})
	assert.Nil(t, err)

	return adr
}

func adbCreateAccountState(_ *testing.T, address state.Address) *state.AccountState {
	return state.NewAccountState(address, state.NewAccount(), mock.HasherMock{})
}

func adbEmulateBalanceTxExecution(acntSrc, acntDest *state.AccountState,
	handler state.AccountsHandler, value *big.Int) error {
	srcVal := acntSrc.Balance()
	destVal := acntDest.Balance()

	if srcVal.Cmp(value) < 0 {
		return errors.New("not enough funds")
	}

	//substract value from src
	err := acntSrc.SetBalance(handler, srcVal.Sub(&srcVal, value))
	if err != nil {
		return err
	}

	//add value to dest
	err = acntDest.SetBalance(handler, destVal.Add(&destVal, value))
	if err != nil {
		return err
	}

	//increment src's nonce
	err = acntSrc.SetNonce(handler, acntSrc.Nonce()+1)
	if err != nil {
		return err
	}

	return nil
}

func adbEmulateBalanceTxSafeExecution(acntSrc, acntDest *state.AccountState,
	handler state.AccountsHandler, value *big.Int) {
	snapshot := handler.Journal().Len()

	err := adbEmulateBalanceTxExecution(acntSrc, acntDest, handler, value)

	if err != nil {
		fmt.Printf("!!!! Error executing tx (value: %v), reverting...\n", value)
		err = handler.Journal().RevertFromSnapshot(snapshot)

		if err != nil {
			panic(err)
		}
	}
}

func TestAccountsDB_RetrieveData_NilRoot_ShouldRetNil(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	//since root is nil, result should be nil and data trie should be nil
	err := adb.RetrieveDataTrie(as)
	assert.Nil(t, err)
	assert.Nil(t, as.DataTrie)
}

func TestAccountsDB_RetrieveData_NilTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)
	as.Account().Root = []byte("12345")

	adb := adbCreateAccountsDB()

	adb.MainTrie = nil

	err := adb.RetrieveDataTrie(as)
	assert.NotNil(t, err)
}

func TestAccountsDB_RetrieveData_BadLength_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)
	as.Account().Root = []byte("12345")

	adb := adbCreateAccountsDB()

	//should return error
	err := adb.RetrieveDataTrie(as)
	assert.NotNil(t, err)
}

func TestAccountsDB_RetrieveData_NotFoundRoot_ShouldReturnErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)
	as.Account().Root = hasher.Compute("12345")

	adb := adbCreateAccountsDB()

	//should return error
	err := adb.RetrieveDataTrie(as)
	assert.NotNil(t, err)
}

func TestAccountsDB_RetrieveData_WithSomeValues_ShouldWork(t *testing.T) {
	//test simulates creation of a new account, data trie retrieval,
	//adding a (key, value) pair in that data trie, commiting changes
	//and then reloading the data trie based on the root hash generated before
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()
	//create a data trie with some values
	dataTrie, err := adb.MainTrie.Recreate(make([]byte, encoding.HashLength), adb.MainTrie.DBW())
	assert.Nil(t, err)
	dataTrie.Update([]byte{65, 66, 67}, []byte{32, 33, 34})
	dataTrie.Update([]byte{68, 69, 70}, []byte{35, 36, 37})
	dataTrie.Commit(nil)

	//link the new created trie to account's data root
	as.Account().Root = dataTrie.Root()

	//should not return error
	err = adb.RetrieveDataTrie(as)
	assert.Nil(t, err)
	assert.NotNil(t, as.DataTrie)

	//verify data
	data, err := as.RetrieveValue([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{32, 33, 34}, data)

	data, err = as.RetrieveValue([]byte{68, 69, 70})
	assert.Nil(t, err)
	assert.Equal(t, []byte{35, 36, 37}, data)
}

func TestAccountsDB_PutCode_NilCodeHash_ShouldRetNil(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	err := adb.PutCode(as, nil)
	assert.Nil(t, err)
	assert.Nil(t, as.CodeHash())
	assert.Nil(t, as.Code)
}

func TestAccountsDB_PutCode_EmptyCodeHash_ShouldRetNil(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	err := adb.PutCode(as, make([]byte, 0))
	assert.Nil(t, err)
	assert.Nil(t, as.CodeHash())
	assert.Nil(t, as.Code)
}

func TestAccountsDB_PutCode_NilTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	adb.MainTrie = nil

	err := adb.PutCode(as, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDB_PutCode_WithSomeValues_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	snapshotRoot := adb.MainTrie.Root()

	err := adb.PutCode(as, []byte("Smart contract code"))
	assert.Nil(t, err)
	assert.NotNil(t, as.CodeHash)
	assert.Equal(t, []byte("Smart contract code"), as.Code)

	fmt.Printf("SC code is at address: %v\n", as.CodeHash())

	//retrieve directly from the trie
	data, err := adb.MainTrie.Get(as.CodeHash())
	assert.Nil(t, err)
	assert.Equal(t, data, as.Code)

	fmt.Printf("SC code is: %v\n", string(data))

	//main root trie should have been modified
	assert.NotEqual(t, snapshotRoot, adb.MainTrie.Root())
}

func TestAccountsDB_SaveData_NilTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()
	as.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})

	adb.MainTrie = nil

	err := adb.SaveData(as)
	assert.NotNil(t, err)
}

func TestAccountsDB_SaveData_NoDirty_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	err := adb.SaveData(as)
	assert.Nil(t, err)
	assert.Equal(t, 0, adb.Journal().Len())
}

func TestAccountsDB_HasAccount_NilTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)

	adb := adbCreateAccountsDB()
	adb.MainTrie = nil

	val, err := adb.HasAccount(*adr)
	assert.NotNil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccount_NotFound_ShouldRetFalse(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)

	adb := adbCreateAccountsDB()

	//should return false
	val, err := adb.HasAccount(*adr)
	assert.Nil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccount_Found_ShouldRetTrue(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)

	adb := adbCreateAccountsDB()
	adb.MainTrie.Update(hasher.Compute(string(adr.Bytes())), []byte{65})

	//should return true
	val, err := adb.HasAccount(*adr)
	assert.Nil(t, err)
	assert.True(t, val)
}

func TestAccountsDB_SaveAccountState_NilTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()
	adb.MainTrie = nil

	err := adb.SaveAccountState(as)
	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountState_NilState_ShouldErr(t *testing.T) {
	t.Parallel()

	adb := adbCreateAccountsDB()
	adb.MainTrie = nil

	err := adb.SaveAccountState(nil)
	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountState_WithSomeValues_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)

	adb := adbCreateAccountsDB()

	//should return error
	err := adb.SaveAccountState(as)
	assert.Nil(t, err)
}

func TestAccountsDB_GetOrCreateAccount_NilTrie_ShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := adbCreateAddress(t, testHash)

	adb := adbCreateAccountsDB()
	adb.MainTrie = nil

	acnt, err := adb.GetOrCreateAccount(*adr)
	assert.NotNil(t, err)
	assert.Nil(t, acnt)

}

func TestAccountsDB_GetOrCreateAccount_ReturnExistingAccnt_ShouldWork(t *testing.T) {
	//test when the account exists
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adb := adbCreateAccountsDB()

	adr := adbCreateAddress(t, testHash)
	as := adbCreateAccountState(t, *adr)
	as.SetBalance(adb, big.NewInt(40))

	adb.SaveAccountState(as)

	acnt, err := adb.GetOrCreateAccount(*adr)
	assert.Nil(t, err)
	assert.NotNil(t, acnt)
	assert.Equal(t, acnt.Balance(), *big.NewInt(40))
}

func TestAccountsDB_GetOrCreateAccount_ReturnNotFoundAccnt_ShouldWork(t *testing.T) {
	//test when the account does not exists
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adb := adbCreateAccountsDB()

	adr := adbCreateAddress(t, testHash)

	//same address of the unsaved account
	acnt, err := adb.GetOrCreateAccount(*adr)
	assert.Nil(t, err)
	assert.NotNil(t, acnt)
	assert.Equal(t, acnt.Balance(), *big.NewInt(0))
}

func TestAccountsDB_Commit_2okAccounts_ShouldWork(t *testing.T) {
	//test creates 2 accounts (one with a data root)
	//verifies that commit saves the new tries and that can be loaded back
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	adr1 := adbCreateAddress(t, testHash1)

	testHash2 := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr2 := adbCreateAddress(t, testHash2)

	adb := adbCreateAccountsDB()

	//first account has the balance of 40
	state1, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	state1.SetBalance(adb, big.NewInt(40))

	//second account has the balance of 50 and some data
	state2, err := adb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)

	state2.SetBalance(adb, big.NewInt(50))
	state2.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	err = adb.SaveData(state2)

	//states are now prepared, committing

	h, err := adb.Commit()
	assert.Nil(t, err)
	fmt.Printf("Result hash: %v\n", base64.StdEncoding.EncodeToString(h))

	fmt.Printf("Data committed! Root: %v\n", base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))

	//reloading a new trie to test if data is inside
	newTrie, err := adb.MainTrie.Recreate(adb.MainTrie.Root(), adb.MainTrie.DBW())
	assert.Nil(t, err)
	marsh := mock.MarshalizerMock{}
	newAdb := state.NewAccountsDB(newTrie, mock.HasherMock{}, &marsh)

	//checking state1
	newState1, err := newAdb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	assert.Equal(t, newState1.Balance(), *big.NewInt(40))

	//checking state2
	newState2, err := newAdb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)
	assert.Equal(t, newState2.Balance(), *big.NewInt(50))
	assert.NotNil(t, newState2.Root)
	//get data
	err = adb.RetrieveDataTrie(newState2)
	assert.Nil(t, err)
	val, err := newState2.RetrieveValue([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{32, 33, 34}, val)
}

func TestAccountsDB_Commit_AccountData_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	adr1 := adbCreateAddress(t, testHash1)

	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	state1, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	hrCreated := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - created account: %v\n", hrCreated)

	state1.SetBalance(adb, big.NewInt(40))
	hrWithBalance := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - account with balance 40: %v\n", hrWithBalance)

	adb.Commit()
	hrCommit := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - commited: %v\n", hrCommit)

	//commit hash == account with balance
	assert.Equal(t, hrCommit, hrWithBalance)

	state1.SetBalance(adb, big.NewInt(0))

	//root hash == hrCreated
	assert.Equal(t, hrCreated, base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
	fmt.Printf("State root - account with balance 0: %v\n", base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))

	adb.RemoveAccount(*adr1)
	//root hash == hrEmpty
	assert.Equal(t, hrEmpty, base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
	fmt.Printf("State root - empty: %v\n", base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
}

func adbrPrintAccount(as *state.AccountState, tag string) {
	bal := as.Balance()
	fmt.Printf("%s address: %s\n", tag, base64.StdEncoding.EncodeToString(as.Addr.Bytes()))
	fmt.Printf("     Nonce: %d\n", as.Nonce())
	fmt.Printf("     Balance: %d\n", bal.Uint64())
	fmt.Printf("     Code hash: %v\n", base64.StdEncoding.EncodeToString(as.CodeHash()))
	fmt.Printf("     Root: %v\n\n", base64.StdEncoding.EncodeToString(as.Root()))
}

func TestAccountsDB_RevertNonceStepByStep_AccountData_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	testHash2 := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr1 := adbCreateAddress(t, testHash1)
	adr2 := adbCreateAddress(t, testHash2)

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.Journal().Len()
	hrCreated1 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.Journal().Len()
	hrCreated2 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Save the preset snapshot id
	snapshotPreSet := adb.Journal().Len()

	//Step 3. Set Nonces and save data
	state1.SetNonce(adb, 40)
	hrWithNonce1 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - account with nonce 40: %v\n", hrWithNonce1)

	state2.SetNonce(adb, 50)
	hrWithNonce2 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - account with nonce 50: %v\n", hrWithNonce2)

	//Test 3.1. current root hash shall not match created root hash hrCreated2
	assert.NotEqual(t, hrCreated2, adb.MainTrie.Root())

	//Step 4. Revert account nonce and test
	adb.Journal().RevertFromSnapshot(snapshotPreSet)

	//Test 4.1. current root hash shall match created root hash hrCreated
	hrFinal := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	assert.Equal(t, hrCreated2, hrFinal)
	fmt.Printf("State root - reverted last 2 nonces set: %v\n", hrFinal)
}

func TestAccountsDB_RevertBalanceStepByStep_AccountData_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	testHash2 := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr1 := adbCreateAddress(t, testHash1)
	adr2 := adbCreateAddress(t, testHash2)

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetOrCreateAccount(*adr1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.Journal().Len()
	hrCreated1 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.GetOrCreateAccount(*adr2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.Journal().Len()
	hrCreated2 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Save the preset snapshot id
	snapshotPreSet := adb.Journal().Len()

	//Step 3. Set balances and save data
	state1.SetBalance(adb, big.NewInt(40))
	hrWithBalance1 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - account with balance 40: %v\n", hrWithBalance1)

	state2.SetBalance(adb, big.NewInt(50))
	hrWithBalance2 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - account with balance 50: %v\n", hrWithBalance2)

	//Test 3.1. current root hash shall not match created root hash hrCreated2
	assert.NotEqual(t, hrCreated2, adb.MainTrie.Root())

	//Step 4. Revert account balances and test
	adb.Journal().RevertFromSnapshot(snapshotPreSet)

	//Test 4.1. current root hash shall match created root hash hrCreated
	hrFinal := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	assert.Equal(t, hrCreated2, hrFinal)
	fmt.Printf("State root - reverted last 2 balance set: %v\n", hrFinal)
}

func TestAccountsDB_RevertCodeStepByStep_AccountData_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	//adr1 puts code hash + code inside trie. adr2 has the same code hash
	//revert should work

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	testHash2 := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr1 := adbCreateAddress(t, testHash1)
	adr2 := adbCreateAddress(t, testHash2)

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetOrCreateAccount(*adr1)
	adb.PutCode(state1, []byte{65, 66, 67})
	assert.Nil(t, err)
	snapshotCreated1 := adb.Journal().Len()
	hrCreated1 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)

	state2, err := adb.GetOrCreateAccount(*adr2)
	adb.PutCode(state2, []byte{65, 66, 67})
	assert.Nil(t, err)
	snapshotCreated2 := adb.Journal().Len()
	hrCreated2 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Step 3. Revert second account
	adb.Journal().RevertFromSnapshot(snapshotCreated1)

	//Test 3.1. current root hash shall match created root hash hrCreated1
	hrCrt := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	assert.Equal(t, hrCreated1, hrCrt)
	fmt.Printf("State root - reverted last account: %v\n", hrCrt)

	//Step 4. Revert first account
	adb.Journal().RevertFromSnapshot(0)

	//Test 4.1. current root hash shall match empty root hash
	hrCrt = base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	assert.Equal(t, hrEmpty, hrCrt)
	fmt.Printf("State root - reverted first account: %v\n", hrCrt)
}

func TestAccountsDB_RevertDataStepByStep_AccountData_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	//adr1 puts data inside trie. adr2 puts the same data
	//revert should work

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	testHash2 := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr1 := adbCreateAddress(t, testHash1)
	adr2 := adbCreateAddress(t, testHash2)

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetOrCreateAccount(*adr1)
	state1.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	adb.SaveData(state1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.Journal().Len()
	hrCreated1 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	hrRoot1 := base64.StdEncoding.EncodeToString(state1.DataTrie.Root())

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)
	fmt.Printf("Data root - 1-st account: %v\n", hrRoot1)

	state2, err := adb.GetOrCreateAccount(*adr2)
	state2.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	adb.SaveData(state2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.Journal().Len()
	hrCreated2 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	hrRoot2 := base64.StdEncoding.EncodeToString(state1.DataTrie.Root())

	fmt.Printf("State root - created 2-nd account: %v\n", hrCreated2)
	fmt.Printf("Data root - 2-nd account: %v\n", hrRoot2)

	//Test 2.1. test that hashes and snapshots ID are different
	assert.NotEqual(t, snapshotCreated2, snapshotCreated1)
	assert.NotEqual(t, hrCreated1, hrCreated2)

	//Test 2.2 test whether the datatrie roots match
	assert.Equal(t, hrRoot1, hrRoot2)

	//Step 3. Revert 2-nd account ant test roots
	adb.Journal().RevertFromSnapshot(snapshotCreated1)
	hrCreated2Rev := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())

	assert.Equal(t, hrCreated1, hrCreated2Rev)

	//Step 4. Revert 1-st account ant test roots
	adb.Journal().RevertFromSnapshot(0)
	hrCreated1Rev := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())

	assert.Equal(t, hrEmpty, hrCreated1Rev)
}

func TestAccountsDB_RevertDataStepByStepWithCommits_AccountData_ShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	//adr1 puts data inside trie. adr2 puts the same data
	//revert should work

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	testHash2 := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr1 := adbCreateAddress(t, testHash1)
	adr2 := adbCreateAddress(t, testHash2)

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	//Step 2. create 2 new accounts
	state1, err := adb.GetOrCreateAccount(*adr1)
	state1.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	adb.SaveData(state1)
	assert.Nil(t, err)
	snapshotCreated1 := adb.Journal().Len()
	hrCreated1 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	hrRoot1 := base64.StdEncoding.EncodeToString(state1.DataTrie.Root())

	fmt.Printf("State root - created 1-st account: %v\n", hrCreated1)
	fmt.Printf("Data root - 1-st account: %v\n", hrRoot1)

	state2, err := adb.GetOrCreateAccount(*adr2)
	state2.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	adb.SaveData(state2)
	assert.Nil(t, err)
	snapshotCreated2 := adb.Journal().Len()
	hrCreated2 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	hrRoot2 := base64.StdEncoding.EncodeToString(state1.DataTrie.Root())

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
	snapshotMod := adb.Journal().Len()
	state2.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 35})
	adb.SaveData(state2)
	hrCreated2p1 := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	hrRoot2p1 := base64.StdEncoding.EncodeToString(state2.DataTrie.Root())

	fmt.Printf("State root - modified 2-nd account: %v\n", hrCreated2p1)
	fmt.Printf("Data root - 2-nd account: %v\n", hrRoot2p1)

	//Test 4.1 test that hashes are different
	assert.NotEqual(t, hrCreated2p1, hrCreated2)

	//Test 4.2 test whether the datatrie roots match/mismatch
	assert.Equal(t, hrRoot1, hrRoot2)
	assert.NotEqual(t, hrRoot2, hrRoot2p1)

	//Step 5. Revert 2-nd account modification
	adb.Journal().RevertFromSnapshot(snapshotMod)
	hrCreated2Rev := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	hrRoot2Rev := base64.StdEncoding.EncodeToString(state2.DataTrie.Root())
	fmt.Printf("State root - reverted 2-nd account: %v\n", hrCreated2Rev)
	fmt.Printf("Data root - 2-nd account: %v\n", hrRoot2Rev)
	assert.Equal(t, hrCommit, hrCreated2Rev)
	assert.Equal(t, hrRoot2, hrRoot2Rev)
}

func TestAccountsDB_ExecBalanceTxExecution(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHashSrc := hasher.Compute("ABCDEFGHIJKLMNOP")
	testHashDest := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adrSrc := adbCreateAddress(t, testHashSrc)
	adrDest := adbCreateAddress(t, testHashDest)

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()

	acntSrc, err := adb.GetOrCreateAccount(*adrSrc)
	assert.Nil(t, err)
	acntDest, err := adb.GetOrCreateAccount(*adrDest)
	assert.Nil(t, err)

	//Set a high balance to src's account
	err = acntSrc.SetBalance(adb, big.NewInt(1000))
	assert.Nil(t, err)

	hrOriginal := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("Original root hash: %s\n", hrOriginal)

	adbrPrintAccount(acntSrc, "Source")
	adbrPrintAccount(acntDest, "Destination")

	fmt.Println("Executing OK transaction...")
	adbEmulateBalanceTxSafeExecution(acntSrc, acntDest, adb, big.NewInt(64))

	hrOK := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("After executing an OK tx root hash: %s\n", hrOK)

	adbrPrintAccount(acntSrc, "Source")
	adbrPrintAccount(acntDest, "Destination")

	fmt.Println("Executing NOK transaction...")
	adbEmulateBalanceTxSafeExecution(acntSrc, acntDest, adb, big.NewInt(10000))

	hrNok := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("After executing a NOK tx root hash: %s\n", hrNok)

	adbrPrintAccount(acntSrc, "Source")
	adbrPrintAccount(acntDest, "Destination")

	assert.NotEqual(t, hrOriginal, hrOK)
	assert.Equal(t, hrOK, hrNok)

}

func TestAccountsDB_ExecALotOfBalanceTxOK(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHashSrc := hasher.Compute("ABCDEFGHIJKLMNOP")
	testHashDest := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adrSrc := adbCreateAddress(t, testHashSrc)
	adrDest := adbCreateAddress(t, testHashDest)

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()

	acntSrc, err := adb.GetOrCreateAccount(*adrSrc)
	assert.Nil(t, err)
	acntDest, err := adb.GetOrCreateAccount(*adrDest)
	assert.Nil(t, err)

	//Set a high balance to src's account
	err = acntSrc.SetBalance(adb, big.NewInt(10000000))
	assert.Nil(t, err)

	hrOriginal := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("Original root hash: %s\n", hrOriginal)

	for i := 1; i <= 1000; i++ {
		err := adbEmulateBalanceTxExecution(acntSrc, acntDest, adb, big.NewInt(int64(i)))

		assert.Nil(t, err)
	}

	adbrPrintAccount(acntSrc, "Source")
	adbrPrintAccount(acntDest, "Destination")
}

func TestAccountsDB_ExecALotOfBalanceTxOKorNOK(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHashSrc := hasher.Compute("ABCDEFGHIJKLMNOP")
	testHashDest := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adrSrc := adbCreateAddress(t, testHashSrc)
	adrDest := adbCreateAddress(t, testHashDest)

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()

	acntSrc, err := adb.GetOrCreateAccount(*adrSrc)
	assert.Nil(t, err)
	acntDest, err := adb.GetOrCreateAccount(*adrDest)
	assert.Nil(t, err)

	//Set a high balance to src's account
	err = acntSrc.SetBalance(adb, big.NewInt(10000000))
	assert.Nil(t, err)

	hrOriginal := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("Original root hash: %s\n", hrOriginal)

	st := time.Now()
	for i := 1; i <= 1000; i++ {
		err := adbEmulateBalanceTxExecution(acntSrc, acntDest, adb, big.NewInt(int64(i)))

		assert.Nil(t, err)

		err = adbEmulateBalanceTxExecution(acntDest, acntSrc, adb, big.NewInt(int64(1000000)))

		assert.NotNil(t, err)
	}

	fmt.Printf("Done in %v\n", time.Now().Sub(st))

	adbrPrintAccount(acntSrc, "Source")
	adbrPrintAccount(acntDest, "Destination")
}

func BenchmarkTxExecution(b *testing.B) {
	hasher := mock.HasherMock{}

	testHashSrc := hasher.Compute("ABCDEFGHIJKLMNOP")
	testHashDest := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adrSrc := adbCreateAddress(b, testHashSrc)
	adrDest := adbCreateAddress(b, testHashDest)

	//Step 1. create accounts objects
	adb := adbCreateAccountsDB()

	acntSrc, err := adb.GetOrCreateAccount(*adrSrc)
	assert.Nil(b, err)
	acntDest, err := adb.GetOrCreateAccount(*adrDest)
	assert.Nil(b, err)

	//Set a high balance to src's account
	err = acntSrc.SetBalance(adb, big.NewInt(10000000))
	assert.Nil(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		adbEmulateBalanceTxSafeExecution(acntSrc, acntDest, adb, big.NewInt(1))
	}
}
