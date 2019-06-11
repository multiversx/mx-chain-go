package state

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
)

func createAccountsDB() (*state.AccountsDB, storage.Storer) {
	marsh := &marshal.JsonMarshalizer{}

	mu := createMemUnit()
	dbw, _ := trie.NewDBWriteCache(mu)
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marsh, &accountFactory{})

	return adb, mu
}

func genAddressJurnalAccountAccountsDB() (state.AddressContainer, state.AccountHandler, *state.AccountsDB, storage.Storer) {
	adr := createDummyAddress()
	adb, mu := createAccountsDB()

	account, _ := state.NewAccount(adr, adb)

	return adr, account, adb, mu
}

func TestAccountsDB_CommitTwoOkAccountsWithRecreationShouldWork(t *testing.T) {
	//test creates 2 accounts (one with a data root)
	//verifies that commit saves the new tries and that can be loaded back
	t.Parallel()

	adr1, _, adb, mu := genAddressJurnalAccountAccountsDB()
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
