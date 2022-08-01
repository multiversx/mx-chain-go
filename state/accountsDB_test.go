package state_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	atomicFlag "github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/factory"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/disabled"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const trieDbOperationDelay = time.Second

func createMockAccountsDBArgs() state.ArgsAccountsDB {
	return state.ArgsAccountsDB{
		Trie: &trieMock.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{}
			},
		},
		Hasher:     &hashingMocks.HasherMock{},
		Marshaller: &testscommon.MarshalizerMock{},
		AccountFactory: &stateMock.AccountsFactoryStub{
			CreateAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return stateMock.NewAccountWrapMock(address), nil
			},
		},
		StoragePruningManager: disabled.NewDisabledStoragePruningManager(),
		ProcessingMode:        common.Normal,
		ProcessStatusHandler:  &testscommon.ProcessStatusHandlerStub{},
	}
}

func generateAccountDBFromTrie(trie common.Trie) *state.AccountsDB {
	args := createMockAccountsDBArgs()
	args.Trie = trie

	accounts, _ := state.NewAccountsDB(args)
	return accounts
}

func generateAccount() *stateMock.AccountWrapMock {
	return stateMock.NewAccountWrapMock(make([]byte, 32))
}

func generateAddressAccountAccountsDB(trie common.Trie) ([]byte, *stateMock.AccountWrapMock, *state.AccountsDB) {
	adr := make([]byte, 32)
	account := stateMock.NewAccountWrapMock(adr)

	adb := generateAccountDBFromTrie(trie)

	return adr, account, adb
}

func getDefaultTrieAndAccountsDb() (common.Trie, *state.AccountsDB) {
	checkpointHashesHolder := hashesHolder.NewCheckpointHashesHolder(10000000, testscommon.HashSize)
	adb, tr, _ := getDefaultStateComponents(checkpointHashesHolder)
	return tr, adb
}

func getDefaultStateComponents(
	hashesHolder trie.CheckpointHashesHolder,
) (*state.AccountsDB, common.Trie, common.StorageManager) {
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:      1000,
		SnapshotsBufferLen:    10,
		SnapshotsGoroutineNum: 1,
	}
	marshaller := &testscommon.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	args := trie.NewTrieStorageManagerArgs{
		MainStorer:             testscommon.NewSnapshotPruningStorerMock(),
		CheckpointsStorer:      testscommon.NewSnapshotPruningStorerMock(),
		Marshalizer:            marshaller,
		Hasher:                 hasher,
		GeneralConfig:          generalCfg,
		CheckpointHashesHolder: hashesHolder,
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
	}
	trieStorage, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(trieStorage, marshaller, hasher, 5)
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(100, testscommon.NewMemDbMock(), marshaller)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, generalCfg.PruningBufferLen)

	argsAccountsDB := state.ArgsAccountsDB{
		Trie:                  tr,
		Hasher:                hasher,
		Marshaller:            marshaller,
		AccountFactory:        factory.NewAccountCreator(),
		StoragePruningManager: spm,
		ProcessingMode:        common.Normal,
		ProcessStatusHandler:  &testscommon.ProcessStatusHandlerStub{},
	}
	adb, _ := state.NewAccountsDB(argsAccountsDB)

	return adb, tr, trieStorage
}

func TestNewAccountsDB(t *testing.T) {
	t.Parallel()

	t.Run("nil trie should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.Trie = nil

		adb, err := state.NewAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilTrie, err)
	})
	t.Run("nil hasher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.Hasher = nil

		adb, err := state.NewAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilHasher, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.Marshaller = nil

		adb, err := state.NewAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilMarshalizer, err)
	})
	t.Run("nil account factory should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.AccountFactory = nil

		adb, err := state.NewAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilAccountFactory, err)
	})
	t.Run("nil storage pruning manager should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.StoragePruningManager = nil

		adb, err := state.NewAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilStoragePruningManager, err)
	})
	t.Run("nil process status handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.ProcessStatusHandler = nil

		adb, err := state.NewAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilProcessStatusHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()

		adb, err := state.NewAccountsDB(args)
		assert.False(t, check.IfNil(adb))
		assert.Nil(t, err)
	})
}

// ------- SaveAccount

func TestAccountsDB_SaveAccountNilAccountShouldErr(t *testing.T) {
	t.Parallel()

	adb := generateAccountDBFromTrie(&trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	})

	err := adb.SaveAccount(nil)
	assert.True(t, errors.Is(err, state.ErrNilAccountHandler))
}

func TestAccountsDB_SaveAccountErrWhenGettingOldAccountShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("trie get err")
	adb := generateAccountDBFromTrie(&trieMock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, expectedErr
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	})

	err := adb.SaveAccount(generateAccount())
	assert.Equal(t, expectedErr, err)
}

func TestAccountsDB_SaveAccountNilOldAccount(t *testing.T) {
	t.Parallel()

	adb := generateAccountDBFromTrie(&trieMock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	})

	acc, _ := state.NewUserAccount([]byte("someAddress"))
	err := adb.SaveAccount(acc)
	assert.Nil(t, err)
	assert.Equal(t, 1, adb.JournalLen())
}

func TestAccountsDB_SaveAccountExistingOldAccount(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewUserAccount([]byte("someAddress"))

	adb := generateAccountDBFromTrie(&trieMock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return (&testscommon.MarshalizerMock{}).Marshal(acc)
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	})

	err := adb.SaveAccount(acc)
	assert.Nil(t, err)
	assert.Equal(t, 1, adb.JournalLen())
}

func TestAccountsDB_SaveAccountSavesCodeAndDataTrieForUserAccount(t *testing.T) {
	t.Parallel()

	updateCalled := 0
	trieStub := &trieMock.TrieStub{
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

	adb := generateAccountDBFromTrie(&trieMock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			updateCalled++
			return nil
		},
		RecreateCalled: func(root []byte) (d common.Trie, err error) {
			return trieStub, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	})

	accCode := []byte("code")
	acc, _ := state.NewUserAccount([]byte("someAddress"))
	acc.SetCode(accCode)
	_ = acc.DataTrieTracker().SaveKeyValue([]byte("key"), []byte("value"))

	err := adb.SaveAccount(acc)
	assert.Nil(t, err)
	assert.Equal(t, 2, updateCalled)
	assert.NotNil(t, acc.GetCodeHash())
	assert.NotNil(t, acc.GetRootHash())
}

func TestAccountsDB_SaveAccountMalfunctionMarshallerShouldErr(t *testing.T) {
	t.Parallel()

	account := generateAccount()
	mockTrie := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	marshaller := &testscommon.MarshalizerMock{}
	args := createMockAccountsDBArgs()
	args.Marshaller = marshaller
	args.Trie = mockTrie

	adb, _ := state.NewAccountsDB(args)

	marshaller.Fail = true

	// should return error
	err := adb.SaveAccount(account)

	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	ts := &trieMock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	_, account, adb := generateAddressAccountAccountsDB(ts)

	// should return error
	err := adb.SaveAccount(account)
	assert.Nil(t, err)
}

// ------- RemoveAccount

func TestAccountsDB_RemoveAccountShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false
	marshaller := &testscommon.MarshalizerMock{}
	trieStub := &trieMock.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return marshaller.Marshal(stateMock.AccountWrapMock{})
		},
		UpdateCalled: func(key, value []byte) error {
			wasCalled = true
			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}

	adr := make([]byte, 32)
	adb := generateAccountDBFromTrie(trieStub)

	err := adb.RemoveAccount(adr)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 2, adb.JournalLen())
}

// ------- LoadAccount

