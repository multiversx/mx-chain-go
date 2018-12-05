package state_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func generateAddressJurnalAccountAccountsDB() (state.AddressContainer, *mock.JournalizedAccountWrapMock, *state.AccountsDB) {
	adr := mock.NewAddressMock()
	jem := mock.NewJournalizedAccountWrapMock(adr)

	adb := accountsDBCreateAccountsDB()

	return adr, jem, adb
}

func accountsDBCreateAccountsDB() *state.AccountsDB {
	marshalizer := mock.MarshalizerMock{}
	adb, _ := state.NewAccountsDB(mock.NewMockTrie(), mock.HasherMock{}, &marshalizer)

	return adb
}

func TestNewAccountsDBWithNilTrieShouldErr(t *testing.T) {
	_, err := state.NewAccountsDB(nil, mock.HasherMock{}, &mock.MarshalizerMock{})
	assert.NotNil(t, err)
}

func TestNewAccountsDBWithNilHasherShouldErr(t *testing.T) {
	_, err := state.NewAccountsDB(mock.NewMockTrie(), nil, &mock.MarshalizerMock{})
	assert.NotNil(t, err)
}

func TestNewAccountsDBWithNilMarshalizerShouldErr(t *testing.T) {
	_, err := state.NewAccountsDB(mock.NewMockTrie(), mock.HasherMock{}, nil)
	assert.NotNil(t, err)
}

//------- PutCode

func TestAccountsDBPutCodeNilCodeHashShouldRetNil(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	err := adb.PutCode(jem, nil)
	assert.Nil(t, err)
	assert.Nil(t, jem.CodeHash())
	assert.Nil(t, jem.Code())
}

func TestAccountsDBPutCodeEmptyCodeHashShouldRetNil(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	err := adb.PutCode(jem, make([]byte, 0))
	assert.Nil(t, err)
	assert.Nil(t, jem.CodeHash())
	assert.Nil(t, jem.Code())
}

func TestAccountsDBPutCodeNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	adb.SetMainTrie(nil)

	err := adb.PutCode(jem, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDBPutCodeMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()
	adb.MainTrie().(*mock.TrieMock).Fail = true

	//should return error
	err := adb.PutCode(jem, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDBPutCodeMalfunction2TrieShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()
	adb.MainTrie().(*mock.TrieMock).FailUpdate = true

	//should return error
	err := adb.PutCode(jem, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDBPutCodeWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	snapshotRoot := adb.RootHash()

	err := adb.PutCode(jem, []byte("Smart contract code"))
	assert.Nil(t, err)
	assert.NotNil(t, jem.CodeHash())
	assert.Equal(t, []byte("Smart contract code"), jem.Code())

	fmt.Printf("SC code is at address: %v\n", jem.CodeHash())

	//retrieve directly from the trie
	data, err := adb.MainTrie().Get(jem.CodeHash())
	assert.Nil(t, err)
	assert.Equal(t, data, jem.Code())

	fmt.Printf("SC code is: %v\n", string(data))

	//main root trie should have been modified
	assert.NotEqual(t, snapshotRoot, adb.RootHash())
}

//------- RemoveAccount

func TestAccountsDBRemoveCodeNilStateShouldErr(t *testing.T) {
	t.Parallel()

	adb := accountsDBCreateAccountsDB()
	adb.SetMainTrie(nil)

	err := adb.RemoveCode([]byte("AAA"))
	assert.NotNil(t, err)
}

func TestAccountsDBRemoveCodeShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	trieStub := mock.TrieStub{}
	trieStub.UpdateCalled = func(key, value []byte) error {
		wasCalled = true
		return nil
	}

	adb := accountsDBCreateAccountsDB()
	adb.SetMainTrie(&trieStub)

	err := adb.RemoveCode([]byte("AAA"))
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

//------- SaveData

func TestAccountsDBSaveDataNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()
	jem.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})

	adb.SetMainTrie(nil)

	err := adb.SaveData(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBSaveDataNoDirtyShouldWork(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	err := adb.SaveData(jem)
	assert.Nil(t, err)
	assert.Equal(t, 0, adb.JournalLen())
}

func TestAccountsDBSaveDataMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	jem.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	adb.MainTrie().(*mock.TrieMock).FailRecreate = true

	err := adb.SaveData(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBSaveDataShouldWork(t *testing.T) {
	t.Parallel()

	adr, _, adb := generateAddressJurnalAccountAccountsDB()

	account := state.NewAccount()
	jaw, err := state.NewJournalizedAccountWrapFromAccountContainer(adr, account, adb)
	assert.Nil(t, err)
	jaw.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})

	err = adb.SaveData(jaw)
	assert.Nil(t, err)
}

