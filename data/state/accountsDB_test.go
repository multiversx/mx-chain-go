package state_test

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/stretchr/testify/assert"
)

func accountsDBCreateAccountsDB() *state.AccountsDB {
	marsh := mock.MarshalizerMock{}
	adb := state.NewAccountsDB(mock.NewMockTrie(), mock.HasherMock{}, &marsh)

	return adb
}

func accountsDBCreateAddress(t *testing.T, buff []byte) *state.Address {
	adr, err := state.FromPubKeyBytes(buff, mock.HasherMock{})
	assert.Nil(t, err)

	return adr
}

func accountsDBCreateAccountState(_ *testing.T, address *state.Address) *state.AccountState {
	return state.NewAccountState(address, state.NewAccount(), mock.HasherMock{})
}

func TestAccountsDBRetrieveCodeNilCodeHashShouldRetNil(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()

	//since code hash is nil, result should be nil and code should be nil
	err := adb.RetrieveCode(as)
	assert.Nil(t, err)
	assert.Nil(t, as.Code())
}

func TestAccountsDBRetrieveCodeNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)
	as.SetCodeHashNoJournal([]byte("12345"))

	adb := accountsDBCreateAccountsDB()

	adb.MainTrie = nil

	err := adb.RetrieveCode(as)
	assert.NotNil(t, err)
}

func TestAccountsDBRetrieveCodeBadLengthShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)
	as.SetCodeHashNoJournal([]byte("12345"))

	adb := accountsDBCreateAccountsDB()

	//should return error
	err := adb.RetrieveCode(as)
	assert.NotNil(t, err)
}

func TestAccountsDBRetrieveCodeMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)
	as.SetCodeHashNoJournal(hasher.Compute("12345"))

	adb := accountsDBCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).Fail = true

	//should return error
	err := adb.RetrieveCode(as)
	assert.NotNil(t, err)
}

func TestAccountsDBRetrieveCodeNotFoundShouldRetNil(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)
	as.SetCodeHashNoJournal(hasher.Compute("12345"))

	adb := accountsDBCreateAccountsDB()

	//should return nil
	err := adb.RetrieveCode(as)
	assert.Nil(t, err)
	assert.Nil(t, as.Code())
}

func TestAccountsDBRetrieveCodeFoundShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)
	as.SetCodeHashNoJournal(hasher.Compute("12345"))

	adb := accountsDBCreateAccountsDB()
	//manually put it in the trie
	err := adb.MainTrie.Update(as.CodeHash(), []byte{68, 69, 70})
	assert.Nil(t, err)

	//should return nil
	err = adb.RetrieveCode(as)
	assert.Nil(t, err)
	assert.Equal(t, []byte{68, 69, 70}, as.Code())
}

func TestAccountsDBRetrieveDataNilRootShouldRetNil(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()

	//since root is nil, result should be nil and data trie should be nil
	err := adb.RetrieveDataTrie(as)
	assert.Nil(t, err)
	assert.Nil(t, as.DataTrie())
}

func TestAccountsDBRetrieveDataNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)
	as.SetRootHashNoJournal([]byte("12345"))

	adb := accountsDBCreateAccountsDB()

	adb.MainTrie = nil

	err := adb.RetrieveDataTrie(as)
	assert.NotNil(t, err)
}

func TestAccountsDBRetrieveDataBadLengthShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)
	as.SetRootHashNoJournal([]byte("12345"))

	adb := accountsDBCreateAccountsDB()

	//should return error
	err := adb.RetrieveDataTrie(as)
	assert.NotNil(t, err)
	fmt.Println(err.Error())
}

func TestAccountsDBRetrieveDataMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)
	as.SetRootHashNoJournal(hasher.Compute("12345"))

	adb := accountsDBCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).Fail = true

	//should return error
	err := adb.RetrieveDataTrie(as)
	assert.NotNil(t, err)
}

func TestAccountsDBRetrieveDataNotFoundRootShouldReturnErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)
	as.SetRootHashNoJournal(hasher.Compute("12345"))

	adb := accountsDBCreateAccountsDB()

	//should return error
	err := adb.RetrieveDataTrie(as)
	assert.NotNil(t, err)
	fmt.Println(err.Error())
}