func TestAccountsDB_LoadAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	adr := make([]byte, 32)
	adb := generateAccountDBFromTrie(trieStub)

	_, err := adb.LoadAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadAccountNotFoundShouldCreateEmpty(t *testing.T) {
	t.Parallel()

	trieStub := &trieMock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}

	adr := make([]byte, 32)
	adb := generateAccountDBFromTrie(trieStub)

	accountExpected := stateMock.NewAccountWrapMock(adr)
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
	dataTrie := &trieMock.TrieStub{}
	marshaller := &testscommon.MarshalizerMock{}

	trieStub := &trieMock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, acc.AddressBytes()) {
				return marshaller.Marshal(acc)
			}
			if bytes.Equal(key, codeHash) {
				return marshaller.Marshal(state.CodeEntry{Code: code})
			}
			return nil, nil
		},
		RecreateCalled: func(root []byte) (d common.Trie, err error) {
			return dataTrie, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}

	adb := generateAccountDBFromTrie(trieStub)
	retrievedAccount, err := adb.LoadAccount(acc.AddressBytes())
	assert.Nil(t, err)

	account, _ := retrievedAccount.(state.UserAccountHandler)
	assert.Equal(t, dataTrie, account.DataTrie())
}

// ------- GetExistingAccount

func TestAccountsDB_GetExistingAccountMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	adr := make([]byte, 32)
	adb := generateAccountDBFromTrie(trieStub)

	_, err := adb.GetExistingAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDB_GetExistingAccountNotFoundShouldRetNil(t *testing.T) {
	t.Parallel()

	trieStub := &trieMock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}

	adr := make([]byte, 32)
	adb := generateAccountDBFromTrie(trieStub)

	account, err := adb.GetExistingAccount(adr)
	assert.Equal(t, state.ErrAccNotFound, err)
	assert.Nil(t, account)
	// no journal entry shall be created
	assert.Equal(t, 0, adb.JournalLen())
}

func TestAccountsDB_GetExistingAccountFoundShouldRetAccount(t *testing.T) {
	t.Parallel()

	acc := generateAccount()
	acc.SetRootHash([]byte("root hash"))
	codeHash := []byte("code hash")
	acc.SetCodeHash(codeHash)
	code := []byte("code")
	dataTrie := &trieMock.TrieStub{}
	marshaller := &testscommon.MarshalizerMock{}

	trieStub := &trieMock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, acc.AddressBytes()) {
				return marshaller.Marshal(acc)
			}
			if bytes.Equal(key, codeHash) {
				return marshaller.Marshal(state.CodeEntry{Code: code})
			}
			return nil, nil
		},
		RecreateCalled: func(root []byte) (d common.Trie, err error) {
			return dataTrie, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}

	adb := generateAccountDBFromTrie(trieStub)
	retrievedAccount, err := adb.GetExistingAccount(acc.AddressBytes())
	assert.Nil(t, err)

	account, _ := retrievedAccount.(state.UserAccountHandler)
	assert.Equal(t, dataTrie, account.DataTrie())
}

// ------- getAccount

func TestAccountsDB_GetAccountAccountNotFound(t *testing.T) {
	t.Parallel()

	tr := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	adr, _, _ := generateAddressAccountAccountsDB(tr)

	// Step 1. Create an account + its DbAccount representation
	testAccount := stateMock.NewAccountWrapMock(adr)
	testAccount.MockValue = 45

	// Step 2. marshalize the DbAccount
	marshaller := &testscommon.MarshalizerMock{}
	buff, err := marshaller.Marshal(testAccount)
	assert.Nil(t, err)

	tr.GetCalled = func(key []byte) (bytes []byte, e error) {
		// whatever the key is, return the same marshalized DbAccount
		return buff, nil
	}

	args := createMockAccountsDBArgs()
	args.Trie = tr
	args.Marshaller = marshaller

	adb, _ := state.NewAccountsDB(args)

	// Step 3. call get, should return a copy of DbAccount, recover an Account object
	recoveredAccount, err := adb.GetAccount(adr)
	assert.Nil(t, err)

	// Step 4. Let's test
	assert.Equal(t, testAccount.MockValue, recoveredAccount.(*stateMock.AccountWrapMock).MockValue)
}

// ------- loadCode