//------- HasAccount

func TestAccountsDBHasAccountNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr, _, adb := generateAddressJurnalAccountAccountsDB()
	adb.SetMainTrie(nil)

	val, err := adb.HasAccount(adr)
	assert.NotNil(t, err)
	assert.False(t, val)
}

func TestAccountsDBHasAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr, _, adb := generateAddressJurnalAccountAccountsDB()
	adb.MainTrie().(*mock.TrieMock).Fail = true

	val, err := adb.HasAccount(adr)
	assert.NotNil(t, err)
	assert.False(t, val)
}

func TestAccountsDBHasAccountNotFoundShouldRetFalse(t *testing.T) {
	t.Parallel()

	adr, _, adb := generateAddressJurnalAccountAccountsDB()

	//should return false
	val, err := adb.HasAccount(adr)
	assert.Nil(t, err)
	assert.False(t, val)
}

func TestAccountsDBHasAccountFoundShouldRetTrue(t *testing.T) {
	t.Parallel()

	adr, _, adb := generateAddressJurnalAccountAccountsDB()

	err := adb.MainTrie().Update(mock.HasherMock{}.Compute(string(adr.Bytes())), []byte{65})
	assert.Nil(t, err)

	//should return true
	val, err := adb.HasAccount(adr)
	assert.Nil(t, err)
	assert.True(t, val)
}

//------- SaveJournalizedAccount

func TestAccountsDBSaveJournalizedAccountNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()
	adb.SetMainTrie(nil)

	err := adb.SaveJournalizedAccount(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBSaveJournalizedAccountNilStateShouldErr(t *testing.T) {
	t.Parallel()

	adb := accountsDBCreateAccountsDB()
	adb.SetMainTrie(nil)

	err := adb.SaveJournalizedAccount(nil)
	assert.NotNil(t, err)
}

func TestAccountsDBSaveJournalizedAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	adb.MainTrie().(*mock.TrieMock).Fail = true

	wasCalled := false
	jem.SaveToDbAccountCalled = func(target state.DbAccountContainer) error {
		wasCalled = true
		return nil
	}

	//should return error
	err := adb.SaveJournalizedAccount(jem)
	assert.NotNil(t, err)
	assert.True(t, wasCalled)
}

func TestAccountsDBSaveJournalizedAccountMalfunctionSaveToDBShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()
	adb.MainTrie().(*mock.TrieMock).Fail = true

	wasCalled := false
	jem.SaveToDbAccountCalled = func(target state.DbAccountContainer) error {
		wasCalled = true
		return errors.New("failure")
	}

	//should return error
	err := adb.SaveJournalizedAccount(jem)
	assert.NotNil(t, err)
	assert.True(t, wasCalled)
}

func TestAccountsDBSaveJournalizedAccountMalfunctionMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	wasCalled := false

	jem.SaveToDbAccountCalled = func(target state.DbAccountContainer) error {
		wasCalled = true
		return nil
	}

	adb.Marshalizer().(*mock.MarshalizerMock).Fail = true

	//should return error
	err := adb.SaveJournalizedAccount(jem)

	assert.NotNil(t, err)
	assert.True(t, wasCalled)
}

func TestAccountsDBSaveJournalizedAccountWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	wasCalled := false

	jem.SaveToDbAccountCalled = func(target state.DbAccountContainer) error {
		wasCalled = true
		return nil
	}

	//should return error
	err := adb.SaveJournalizedAccount(jem)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

//------- RemoveAccount

func TestAccountsDBRemoveAccountNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr, _, adb := generateAddressJurnalAccountAccountsDB()
	adb.SetMainTrie(nil)

	err := adb.RemoveAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDBRemoveAccountShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	trieStub := mock.TrieStub{}
	trieStub.UpdateCalled = func(key, value []byte) error {
		wasCalled = true
		return nil
	}

	adr, _, adb := generateAddressJurnalAccountAccountsDB()
	adb.SetMainTrie(&trieStub)

	err := adb.RemoveAccount(adr)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

//------- GetJournalizedAccount

func TestAccountsDBGetJournalizedAccountNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr, _, adb := generateAddressJurnalAccountAccountsDB()
	adb.SetMainTrie(nil)

	_, err := adb.GetJournalizedAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDBGetJournalizedAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	trieMock := mock.NewMockTrie()
	trieMock.FailGet = true

	adr, _, adb := generateAddressJurnalAccountAccountsDB()
	adb.MainTrie().(*mock.TrieMock).FailGet = true

	_, err := adb.GetJournalizedAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDBGetJournalizedAccountNotFoundShouldCreateEmpty(t *testing.T) {
	t.Parallel()

	trieMock := mock.NewMockTrie()

	adr, _, adb := generateAddressJurnalAccountAccountsDB()
	adb.SetMainTrie(trieMock)

	account, err := adb.GetJournalizedAccount(adr)
	assert.Nil(t, err)

	assert.Equal(t, uint64(0), account.Nonce())
	assert.Equal(t, *big.NewInt(0), account.Balance())
	assert.Equal(t, []byte(nil), account.CodeHash())
	assert.Equal(t, []byte(nil), account.RootHash())
	assert.Equal(t, adr, account.AddressContainer())
}

//------- getDbAccount

func TestAccountsDBGetDbAccountAccountNotFound(t *testing.T) {
	t.Parallel()

	trieMock := mock.TrieStub{}

	//Step 1. Create an account + its DbAccount representation
	testAccount := state.NewAccount()
	testAccount.SetNonce(1)
	testAccount.SetBalance(*big.NewInt(45))

	testDbAccount := state.DbAccount{}
	err := testAccount.SaveToDbAccount(&testDbAccount)
	assert.Nil(t, err)

	//Step 2. marshalize the DbAccount
	marshalizer := mock.MarshalizerMock{}
	buff, err := marshalizer.Marshal(testDbAccount)
	assert.Nil(t, err)

	trieMock.GetCalled = func(key []byte) (bytes []byte, e error) {
		//whatever the key is, return the same marshalized DbAccount
		return buff, nil
	}

	adr, _, adb := generateAddressJurnalAccountAccountsDB()
	adb.SetMainTrie(&trieMock)

	//Step 3. call get, should return a copy of DbAccount, recover an Account object
	dbAccount, err := adb.GetDbAccount(adr)
	assert.Nil(t, err)
	recoveredAccount := state.NewAccount()
	err = recoveredAccount.LoadFromDbAccount(dbAccount)
	assert.Nil(t, err)

	//Step 4. Let's test
	assert.Equal(t, testAccount.Nonce(), recoveredAccount.Nonce())
	assert.Equal(t, testAccount.Balance(), recoveredAccount.Balance())

}

