package state_test

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/stretchr/testify/assert"
)

func generateAccountDBFromTrie(trie trie.Trie) *state.AccountsDB {
	accnt, _ := state.NewAccountsDB(trie, mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
			return mock.NewAccountWrapMock(address, tracker), nil
		},
	})
	return accnt
}

func generateAccount() *mock.AccountWrapMock {
	adr := mock.NewAddressMock()
	return mock.NewAccountWrapMock(adr, nil)
}

func generateAddressAccountAccountsDB() (state.AddressContainer, *mock.AccountWrapMock, *state.AccountsDB) {
	adr := mock.NewAddressMock()
	account := mock.NewAccountWrapMock(adr, nil)

	adb := accountsDBCreateAccountsDB()

	return adr, account, adb
}

func accountsDBCreateAccountsDB() *state.AccountsDB {
	marshalizer := mock.MarshalizerMock{}
	adb, _ := state.NewAccountsDB(&mock.TrieStub{}, mock.HasherMock{}, &marshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
			return mock.NewAccountWrapMock(address, tracker), nil
		},
	})

	return adb
}

//------- NewAccountsDB

func TestNewAccountsDB_WithNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewAccountsDB(
		nil,
		mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
	)

	assert.Nil(t, adb)
	assert.Equal(t, state.ErrNilTrie, err)
}

func TestNewAccountsDB_WithNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewAccountsDB(
		&mock.TrieStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
	)

	assert.Nil(t, adb)
	assert.Equal(t, state.ErrNilHasher, err)
}

func TestNewAccountsDB_WithNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewAccountsDB(
		&mock.TrieStub{},
		mock.HasherMock{},
		nil,
		&mock.AccountsFactoryStub{},
	)

	assert.Nil(t, adb)
	assert.Equal(t, state.ErrNilMarshalizer, err)
}

func TestNewAccountsDB_WithNilAddressFactoryShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewAccountsDB(
		&mock.TrieStub{},
		mock.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
	)

	assert.Nil(t, adb)
	assert.Equal(t, state.ErrNilAccountFactory, err)
}

func TestNewAccountsDB_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	adb, err := state.NewAccountsDB(
		&mock.TrieStub{},
		mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
	)

	assert.NotNil(t, adb)
	assert.Nil(t, err)
}

//------- PutCode

func TestAccountsDB_PutCodeNilCodeHashShouldRetNil(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB()

	err := adb.PutCode(account, nil)
	assert.Equal(t, state.ErrNilCode, err)
}

func TestAccountsDB_PutCodeEmptyCodeShouldWork(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB()

	wasCalled := false
	account.SetCodeHashWithJournalCalled = func(codeHash []byte) error {
		wasCalled = true
		account.SetCodeHash(codeHash)
		return nil
	}

	err := adb.PutCode(account, make([]byte, 0))
	assert.Nil(t, err)
	assert.Equal(t, []byte{}, account.GetCode())
	assert.Equal(t, true, wasCalled)
}

func TestAccountsDB_PutCodeMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	account := generateAccount()
	mockTrie := &mock.TrieStub{}
	adb := generateAccountDBFromTrie(mockTrie)

	//should return error
	err := adb.PutCode(account, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDB_PutCodeMalfunction2TrieShouldErr(t *testing.T) {
	t.Parallel()

	account := generateAccount()

	mockTrie := &mock.TrieStub{}
	adb := generateAccountDBFromTrie(mockTrie)

	//should return error
	err := adb.PutCode(account, []byte{65})
	assert.NotNil(t, err)
}

func TestAccountsDB_PutCodeWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false
	account := generateAccount()
	account.SetCodeHashWithJournalCalled = func(codeHash []byte) error {
		wasCalled = true
		account.SetCodeHash(codeHash)
		return nil
	}
	mockTrie := &mock.TrieStub{}
	adb := generateAccountDBFromTrie(mockTrie)

	//snapshotRoot, err := adb.RootHash()
	//assert.Nil(t, err)

	err := adb.PutCode(account, []byte("Smart contract code"))
	assert.Nil(t, err)
	assert.NotNil(t, account.GetCodeHash())
	assert.Equal(t, []byte("Smart contract code"), account.GetCode())

	fmt.Printf("SC code is at address: %v\n", account.GetCodeHash())

	//retrieve directly from the trie
	data, err := mockTrie.Get(account.GetCodeHash())
	assert.Nil(t, err)
	assert.Equal(t, data, account.GetCode())

	fmt.Printf("SC code is: %v\n", string(data))

	//main root trie should have been modified
	//TODO fix this
	wasCalled = wasCalled
	//assert.NotEqual(t, snapshotRoot, adb.RootHash())
	//assert.True(t, wasCalled)
}