func TestAccountsDB_LoadCodeWrongHashLengthShouldErr(t *testing.T) {
	t.Parallel()

	tr := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	_, account, adb := generateAddressAccountAccountsDB(tr)

	account.SetCodeHash([]byte("AAAA"))

	err := adb.LoadCode(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadCodeMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	adr := make([]byte, 32)
	account := generateAccount()
	mockTrie := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	adb := generateAccountDBFromTrie(mockTrie)

	// just search a hash. Any hash will do
	account.SetCodeHash(adr)

	err := adb.LoadCode(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadCodeOkValsShouldWork(t *testing.T) {
	t.Parallel()

	tr := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	adr, account, _ := generateAddressAccountAccountsDB(tr)
	marshaller := &testscommon.MarshalizerMock{}

	trieStub := &trieMock.TrieStub{
		GetCalled: func(key []byte) (bytes []byte, e error) {
			// will return adr.Bytes() so its hash will correspond to adr.Hash()
			return marshaller.Marshal(&state.CodeEntry{Code: adr})
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}

	args := createMockAccountsDBArgs()
	args.Trie = trieStub
	args.Marshaller = marshaller

	adb, _ := state.NewAccountsDB(args)

	// just search a hash. Any hash will do
	account.SetCodeHash(adr)

	err := adb.LoadCode(account)
	assert.Nil(t, err)
	assert.Equal(t, adr, state.GetCode(account))
}

// ------- RetrieveData

func TestAccountsDB_LoadDataNilRootShouldRetNil(t *testing.T) {
	t.Parallel()

	tr := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	_, account, adb := generateAddressAccountAccountsDB(tr)

	// since root is nil, result should be nil and data trie should be nil
	err := adb.LoadDataTrie(account)
	assert.Nil(t, err)
	assert.Nil(t, account.DataTrie())
}

func TestAccountsDB_LoadDataBadLengthShouldErr(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB(&trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	})

	account.SetRootHash([]byte("12345"))

	// should return error
	err := adb.LoadDataTrie(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadDataMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	account := generateAccount()
	account.SetRootHash([]byte("12345"))

	mockTrie := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	adb := generateAccountDBFromTrie(mockTrie)

	// should return error
	err := adb.LoadDataTrie(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadDataNotFoundRootShouldReturnErr(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB(&trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	})

	rootHash := make([]byte, (&hashingMocks.HasherMock{}).Size())
	rootHash[0] = 1
	account.SetRootHash(rootHash)

	// should return error
	err := adb.LoadDataTrie(account)
	assert.NotNil(t, err)
	fmt.Println(err.Error())
}

func TestAccountsDB_LoadDataWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	rootHash := make([]byte, (&hashingMocks.HasherMock{}).Size())
	rootHash[0] = 1
	keyRequired := []byte{65, 66, 67}
	val := []byte{32, 33, 34}

	trieVal := append(val, keyRequired...)
	trieVal = append(trieVal, []byte("identifier")...)

	dataTrie := &trieMock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, keyRequired) {
				return trieVal, nil
			}

			return nil, nil
		},
	}

	account := generateAccount()
	mockTrie := &trieMock.TrieStub{
		RecreateCalled: func(root []byte) (trie common.Trie, e error) {
			if !bytes.Equal(root, rootHash) {
				return nil, errors.New("bad root hash")
			}

			return dataTrie, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	adb := generateAccountDBFromTrie(mockTrie)

	account.SetRootHash(rootHash)

	// should not return error
	err := adb.LoadDataTrie(account)
	assert.Nil(t, err)

	// verify data
	dataRecov, err := account.DataTrieTracker().RetrieveValue(keyRequired)
	assert.Nil(t, err)
	assert.Equal(t, val, dataRecov)
}

// ------- Commit

func TestAccountsDB_CommitShouldCallCommitFromTrie(t *testing.T) {
	t.Parallel()

	commitCalled := 0
	marshaller := &testscommon.MarshalizerMock{}
	serializedAccount, _ := marshaller.Marshal(stateMock.AccountWrapMock{})
	trieStub := trieMock.TrieStub{
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
		RecreateCalled: func(root []byte) (trie common.Trie, err error) {
			return &trieMock.TrieStub{
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
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}

	adb := generateAccountDBFromTrie(&trieStub)

	accnt, _ := adb.LoadAccount(make([]byte, 32))
	_ = accnt.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue([]byte("dog"), []byte("puppy"))
	_ = adb.SaveAccount(accnt)

	_, err := adb.Commit()
	assert.Nil(t, err)
	// one commit for the JournalEntryData and one commit for the main trie
	assert.Equal(t, 2, commitCalled)
}

// ------- RecreateTrie

func TestAccountsDB_RecreateTrieMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	wasCalled := false

	errExpected := errors.New("failure")
	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	trieStub.RecreateCalled = func(root []byte) (tree common.Trie, e error) {
		wasCalled = true
		return nil, errExpected
	}

	adb := generateAccountDBFromTrie(trieStub)

	err := adb.RecreateTrie(nil)
	assert.Equal(t, errExpected, err)
	assert.True(t, wasCalled)
}

func TestAccountsDB_RecreateTrieOutputsNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	wasCalled := false

	trieStub := trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	trieStub.RecreateCalled = func(root []byte) (tree common.Trie, e error) {
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

	trieStub := trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
		RecreateCalled: func(root []byte) (common.Trie, error) {
			wasCalled = true
			return &trieMock.TrieStub{}, nil
		},
	}

	adb := generateAccountDBFromTrie(&trieStub)
	err := adb.RecreateTrie(nil)

	assert.Nil(t, err)
	assert.True(t, wasCalled)

}

func TestAccountsDB_SnapshotState(t *testing.T) {
	t.Parallel()

	takeSnapshotWasCalled := false
	snapshotMut := sync.Mutex{}
	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				TakeSnapshotCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ chan error, _ common.SnapshotStatisticsHandler, _ uint32) {
					snapshotMut.Lock()
					takeSnapshotWasCalled = true
					snapshotMut.Unlock()
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.SnapshotState([]byte("roothash"))
	time.Sleep(time.Second)

	snapshotMut.Lock()
	assert.True(t, takeSnapshotWasCalled)
	snapshotMut.Unlock()
}

func TestAccountsDB_SnapshotStateOnAClosedStorageManagerShouldNotMarkActiveDB(t *testing.T) {
	t.Parallel()

	mut := sync.RWMutex{}
	lastSnapshotStartedWasPut := false
	activeDBWasPut := false
	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				ShouldTakeSnapshotCalled: func() bool {
					return true
				},
				TakeSnapshotCalled: func(_ []byte, _ []byte, ch chan core.KeyValueHolder, _ chan error, stats common.SnapshotStatisticsHandler, _ uint32) {
					close(ch)
					stats.SnapshotFinished()
				},
				IsClosedCalled: func() bool {
					return true
				},
				PutCalled: func(key []byte, val []byte) error {
					mut.Lock()
					defer mut.Unlock()

					if string(key) == state.LastSnapshotStarted {
						lastSnapshotStartedWasPut = true
					}

					return nil
				},
				PutInEpochCalled: func(key []byte, val []byte, epoch uint32) error {
					mut.Lock()
					defer mut.Unlock()

					if string(key) == common.ActiveDBKey {
						activeDBWasPut = true
					}

					return nil
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.SnapshotState([]byte("roothash"))
	time.Sleep(time.Second)

	mut.RLock()
	defer mut.RUnlock()
	assert.True(t, lastSnapshotStartedWasPut)
	assert.False(t, activeDBWasPut)
}

func TestAccountsDB_SnapshotStateWithErrorsShouldNotMarkActiveDB(t *testing.T) {
	t.Parallel()

	mut := sync.RWMutex{}
	lastSnapshotStartedWasPut := false
	activeDBWasPut := false
	expectedErr := errors.New("expected error")
	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				ShouldTakeSnapshotCalled: func() bool {
					return true
				},
				TakeSnapshotCalled: func(_ []byte, _ []byte, ch chan core.KeyValueHolder, errChan chan error, stats common.SnapshotStatisticsHandler, _ uint32) {
					errChan <- expectedErr
					close(ch)
					stats.SnapshotFinished()
				},
				IsClosedCalled: func() bool {
					return false
				},
				PutCalled: func(key []byte, val []byte) error {
					mut.Lock()
					defer mut.Unlock()

					if string(key) == state.LastSnapshotStarted {
						lastSnapshotStartedWasPut = true
					}

					return nil
				},
				PutInEpochCalled: func(key []byte, val []byte, epoch uint32) error {
					mut.Lock()
					defer mut.Unlock()

					if string(key) == common.ActiveDBKey {
						activeDBWasPut = true
					}

					return nil
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.SnapshotState([]byte("roothash"))
	time.Sleep(time.Second)

	mut.RLock()
	defer mut.RUnlock()
	assert.True(t, lastSnapshotStartedWasPut)
	assert.False(t, activeDBWasPut)
}

func TestAccountsDB_SnapshotStateGetLatestStorageEpochErrDoesNotSnapshot(t *testing.T) {
	t.Parallel()

	takeSnapshotCalled := false
	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				GetLatestStorageEpochCalled: func() (uint32, error) {
					return 0, fmt.Errorf("new error")
				},
				TakeSnapshotCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ chan error, _ common.SnapshotStatisticsHandler, _ uint32) {
					takeSnapshotCalled = true
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.SnapshotState([]byte("roothash"))
	time.Sleep(time.Second)

	assert.False(t, takeSnapshotCalled)
}

func TestAccountsDB_SnapshotStateSnapshotSameRootHash(t *testing.T) {
	t.Parallel()

	rootHash1 := []byte("rootHash1")
	rootHash2 := []byte("rootHash2")
	latestEpoch := uint32(0)
	snapshotMutex := sync.RWMutex{}
	takeSnapshotCalled := 0
	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				GetLatestStorageEpochCalled: func() (uint32, error) {
					return latestEpoch, nil
				},
				TakeSnapshotCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ chan error, _ common.SnapshotStatisticsHandler, _ uint32) {
					snapshotMutex.Lock()
					takeSnapshotCalled++
					snapshotMutex.Unlock()
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	waitForOpToFinish := time.Millisecond * 100

	// snapshot rootHash1 and epoch 0
	adb.SnapshotState(rootHash1)
	time.Sleep(waitForOpToFinish)
	snapshotMutex.Lock()
	assert.Equal(t, 1, takeSnapshotCalled)
	snapshotMutex.Unlock()

	// snapshot rootHash1 and epoch 1
	latestEpoch = 1
	adb.SnapshotState(rootHash1)
	time.Sleep(waitForOpToFinish)
	snapshotMutex.Lock()
	assert.Equal(t, 2, takeSnapshotCalled)
	snapshotMutex.Unlock()

	// snapshot rootHash1 and epoch 0 again
	latestEpoch = 0
	adb.SnapshotState(rootHash1)
	time.Sleep(waitForOpToFinish)
	snapshotMutex.Lock()
	assert.Equal(t, 3, takeSnapshotCalled)
	snapshotMutex.Unlock()

	// snapshot rootHash1 and epoch 0 again
	adb.SnapshotState(rootHash1)
	time.Sleep(waitForOpToFinish)
	snapshotMutex.Lock()
	assert.Equal(t, 3, takeSnapshotCalled)
	snapshotMutex.Unlock()

	// snapshot rootHash2 and epoch 0
	adb.SnapshotState(rootHash2)
	time.Sleep(waitForOpToFinish)
	snapshotMutex.Lock()
	assert.Equal(t, 4, takeSnapshotCalled)
	snapshotMutex.Unlock()

	// snapshot rootHash2 and epoch 1
	latestEpoch = 1
	adb.SnapshotState(rootHash2)
	time.Sleep(waitForOpToFinish)
	snapshotMutex.Lock()
	assert.Equal(t, 5, takeSnapshotCalled)
	snapshotMutex.Unlock()

	// snapshot rootHash2 and epoch 1 again
	latestEpoch = 1
	adb.SnapshotState(rootHash2)
	time.Sleep(waitForOpToFinish)
	snapshotMutex.Lock()
	assert.Equal(t, 5, takeSnapshotCalled)
	snapshotMutex.Unlock()
}

func TestAccountsDB_SetStateCheckpointWithDataTries(t *testing.T) {
	t.Parallel()

	tr, adb := getDefaultTrieAndAccountsDb()

	accountsAddresses := generateAccounts(t, 3, adb)
	newHashes := modifyDataTries(t, accountsAddresses, adb)
	rootHash, _ := adb.Commit()

	adb.SetStateCheckpoint(rootHash)
	time.Sleep(time.Second)

	trieDb := tr.GetStorageManager()
	err := trieDb.Remove(rootHash)
	assert.Nil(t, err)
	for hash := range newHashes {
		err = trieDb.Remove([]byte(hash))
		assert.Nil(t, err)
	}

	val, err := trieDb.Get(rootHash)
	assert.NotNil(t, val)
	assert.Nil(t, err)

	for hash := range newHashes {
		val, err = trieDb.Get([]byte(hash))
		assert.NotNil(t, val)
		assert.Nil(t, err)
	}
}

func TestAccountsDB_SetStateCheckpoint(t *testing.T) {
	t.Parallel()

	setCheckPointWasCalled := false
	snapshotMut := sync.Mutex{}
	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				SetCheckpointCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ chan error, _ common.SnapshotStatisticsHandler) {
					snapshotMut.Lock()
					setCheckPointWasCalled = true
					snapshotMut.Unlock()
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.SetStateCheckpoint([]byte("roothash"))
	time.Sleep(time.Second)

	snapshotMut.Lock()
	assert.True(t, setCheckPointWasCalled)
	snapshotMut.Unlock()
}

func TestAccountsDB_IsPruningEnabled(t *testing.T) {
	t.Parallel()

	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				IsPruningEnabledCalled: func() bool {
					return true
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	res := adb.IsPruningEnabled()

	assert.Equal(t, true, res)
}

func TestAccountsDB_RevertToSnapshotOutOfBounds(t *testing.T) {
	t.Parallel()

	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)

	err := adb.RevertToSnapshot(1)
	assert.Equal(t, state.ErrSnapshotValueOutOfBounds, err)
}

func TestAccountsDB_RevertToSnapshotShouldWork(t *testing.T) {
	t.Parallel()

	_, adb := getDefaultTrieAndAccountsDb()

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

	_, adb := getDefaultTrieAndAccountsDb()

	err := adb.RevertToSnapshot(0)
	assert.Nil(t, err)

	rootHash, err := adb.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, make([]byte, 32), rootHash)
}

func TestAccountsDB_SaveAccountWithoutLoading(t *testing.T) {
	_, adb := getDefaultTrieAndAccountsDb()

	key := []byte("key")
	key1 := []byte("key1")
	value := []byte("value")

	address := make([]byte, 32)
	account, err := adb.LoadAccount(address)
	assert.Nil(t, err)
	userAcc := account.(state.UserAccountHandler)

	err = userAcc.DataTrieTracker().SaveKeyValue(key, value)
	assert.Nil(t, err)
	err = adb.SaveAccount(userAcc)
	assert.Nil(t, err)
	_, err = adb.Commit()
	assert.Nil(t, err)

	err = userAcc.DataTrieTracker().SaveKeyValue(key1, value)
	assert.Nil(t, err)
	err = adb.SaveAccount(userAcc)
	assert.Nil(t, err)
	_, err = adb.Commit()
	assert.Nil(t, err)

	account, err = adb.LoadAccount(address)
	assert.Nil(t, err)
	userAcc = account.(state.UserAccountHandler)

	returnedVal, err := userAcc.DataTrieTracker().RetrieveValue(key)
	assert.Nil(t, err)
	assert.Equal(t, value, returnedVal)

	returnedVal, err = userAcc.DataTrieTracker().RetrieveValue(key1)
	assert.Nil(t, err)
	assert.Equal(t, value, returnedVal)

}

func TestAccountsDB_RecreateTrieInvalidatesJournalEntries(t *testing.T) {
	t.Parallel()

	_, adb := getDefaultTrieAndAccountsDb()

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
	trieStub := &trieMock.TrieStub{
		RootCalled: func() (i []byte, err error) {
			rootHashCalled = true
			return rootHash, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
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
	trieStub := &trieMock.TrieStub{
		GetAllLeavesOnChannelCalled: func(ch chan core.KeyValueHolder, ctx context.Context, rootHash []byte) error {
			getAllLeavesCalled = true
			close(ch)

			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}

	adb := generateAccountDBFromTrie(trieStub)

	leavesChannel := make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity)
	err := adb.GetAllLeaves(leavesChannel, context.Background(), []byte("root hash"))
	assert.Nil(t, err)
	assert.True(t, getAllLeavesCalled)
}

func checkCodeEntry(
	codeHash []byte,
	expectedCode []byte,
	expectedNumReferences uint32,
	marshaller marshal.Marshalizer,
	tr common.Trie,
	t *testing.T,
) {
	val, err := tr.Get(codeHash)
	assert.Nil(t, err)
	assert.NotNil(t, val)

	var codeEntry state.CodeEntry
	err = marshaller.Unmarshal(&codeEntry, val)
	assert.Nil(t, err)

	assert.Equal(t, expectedCode, codeEntry.Code)
	assert.Equal(t, expectedNumReferences, codeEntry.NumReferences)
}

func TestAccountsDB_SaveAccountSavesCodeIfCodeHashIsSet(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	tr, adb := getDefaultTrieAndAccountsDb()

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	codeHash := hasher.Compute(string(code))
	userAcc.SetCodeHash(codeHash)

	_ = adb.SaveAccount(acc)

	checkCodeEntry(codeHash, code, 1, marshaller, tr, t)
}

func TestAccountsDB_saveCode_OldCodeAndNewCodeAreNil(t *testing.T) {
	t.Parallel()

	_, adb := getDefaultTrieAndAccountsDb()

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	err := adb.SaveAccount(userAcc)
	assert.Nil(t, err)
	assert.Nil(t, userAcc.GetCodeHash())
}

func TestAccountsDB_saveCode_OldCodeIsNilAndNewCodeIsNotNilAndRevert(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	tr, adb := getDefaultTrieAndAccountsDb()

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	expectedCodeHash := hasher.Compute(string(code))

	err := adb.SaveAccount(userAcc)
	assert.Nil(t, err)
	assert.Equal(t, expectedCodeHash, userAcc.GetCodeHash())

	checkCodeEntry(userAcc.GetCodeHash(), code, 1, marshaller, tr, t)

	err = adb.RevertToSnapshot(1)
	assert.Nil(t, err)

	val, err := tr.Get(expectedCodeHash)
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestAccountsDB_saveCode_OldCodeIsNilAndNewCodeAlreadyExistsAndRevert(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	tr, adb := getDefaultTrieAndAccountsDb()

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	expectedCodeHash := hasher.Compute(string(code))
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

	checkCodeEntry(userAcc.GetCodeHash(), code, 2, marshaller, tr, t)

	err = adb.RevertToSnapshot(journalLen)
	assert.Nil(t, err)

	checkCodeEntry(expectedCodeHash, code, 1, marshaller, tr, t)
}

func TestAccountsDB_saveCode_OldCodeExistsAndNewCodeIsNilAndRevert(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	tr, adb := getDefaultTrieAndAccountsDb()

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	oldCodeHash := hasher.Compute(string(code))
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

	checkCodeEntry(oldCodeHash, code, 1, marshaller, tr, t)
}

func TestAccountsDB_saveCode_OldCodeExistsAndNewCodeExistsAndRevert(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	tr, adb := getDefaultTrieAndAccountsDb()

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)
	oldCode := []byte("code1")
	userAcc.SetCode(oldCode)
	oldCodeHash := hasher.Compute(string(oldCode))
	_ = adb.SaveAccount(userAcc)

	addr1 := make([]byte, 32)
	addr1[0] = 1
	acc, _ = adb.LoadAccount(addr1)
	userAcc = acc.(state.UserAccountHandler)
	newCode := []byte("code2")
	userAcc.SetCode(newCode)
	newCodeHash := hasher.Compute(string(newCode))
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

	checkCodeEntry(newCodeHash, newCode, 2, marshaller, tr, t)

	err = adb.RevertToSnapshot(journalLen)
	assert.Nil(t, err)

	checkCodeEntry(oldCodeHash, oldCode, 1, marshaller, tr, t)
	checkCodeEntry(newCodeHash, newCode, 1, marshaller, tr, t)
}

func TestAccountsDB_saveCode_OldCodeIsReferencedMultipleTimesAndNewCodeIsNilAndRevert(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	tr, adb := getDefaultTrieAndAccountsDb()

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	oldCodeHash := hasher.Compute(string(code))
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

	checkCodeEntry(oldCodeHash, code, 1, marshaller, tr, t)

	err = adb.RevertToSnapshot(journalLen)
	assert.Nil(t, err)

	checkCodeEntry(oldCodeHash, code, 2, marshaller, tr, t)
}

func TestAccountsDB_RemoveAccountAlsoRemovesCodeAndRevertsCorrectly(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	tr, adb := getDefaultTrieAndAccountsDb()

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	oldCodeHash := hasher.Compute(string(code))
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

	marshaller := &testscommon.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	ewl := stateMock.NewEvictionWaitingList(100, testscommon.NewMemDbMock(), marshaller)
	args := trie.NewTrieStorageManagerArgs{
		MainStorer:             testscommon.CreateMemUnit(),
		CheckpointsStorer:      testscommon.CreateMemUnit(),
		Marshalizer:            marshaller,
		Hasher:                 hasher,
		GeneralConfig:          config.TrieStorageManagerConfig{SnapshotsGoroutineNum: 1},
		CheckpointHashesHolder: &trieMock.CheckpointHashesHolderStub{},
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
	}
	storageManager, _ := trie.NewTrieStorageManager(args)
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(storageManager, marshaller, hasher, maxTrieLevelInMemory)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 5)

	argsAccountsDB := createMockAccountsDBArgs()
	argsAccountsDB.Trie = tr
	argsAccountsDB.Hasher = hasher
	argsAccountsDB.Marshaller = marshaller
	argsAccountsDB.AccountFactory = factory.NewAccountCreator()
	argsAccountsDB.StoragePruningManager = spm

	adb, _ := state.NewAccountsDB(argsAccountsDB)

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	_ = adb.SaveAccount(userAcc)
	rootHash, _ := adb.Commit()

	rootHash1 := append(rootHash, byte(state.NewRoot))
	hashesForEviction := ewl.Cache[string(rootHash1)]
	assert.Equal(t, 3, len(hashesForEviction))

	acc, _ = adb.LoadAccount(addr)
	userAcc = acc.(state.UserAccountHandler)
	userAcc.SetCode(nil)
	_ = adb.SaveAccount(userAcc)
	_, _ = adb.Commit()

	rootHash2 := append(rootHash, byte(state.OldRoot))
	hashesForEviction = ewl.Cache[string(rootHash2)]
	assert.Equal(t, 3, len(hashesForEviction))
}

func TestAccountsDB_RemoveAccountSetsObsoleteHashes(t *testing.T) {
	t.Parallel()

	_, adb := getDefaultTrieAndAccountsDb()

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

	err := adb.RemoveAccount(addr)
	obsoleteHashes := adb.GetObsoleteHashes()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(obsoleteHashes))
	assert.Equal(t, hashes, obsoleteHashes[string(hashes[0])])

	err = adb.RevertToSnapshot(snapshot)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(adb.GetObsoleteHashes()))
}

func TestAccountsDB_RemoveAccountMarksObsoleteHashesForEviction(t *testing.T) {
	t.Parallel()

	maxTrieLevelInMemory := uint(5)
	marshaller := &testscommon.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	ewl := stateMock.NewEvictionWaitingList(100, testscommon.NewMemDbMock(), marshaller)
	args := trie.NewTrieStorageManagerArgs{
		MainStorer:             testscommon.CreateMemUnit(),
		CheckpointsStorer:      testscommon.CreateMemUnit(),
		Marshalizer:            marshaller,
		Hasher:                 hasher,
		GeneralConfig:          config.TrieStorageManagerConfig{SnapshotsGoroutineNum: 1},
		CheckpointHashesHolder: &trieMock.CheckpointHashesHolderStub{},
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
	}
	storageManager, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(storageManager, marshaller, hasher, maxTrieLevelInMemory)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 5)

	argsAccountsDB := createMockAccountsDBArgs()
	argsAccountsDB.Trie = tr
	argsAccountsDB.Hasher = hasher
	argsAccountsDB.Marshaller = marshaller
	argsAccountsDB.AccountFactory = factory.NewAccountCreator()
	argsAccountsDB.StoragePruningManager = spm

	adb, _ := state.NewAccountsDB(argsAccountsDB)

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

	err := adb.RemoveAccount(addr)
	obsoleteHashes := adb.GetObsoleteHashes()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(obsoleteHashes))
	assert.Equal(t, hashes, obsoleteHashes[string(hashes[0])])

	_, _ = adb.Commit()
	rootHash = append(rootHash, byte(state.OldRoot))
	oldHashes := ewl.Cache[string(rootHash)]
	assert.Equal(t, 5, len(oldHashes))
}

func TestAccountsDB_TrieDatabasePruning(t *testing.T) {
	t.Parallel()

	tr, adb := getDefaultTrieAndAccountsDb()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("ddog"), []byte("cat"))
	_ = tr.Commit()

	rootHash, _ := tr.RootHash()

	_ = tr.Update([]byte("dog"), []byte("doee"))
	oldHashes := tr.GetObsoleteHashes()
	assert.Equal(t, 4, len(oldHashes))
	_, err := adb.Commit()
	assert.Nil(t, err)

	adb.CancelPrune(rootHash, state.NewRoot)
	adb.PruneTrie(rootHash, state.OldRoot, state.NewPruningHandler(state.EnableDataRemoval))
	time.Sleep(trieDbOperationDelay)

	for i := range oldHashes {
		encNode, errGet := tr.GetStorageManager().Get(oldHashes[i])
		assert.Nil(t, encNode)
		assert.NotNil(t, errGet)
	}
}