//------- loadCode
func TestAccountsDBLoadCodeNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr, jem, adb := generateAddressJurnalAccountAccountsDB()

	adb.SetMainTrie(nil)

	//just search a hash. Any hash will do
	jem.SetCodeHash(adr.Hash())

	err := adb.LoadCode(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBLoadCodeWrongHashLengthShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	jem.SetCodeHash([]byte("AAAA"))

	err := adb.LoadCode(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBLoadCodeMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr, jem, adb := generateAddressJurnalAccountAccountsDB()

	adb.MainTrie().(*mock.TrieMock).FailGet = true

	//just search a hash. Any hash will do
	jem.SetCodeHash(adr.Hash())

	err := adb.LoadCode(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBLoadCodeWrongCodeShouldErr(t *testing.T) {
	t.Parallel()

	adr, jem, adb := generateAddressJurnalAccountAccountsDB()

	trieStub := mock.TrieStub{}
	trieStub.GetCalled = func(key []byte) (bytes []byte, e error) {
		//will return something different of adr.Hash()
		return []byte("AAA"), nil
	}
	adb.SetMainTrie(&trieStub)

	//just search a hash. Any hash will do
	jem.SetCodeHash(adr.Hash())

	err := adb.LoadCode(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBLoadCodeOkValsShouldWork(t *testing.T) {
	t.Parallel()

	adr, jem, adb := generateAddressJurnalAccountAccountsDB()

	trieStub := mock.TrieStub{}
	trieStub.GetCalled = func(key []byte) (bytes []byte, e error) {
		//will return adr.Bytes() so its hash will correspond to adr.Hash()
		return adr.Bytes(), nil
	}
	adb.SetMainTrie(&trieStub)

	//just search a hash. Any hash will do
	jem.SetCodeHash(adr.Hash())

	err := adb.LoadCode(jem)
	assert.Nil(t, err)
	assert.Equal(t, adr.Bytes(), jem.Code())
}

//------- RetrieveData

func TestAccountsDBLoadDataNilRootShouldRetNil(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	//since root is nil, result should be nil and data trie should be nil
	err := adb.LoadDataTrie(jem)
	assert.Nil(t, err)
	assert.Nil(t, jem.DataTrie())
}

func TestAccountsDBLoadDataNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()
	jem.SetRootHash([]byte("12345"))

	adb.SetMainTrie(nil)

	err := adb.LoadDataTrie(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBLoadDataBadLengthShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	jem.SetRootHash([]byte("12345"))

	//should return error
	err := adb.LoadDataTrie(jem)
	assert.NotNil(t, err)
	fmt.Println(err.Error())
}

func TestAccountsDBLoadDataMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	jem.SetRootHash([]byte("12345"))

	adb.MainTrie().(*mock.TrieMock).Fail = true

	//should return error
	err := adb.LoadDataTrie(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBLoadDataNotFoundRootShouldReturnErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	rootHash := make([]byte, state.AdrLen)
	rootHash[0] = 1
	jem.SetRootHash(rootHash)

	//should return error
	err := adb.LoadDataTrie(jem)
	assert.NotNil(t, err)
	fmt.Println(err.Error())
}

func TestAccountsDBLoadDataWithSomeValuesShouldWork(t *testing.T) {
	//test simulates creation of a new account, data trie retrieval,
	//adding a (key, value) pair in that data trie, commiting changes
	//and then reloading the data trie based on the root hash generated before
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	rootHash := make([]byte, state.AdrLen)
	rootHash[0] = 1
	jem.SetRootHash(rootHash)

	//create a data trie with some values
	dataTrie, err := adb.MainTrie().Recreate(make([]byte, encoding.HashLength), adb.MainTrie().DBW())
	assert.Nil(t, err)
	err = dataTrie.Update([]byte{65, 66, 67}, []byte{32, 33, 34})
	assert.Nil(t, err)
	err = dataTrie.Update([]byte{68, 69, 70}, []byte{35, 36, 37})
	assert.Nil(t, err)
	_, err = dataTrie.Commit(nil)
	assert.Nil(t, err)

	//link the new created trie to account's data root
	jem.SetRootHash(dataTrie.Root())

	//should not return error
	err = adb.LoadDataTrie(jem)
	assert.Nil(t, err)
	assert.NotNil(t, jem.DataTrie())

	//verify data
	data, err := jem.RetrieveValue([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{32, 33, 34}, data)

	data, err = jem.RetrieveValue([]byte{68, 69, 70})
	assert.Nil(t, err)
	assert.Equal(t, []byte{35, 36, 37}, data)
}

//------- Commit

func TestAccountsDBCommitWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	commitCalled := 0

	trieStub := mock.TrieStub{}
	trieStub.CommitCalled = func(onleaf trie.LeafCallback) (root []byte, err error) {
		commitCalled++

		return nil, nil
	}
	adb.SetMainTrie(&trieStub)

	adb.AddJournalEntry(state.NewJournalEntryData(jem, &trieStub))

	_, err := adb.Commit()
	assert.Nil(t, err)
	//one commit for the JournalEntryData and one commit for the main trie
	assert.Equal(t, 2, commitCalled)
}

//------- Functionality test

func TestAccountsDBTestCreateModifyComitSaveGet(t *testing.T) {
	t.Parallel()

	trieMock := mock.NewMockTrie()

	adr, _, adb := generateAddressJurnalAccountAccountsDB()

	adb.SetMainTrie(trieMock)

	//Step 1. get a fresh new account
	journalizedAccount, err := adb.GetJournalizedAccount(adr)
	assert.Nil(t, err)
	err = journalizedAccount.SetNonceWithJournal(34)
	assert.Nil(t, err)
	err = journalizedAccount.SetBalanceWithJournal(*big.NewInt(45))
	assert.Nil(t, err)
	err = adb.PutCode(journalizedAccount, []byte("Test SC code to be executed"))
	assert.Nil(t, err)
	journalizedAccount.SaveKeyValue([]byte("a key"), []byte("a value"))
	err = adb.SaveData(journalizedAccount)
	assert.Nil(t, err)

	//Step 2. Commit, as to save the data inside the trie
	rootHash, err := adb.Commit()
	assert.Nil(t, err)
	fmt.Printf("Root hash: %v\n", rootHash)

	//Step 3. Load data and test
	recoveredAccount, err := adb.GetJournalizedAccount(adr)
	assert.Nil(t, err)

	assert.Equal(t, uint64(34), recoveredAccount.Nonce())
	assert.Equal(t, *big.NewInt(45), recoveredAccount.Balance())
	assert.Equal(t, []byte("Test SC code to be executed"), recoveredAccount.Code())
	value, err := recoveredAccount.RetrieveValue([]byte("a key"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("a value"), value)
}
