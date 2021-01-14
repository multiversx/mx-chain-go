package state_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateAccountDBFromTrie(trie data.Trie) *state.AccountsDB {
	accnt, _ := state.NewAccountsDB(trie, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return mock.NewAccountWrapMock(address), nil
		},
	})
	return accnt
}

func generateAccount() *mock.AccountWrapMock {
	return mock.NewAccountWrapMock(make([]byte, 32))
}

func generateAddressAccountAccountsDB(trie data.Trie) ([]byte, *mock.AccountWrapMock, *state.AccountsDB) {
	adr := make([]byte, 32)
	account := mock.NewAccountWrapMock(adr)

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

	assert.True(t, check.IfNil(adb))
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

	assert.True(t, check.IfNil(adb))
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

	assert.True(t, check.IfNil(adb))
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

	assert.True(t, check.IfNil(adb))
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

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))
}

func TestNewAccountsDB_SetsNumCheckpoints(t *testing.T) {
	t.Parallel()

	numCheckpointsKey := []byte("state checkpoint")
	numCheckpoints := uint32(121)
	db := mock.NewMemDbMock()

	numCheckpointsVal := make([]byte, 4)
	binary.BigEndian.PutUint32(numCheckpointsVal, numCheckpoints)
	_ = db.Put(numCheckpointsKey, numCheckpointsVal)

	adb, _ := state.NewAccountsDB(
		&mock.TrieStub{
			DatabaseCalled: func() data.DBWriteCacher {
				return db
			},
		},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
	)

	assert.Equal(t, numCheckpoints, adb.GetNumCheckpoints())
}

func TestAccountsDB_SetStateCheckpointSavesNumCheckpoints(t *testing.T) {
	t.Parallel()

	numCheckpointsKey := []byte("state checkpoint")
	numCheckpoints := 50
	db := mock.NewMemDbMock()
	adb, _ := state.NewAccountsDB(
		&mock.TrieStub{
			DatabaseCalled: func() data.DBWriteCacher {
				return db
			},
			RecreateCalled: func(root []byte) (data.Trie, error) {
				return &mock.TrieStub{
					GetAllLeavesOnChannelCalled: func(_ []byte) (chan core.KeyValueHolder, error) {
						ch := make(chan core.KeyValueHolder)
						close(ch)

						return ch, nil
					},
				}, nil
			},
		},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
	)

	for i := 0; i < numCheckpoints; i++ {
		adb.SetStateCheckpoint([]byte("rootHash"), context.Background())
	}

	time.Sleep(time.Second * 2)

	val, err := db.Get(numCheckpointsKey)
	assert.Nil(t, err)

	numCheckpointsRecovered := binary.BigEndian.Uint32(val)
	assert.Equal(t, uint32(numCheckpoints), numCheckpointsRecovered)
	assert.Equal(t, uint32(numCheckpoints), adb.GetNumCheckpoints())
}

//------- SaveAccount

func TestAccountsDB_SaveAccountNilAccountShouldErr(t *testing.T) {
	t.Parallel()

	adb := generateAccountDBFromTrie(&mock.TrieStub{})

	err := adb.SaveAccount(nil)
	assert.True(t, errors.Is(err, state.ErrNilAccountHandler))
}

func TestAccountsDB_SaveAccountErrWhenGettingOldAccountShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("trie get err")
	adb := generateAccountDBFromTrie(&mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, expectedErr
		},
	})

	err := adb.SaveAccount(generateAccount())
	assert.Equal(t, expectedErr, err)
}

func TestAccountsDB_SaveAccountNilOldAccount(t *testing.T) {
	t.Parallel()

	adb := generateAccountDBFromTrie(&mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
	})

	err := adb.SaveAccount(generateAccount())
	assert.Nil(t, err)
	assert.Equal(t, 1, adb.JournalLen())
}