func TestAccountsDB_PruningAndPruningCancellingOnTrieRollback(t *testing.T) {
	t.Parallel()

	testVals := []struct {
		key   []byte
		value []byte
	}{
		{[]byte("doe"), []byte("reindeer")},
		{[]byte("dog"), []byte("puppy")},
		{[]byte("dogglesworth"), []byte("cat")},
		{[]byte("horse"), []byte("stallion")},
	}
	tr, adb := getDefaultTrieAndAccountsDb()

	rootHashes := make([][]byte, 0)
	for _, testVal := range testVals {
		_ = tr.Update(testVal.key, testVal.value)
		_, _ = adb.Commit()

		rootHash, _ := tr.RootHash()
		rootHashes = append(rootHashes, rootHash)
	}

	for i := 0; i < len(rootHashes); i++ {
		_, err := tr.Recreate(rootHashes[i])
		assert.Nil(t, err)
	}

	adb.CancelPrune(rootHashes[0], state.NewRoot)
	finalizeTrieState(t, 1, tr, adb, rootHashes)
	finalizeTrieState(t, 2, tr, adb, rootHashes)
	rollbackTrieState(t, 3, tr, adb, rootHashes)

	_, err := tr.Recreate(rootHashes[2])
	assert.Nil(t, err)
}