func TestAccountsDBRetrieveDataWithSomeValuesShouldWork(t *testing.T) {
	//test simulates creation of a new account, data trie retrieval,
	//adding a (key, value) pair in that data trie, commiting changes
	//and then reloading the data trie based on the root hash generated before
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()
	//create a data trie with some values
	dataTrie, err := adb.MainTrie.Recreate(make([]byte, encoding.HashLength), adb.MainTrie.DBW())
	assert.Nil(t, err)
	err = dataTrie.Update([]byte{65, 66, 67}, []byte{32, 33, 34})
	assert.Nil(t, err)
	err = dataTrie.Update([]byte{68, 69, 70}, []byte{35, 36, 37})
	assert.Nil(t, err)
	_, err = dataTrie.Commit(nil)
	assert.Nil(t, err)

	//link the new created trie to account's data root
	as.SetRootHashNoJournal(dataTrie.Root())

	//should not return error
	err = adb.RetrieveDataTrie(as)
	assert.Nil(t, err)
	assert.NotNil(t, as.DataTrie())

	//verify data
	data, err := as.RetrieveValue([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{32, 33, 34}, data)

	data, err = as.RetrieveValue([]byte{68, 69, 70})
	assert.Nil(t, err)
	assert.Equal(t, []byte{35, 36, 37}, data)
}

func TestAccountsDBPutCodeNilCodeHashShouldRetNil(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()

	err := adb.PutCode(as, nil)
	assert.Nil(t, err)
	assert.Nil(t, as.CodeHash())
	assert.Nil(t, as.Code())
}

func TestAccountsDBPutCodeEmptyCodeHashShouldRetNil(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()

	err := adb.PutCode(as, make([]byte, 0))
	assert.Nil(t, err)
	assert.Nil(t, as.CodeHash())
	assert.Nil(t, as.Code())
}

func TestAccountsDBPutCodeNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()

	adb.MainTrie = nil

	err := adb.PutCode(as, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDBPutCodeMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).Fail = true

	//should return error
	err := adb.PutCode(as, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDBPutCodeMalfunction2TrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).FailUpdate = true

	//should return error
	err := adb.PutCode(as, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDBPutCodeWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()

	snapshotRoot := adb.MainTrie.Root()

	err := adb.PutCode(as, []byte("Smart contract code"))
	assert.Nil(t, err)
	assert.NotNil(t, as.CodeHash())
	assert.Equal(t, []byte("Smart contract code"), as.Code())

	fmt.Printf("SC code is at address: %v\n", as.CodeHash())

	//retrieve directly from the trie
	data, err := adb.MainTrie.Get(as.CodeHash())
	assert.Nil(t, err)
	assert.Equal(t, data, as.Code())

	fmt.Printf("SC code is: %v\n", string(data))

	//main root trie should have been modified
	assert.NotEqual(t, snapshotRoot, adb.MainTrie.Root())
}

func TestAccountsDBSaveDataNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()
	as.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})

	adb.MainTrie = nil

	err := adb.SaveData(as)
	assert.NotNil(t, err)
}

func TestAccountsDBSaveDataNoDirtyShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()

	err := adb.SaveData(as)
	assert.Nil(t, err)
	assert.Equal(t, 0, adb.JournalLen())
}

func TestAccountsDBSaveDataMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()
	as.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	adb.MainTrie.(*mock.TrieMock).FailRecreate = true

	err := adb.SaveData(as)
	assert.NotNil(t, err)
}

func TestAccountsDBHasAccountStateNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)

	adb := accountsDBCreateAccountsDB()
	adb.MainTrie = nil

	val, err := adb.HasAccountState(adr)
	assert.NotNil(t, err)
	assert.False(t, val)
}

func TestAccountsDBHasAccountStateMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)

	adb := accountsDBCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).Fail = true

	//should return error
	val, err := adb.HasAccountState(adr)
	assert.NotNil(t, err)
	assert.False(t, val)
}

func TestAccountsDBHasAccountStateNotFoundShouldRetFalse(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)

	adb := accountsDBCreateAccountsDB()

	//should return false
	val, err := adb.HasAccountState(adr)
	assert.Nil(t, err)
	assert.False(t, val)
}

func TestAccountsDBHasAccountStateFoundShouldRetTrue(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)

	adb := accountsDBCreateAccountsDB()
	err := adb.MainTrie.Update(hasher.Compute(string(adr.Bytes())), []byte{65})
	assert.Nil(t, err)

	//should return true
	val, err := adb.HasAccountState(adr)
	assert.Nil(t, err)
	assert.True(t, val)
}

func TestAccountsDBSaveAccountStateNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()
	adb.MainTrie = nil

	err := adb.SaveAccountState(as)
	assert.NotNil(t, err)
}

func TestAccountsDBSaveAccountStateNilStateShouldErr(t *testing.T) {
	t.Parallel()

	adb := accountsDBCreateAccountsDB()
	adb.MainTrie = nil

	err := adb.SaveAccountState(nil)
	assert.NotNil(t, err)
}

func TestAccountsDBSaveAccountStateMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).Fail = true

	//should return error
	err := adb.SaveAccountState(as)
	assert.NotNil(t, err)
}

func TestAccountsDBSaveAccountStateMalfunctionMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()
	adb.Marsh().(*mock.MarshalizerMock).Fail = true
	defer func() {
		//call defer to reset the fail pointer so other tests will work as expected
		adb.Marsh().(*mock.MarshalizerMock).Fail = false
	}()

	//should return error
	err := adb.SaveAccountState(as)

	assert.NotNil(t, err)
}

func TestAccountsDBSaveAccountStateWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()

	//should return error
	err := adb.SaveAccountState(as)
	assert.Nil(t, err)
}

func TestAccountsDBGetOrCreateAccountStateNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)

	adb := accountsDBCreateAccountsDB()
	adb.MainTrie = nil

	acnt, err := adb.GetOrCreateAccountState(adr)
	assert.NotNil(t, err)
	assert.Nil(t, acnt)

}

func TestAccountsDBGetOrCreateAccountStateMalfunctionTrieShouldErr(t *testing.T) {
	//test when there is no such account

	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)

	adb := accountsDBCreateAccountsDB()
	adb.MainTrie.(*mock.TrieMock).FailGet = true
	adb.MainTrie.(*mock.TrieMock).TTFGet = 1
	adb.MainTrie.(*mock.TrieMock).FailUpdate = true
	adb.MainTrie.(*mock.TrieMock).TTFUpdate = 0

	acnt, err := adb.GetOrCreateAccountState(adr)
	assert.NotNil(t, err)
	assert.Nil(t, acnt)

}

func TestAccountsDBGetOrCreateAccountStateMalfunction2TrieShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()
	err := adb.SaveAccountState(as)
	assert.Nil(t, err)

	adb.MainTrie.(*mock.TrieMock).FailGet = true
	adb.MainTrie.(*mock.TrieMock).TTFGet = 0

	acnt, err := adb.GetOrCreateAccountState(adr)
	assert.NotNil(t, err)
	assert.Nil(t, acnt)

}

func TestAccountsDBGetOrCreateAccountStateMalfunctionMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)

	adb := accountsDBCreateAccountsDB()
	err := adb.SaveAccountState(as)
	assert.Nil(t, err)

	adb.Marsh().(*mock.MarshalizerMock).Fail = true
	defer func() {
		adb.Marsh().(*mock.MarshalizerMock).Fail = false
	}()

	acnt, err := adb.GetOrCreateAccountState(adr)
	assert.NotNil(t, err)
	assert.Nil(t, acnt)

}

func TestAccountsDBGetOrCreateAccountStateReturnExistingAccntShouldWork(t *testing.T) {
	//test when the account exists
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adb := accountsDBCreateAccountsDB()

	adr := accountsDBCreateAddress(t, testHash)
	as := accountsDBCreateAccountState(t, adr)
	err := as.SetBalance(adb, big.NewInt(40))
	assert.Nil(t, err)

	err = adb.SaveAccountState(as)
	assert.Nil(t, err)

	acnt, err := adb.GetOrCreateAccountState(adr)
	assert.Nil(t, err)
	assert.NotNil(t, acnt)
	assert.Equal(t, acnt.Balance(), *big.NewInt(40))
}

func TestAccountsDBGetOrCreateAccountStateReturnNotFoundAccntShouldWork(t *testing.T) {
	//test when the account does not exists
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash := hasher.Compute("ABCDEFGHIJKLMNOP")

	adb := accountsDBCreateAccountsDB()

	adr := accountsDBCreateAddress(t, testHash)

	//same address of the unsaved account
	acnt, err := adb.GetOrCreateAccountState(adr)
	assert.Nil(t, err)
	assert.NotNil(t, acnt)
	assert.Equal(t, acnt.Balance(), *big.NewInt(0))
}

func TestAccountsDBCommitMalfunctionTrieShouldErr(t *testing.T) {
	//test creates 2 accounts (one with a data root)
	//verifies that commit saves the new tries and that can be loaded back
	//second data trie malfunctions
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	adr1 := accountsDBCreateAddress(t, testHash1)

	testHash2 := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr2 := accountsDBCreateAddress(t, testHash2)

	adb := accountsDBCreateAccountsDB()

	//first account has the balance of 40
	state1, err := adb.GetOrCreateAccountState(adr1)
	assert.Nil(t, err)
	err = state1.SetBalance(adb, big.NewInt(40))
	assert.Nil(t, err)

	//second account has the balance of 50 and some data
	state2, err := adb.GetOrCreateAccountState(adr2)
	assert.Nil(t, err)

	err = state2.SetBalance(adb, big.NewInt(50))
	assert.Nil(t, err)
	state2.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	err = adb.SaveData(state2)
	assert.Nil(t, err)

	//states are now prepared, committing

	state2.DataTrie().(*mock.TrieMock).Fail = true

	_, err = adb.Commit()
	assert.NotNil(t, err)
}