func TestAccountsDB_SaveAccountExistingOldAccount(t *testing.T) {
	t.Parallel()

	adb := generateAccountDBFromTrie(&mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return (&mock.MarshalizerMock{}).Marshal(generateAccount())
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
	})

	err := adb.SaveAccount(generateAccount())
	assert.Nil(t, err)
	assert.Equal(t, 1, adb.JournalLen())
}

func TestAccountsDB_SaveAccountSavesCodeAndDataTrieForUserAccount(t *testing.T) {
	t.Parallel()

	updateCalled := 0
	trieStub := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		RootCalled: func() (i []byte, err error) {
			return []byte("rootHash"), nil
		},
	}

	adb := generateAccountDBFromTrie(&mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			updateCalled++
			return nil
		},
		RecreateCalled: func(root []byte) (d data.Trie, err error) {
			return trieStub, nil
		},
	})

	accCode := []byte("code")
	acc := generateAccount()
	acc.SetCode(accCode)
	_ = acc.DataTrieTracker().SaveKeyValue([]byte("key"), []byte("value"))

	err := adb.SaveAccount(acc)
	assert.Nil(t, err)
	assert.Equal(t, 2, updateCalled)
	assert.NotNil(t, acc.GetCodeHash())
	assert.NotNil(t, acc.GetRootHash())
}

func TestAccountsDB_SaveAccountMalfunctionMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	account := generateAccount()
	mockTrie := &mock.TrieStub{}
	marshalizer := &mock.MarshalizerMock{}
	adb, _ := state.NewAccountsDB(mockTrie, &mock.HasherMock{}, marshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return mock.NewAccountWrapMock(address), nil
		},
	})

	marshalizer.Fail = true

	//should return error
	err := adb.SaveAccount(account)

	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	ts := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
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
	marshalizer := &mock.MarshalizerMock{}
	trieStub := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return marshalizer.Marshal(mock.AccountWrapMock{})
		},
		UpdateCalled: func(key, value []byte) error {
			wasCalled = true
			return nil
		},
	}

	adr := make([]byte, 32)
	adb := generateAccountDBFromTrie(trieStub)

	err := adb.RemoveAccount(adr)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 2, adb.JournalLen())
}

//------- LoadAccount

func TestAccountsDB_LoadAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	trieStub := &mock.TrieStub{}
	adr := make([]byte, 32)
	adb := generateAccountDBFromTrie(trieStub)

	_, err := adb.LoadAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadAccountNotFoundShouldCreateEmpty(t *testing.T) {
	t.Parallel()

	trieStub := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
	}

	adr := make([]byte, 32)
	adb := generateAccountDBFromTrie(trieStub)

	accountExpected := mock.NewAccountWrapMock(adr)
	accountRecovered, err := adb.LoadAccount(adr)

	assert.Equal(t, accountExpected, accountRecovered)
	assert.Nil(t, err)
}

func TestAccountsDB_LoadAccountExistingShouldLoadDataTrie(t *testing.T) {
	t.Parallel()

	acc := generateAccount()
	acc.SetRootHash([]byte("root hash"))
	codeHash := []byte("code hash")
	acc.SetCodeHash(codeHash)
	code := []byte("code")
	dataTrie := &mock.TrieStub{}
	marshalizer := &mock.MarshalizerMock{}

	trieStub := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, acc.AddressBytes()) {
				return marshalizer.Marshal(acc)
			}
			if bytes.Equal(key, codeHash) {
				return marshalizer.Marshal(state.CodeEntry{Code: code})
			}
			return nil, nil
		},
		RecreateCalled: func(root []byte) (d data.Trie, err error) {
			return dataTrie, nil
		},
	}

	adb := generateAccountDBFromTrie(trieStub)
	retrievedAccount, err := adb.LoadAccount(acc.AddressBytes())
	assert.Nil(t, err)

	account, _ := retrievedAccount.(state.UserAccountHandler)
	assert.Equal(t, dataTrie, account.DataTrie())
}

//------- GetExistingAccount