func finalizeTrieState(t *testing.T, index int, tr common.Trie, adb state.AccountsAdapter, rootHashes [][]byte) {
	adb.PruneTrie(rootHashes[index-1], state.OldRoot, state.NewPruningHandler(state.EnableDataRemoval))
	adb.CancelPrune(rootHashes[index], state.NewRoot)
	time.Sleep(trieDbOperationDelay)

	_, err := tr.Recreate(rootHashes[index-1])
	assert.NotNil(t, err)
}

func rollbackTrieState(t *testing.T, index int, tr common.Trie, adb state.AccountsAdapter, rootHashes [][]byte) {
	adb.PruneTrie(rootHashes[index], state.NewRoot, state.NewPruningHandler(state.EnableDataRemoval))
	adb.CancelPrune(rootHashes[index-1], state.OldRoot)
	time.Sleep(trieDbOperationDelay)

	_, err := tr.Recreate(rootHashes[index])
	assert.NotNil(t, err)
}

func TestAccountsDB_Prune(t *testing.T) {
	t.Parallel()

	tr, adb := getDefaultTrieAndAccountsDb()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_, _ = adb.Commit()
	rootHash, _ := tr.RootHash()

	_ = tr.Update([]byte("dog"), []byte("value of dog"))
	_, _ = adb.Commit()

	adb.CancelPrune(rootHash, state.NewRoot)
	adb.PruneTrie(rootHash, state.OldRoot, state.NewPruningHandler(state.EnableDataRemoval))

	val, err := tr.GetStorageManager().Get(rootHash)
	assert.Nil(t, val)
	assert.Equal(t, trie.ErrKeyNotFound, err)
}

