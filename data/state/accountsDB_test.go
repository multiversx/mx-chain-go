package state_test

import (
	"fmt"
	"github.com/pkg/errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/stretchr/testify/assert"
)

func generateAccountDBFromTrie(trie trie.PatriciaMerkelTree) state.AccountsDB {
	return state.Generate(trie, mock.HasherMock{}, &mock.MarshalizerMock{}, state.NewJournal())
}

func generateAccountDBFromTrieAndMarshalizer(trie trie.PatriciaMerkelTree,
	marshalizer marshal.Marshalizer) state.AccountsDB {
	return state.Generate(trie, mock.HasherMock{}, marshalizer, state.NewJournal())
}

func generateJem() *mock.JournalizedAccountWrapMock {
	adr := mock.NewAddressMock()
	return mock.NewJournalizedAccountWrapMock(adr)
}

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
	assert.Nil(t, jem.CodeHash)
	assert.Nil(t, jem.Code())
}

func TestAccountsDBPutCodeEmptyCodeHashShouldRetNil(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	err := adb.PutCode(jem, make([]byte, 0))
	assert.Nil(t, err)
	assert.Nil(t, jem.CodeHash)
	assert.Nil(t, jem.Code())
}

func TestAccountsDBPutCodeNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	jem := generateJem()
	adb := generateAccountDBFromTrie(nil)

	err := adb.PutCode(jem, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDBPutCodeMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	jem := generateJem()
	mockTrie := mock.NewMockTrie()
	adb := generateAccountDBFromTrie(mockTrie)

	mockTrie.Fail = true

	//should return error
	err := adb.PutCode(jem, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDBPutCodeMalfunction2TrieShouldErr(t *testing.T) {
	t.Parallel()

	jem := generateJem()

	mockTrie := mock.NewMockTrie()
	adb := generateAccountDBFromTrie(mockTrie)

	mockTrie.FailUpdate = true

	//should return error
	err := adb.PutCode(jem, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDBPutCodeWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	jem := generateJem()
	mockTrie := mock.NewMockTrie()
	adb := generateAccountDBFromTrie(mockTrie)

	snapshotRoot := adb.RootHash()

	err := adb.PutCode(jem, []byte("Smart contract code"))
	assert.Nil(t, err)
	assert.NotNil(t, jem.CodeHash)
	assert.Equal(t, []byte("Smart contract code"), jem.Code())

	fmt.Printf("SC code is at address: %v\n", jem.CodeHash)

	//retrieve directly from the trie
	data, err := mockTrie.Get(jem.CodeHash)
	assert.Nil(t, err)
	assert.Equal(t, data, jem.Code())

	fmt.Printf("SC code is: %v\n", string(data))

	//main root trie should have been modified
	assert.NotEqual(t, snapshotRoot, adb.RootHash())
}

//------- RemoveAccount

func TestAccountsDBRemoveCodeNilStateShouldErr(t *testing.T) {
	t.Parallel()

	adb := generateAccountDBFromTrie(nil)

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

	adb := generateAccountDBFromTrie(&trieStub)

	err := adb.RemoveCode([]byte("AAA"))
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

//------- SaveData

func TestAccountsDBSaveDataNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	jem := generateJem()

	jem.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})

	adb := generateAccountDBFromTrie(nil)

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

	jem := generateJem()
	mockTrie := mock.NewMockTrie()
	adb := generateAccountDBFromTrie(mockTrie)

	jem.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	mockTrie.FailRecreate = true

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

	adr := mock.NewAddressMock()

	adb := generateAccountDBFromTrie(nil)

	val, err := adb.HasAccount(adr)
	assert.NotNil(t, err)
	assert.False(t, val)
}

func TestAccountsDBHasAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()

	mockTrie := mock.NewMockTrie()
	adb := generateAccountDBFromTrie(mockTrie)

	mockTrie.Fail = true

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
	adr := mock.NewAddressMock()
	mockTrie := mock.NewMockTrie()

	adb := generateAccountDBFromTrie(mockTrie)

	err := mockTrie.Update(mock.HasherMock{}.Compute(string(adr.Bytes())), []byte{65})
	assert.Nil(t, err)

	//should return true
	val, err := adb.HasAccount(adr)
	assert.Nil(t, err)
	assert.True(t, val)
}

//------- SaveJournalizedAccount

func TestAccountsDBSaveJournalizedAccountNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	jem := generateJem()
	adb := generateAccountDBFromTrie(nil)

	err := adb.SaveJournalizedAccount(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBSaveJournalizedAccountNilStateShouldErr(t *testing.T) {
	t.Parallel()

	adb := generateAccountDBFromTrie(nil)

	err := adb.SaveJournalizedAccount(nil)
	assert.NotNil(t, err)
}

func TestAccountsDBSaveJournalizedAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	jem := generateJem()
	mockTrie := mock.NewMockTrie()
	adb := generateAccountDBFromTrie(mockTrie)

	mockTrie.Fail = true

	//should return error
	err := adb.SaveJournalizedAccount(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBSaveJournalizedAccountMalfunctionMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	jem := generateJem()
	mockTrie := mock.NewMockTrie()
	marshalizer := &mock.MarshalizerMock{}
	adb := generateAccountDBFromTrieAndMarshalizer(mockTrie, marshalizer)

	marshalizer.Fail = true

	//should return error
	err := adb.SaveJournalizedAccount(jem)

	assert.NotNil(t, err)
}

func TestAccountsDBSaveJournalizedAccountWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	//should return error
	err := adb.SaveJournalizedAccount(jem)
	assert.Nil(t, err)
}

//------- RemoveAccount

func TestAccountsDBRemoveAccountNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr := &mock.AddressMock{}
	adb := generateAccountDBFromTrie(nil)

	err := adb.RemoveAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDBRemoveAccountShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	trieStub := &mock.TrieStub{}
	trieStub.UpdateCalled = func(key, value []byte) error {
		wasCalled = true
		return nil
	}

	adr := mock.NewAddressMock()
	adb := generateAccountDBFromTrie(trieStub)

	err := adb.RemoveAccount(adr)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

//------- GetJournalizedAccount

func TestAccountsDBGetJournalizedAccountNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	adb := generateAccountDBFromTrie(nil)

	_, err := adb.GetJournalizedAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDBGetJournalizedAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	trieMock := mock.NewMockTrie()

	adr := mock.NewAddressMock()
	adb := generateAccountDBFromTrie(trieMock)
	trieMock.FailGet = true

	_, err := adb.GetJournalizedAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDBGetJournalizedAccountNotFoundShouldCreateEmpty(t *testing.T) {
	t.Parallel()

	trieMock := mock.NewMockTrie()

	adr := mock.NewAddressMock()
	adb := generateAccountDBFromTrie(trieMock)

	account, err := adb.GetJournalizedAccount(adr)
	assert.Nil(t, err)

	assert.Equal(t, uint64(0), account.BaseAccount().Nonce)
	assert.Equal(t, *big.NewInt(0), account.BaseAccount().Balance)
	assert.Equal(t, []byte(nil), account.BaseAccount().CodeHash)
	assert.Equal(t, []byte(nil), account.BaseAccount().RootHash)
	assert.Equal(t, adr, account.AddressContainer())
}

//------- getAccount

func TestAccountsDBGetAccountAccountNotFound(t *testing.T) {
	t.Parallel()

	trieMock := mock.TrieStub{}

	//Step 1. Create an account + its DbAccount representation
	testAccount := state.NewAccount()
	testAccount.Nonce = 1
	testAccount.Balance = *big.NewInt(45)

	//Step 2. marshalize the DbAccount
	marshalizer := mock.MarshalizerMock{}
	buff, err := marshalizer.Marshal(testAccount)
	assert.Nil(t, err)

	trieMock.GetCalled = func(key []byte) (bytes []byte, e error) {
		//whatever the key is, return the same marshalized DbAccount
		return buff, nil
	}

	adr, _, adb := generateAddressJurnalAccountAccountsDB()

	adb, _ = state.NewAccountsDB(&trieMock, mock.HasherMock{}, &marshalizer)

	//Step 3. call get, should return a copy of DbAccount, recover an Account object
	recoveredAccount, err := adb.GetAccount(adr)
	assert.Nil(t, err)

	//Step 4. Let's test
	assert.Equal(t, testAccount.Nonce, recoveredAccount.Nonce)
	assert.Equal(t, testAccount.Balance, recoveredAccount.Balance)

}