//------- RemoveAccount

func TestAccountsDB_RemoveCodeShouldWork(t *testing.T) {
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

func TestAccountsDB_SaveDataNoDirtyShouldWork(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB()
	err := adb.SaveDataTrie(account)

	assert.Nil(t, err)
	assert.Equal(t, 0, adb.JournalLen())
}

func TestAccountsDB_SaveDataMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	account := generateAccount()
	mockTrie := &mock.TrieStub{}
	adb := generateAccountDBFromTrie(mockTrie)

	err := adb.SaveDataTrie(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_SaveDataTrieShouldWork(t *testing.T) {
	t.Parallel()

	_, _, adb := generateAddressAccountAccountsDB()

	wasCalled := true

	account := generateAccount()
	account.SetRootHashWithJournalCalled = func(bytes []byte) error {
		wasCalled = true
		account.SetRootHash(bytes)
		return nil
	}
	account.DataTrieTracker().SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})

	err := adb.SaveDataTrie(account)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

//------- HasAccount

func TestAccountsDB_HasAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()

	mockTrie := &mock.TrieStub{}
	adb := generateAccountDBFromTrie(mockTrie)

	val, err := adb.HasAccount(adr)
	assert.NotNil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccountNotFoundShouldRetFalse(t *testing.T) {
	t.Parallel()

	adr, _, adb := generateAddressAccountAccountsDB()

	//should return false
	val, err := adb.HasAccount(adr)
	assert.Nil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccountFoundShouldRetTrue(t *testing.T) {
	t.Parallel()
	adr := mock.NewAddressMock()
	mockTrie := &mock.TrieStub{}

	adb := generateAccountDBFromTrie(mockTrie)

	err := mockTrie.Update(adr.Bytes(), []byte{65})
	assert.Nil(t, err)

	//should return true
	val, err := adb.HasAccount(adr)
	assert.Nil(t, err)
	assert.True(t, val)
}

//------- SaveJournalizedAccount

func TestAccountsDB_SaveJournalizedAccountNilStateShouldErr(t *testing.T) {
	t.Parallel()

	adb := generateAccountDBFromTrie(nil)

	err := adb.SaveAccount(nil)
	assert.NotNil(t, err)
}

func TestAccountsDB_SaveJournalizedAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	account := generateAccount()
	mockTrie := &mock.TrieStub{}
	adb := generateAccountDBFromTrie(mockTrie)

	//should return error
	err := adb.SaveAccount(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_SaveJournalizedAccountMalfunctionMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	account := generateAccount()
	mockTrie := &mock.TrieStub{}
	marshalizer := &mock.MarshalizerMock{}
	adb, _ := state.NewAccountsDB(mockTrie, mock.HasherMock{}, marshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
			return mock.NewAccountWrapMock(address, tracker), nil
		},
	})

	marshalizer.Fail = true

	//should return error
	err := adb.SaveAccount(account)

	assert.NotNil(t, err)
}

func TestAccountsDB_SaveJournalizedAccountWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB()

	//should return error
	err := adb.SaveAccount(account)
	assert.Nil(t, err)
}

//------- RemoveAccount

func TestAccountsDB_RemoveAccountShouldWork(t *testing.T) {
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

func TestAccountsDB_GetJournalizedAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	trieMock := &mock.TrieStub{}

	adr := mock.NewAddressMock()
	adb := generateAccountDBFromTrie(trieMock)

	_, err := adb.GetAccountWithJournal(adr)
	assert.NotNil(t, err)
}

func TestAccountsDB_GetJournalizedAccountNotFoundShouldCreateEmpty(t *testing.T) {
	t.Parallel()

	trieMock := &mock.TrieStub{}

	adr := mock.NewAddressMock()
	adb := generateAccountDBFromTrie(trieMock)

	accountExpected := mock.NewAccountWrapMock(adr, adb)
	accountRecovered, err := adb.GetAccountWithJournal(adr)

	assert.Equal(t, accountExpected, accountRecovered)
	assert.Nil(t, err)
}

//------- GetExistingAccount

func TestAccountsDB_GetExistingAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	trieMock := &mock.TrieStub{}

	adr := mock.NewAddressMock()
	adb := generateAccountDBFromTrie(trieMock)

	_, err := adb.GetExistingAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDB_GetExistingAccountNotFoundShouldRetNil(t *testing.T) {
	t.Parallel()

	trieMock := &mock.TrieStub{}

	adr := mock.NewAddressMock()
	adb := generateAccountDBFromTrie(trieMock)

	account, err := adb.GetExistingAccount(adr)
	assert.Equal(t, state.ErrAccNotFound, err)
	assert.Nil(t, account)
	//no journal entry shall be created
	assert.Equal(t, 0, adb.JournalLen())
}