func TestAccountsDB_CommitAddsDirtyHashesToCheckpointHashesHolder(t *testing.T) {
	t.Parallel()

	newHashes := make(common.ModifiedHashes)
	var rootHash []byte
	checkpointHashesHolder := &trieMock.CheckpointHashesHolderStub{
		PutCalled: func(rH []byte, hashes common.ModifiedHashes) bool {
			assert.True(t, len(rH) != 0)
			assert.True(t, len(hashes) != 0)
			assert.Equal(t, rootHash, rH)
			assert.Equal(t, len(newHashes), len(hashes))

			for key := range hashes {
				_, ok := newHashes[key]
				assert.True(t, ok)
			}

			return false
		},
	}

	adb, tr, _ := getDefaultStateComponents(checkpointHashesHolder)

	accountsAddresses := generateAccounts(t, 3, adb)
	newHashes, _ = tr.GetDirtyHashes()
	rootHash, _ = tr.RootHash()
	_, err := adb.Commit()
	assert.Nil(t, err)

	newHashes = modifyDataTries(t, accountsAddresses, adb)
	_ = generateAccounts(t, 2, adb)
	newHashesMainTrie, _ := tr.GetDirtyHashes()
	mergeMaps(newHashes, newHashesMainTrie)
	rootHash, _ = tr.RootHash()

	_, err = adb.Commit()
	assert.Nil(t, err)
}

func mergeMaps(map1 common.ModifiedHashes, map2 common.ModifiedHashes) {
	for key, val := range map2 {
		map1[key] = val
	}
}

func TestAccountsDB_CommitSetsStateCheckpointIfCheckpointHashesHolderIsFull(t *testing.T) {
	t.Parallel()

	newHashes := make(common.ModifiedHashes)
	numRemoveCalls := 0
	checkpointHashesHolder := &trieMock.CheckpointHashesHolderStub{
		PutCalled: func(_ []byte, _ common.ModifiedHashes) bool {
			return true
		},
		RemoveCalled: func(hash []byte) {
			_, ok := newHashes[string(hash)]
			assert.True(t, ok)
			numRemoveCalls++
		},
	}

	adb, tr, trieStorage := getDefaultStateComponents(checkpointHashesHolder)

	accountsAddresses := generateAccounts(t, 3, adb)
	newHashes = modifyDataTries(t, accountsAddresses, adb)
	newHashesMainTrie, _ := tr.GetDirtyHashes()
	mergeMaps(newHashes, newHashesMainTrie)

	_, err := adb.Commit()
	for trieStorage.IsPruningBlocked() {
		time.Sleep(10 * time.Millisecond)
	}
	assert.Nil(t, err)
	assert.Equal(t, len(newHashes), numRemoveCalls)
}

func TestAccountsDB_SnapshotStateCleansCheckpointHashesHolder(t *testing.T) {
	t.Parallel()

	removeCommitedCalled := false
	checkpointHashesHolder := &trieMock.CheckpointHashesHolderStub{
		PutCalled: func(_ []byte, _ common.ModifiedHashes) bool {
			return false
		},
		RemoveCommittedCalled: func(_ []byte) {
			removeCommitedCalled = true
		},
		ShouldCommitCalled: func(_ []byte) bool {
			return false
		},
	}
	adb, tr, trieStorage := getDefaultStateComponents(checkpointHashesHolder)
	_ = trieStorage.Put([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal))

	accountsAddresses := generateAccounts(t, 3, adb)
	newHashes := modifyDataTries(t, accountsAddresses, adb)
	newHashesMainTrie, _ := tr.GetDirtyHashes()
	mergeMaps(newHashes, newHashesMainTrie)

	rootHash, _ := adb.Commit()
	adb.SnapshotState(rootHash)
	for trieStorage.IsPruningBlocked() {
		time.Sleep(10 * time.Millisecond)
	}

	assert.True(t, removeCommitedCalled)
}

func TestAccountsDB_SetStateCheckpointCommitsOnlyMissingData(t *testing.T) {
	t.Parallel()

	checkpointHashesHolder := hashesHolder.NewCheckpointHashesHolder(100000, testscommon.HashSize)
	adb, tr, trieStorage := getDefaultStateComponents(checkpointHashesHolder)

	accountsAddresses := generateAccounts(t, 3, adb)
	rootHash, _ := tr.RootHash()

	_, err := adb.Commit()
	assert.Nil(t, err)
	checkpointHashesHolder.RemoveCommitted(rootHash)

	newHashes := modifyDataTries(t, accountsAddresses, adb)

	_ = generateAccounts(t, 2, adb)

	newHashesMainTrie, _ := tr.GetDirtyHashes()
	mergeMaps(newHashes, newHashesMainTrie)
	rootHash, _ = adb.Commit()

	adb.SetStateCheckpoint(rootHash)
	for trieStorage.IsPruningBlocked() {
		time.Sleep(10 * time.Millisecond)
	}

	allStateHashes := make([][]byte, 0)
	mainTrieHashes, _ := tr.GetAllHashes()
	allStateHashes = append(allStateHashes, mainTrieHashes...)

	acc, _ := adb.LoadAccount(accountsAddresses[0])
	dataTrie1Hashes, _ := acc.(state.UserAccountHandler).DataTrie().GetAllHashes()
	allStateHashes = append(allStateHashes, dataTrie1Hashes...)

	acc, _ = adb.LoadAccount(accountsAddresses[1])
	dataTrie2Hashes, _ := acc.(state.UserAccountHandler).DataTrie().GetAllHashes()
	allStateHashes = append(allStateHashes, dataTrie2Hashes...)

	for _, hash := range allStateHashes {
		err = trieStorage.Remove(hash)
		assert.Nil(t, err)
	}

	numPresent := 0
	numAbsent := 0
	for _, hash := range allStateHashes {
		_, ok := newHashes[string(hash)]
		if ok {
			val, errGet := trieStorage.Get(hash)
			assert.Nil(t, errGet)
			assert.NotNil(t, val)
			numPresent++
			continue
		}

		val, errGet := trieStorage.Get(hash)
		assert.Nil(t, val)
		assert.NotNil(t, errGet)
		numAbsent++
	}

	assert.Equal(t, len(newHashes), numPresent)
	if len(allStateHashes) > len(newHashes) {
		assert.True(t, numAbsent > 0)
	}
}