func TestAccountsDBCommitMalfunctionMainTrieShouldErr(t *testing.T) {
	//test creates 2 accounts (one with a data root)
	//verifies that commit saves the new tries and that can be loaded back
	//main trie malfunctions
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	adr1 := accountsDBCreateAddress(t, testHash1)

	testHash2 := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr2 := accountsDBCreateAddress(t, testHash2)

	adb := accountsDBCreateAccountsDB()

	//first account has the balance of 40
	state1, err := adb.GetOrCreateAccountState(adr1)
	assert.Nil(t, err)
	err = state1.SetBalance(adb, big.NewInt(40))
	assert.Nil(t, err)

	//second account has the balance of 50 and some data
	state2, err := adb.GetOrCreateAccountState(adr2)
	assert.Nil(t, err)

	err = state2.SetBalance(adb, big.NewInt(50))
	assert.Nil(t, err)
	state2.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	err = adb.SaveData(state2)
	assert.Nil(t, err)

	//states are now prepared, committing

	adb.MainTrie.(*mock.TrieMock).Fail = true

	_, err = adb.Commit()
	assert.NotNil(t, err)
}

func TestAccountsDBCommit2okAccountsShouldWork(t *testing.T) {
	//test creates 2 accounts (one with a data root)
	//verifies that commit saves the new tries and that can be loaded back
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	adr1 := accountsDBCreateAddress(t, testHash1)

	testHash2 := hasher.Compute("ABCDEFGHIJKLMNOPQ")
	adr2 := accountsDBCreateAddress(t, testHash2)

	adb := accountsDBCreateAccountsDB()

	//first account has the balance of 40
	state1, err := adb.GetOrCreateAccountState(adr1)
	assert.Nil(t, err)
	err = state1.SetBalance(adb, big.NewInt(40))
	assert.Nil(t, err)

	//second account has the balance of 50 and some data
	state2, err := adb.GetOrCreateAccountState(adr2)
	assert.Nil(t, err)

	err = state2.SetBalance(adb, big.NewInt(50))
	assert.Nil(t, err)
	state2.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	err = adb.SaveData(state2)
	assert.Nil(t, err)

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
	newState1, err := newAdb.GetOrCreateAccountState(adr1)
	assert.Nil(t, err)
	assert.Equal(t, newState1.Balance(), *big.NewInt(40))

	//checking state2
	newState2, err := newAdb.GetOrCreateAccountState(adr2)
	assert.Nil(t, err)
	assert.Equal(t, newState2.Balance(), *big.NewInt(50))
	assert.NotNil(t, newState2.RootHash)
	//get data
	err = adb.RetrieveDataTrie(newState2)
	assert.Nil(t, err)
	val, err := newState2.RetrieveValue([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{32, 33, 34}, val)
}

func TestAccountsDBCommitAccountDataShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}

	testHash1 := hasher.Compute("ABCDEFGHIJKLMNOP")
	adr1 := accountsDBCreateAddress(t, testHash1)

	adb := accountsDBCreateAccountsDB()
	hrEmpty := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - empty: %v\n", hrEmpty)

	state1, err := adb.GetOrCreateAccountState(adr1)
	assert.Nil(t, err)
	hrCreated := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - created account: %v\n", hrCreated)

	err = state1.SetBalance(adb, big.NewInt(40))
	assert.Nil(t, err)
	hrWithBalance := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - account with balance 40: %v\n", hrWithBalance)

	_, err = adb.Commit()
	assert.Nil(t, err)
	hrCommit := base64.StdEncoding.EncodeToString(adb.MainTrie.Root())
	fmt.Printf("State root - commited: %v\n", hrCommit)

	//commit hash == account with balance
	assert.Equal(t, hrCommit, hrWithBalance)

	err = state1.SetBalance(adb, big.NewInt(0))
	assert.Nil(t, err)

	//root hash == hrCreated
	assert.Equal(t, hrCreated, base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
	fmt.Printf("State root - account with balance 0: %v\n", base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))

	err = adb.RemoveAccount(adr1)
	assert.Nil(t, err)
	//root hash == hrEmpty
	assert.Equal(t, hrEmpty, base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
	fmt.Printf("State root - empty: %v\n", base64.StdEncoding.EncodeToString(adb.MainTrie.Root()))
}