func TestAccountsDB_GetExistingAccountFoundShouldRetAccount(t *testing.T) {
	t.Parallel()

	trieMock := &mock.TrieStub{}

	adr := mock.NewAddressMock()
	adb := generateAccountDBFromTrie(trieMock)

	value := 45
	//create a new account
	acnt, _ := adb.GetAccountWithJournal(adr)
	acnt.(*mock.AccountWrapMock).MockValue = value
	adb.SaveAccount(acnt)

	_, err := adb.Commit()
	assert.Nil(t, err)

	account, err := adb.GetExistingAccount(adr)
	assert.Nil(t, err)
	assert.NotNil(t, account)

	accountReal, ok := account.(*mock.AccountWrapMock)
	assert.Equal(t, true, ok)

	assert.Equal(t, value, accountReal.MockValue)
	//no journal entry shall be created
	assert.Equal(t, 0, adb.JournalLen())
}

func TestAccountsDB_GetExistingAccountConcurrentlyShouldWork(t *testing.T) {
	t.Parallel()

	trieMock := &mock.TrieStub{}

	adb := generateAccountDBFromTrie(trieMock)

	wg := sync.WaitGroup{}
	wg.Add(2000)

	addresses := make([]state.AddressContainer, 0)

	//generating 2000 different addresses
	for len(addresses) < 2000 {
		addr := mock.NewAddressMock()

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

//------- getAccount

func TestAccountsDB_GetAccountAccountNotFound(t *testing.T) {
	t.Parallel()

	trieMock := mock.TrieStub{}
	adr, _, adb := generateAddressAccountAccountsDB()

	//Step 1. Create an account + its DbAccount representation
	testAccount := mock.NewAccountWrapMock(adr, adb)
	testAccount.MockValue = 45

	//Step 2. marshalize the DbAccount
	marshalizer := mock.MarshalizerMock{}
	buff, err := marshalizer.Marshal(testAccount)
	assert.Nil(t, err)

	trieMock.GetCalled = func(key []byte) (bytes []byte, e error) {
		//whatever the key is, return the same marshalized DbAccount
		return buff, nil
	}

	adb, _ = state.NewAccountsDB(&trieMock, mock.HasherMock{}, &marshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
			return mock.NewAccountWrapMock(address, tracker), nil
		},
	})

	//Step 3. call get, should return a copy of DbAccount, recover an Account object
	recoveredAccount, err := adb.GetAccount(adr)
	assert.Nil(t, err)

	//Step 4. Let's test
	assert.Equal(t, testAccount.MockValue, recoveredAccount.(*mock.AccountWrapMock).MockValue)
}

//------- loadCode