func TestAccountsDB_CheckpointHashesHolderReceivesOnly32BytesData(t *testing.T) {
	t.Parallel()

	putCalled := false
	checkpointHashesHolder := &trieMock.CheckpointHashesHolderStub{
		PutCalled: func(rootHash []byte, hashes common.ModifiedHashes) bool {
			putCalled = true
			assert.Equal(t, 32, len(rootHash))
			for key := range hashes {
				assert.Equal(t, 32, len(key))
			}
			return false
		},
	}
	adb, _, _ := getDefaultStateComponents(checkpointHashesHolder)

	accountsAddresses := generateAccounts(t, 3, adb)
	_ = modifyDataTries(t, accountsAddresses, adb)

	_, _ = adb.Commit()
	assert.True(t, putCalled)
}

func TestAccountsDB_PruneRemovesDataFromCheckpointHashesHolder(t *testing.T) {
	t.Parallel()

	newHashes := make(common.ModifiedHashes)
	removeCalled := 0
	checkpointHashesHolder := &trieMock.CheckpointHashesHolderStub{
		RemoveCalled: func(hash []byte) {
			_, ok := newHashes[string(hash)]
			assert.True(t, ok)
			removeCalled++
		},
	}
	adb, tr, _ := getDefaultStateComponents(checkpointHashesHolder)

	accountsAddresses := generateAccounts(t, 3, adb)
	newHashes, _ = tr.GetDirtyHashes()
	rootHash, _ := tr.RootHash()
	_, err := adb.Commit()
	assert.Nil(t, err)

	_ = modifyDataTries(t, accountsAddresses, adb)
	_ = generateAccounts(t, 2, adb)
	_, err = adb.Commit()
	assert.Nil(t, err)

	adb.CancelPrune(rootHash, state.NewRoot)
	adb.PruneTrie(rootHash, state.OldRoot, state.NewPruningHandler(state.EnableDataRemoval))
	assert.True(t, removeCalled > 0)
}

func generateAccounts(t *testing.T, numAccounts int, adb state.AccountsAdapter) [][]byte {
	accountsAddresses := make([][]byte, numAccounts)
	for i := 0; i < numAccounts; i++ {
		addr := generateRandomByteArray(32)
		acc, err := adb.LoadAccount(addr)
		require.Nil(t, err)
		require.NotNil(t, acc)
		err = adb.SaveAccount(acc)
		require.Nil(t, err)

		accountsAddresses[i] = addr
	}

	return accountsAddresses
}

func generateRandomByteArray(size int) []byte {
	r := make([]byte, size)
	_, _ = rand.Read(r)
	return r
}

func modifyDataTries(t *testing.T, accountsAddresses [][]byte, adb *state.AccountsDB) common.ModifiedHashes {
	acc, _ := adb.LoadAccount(accountsAddresses[0])
	err := acc.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue([]byte("key1"), []byte("value1"))
	assert.Nil(t, err)
	err = acc.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue([]byte("key2"), []byte("value2"))
	assert.Nil(t, err)
	_ = adb.SaveAccount(acc)
	newHashes, _ := acc.(state.UserAccountHandler).DataTrie().GetDirtyHashes()

	acc, _ = adb.LoadAccount(accountsAddresses[1])
	err = acc.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue([]byte("key2"), []byte("value2"))
	assert.Nil(t, err)
	_ = adb.SaveAccount(acc)
	newHashesDataTrie, _ := acc.(state.UserAccountHandler).DataTrie().GetDirtyHashes()
	mergeMaps(newHashes, newHashesDataTrie)

	return newHashes
}

func TestAccountsDB_GetCode(t *testing.T) {
	t.Parallel()

	maxTrieLevelInMemory := uint(5)
	marshaller := &testscommon.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	args := trie.NewTrieStorageManagerArgs{
		MainStorer:             testscommon.CreateMemUnit(),
		CheckpointsStorer:      testscommon.CreateMemUnit(),
		Marshalizer:            marshaller,
		Hasher:                 hasher,
		GeneralConfig:          config.TrieStorageManagerConfig{SnapshotsGoroutineNum: 1},
		CheckpointHashesHolder: &trieMock.CheckpointHashesHolderStub{},
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
	}
	storageManager, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(storageManager, marshaller, hasher, maxTrieLevelInMemory)
	spm := disabled.NewDisabledStoragePruningManager()
	argsAccountsDB := createMockAccountsDBArgs()
	argsAccountsDB.Trie = tr
	argsAccountsDB.Hasher = hasher
	argsAccountsDB.Marshaller = marshaller
	argsAccountsDB.AccountFactory = factory.NewAccountCreator()
	argsAccountsDB.StoragePruningManager = spm

	adb, _ := state.NewAccountsDB(argsAccountsDB)

	address := make([]byte, 32)
	acc, err := adb.LoadAccount(address)
	require.Nil(t, err)

	userAcc := acc.(state.UserAccountHandler)
	codeLen := 100000
	code := make([]byte, codeLen)

	n, err := rand.Read(code)
	require.Nil(t, err)
	require.Equal(t, codeLen, n)

	userAcc.SetCode(code)
	err = adb.SaveAccount(userAcc)
	require.Nil(t, err)

	codeHash := hasher.Compute(string(code))

	retrievedCode := adb.GetCode(codeHash)
	assert.Equal(t, retrievedCode, code)
}

func TestAccountsDB_ImportAccount(t *testing.T) {
	t.Parallel()

	_, adb := getDefaultTrieAndAccountsDb()

	address := make([]byte, 32)
	acc, err := adb.LoadAccount(address)
	require.Nil(t, err)

	retrievedAcc, err := adb.GetExistingAccount(address)
	assert.NotNil(t, err)
	assert.Nil(t, retrievedAcc)

	err = adb.ImportAccount(acc)
	assert.Nil(t, err)

	retrievedAcc, err = adb.GetExistingAccount(address)
	assert.NotNil(t, retrievedAcc)
	assert.Nil(t, err)
}

func TestAccountsDB_RecreateAllTries(t *testing.T) {
	t.Parallel()

	_, adb := getDefaultTrieAndAccountsDb()

	addresses := generateAccounts(t, 2, adb)
	_ = modifyDataTries(t, addresses, adb)

	rootHash, _ := adb.Commit()

	tries, err := adb.RecreateAllTries(rootHash)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(tries))

	_, ok := tries[string(rootHash)]
	assert.True(t, ok)

	for i := 0; i < len(addresses); i++ {
		acc, _ := adb.LoadAccount(addresses[i])
		rootHash = acc.(state.UserAccountHandler).GetRootHash()

		_, ok = tries[string(rootHash)]
		assert.True(t, ok)
	}
}

func TestAccountsDB_GetTrie(t *testing.T) {
	t.Parallel()

	_, adb := getDefaultTrieAndAccountsDb()

	addresses := generateAccounts(t, 2, adb)
	_ = modifyDataTries(t, addresses, adb)

	_, _ = adb.Commit()

	for i := 0; i < len(addresses); i++ {
		acc, _ := adb.LoadAccount(addresses[i])
		rootHash := acc.(state.UserAccountHandler).GetRootHash()

		tr, err := adb.GetTrie(rootHash)
		assert.Nil(t, err)
		assert.NotNil(t, tr)
	}
}