func TestAccountsDB_GetExistingAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	trieStub := &mock.TrieStub{}
	adr := make([]byte, 32)
	adb := generateAccountDBFromTrie(trieStub)

	_, err := adb.GetExistingAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDB_GetExistingAccountNotFoundShouldRetNil(t *testing.T) {
	t.Parallel()

	trieStub := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, nil
		},
	}

	adr := make([]byte, 32)
	adb := generateAccountDBFromTrie(trieStub)

	account, err := adb.GetExistingAccount(adr)
	assert.Equal(t, state.ErrAccNotFound, err)
	assert.Nil(t, account)
	//no journal entry shall be created
	assert.Equal(t, 0, adb.JournalLen())
}

func TestAccountsDB_GetExistingAccountFoundShouldRetAccount(t *testing.T) {
	t.Parallel()

	acc := generateAccount()
	acc.SetRootHash([]byte("root hash"))
	codeHash := []byte("code hash")
	acc.SetCodeHash(codeHash)
	code := []byte("code")
	dataTrie := &mock.TrieStub{}
	marshalizer := &mock.MarshalizerMock{}

	trieStub := &mock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, acc.AddressBytes()) {
				return marshalizer.Marshal(acc)
			}
			if bytes.Equal(key, codeHash) {
				return marshalizer.Marshal(state.CodeEntry{Code: code})
			}
			return nil, nil
		},
		RecreateCalled: func(root []byte) (d data.Trie, err error) {
			return dataTrie, nil
		},
	}

	adb := generateAccountDBFromTrie(trieStub)
	retrievedAccount, err := adb.GetExistingAccount(acc.AddressBytes())
	assert.Nil(t, err)

	account, _ := retrievedAccount.(state.UserAccountHandler)
	assert.Equal(t, dataTrie, account.DataTrie())
}

//------- getAccount