//------- loadCode
func TestAccountsDBLoadCodeNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr, jem, _ := generateAddressJurnalAccountAccountsDB()

	adb := generateAccountDBFromTrie(nil)

	//just search a hash. Any hash will do
	jem.CodeHash = adr.Hash()

	err := adb.LoadCode(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBLoadCodeWrongHashLengthShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	jem.CodeHash = []byte("AAAA")

	err := adb.LoadCode(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBLoadCodeMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	jem := generateJem()
	mockTrie := mock.NewMockTrie()
	adb := generateAccountDBFromTrie(mockTrie)

	mockTrie.FailGet = true

	//just search a hash. Any hash will do
	jem.CodeHash = adr.Hash()

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
	marshalizer := mock.MarshalizerMock{}
	adb, _ = state.NewAccountsDB(&trieStub, mock.HasherMock{}, &marshalizer)

	//just search a hash. Any hash will do
	jem.CodeHash = adr.Hash()

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

	jem := generateJem()
	jem.RootHash = []byte("12345")

	adb := generateAccountDBFromTrie(nil)

	err := adb.LoadDataTrie(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBLoadDataBadLengthShouldErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	jem.RootHash = []byte("12345")

	//should return error
	err := adb.LoadDataTrie(jem)
	assert.NotNil(t, err)
	fmt.Println(err.Error())
}

func TestAccountsDBLoadDataMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	jem := generateJem()
	jem.RootHash = []byte("12345")

	mockTrie := mock.NewMockTrie()
	adb := generateAccountDBFromTrie(mockTrie)

	mockTrie.Fail = true

	//should return error
	err := adb.LoadDataTrie(jem)
	assert.NotNil(t, err)
}

func TestAccountsDBLoadDataNotFoundRootShouldReturnErr(t *testing.T) {
	t.Parallel()

	_, jem, adb := generateAddressJurnalAccountAccountsDB()

	rootHash := make([]byte, mock.HasherMock{}.Size())
	rootHash[0] = 1
	jem.RootHash = rootHash

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

	jem := generateJem()
	mockTrie := mock.NewMockTrie()
	adb := generateAccountDBFromTrie(mockTrie)

	rootHash := make([]byte, mock.HasherMock{}.Size())
	rootHash[0] = 1
	jem.RootHash = rootHash

	//create a data trie with some values
	dataTrie, err := mockTrie.Recreate(make([]byte, encoding.HashLength), mockTrie.DBW())
	assert.Nil(t, err)
	err = dataTrie.Update([]byte{65, 66, 67}, []byte{32, 33, 34})
	assert.Nil(t, err)
	err = dataTrie.Update([]byte{68, 69, 70}, []byte{35, 36, 37})
	assert.Nil(t, err)
	_, err = dataTrie.Commit(nil)
	assert.Nil(t, err)

	//link the new created trie to account's data root
	jem.RootHash = dataTrie.Root()

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

	jem := generateJem()

	commitCalled := 0

	trieStub := mock.TrieStub{}
	trieStub.CommitCalled = func(onleaf trie.LeafCallback) (root []byte, err error) {
		commitCalled++

		return nil, nil
	}
	adb := generateAccountDBFromTrie(&trieStub)

	adb.AddJournalEntry(state.NewJournalEntryData(jem, &trieStub))

	_, err := adb.Commit()
	assert.Nil(t, err)
	//one commit for the JournalEntryData and one commit for the main trie
	assert.Equal(t, 2, commitCalled)
}

//------- RecreateTrie

func TestAccountsDBRecreateTrieMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	wasCalled := false

	trieStub := mock.TrieStub{}
	trieStub.RecreateCalled = func(root []byte, dbw trie.DBWriteCacher) (tree trie.PatriciaMerkelTree, e error) {
		wasCalled = true
		return nil, errors.New("failure")
	}
	trieStub.DBWCalled = func() trie.DBWriteCacher {
		return nil
	}

	adb := generateAccountDBFromTrie(&trieStub)

	err := adb.RecreateTrie(nil)
	assert.NotNil(t, err)
	assert.True(t, wasCalled)
}

func TestAccountsDBRecreateTriOkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	trieStub := mock.TrieStub{}
	trieStub.RecreateCalled = func(root []byte, dbw trie.DBWriteCacher) (tree trie.PatriciaMerkelTree, e error) {
		wasCalled = true
		return nil, nil
	}
	trieStub.DBWCalled = func() trie.DBWriteCacher {
		return nil
	}

	adb := generateAccountDBFromTrie(&trieStub)

	err := adb.RecreateTrie(nil)
	assert.Nil(t, err)
	assert.True(t, wasCalled)

}

//------- Functionality test

func TestAccountsDBTestCreateModifyComitSaveGet(t *testing.T) {
	t.Parallel()

	trieMock := mock.NewMockTrie()

	adr := mock.NewAddressMock()

	adb := generateAccountDBFromTrie(trieMock)

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

	assert.Equal(t, uint64(34), recoveredAccount.BaseAccount().Nonce)
	assert.Equal(t, *big.NewInt(45), recoveredAccount.BaseAccount().Balance)
	assert.Equal(t, []byte("Test SC code to be executed"), recoveredAccount.Code())
	value, err := recoveredAccount.RetrieveValue([]byte("a key"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("a value"), value)
}