func TestAccountsDB_LoadCodeWrongHashLengthShouldErr(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB()

	account.SetCodeHash([]byte("AAAA"))

	err := adb.LoadCode(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadCodeMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr := mock.NewAddressMock()
	account := generateAccount()
	mockTrie := &mock.TrieStub{}
	adb := generateAccountDBFromTrie(mockTrie)

	//just search a hash. Any hash will do
	account.SetCodeHash(adr.Bytes())

	err := adb.LoadCode(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadCodeOkValsShouldWork(t *testing.T) {
	t.Parallel()

	adr, account, adb := generateAddressAccountAccountsDB()

	trieStub := mock.TrieStub{}
	trieStub.GetCalled = func(key []byte) (bytes []byte, e error) {
		//will return adr.Bytes() so its hash will correspond to adr.Hash()
		return adr.Bytes(), nil
	}
	marshalizer := mock.MarshalizerMock{}
	adb, _ = state.NewAccountsDB(&trieStub, mock.HasherMock{}, &marshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
			return mock.NewAccountWrapMock(address, tracker), nil
		},
	})

	//just search a hash. Any hash will do
	account.SetCodeHash(adr.Bytes())

	err := adb.LoadCode(account)
	assert.Nil(t, err)
	assert.Equal(t, adr.Bytes(), account.GetCode())
}

//------- RetrieveData

func TestAccountsDB_LoadDataNilRootShouldRetNil(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB()

	//since root is nil, result should be nil and data trie should be nil
	err := adb.LoadDataTrie(account)
	assert.Nil(t, err)
	assert.Nil(t, account.DataTrie())
}

func TestAccountsDB_LoadDataBadLengthShouldErr(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB()

	account.SetRootHash([]byte("12345"))

	//should return error
	err := adb.LoadDataTrie(account)
	assert.NotNil(t, err)
	fmt.Println(err.Error())
}

func TestAccountsDB_LoadDataMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	account := generateAccount()
	account.SetRootHash([]byte("12345"))

	mockTrie := &mock.TrieStub{}
	adb := generateAccountDBFromTrie(mockTrie)

	//should return error
	err := adb.LoadDataTrie(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadDataNotFoundRootShouldReturnErr(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB()

	rootHash := make([]byte, mock.HasherMock{}.Size())
	rootHash[0] = 1
	account.SetRootHash(rootHash)

	//should return error
	err := adb.LoadDataTrie(account)
	assert.NotNil(t, err)
	fmt.Println(err.Error())
}

func TestAccountsDB_LoadDataWithSomeValuesShouldWork(t *testing.T) {
	//test simulates creation of a new account, data trie retrieval,
	//adding a (key, value) pair in that data trie, commiting changes
	//and then reloading the data trie based on the root hash generated before
	t.Parallel()

	account := generateAccount()
	//mockTrie := &mock.TrieStub{}
	//adb := generateAccountDBFromTrie(mockTrie)

	rootHash := make([]byte, mock.HasherMock{}.Size())
	rootHash[0] = 1
	account.SetRootHash(rootHash)

	//create a data trie with some values
	//TODO fix this
	//dataTrie, err := mockTrie.Recreate(make([]byte, encoding.HashLength), mockTrie.DBW())
	//assert.Nil(t, err)
	//err = dataTrie.Update([]byte{65, 66, 67}, []byte{32, 33, 34})
	//assert.Nil(t, err)
	//err = dataTrie.Update([]byte{68, 69, 70}, []byte{35, 36, 37})
	//assert.Nil(t, err)
	//_, err = dataTrie.Commit(nil)
	//assert.Nil(t, err)
	//
	////link the new created trie to account's data root
	//account.SetRootHash(dataTrie.Root())
	//
	////should not return error
	//err = adb.LoadDataTrie(account)
	//assert.Nil(t, err)
	//assert.NotNil(t, account.DataTrie())
	//
	////verify data
	//data, err := account.DataTrieTracker().RetrieveValue([]byte{65, 66, 67})
	//assert.Nil(t, err)
	//assert.Equal(t, []byte{32, 33, 34}, data)
	//
	//data, err = account.DataTrieTracker().RetrieveValue([]byte{68, 69, 70})
	//assert.Nil(t, err)
	//assert.Equal(t, []byte{35, 36, 37}, data)
}

//------- Commit

func TestAccountsDB_CommitShouldCallCommitFromTrie(t *testing.T) {
	t.Parallel()

	account := generateAccount()
	commitCalled := 0

	trieStub := mock.TrieStub{
		CommitCalled: func() error {
			commitCalled++

			return nil
		},
		RootCalled: func() (i []byte, e error) {
			return nil, nil
		},
	}

	adb := generateAccountDBFromTrie(&trieStub)

	entry, _ := state.NewBaseJournalEntryData(account, &trieStub)
	adb.Journalize(entry)

	_, err := adb.Commit()
	assert.Nil(t, err)
	//one commit for the JournalEntryData and one commit for the main trie
	assert.Equal(t, 2, commitCalled)
}

//------- RecreateTrie

func TestAccountsDB_RecreateTrieMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	wasCalled := false

	errExpected := errors.New("failure")
	trieStub := mock.TrieStub{}
	trieStub.RecreateCalled = func(root []byte) (tree trie.Trie, e error) {
		wasCalled = true
		return nil, errExpected
	}

	adb := generateAccountDBFromTrie(&trieStub)

	err := adb.RecreateTrie(nil)
	assert.Equal(t, errExpected, err)
	assert.True(t, wasCalled)
}

func TestAccountsDB_RecreateTrieOutputsNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	wasCalled := false

	trieStub := mock.TrieStub{}
	trieStub.RecreateCalled = func(root []byte) (tree trie.Trie, e error) {
		wasCalled = true
		return nil, nil
	}

	adb := generateAccountDBFromTrie(&trieStub)
	err := adb.RecreateTrie(nil)

	assert.Equal(t, state.ErrNilTrie, err)
	assert.True(t, wasCalled)

}

func TestAccountsDB_RecreateTrieOkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	trieStub := mock.TrieStub{}
	trieStub.RecreateCalled = func(root []byte) (tree trie.Trie, e error) {
		wasCalled = true
		return &mock.TrieStub{}, nil
	}

	adb := generateAccountDBFromTrie(&trieStub)
	err := adb.RecreateTrie(nil)

	assert.Nil(t, err)
	assert.True(t, wasCalled)

}
