package state_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/stretchr/testify/assert"
)

func generateAccountDBFromTrie(trie data.Trie) *state.AccountsDB {
	accnt, _ := state.NewAccountsDB(trie, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{
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

func generateAddressAccountAccountsDB(trie data.Trie) (state.AddressContainer, *mock.AccountWrapMock, *state.AccountsDB) {
	adr := mock.NewAddressMock()
	account := mock.NewAccountWrapMock(adr, nil)

	adb := generateAccountDBFromTrie(trie)

	return adr, account, adb
}

//------- NewAccountsDB

func TestNewAccountsDB_WithNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewAccountsDB(
		nil,
		&mock.HasherMock{},
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
		&mock.HasherMock{},
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
		&mock.HasherMock{},
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
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
	)

	assert.NotNil(t, adb)
	assert.Nil(t, err)
}

//------- PutCode

func TestAccountsDB_PutCodeNilCodeHashShouldRetNil(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB(&mock.TrieStub{})

	err := adb.PutCode(account, nil)
	assert.Equal(t, state.ErrNilCode, err)
}

func TestAccountsDB_PutCodeEmptyCodeShouldWork(t *testing.T) {
	t.Parallel()

	updateCalled := false
	ts := &mock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			updateCalled = true
			return nil
		},
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, nil
		},
	}
	_, account, adb := generateAddressAccountAccountsDB(ts)

	setCodeCalled := false
	account.SetCodeHashWithJournalCalled = func(codeHash []byte) error {
		setCodeCalled = true
		account.SetCodeHash(codeHash)
		return nil
	}

	err := adb.PutCode(account, make([]byte, 0))
	assert.Nil(t, err)
	assert.True(t, updateCalled)
	assert.True(t, setCodeCalled)
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

	ts := &mock.TrieStub{
		RecreateCalled: func(root []byte) (i data.Trie, e error) {
			return nil, nil
		},
	}
	_, account, adb := generateAddressAccountAccountsDB(ts)
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

	ts := &mock.TrieStub{
		RecreateCalled: func(root []byte) (i data.Trie, e error) {
			dataTrie := mock.TrieStub{
				UpdateCalled: func(key, value []byte) error {
					return nil
				},
				RootCalled: func() (i []byte, e error) {
					return make([]byte, 0), nil
				},
			}

			return &dataTrie, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		RootCalled: func() (i []byte, e error) {
			return make([]byte, 0), nil
		},
	}
	_, _, adb := generateAddressAccountAccountsDB(ts)

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

	ts := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, nil
		},
	}
	adr, _, adb := generateAddressAccountAccountsDB(ts)

	//should return false
	val, err := adb.HasAccount(adr)
	assert.Nil(t, err)
	assert.False(t, val)
}

func TestAccountsDB_HasAccountFoundShouldRetTrue(t *testing.T) {
	t.Parallel()

	ts := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return make([]byte, 0), nil
		},
	}
	adr, _, adb := generateAddressAccountAccountsDB(ts)

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
	adb, _ := state.NewAccountsDB(mockTrie, &mock.HasherMock{}, marshalizer, &mock.AccountsFactoryStub{
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

	ts := &mock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
	}
	_, account, adb := generateAddressAccountAccountsDB(ts)

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

	trieMock := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
	}

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

	trieMock := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, nil
		},
	}

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

	expectedValue := 45
	adr := mock.NewAddressMock()
	accnt := mock.NewAccountWrapMock(adr, nil)
	accnt.MockValue = expectedValue
	marshalizer := &mock.MarshalizerMock{}
	buffExpected, _ := marshalizer.Marshal(accnt)

	trieMock := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return buffExpected, nil
		},
	}

	adb := generateAccountDBFromTrie(trieMock)
	accntRecov, err := adb.GetAccountWithJournal(adr)

	assert.Nil(t, err)
	assert.Equal(t, expectedValue, accntRecov.(*mock.AccountWrapMock).MockValue)
	//no journal entry shall be created
	assert.Equal(t, 0, adb.JournalLen())
}

//------- getAccount

func TestAccountsDB_GetAccountAccountNotFound(t *testing.T) {
	t.Parallel()

	trieMock := mock.TrieStub{}
	adr, _, adb := generateAddressAccountAccountsDB(&mock.TrieStub{})

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

	adb, _ = state.NewAccountsDB(&trieMock, &mock.HasherMock{}, &marshalizer, &mock.AccountsFactoryStub{
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

	_, account, adb := generateAddressAccountAccountsDB(&mock.TrieStub{})

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

	adr, account, adb := generateAddressAccountAccountsDB(&mock.TrieStub{})

	trieStub := mock.TrieStub{}
	trieStub.GetCalled = func(key []byte) (bytes []byte, e error) {
		//will return adr.Bytes() so its hash will correspond to adr.Hash()
		return adr.Bytes(), nil
	}
	marshalizer := mock.MarshalizerMock{}
	adb, _ = state.NewAccountsDB(&trieStub, &mock.HasherMock{}, &marshalizer, &mock.AccountsFactoryStub{
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

	_, account, adb := generateAddressAccountAccountsDB(&mock.TrieStub{})

	//since root is nil, result should be nil and data trie should be nil
	err := adb.LoadDataTrie(account)
	assert.Nil(t, err)
	assert.Nil(t, account.DataTrie())
}

func TestAccountsDB_LoadDataBadLengthShouldErr(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB(&mock.TrieStub{})

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

	_, account, adb := generateAddressAccountAccountsDB(&mock.TrieStub{})

	rootHash := make([]byte, mock.HasherMock{}.Size())
	rootHash[0] = 1
	account.SetRootHash(rootHash)

	//should return error
	err := adb.LoadDataTrie(account)
	assert.NotNil(t, err)
	fmt.Println(err.Error())
}

func TestAccountsDB_LoadDataWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	rootHash := make([]byte, mock.HasherMock{}.Size())
	rootHash[0] = 1
	keyRequired := []byte{65, 66, 67}
	val := []byte{32, 33, 34}

	trieVal := append(val, keyRequired...)
	trieVal = append(trieVal, []byte("identifier")...)

	dataTrie := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, keyRequired) {
				return trieVal, nil
			}

			return nil, nil
		},
	}

	account := generateAccount()
	mockTrie := &mock.TrieStub{
		RecreateCalled: func(root []byte) (trie data.Trie, e error) {
			if !bytes.Equal(root, rootHash) {
				return nil, errors.New("bad root hash")
			}

			return dataTrie, nil
		},
	}
	adb := generateAccountDBFromTrie(mockTrie)

	account.SetRootHash(rootHash)

	//should not return error
	err := adb.LoadDataTrie(account)
	assert.Nil(t, err)

	//verify data
	dataRecov, err := account.DataTrieTracker().RetrieveValue(keyRequired)
	assert.Nil(t, err)
	assert.Equal(t, val, dataRecov)
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
	trieStub.RecreateCalled = func(root []byte) (tree data.Trie, e error) {
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
	trieStub.RecreateCalled = func(root []byte) (tree data.Trie, e error) {
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
	trieStub.RecreateCalled = func(root []byte) (tree data.Trie, e error) {
		wasCalled = true
		return &mock.TrieStub{}, nil
	}

	adb := generateAccountDBFromTrie(&trieStub)
	err := adb.RecreateTrie(nil)

	assert.Nil(t, err)
	assert.True(t, wasCalled)

}