func TestAccountsDB_Close(t *testing.T) {
	t.Parallel()

	closeCalled := false
	tr := &trieMock.TrieStub{
		CloseCalled: func() error {
			closeCalled = true
			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}
	marshaller := &testscommon.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(100, testscommon.NewMemDbMock(), marshaller)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)

	argsAccountsDB := createMockAccountsDBArgs()
	argsAccountsDB.Trie = tr
	argsAccountsDB.Hasher = hasher
	argsAccountsDB.Marshaller = marshaller
	argsAccountsDB.AccountFactory = factory.NewAccountCreator()
	argsAccountsDB.StoragePruningManager = spm

	adb, _ := state.NewAccountsDB(argsAccountsDB)

	err := adb.Close()
	assert.Nil(t, err)
	assert.True(t, closeCalled)
}

func TestAccountsDB_GetAccountFromBytesInvalidAddress(t *testing.T) {
	t.Parallel()

	_, adb := getDefaultTrieAndAccountsDb()

	acc, err := adb.GetAccountFromBytes([]byte{}, []byte{})
	assert.Nil(t, acc)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), state.ErrNilAddress.Error()))
}

func TestAccountsDB_GetAccountFromBytes(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerMock{}
	adr := make([]byte, 32)
	accountExpected, _ := state.NewUserAccount(adr)
	accountBytes, _ := marshaller.Marshal(accountExpected)
	_, adb := getDefaultTrieAndAccountsDb()

	acc, err := adb.GetAccountFromBytes(adr, accountBytes)
	assert.Nil(t, err)
	assert.Equal(t, accountExpected, acc)
}

func TestAccountsDB_GetAccountFromBytesShouldLoadDataTrie(t *testing.T) {
	t.Parallel()

	acc := generateAccount()
	acc.SetRootHash([]byte("root hash"))
	dataTrie := &trieMock.TrieStub{}
	marshaller := &testscommon.MarshalizerMock{}
	serializerAcc, _ := marshaller.Marshal(acc)

	trieStub := &trieMock.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, acc.AddressBytes()) {
				return serializerAcc, nil
			}
			return nil, nil
		},
		RecreateCalled: func(root []byte) (d common.Trie, err error) {
			return dataTrie, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{}
		},
	}

	adb := generateAccountDBFromTrie(trieStub)
	retrievedAccount, err := adb.GetAccountFromBytes(acc.AddressBytes(), serializerAcc)
	assert.Nil(t, err)

	account, _ := retrievedAccount.(state.UserAccountHandler)
	assert.Equal(t, dataTrie, account.DataTrie())
}

func TestAccountsDB_NewAccountsDbStartsSnapshotAfterRestart(t *testing.T) {
	t.Parallel()

	rootHash := []byte("rootHash")
	takeSnapshotCalled := atomicFlag.Flag{}
	trieStub := &trieMock.TrieStub{
		RootCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &testscommon.StorageManagerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, []byte(common.ActiveDBKey)) {
						return nil, fmt.Errorf("key not found")
					}
					return []byte("rootHash"), nil
				},
				ShouldTakeSnapshotCalled: func() bool {
					return true
				},
				TakeSnapshotCalled: func(_ []byte, _ []byte, _ chan core.KeyValueHolder, _ chan error, _ common.SnapshotStatisticsHandler, _ uint32) {
					takeSnapshotCalled.SetValue(true)
				},
				GetLatestStorageEpochCalled: func() (uint32, error) {
					return 1, nil
				},
			}
		},
	}

	_ = generateAccountDBFromTrie(trieStub)
	time.Sleep(time.Second)
	assert.True(t, takeSnapshotCalled.IsSet())
}

func BenchmarkAccountsDb_GetCodeEntry(b *testing.B) {
	maxTrieLevelInMemory := uint(5)
	marshaller := &testscommon.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	args := trie.NewTrieStorageManagerArgs{
		MainStorer:             testscommon.CreateMemUnit(),
		CheckpointsStorer:      testscommon.CreateMemUnit(),
		Marshalizer:            marshaller,
		Hasher:                 hasher,
		GeneralConfig:          config.TrieStorageManagerConfig{SnapshotsGoroutineNum: 1},
		CheckpointHashesHolder: &trieMock.CheckpointHashesHolderStub{},
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
	}
	storageManager, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(storageManager, marshaller, hasher, maxTrieLevelInMemory)
	spm := disabled.NewDisabledStoragePruningManager()

	argsAccountsDB := createMockAccountsDBArgs()
	argsAccountsDB.Trie = tr
	argsAccountsDB.Hasher = hasher
	argsAccountsDB.Marshaller = marshaller
	argsAccountsDB.AccountFactory = factory.NewAccountCreator()
	argsAccountsDB.StoragePruningManager = spm

	adb, _ := state.NewAccountsDB(argsAccountsDB)

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
		entry, _ := state.GetCodeEntry(codeHash, tr, marshaller)
		assert.Equal(b, code, entry.Code)
	}
}

func TestAccountsDB_waitForCompletionIfRunningInImportDB(t *testing.T) {
	t.Parallel()

	t.Run("not in import db", func(t *testing.T) {
		t.Parallel()

		argsAccountsDB := createMockAccountsDBArgs()
		argsAccountsDB.ProcessingMode = common.Normal
		argsAccountsDB.ProcessStatusHandler = &testscommon.ProcessStatusHandlerStub{
			SetIdleCalled: func() {
				assert.Fail(t, "should have not called set idle ")
			},
		}
		adb, _ := state.NewAccountsDB(argsAccountsDB)
		adb.WaitForCompletionIfRunningInImportDB(&trieMock.MockStatistics{})
	})
	t.Run("in import db", func(t *testing.T) {
		t.Parallel()

		argsAccountsDB := createMockAccountsDBArgs()
		argsAccountsDB.ProcessingMode = common.ImportDb
		idleWasSet := false
		argsAccountsDB.ProcessStatusHandler = &testscommon.ProcessStatusHandlerStub{
			SetIdleCalled: func() {
				idleWasSet = true
			},
		}

		stats := &trieMock.MockStatistics{}
		waitForSnapshotsToFinishCalled := false
		stats.WaitForSnapshotsToFinishCalled = func() {
			waitForSnapshotsToFinishCalled = true
		}
		adb, _ := state.NewAccountsDB(argsAccountsDB)
		adb.WaitForCompletionIfRunningInImportDB(stats)
		assert.True(t, idleWasSet)
		assert.True(t, waitForSnapshotsToFinishCalled)
	})
}

func TestEmptyErrChanReturningHadContained(t *testing.T) {
	t.Parallel()

	t.Run("empty chan should return false", func(t *testing.T) {
		t.Parallel()

		t.Run("unbuffered chan", func(t *testing.T) {
			t.Parallel()

			errChan := make(chan error)
			assert.False(t, state.EmptyErrChanReturningHadContained(errChan))
			assert.Equal(t, 0, len(errChan))
		})
		t.Run("buffered chan", func(t *testing.T) {
			t.Parallel()

			for i := 1; i < 10; i++ {
				errChan := make(chan error, i)
				assert.False(t, state.EmptyErrChanReturningHadContained(errChan))
				assert.Equal(t, 0, len(errChan))
			}
		})
	})
	t.Run("chan containing elements should return true", func(t *testing.T) {
		t.Parallel()

		t.Run("unbuffered chan", func(t *testing.T) {
			t.Parallel()

			errChan := make(chan error)
			go func() {
				errChan <- errors.New("test")
			}()

			time.Sleep(time.Second) // allow the go routine to start

			assert.True(t, state.EmptyErrChanReturningHadContained(errChan))
			assert.Equal(t, 0, len(errChan))
		})
		t.Run("buffered chan", func(t *testing.T) {
			t.Parallel()

			for i := 1; i < 10; i++ {
				errChan := make(chan error, i)
				for j := 0; j < i; j++ {
					errChan <- errors.New("test")
				}

				assert.True(t, state.EmptyErrChanReturningHadContained(errChan))
				assert.Equal(t, 0, len(errChan))
			}
		})
	})
}