func TestAccountsDB_GetAccountAccountNotFound(t *testing.T) {
	t.Parallel()

	trieMock := mock.TrieStub{}
	adr, _, adb := generateAddressAccountAccountsDB(&mock.TrieStub{})

	//Step 1. Create an account + its DbAccount representation
	testAccount := mock.NewAccountWrapMock(adr)
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
		CreateAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return mock.NewAccountWrapMock(address), nil
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

	adr := make([]byte, 32)
	account := generateAccount()
	mockTrie := &mock.TrieStub{}
	adb := generateAccountDBFromTrie(mockTrie)

	//just search a hash. Any hash will do
	account.SetCodeHash(adr)

	err := adb.LoadCode(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadCodeOkValsShouldWork(t *testing.T) {
	t.Parallel()

	adr, account, _ := generateAddressAccountAccountsDB(&mock.TrieStub{})
	marshalizer := mock.MarshalizerMock{}

	trieStub := mock.TrieStub{
		GetCalled: func(key []byte) (bytes []byte, e error) {
			//will return adr.Bytes() so its hash will correspond to adr.Hash()
			return marshalizer.Marshal(&state.CodeEntry{Code: adr})
		},
	}

	adb, _ := state.NewAccountsDB(&trieStub, &mock.HasherMock{}, &marshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return mock.NewAccountWrapMock(address), nil
		},
	})

	//just search a hash. Any hash will do
	account.SetCodeHash(adr)

	err := adb.LoadCode(account)
	assert.Nil(t, err)
	assert.Equal(t, adr, state.GetCode(account))
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

	commitCalled := 0
	marshalizer := &mock.MarshalizerMock{}
	serializedAccount, _ := marshalizer.Marshal(mock.AccountWrapMock{})
	trieStub := mock.TrieStub{
		CommitCalled: func() error {
			commitCalled++

			return nil
		},
		RootCalled: func() (i []byte, e error) {
			return nil, nil
		},
		GetCalled: func(key []byte) (i []byte, err error) {
			return serializedAccount, nil
		},
		RecreateCalled: func(root []byte) (trie data.Trie, err error) {
			return &mock.TrieStub{
				GetCalled: func(key []byte) (i []byte, err error) {
					return []byte("doge"), nil
				},
				UpdateCalled: func(key, value []byte) error {
					return nil
				},
				CommitCalled: func() error {
					commitCalled++

					return nil
				},
			}, nil
		},
	}

	adb := generateAccountDBFromTrie(&trieStub)

	state2, _ := adb.LoadAccount(make([]byte, 32))
	_ = state2.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue([]byte("dog"), []byte("puppy"))
	_ = adb.SaveAccount(state2)

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

func TestAccountsDB_CancelPrune(t *testing.T) {
	t.Parallel()

	cancelPruneWasCalled := false
	trieStub := &mock.TrieStub{
		CancelPruneCalled: func(rootHash []byte, identifier data.TriePruningIdentifier) {
			cancelPruneWasCalled = true
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.CancelPrune([]byte("roothash"), data.OldRoot)

	assert.True(t, cancelPruneWasCalled)
}

func TestAccountsDB_PruneTrie(t *testing.T) {
	t.Parallel()

	pruneTrieWasCalled := atomic.Flag{}
	trieStub := &mock.TrieStub{
		PruneCalled: func(rootHash []byte, identifier data.TriePruningIdentifier) {
			pruneTrieWasCalled.Set()
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.PruneTrie([]byte("roothash"), data.OldRoot)
	assert.True(t, pruneTrieWasCalled.IsSet())
}

func TestAccountsDB_SnapshotState(t *testing.T) {
	t.Parallel()

	takeSnapshotWasCalled := false
	snapshotMut := sync.Mutex{}
	trieStub := &mock.TrieStub{
		TakeSnapshotCalled: func(rootHash []byte) {
			snapshotMut.Lock()
			takeSnapshotWasCalled = true
			snapshotMut.Unlock()
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.SnapshotState([]byte("roothash"), context.Background())
	time.Sleep(time.Second)

	snapshotMut.Lock()
	assert.True(t, takeSnapshotWasCalled)
	snapshotMut.Unlock()
}

func TestAccountsDB_SetStateCheckpoint(t *testing.T) {
	t.Parallel()

	setCheckPointWasCalled := false
	snapshotMut := sync.Mutex{}
	trieStub := &mock.TrieStub{
		SetCheckpointCalled: func(rootHash []byte) {
			snapshotMut.Lock()
			setCheckPointWasCalled = true
			snapshotMut.Unlock()
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.SetStateCheckpoint([]byte("roothash"), context.Background())
	time.Sleep(time.Second)

	snapshotMut.Lock()
	assert.True(t, setCheckPointWasCalled)
	snapshotMut.Unlock()
}

func TestAccountsDB_IsPruningEnabled(t *testing.T) {
	t.Parallel()

	trieStub := &mock.TrieStub{
		IsPruningEnabledCalled: func() bool {
			return true
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	res := adb.IsPruningEnabled()

	assert.Equal(t, true, res)
}

func TestAccountsDB_RevertToSnapshotOutOfBounds(t *testing.T) {
	t.Parallel()

	trieStub := &mock.TrieStub{}
	adb := generateAccountDBFromTrie(trieStub)

	err := adb.RevertToSnapshot(1)
	assert.Equal(t, state.ErrSnapshotValueOutOfBounds, err)
}

func TestAccountsDB_RevertToSnapshotShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	accFactory := factory.NewAccountCreator()
	storageManager, _ := trie.NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(storageManager, marshalizer, hsh, maxTrieLevelInMemory)

	adb, _ := state.NewAccountsDB(tr, hsh, marshalizer, accFactory)

	acc, _ := adb.LoadAccount(make([]byte, 32))
	acc.(state.UserAccountHandler).SetCode([]byte("code"))
	_ = adb.SaveAccount(acc)

	err := adb.RevertToSnapshot(0)
	assert.Nil(t, err)

	expectedRoot := make([]byte, 32)
	root, err := adb.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, expectedRoot, root)
}

func TestAccountsDB_RevertToSnapshotWithoutLastRootHashSet(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	accFactory := factory.NewAccountCreator()
	storageManager, _ := trie.NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(storageManager, marshalizer, hsh, maxTrieLevelInMemory)
	adb, _ := state.NewAccountsDB(tr, hsh, marshalizer, accFactory)

	err := adb.RevertToSnapshot(0)
	assert.Nil(t, err)

	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, make([]byte, 32), rootHash)
}

func TestAccountsDB_RecreateTrieInvalidatesJournalEntries(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	accFactory := factory.NewAccountCreator()
	storageManager, _ := trie.NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(storageManager, marshalizer, hsh, maxTrieLevelInMemory)
	adb, _ := state.NewAccountsDB(tr, hsh, marshalizer, accFactory)

	address := make([]byte, 32)
	key := []byte("key")
	value := []byte("value")

	acc, _ := adb.LoadAccount(address)
	_ = adb.SaveAccount(acc)
	rootHash, _ := adb.Commit()

	acc, _ = adb.LoadAccount(address)
	acc.(state.UserAccountHandler).SetCode([]byte("code"))
	_ = adb.SaveAccount(acc)

	acc, _ = adb.LoadAccount(address)
	acc.(state.UserAccountHandler).IncreaseNonce(1)
	_ = adb.SaveAccount(acc)

	acc, _ = adb.LoadAccount(address)
	_ = acc.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue(key, value)
	_ = adb.SaveAccount(acc)

	assert.Equal(t, 5, adb.JournalLen())
	err := adb.RecreateTrie(rootHash)
	assert.Nil(t, err)
	assert.Equal(t, 0, adb.JournalLen())
}

func TestAccountsDB_RootHash(t *testing.T) {
	t.Parallel()

	rootHashCalled := false
	rootHash := []byte("root hash")
	trieStub := &mock.TrieStub{
		RootCalled: func() (i []byte, err error) {
			rootHashCalled = true
			return rootHash, nil
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	res, err := adb.RootHash()
	assert.Nil(t, err)
	assert.True(t, rootHashCalled)
	assert.Equal(t, rootHash, res)
}

func TestAccountsDB_GetAllLeaves(t *testing.T) {
	t.Parallel()

	getAllLeavesCalled := false
	trieStub := &mock.TrieStub{
		GetAllLeavesOnChannelCalled: func(rootHash []byte) (chan core.KeyValueHolder, error) {
			getAllLeavesCalled = true

			ch := make(chan core.KeyValueHolder)
			close(ch)

			return ch, nil
		},
	}

	adb := generateAccountDBFromTrie(trieStub)
	_, err := adb.GetAllLeaves([]byte("root hash"), context.Background())
	assert.Nil(t, err)
	assert.True(t, getAllLeavesCalled)
}

func getTestAccountsDbAndTrie(marshalizer marshal.Marshalizer, hsh hashing.Hasher) (*state.AccountsDB, data.Trie) {
	accFactory := factory.NewAccountCreator()
	storageManager, _ := trie.NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(storageManager, marshalizer, hsh, maxTrieLevelInMemory)
	adb, _ := state.NewAccountsDB(tr, hsh, marshalizer, accFactory)

	return adb, tr
}

func checkCodeEntry(
	codeHash []byte,
	expectedCode []byte,
	expectedNumReferences uint32,
	marshalizer marshal.Marshalizer,
	tr data.Trie,
	t *testing.T,
) {
	val, err := tr.Get(codeHash)
	assert.Nil(t, err)
	assert.NotNil(t, val)

	var codeEntry state.CodeEntry
	err = marshalizer.Unmarshal(&codeEntry, val)
	assert.Nil(t, err)

	assert.Equal(t, expectedCode, codeEntry.Code)
	assert.Equal(t, expectedNumReferences, codeEntry.NumReferences)
}

func TestAccountsDB_SaveAccountSavesCodeIfCodeHashIsSet(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	adb, tr := getTestAccountsDbAndTrie(marshalizer, hsh)

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	codeHash := hsh.Compute(string(code))
	userAcc.SetCodeHash(codeHash)

	_ = adb.SaveAccount(acc)

	checkCodeEntry(codeHash, code, 1, marshalizer, tr, t)
}

func TestAccountsDB_saveCode_OldCodeAndNewCodeAreNil(t *testing.T) {
	t.Parallel()

	adb, _ := getTestAccountsDbAndTrie(&mock.MarshalizerMock{}, mock.HasherMock{})

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	err := adb.SaveAccount(userAcc)
	assert.Nil(t, err)
	assert.Nil(t, userAcc.GetCodeHash())
}

func TestAccountsDB_saveCode_OldCodeIsNilAndNewCodeIsNotNilAndRevert(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	adb, tr := getTestAccountsDbAndTrie(marshalizer, hsh)

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	expectedCodeHash := hsh.Compute(string(code))

	err := adb.SaveAccount(userAcc)
	assert.Nil(t, err)
	assert.Equal(t, expectedCodeHash, userAcc.GetCodeHash())

	checkCodeEntry(userAcc.GetCodeHash(), code, 1, marshalizer, tr, t)

	err = adb.RevertToSnapshot(1)
	assert.Nil(t, err)

	val, err := tr.Get(expectedCodeHash)
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestAccountsDB_saveCode_OldCodeIsNilAndNewCodeAlreadyExistsAndRevert(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	adb, tr := getTestAccountsDbAndTrie(marshalizer, hsh)

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	expectedCodeHash := hsh.Compute(string(code))
	_ = adb.SaveAccount(userAcc)

	journalLen := adb.JournalLen()

	addr1 := make([]byte, 32)
	addr1[0] = 1
	acc, _ = adb.LoadAccount(addr1)
	userAcc = acc.(state.UserAccountHandler)
	userAcc.SetCode(code)

	err := adb.SaveAccount(userAcc)
	assert.Nil(t, err)
	assert.Equal(t, expectedCodeHash, userAcc.GetCodeHash())

	checkCodeEntry(userAcc.GetCodeHash(), code, 2, marshalizer, tr, t)

	err = adb.RevertToSnapshot(journalLen)
	assert.Nil(t, err)

	checkCodeEntry(expectedCodeHash, code, 1, marshalizer, tr, t)
}

func TestAccountsDB_saveCode_OldCodeExistsAndNewCodeIsNilAndRevert(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	adb, tr := getTestAccountsDbAndTrie(marshalizer, hsh)

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	oldCodeHash := hsh.Compute(string(code))
	_ = adb.SaveAccount(userAcc)

	journalLen := adb.JournalLen()

	acc, _ = adb.LoadAccount(addr)
	userAcc = acc.(state.UserAccountHandler)
	userAcc.SetCode(nil)

	err := adb.SaveAccount(userAcc)
	assert.Nil(t, err)
	assert.Nil(t, userAcc.GetCodeHash())

	val, err := tr.Get(oldCodeHash)
	assert.Nil(t, err)
	assert.Nil(t, val)

	err = adb.RevertToSnapshot(journalLen)
	assert.Nil(t, err)

	checkCodeEntry(oldCodeHash, code, 1, marshalizer, tr, t)
}

func TestAccountsDB_saveCode_OldCodeExistsAndNewCodeExistsAndRevert(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	adb, tr := getTestAccountsDbAndTrie(marshalizer, hsh)

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)
	oldCode := []byte("code1")
	userAcc.SetCode(oldCode)
	oldCodeHash := hsh.Compute(string(oldCode))
	_ = adb.SaveAccount(userAcc)

	addr1 := make([]byte, 32)
	addr1[0] = 1
	acc, _ = adb.LoadAccount(addr1)
	userAcc = acc.(state.UserAccountHandler)
	newCode := []byte("code2")
	userAcc.SetCode(newCode)
	newCodeHash := hsh.Compute(string(newCode))
	_ = adb.SaveAccount(userAcc)

	journalLen := adb.JournalLen()

	acc, _ = adb.LoadAccount(addr)
	userAcc = acc.(state.UserAccountHandler)
	userAcc.SetCode(newCode)
	err := adb.SaveAccount(userAcc)
	assert.Nil(t, err)
	assert.Equal(t, newCodeHash, userAcc.GetCodeHash())

	val, err := tr.Get(oldCodeHash)
	assert.Nil(t, err)
	assert.Nil(t, val)

	checkCodeEntry(newCodeHash, newCode, 2, marshalizer, tr, t)

	err = adb.RevertToSnapshot(journalLen)
	assert.Nil(t, err)

	checkCodeEntry(oldCodeHash, oldCode, 1, marshalizer, tr, t)
	checkCodeEntry(newCodeHash, newCode, 1, marshalizer, tr, t)
}

func TestAccountsDB_saveCode_OldCodeIsReferencedMultipleTimesAndNewCodeIsNilAndRevert(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	adb, tr := getTestAccountsDbAndTrie(marshalizer, hsh)

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	oldCodeHash := hsh.Compute(string(code))
	_ = adb.SaveAccount(userAcc)

	addr1 := make([]byte, 32)
	addr1[0] = 1
	acc, _ = adb.LoadAccount(addr1)
	userAcc = acc.(state.UserAccountHandler)
	userAcc.SetCode(code)
	_ = adb.SaveAccount(userAcc)

	journalLen := adb.JournalLen()

	acc, _ = adb.LoadAccount(addr1)
	userAcc = acc.(state.UserAccountHandler)
	userAcc.SetCode(nil)
	err := adb.SaveAccount(userAcc)
	assert.Nil(t, err)
	assert.Nil(t, userAcc.GetCodeHash())

	checkCodeEntry(oldCodeHash, code, 1, marshalizer, tr, t)

	err = adb.RevertToSnapshot(journalLen)
	assert.Nil(t, err)

	checkCodeEntry(oldCodeHash, code, 2, marshalizer, tr, t)
}

func TestAccountsDB_RemoveAccountAlsoRemovesCodeAndRevertsCorrectly(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	adb, tr := getTestAccountsDbAndTrie(marshalizer, hsh)

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	oldCodeHash := hsh.Compute(string(code))
	_ = adb.SaveAccount(userAcc)

	snapshot := adb.JournalLen()

	val, err := tr.Get(oldCodeHash)
	assert.Nil(t, err)
	assert.NotNil(t, val)

	err = adb.RemoveAccount(addr)
	assert.Nil(t, err)

	val, err = tr.Get(oldCodeHash)
	assert.Nil(t, err)
	assert.Nil(t, val)

	_ = adb.RevertToSnapshot(snapshot)

	val, err = tr.Get(oldCodeHash)
	assert.Nil(t, err)
	assert.NotNil(t, val)
}

func TestAccountsDB_MainTrieAutomaticallyMarksCodeUpdatesForEviction(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	db := mock.NewMemDbMock()
	ewl, _ := mock.NewEvictionWaitingList(100, mock.NewMemDbMock(), marshalizer)
	storageManager, _ := trie.NewTrieStorageManager(db, marshalizer, hsh, config.DBConfig{}, ewl, config.TrieStorageManagerConfig{})
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(storageManager, marshalizer, hsh, maxTrieLevelInMemory)
	adb, _ := state.NewAccountsDB(tr, hsh, marshalizer, factory.NewAccountCreator())

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	_ = adb.SaveAccount(userAcc)
	rootHash, _ := adb.Commit()

	rootHash1 := append(rootHash, byte(data.NewRoot))
	hashesForEviction := ewl.Cache[string(rootHash1)]
	assert.Equal(t, 3, len(hashesForEviction))

	acc, _ = adb.LoadAccount(addr)
	userAcc = acc.(state.UserAccountHandler)
	userAcc.SetCode(nil)
	_ = adb.SaveAccount(userAcc)
	_, _ = adb.Commit()

	rootHash2 := append(rootHash, byte(data.OldRoot))
	hashesForEviction = ewl.Cache[string(rootHash2)]
	assert.Equal(t, 3, len(hashesForEviction))
}

func TestAccountsDB_RemoveAccountSetsObsoleteHashes(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	adb, _ := getTestAccountsDbAndTrie(marshalizer, hsh)

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)
	_ = userAcc.DataTrieTracker().SaveKeyValue([]byte("key"), []byte("value"))

	_ = adb.SaveAccount(userAcc)
	_, _ = adb.Commit()

	acc, _ = adb.LoadAccount(addr)
	userAcc = acc.(state.UserAccountHandler)
	userAcc.SetCode([]byte("code"))
	snapshot := adb.JournalLen()
	hashes, _ := userAcc.DataTrieTracker().DataTrie().GetAllHashes()
	dataTrieRootHash := string(hashes[0])

	err := adb.RemoveAccount(addr)
	obsoleteHashes := adb.GetObsoleteHashes()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(obsoleteHashes))
	assert.Equal(t, hashes, obsoleteHashes[dataTrieRootHash])

	err = adb.RevertToSnapshot(snapshot)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(adb.GetObsoleteHashes()))
}

func TestAccountsDB_RemoveAccountMarksObsoleteHashesForEviction(t *testing.T) {
	t.Parallel()

	maxTrieLevelInMemory := uint(5)
	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	db := mock.NewMemDbMock()

	ewl, _ := mock.NewEvictionWaitingList(100, mock.NewMemDbMock(), marshalizer)
	storageManager, _ := trie.NewTrieStorageManager(db, marshalizer, hsh, config.DBConfig{}, ewl, config.TrieStorageManagerConfig{})
	tr, _ := trie.NewTrie(storageManager, marshalizer, hsh, maxTrieLevelInMemory)
	adb, _ := state.NewAccountsDB(tr, hsh, marshalizer, factory.NewAccountCreator())

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)
	_ = userAcc.DataTrieTracker().SaveKeyValue([]byte("key"), []byte("value"))
	_ = adb.SaveAccount(userAcc)

	addr1 := make([]byte, 32)
	addr1[0] = 1
	acc, _ = adb.LoadAccount(addr1)
	_ = adb.SaveAccount(acc)

	rootHash, _ := adb.Commit()
	hashes, _ := userAcc.DataTrieTracker().DataTrie().GetAllHashes()
	dataTrieRootHash := string(hashes[0])

	err := adb.RemoveAccount(addr)
	obsoleteHashes := adb.GetObsoleteHashes()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(obsoleteHashes))
	assert.Equal(t, hashes, obsoleteHashes[dataTrieRootHash])

	_, _ = adb.Commit()
	rootHash = append(rootHash, byte(data.OldRoot))
	oldHashes := ewl.Cache[string(rootHash)]
	assert.Equal(t, 5, len(oldHashes))
}

func BenchmarkAccountsDb_GetCodeEntry(b *testing.B) {
	maxTrieLevelInMemory := uint(5)
	marshalizer := &mock.MarshalizerMock{}
	hasher := mock.HasherMock{}
	db := mock.NewMemDbMock()

	ewl, _ := mock.NewEvictionWaitingList(100, mock.NewMemDbMock(), marshalizer)
	storageManager, _ := trie.NewTrieStorageManager(db, marshalizer, hasher, config.DBConfig{}, ewl, config.TrieStorageManagerConfig{})
	tr, _ := trie.NewTrie(storageManager, marshalizer, hasher, maxTrieLevelInMemory)
	adb, _ := state.NewAccountsDB(tr, hasher, marshalizer, factory.NewAccountCreator())

	address := make([]byte, 32)
	acc, err := adb.LoadAccount(address)
	require.Nil(b, err)

	userAcc := acc.(state.UserAccountHandler)
	codeLen := 100000
	code := make([]byte, codeLen)

	n, err := rand.Read(code)
	require.Nil(b, err)
	require.Equal(b, codeLen, n)

	userAcc.SetCode(code)
	err = adb.SaveAccount(userAcc)
	require.Nil(b, err)

	codeHash := hasher.Compute(string(code))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry, _ := state.GetCodeEntry(codeHash, tr, marshalizer)
		assert.Equal(b, code, entry.Code)
	}
}
