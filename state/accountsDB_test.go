package state_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	mathRand "math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/iteratorChannelsProvider"
	"github.com/multiversx/mx-chain-go/state/lastSnapshotMarker"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/disabled"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/evictionWaitingList"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/dataTrieMigrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const trieDbOperationDelay = time.Second

func createMockAccountsDBArgs() state.ArgsAccountsDB {
	accCreator := &stateMock.AccountsFactoryStub{
		CreateAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return stateMock.NewAccountWrapMock(address), nil
		},
	}

	snapshotsManager, _ := state.NewSnapshotsManager(state.ArgsNewSnapshotsManager{
		ProcessingMode:       common.Normal,
		Marshaller:           &marshallerMock.MarshalizerMock{},
		AddressConverter:     &testscommon.PubkeyConverterMock{},
		ProcessStatusHandler: &testscommon.ProcessStatusHandlerStub{},
		StateMetrics:         &stateMock.StateMetricsStub{},
		AccountFactory:       accCreator,
		ChannelsProvider:     iteratorChannelsProvider.NewUserStateIteratorChannelsProvider(),
		LastSnapshotMarker:   lastSnapshotMarker.NewLastSnapshotMarker(),
		StateStatsHandler:    statistics.NewStateStatistics(),
	})

	return state.ArgsAccountsDB{
		Trie: &trieMock.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &storageManager.StorageManagerStub{}
			},
		},
		Hasher:                &hashingMocks.HasherMock{},
		Marshaller:            &marshallerMock.MarshalizerMock{},
		AccountFactory:        accCreator,
		StoragePruningManager: disabled.NewDisabledStoragePruningManager(),
		AddressConverter:      &testscommon.PubkeyConverterMock{},
		SnapshotsManager:      snapshotsManager,
	}
}

func createUserAcc(address []byte) state.UserAccountHandler {
	acc, _ := accounts.NewUserAccount(address, &trieMock.DataTrieTrackerStub{}, &trieMock.TrieLeafParserStub{})
	return acc
}

func generateAccountDBFromTrie(trie common.Trie) *state.AccountsDB {
	args := createMockAccountsDBArgs()
	args.Trie = trie

	adb, _ := state.NewAccountsDB(args)
	return adb
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
	adb, tr, _ := getDefaultStateComponents(testscommon.NewSnapshotPruningStorerMock(), &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	return tr, adb
}

func getDefaultTrieAndAccountsDbWithCustomDB(db common.BaseStorer) (common.Trie, *state.AccountsDB) {
	adb, tr, _ := getDefaultStateComponents(db, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	return tr, adb
}

func getDefaultStateComponents(
	db common.BaseStorer,
	enableEpochsHandler common.EnableEpochsHandler,
) (*state.AccountsDB, common.Trie, common.StorageManager) {
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:      1000,
		SnapshotsBufferLen:    10,
		SnapshotsGoroutineNum: 1,
	}
	marshaller := &marshallerMock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	args := storage.GetStorageManagerArgs()
	args.MainStorer = db
	trieStorage, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(trieStorage, marshaller, hasher, enableEpochsHandler, 5)
	ewlArgs := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: 100,
		HashesSize:     10000,
	}
	ewl, _ := evictionWaitingList.NewMemoryEvictionWaitingList(ewlArgs)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, generalCfg.PruningBufferLen)
	argsAccCreator := factory.ArgsAccountCreator{
		Hasher:              hasher,
		Marshaller:          marshaller,
		EnableEpochsHandler: enableEpochsHandler,
	}
	accCreator, _ := factory.NewAccountCreator(argsAccCreator)

	snapshotsManager, _ := state.NewSnapshotsManager(state.ArgsNewSnapshotsManager{
		ProcessingMode:       common.Normal,
		Marshaller:           marshaller,
		AddressConverter:     &testscommon.PubkeyConverterMock{},
		ProcessStatusHandler: &testscommon.ProcessStatusHandlerStub{},
		StateMetrics:         &stateMock.StateMetricsStub{},
		AccountFactory:       accCreator,
		ChannelsProvider:     iteratorChannelsProvider.NewUserStateIteratorChannelsProvider(),
		LastSnapshotMarker:   lastSnapshotMarker.NewLastSnapshotMarker(),
		StateStatsHandler:    statistics.NewStateStatistics(),
	})

	argsAccountsDB := state.ArgsAccountsDB{
		Trie:                  tr,
		Hasher:                hasher,
		Marshaller:            marshaller,
		AccountFactory:        accCreator,
		StoragePruningManager: spm,
		AddressConverter:      &testscommon.PubkeyConverterMock{},
		SnapshotsManager:      snapshotsManager,
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
	t.Run("nil address converter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.AddressConverter = nil

		adb, err := state.NewAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilAddressConverter, err)
	})
	t.Run("nil snapshots manager should error", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()
		args.SnapshotsManager = nil

		adb, err := state.NewAccountsDB(args)
		assert.True(t, check.IfNil(adb))
		assert.Equal(t, state.ErrNilSnapshotsManager, err)
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
			return &storageManager.StorageManagerStub{}
		},
	})

	err := adb.SaveAccount(nil)
	assert.True(t, errors.Is(err, state.ErrNilAccountHandler))
}

func TestAccountsDB_SaveAccountErrWhenGettingOldAccountShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("trie get err")
	adb := generateAccountDBFromTrie(&trieMock.TrieStub{
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			return nil, 0, expectedErr
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
		},
	})

	err := adb.SaveAccount(generateAccount())
	assert.Equal(t, expectedErr, err)
}

func TestAccountsDB_SaveAccountNilOldAccount(t *testing.T) {
	t.Parallel()

	adb := generateAccountDBFromTrie(&trieMock.TrieStub{
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			return nil, 0, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
		},
	})

	acc := createUserAcc([]byte("someAddress"))
	err := adb.SaveAccount(acc)
	assert.Nil(t, err)
	assert.Equal(t, 1, adb.JournalLen())
}

func TestAccountsDB_SaveAccountExistingOldAccount(t *testing.T) {
	t.Parallel()

	acc := createUserAcc([]byte("someAddress"))
	adb := generateAccountDBFromTrie(&trieMock.TrieStub{
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			serializedAcc, err := (&marshallerMock.MarshalizerMock{}).Marshal(acc)
			return serializedAcc, 0, err
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
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
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			return nil, 0, nil
		},
		UpdateWithVersionCalled: func(key, value []byte, version core.TrieNodeVersion) error {
			return nil
		},
		RootCalled: func() (i []byte, err error) {
			return []byte("rootHash"), nil
		},
	}

	adb := generateAccountDBFromTrie(&trieMock.TrieStub{
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			return nil, 0, nil
		},
		UpdateCalled: func(key, value []byte) error {
			updateCalled++
			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
		},
	})

	dtt := &trieMock.DataTrieTrackerStub{
		SaveDirtyDataCalled: func(_ common.Trie) ([]core.TrieData, error) {
			return []core.TrieData{
				{
					Key:     []byte("key"),
					Value:   []byte("value"),
					Version: 0,
				},
			}, nil
		},
		DataTrieCalled: func() common.Trie {
			return trieStub
		},
	}

	accCode := []byte("code")
	acc, _ := accounts.NewUserAccount([]byte("someAddress"), dtt, &trieMock.TrieLeafParserStub{})
	acc.SetCode(accCode)
	_ = acc.SaveKeyValue([]byte("key"), []byte("value"))

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
			return &storageManager.StorageManagerStub{}
		},
	}
	marshaller := &marshallerMock.MarshalizerMock{}
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
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			return nil, 0, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
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
	marshaller := &marshallerMock.MarshalizerMock{}
	trieStub := &trieMock.TrieStub{
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			serializedAcc, err := marshaller.Marshal(stateMock.AccountWrapMock{})
			return serializedAcc, 0, err
		},
		DeleteCalled: func(key []byte) error {
			wasCalled = true
			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
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
			return &storageManager.StorageManagerStub{}
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
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			return nil, 0, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
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
	marshaller := &marshallerMock.MarshalizerMock{}

	trieStub := &trieMock.TrieStub{
		GetCalled: func(key []byte) ([]byte, uint32, error) {
			if bytes.Equal(key, acc.AddressBytes()) {
				serializedAcc, err := marshaller.Marshal(acc)
				return serializedAcc, 0, err
			}
			if bytes.Equal(key, codeHash) {
				serializedCodeEntry, err := marshaller.Marshal(state.CodeEntry{Code: code})
				return serializedCodeEntry, 0, err
			}
			return nil, 0, nil
		},
		RecreateCalled: func(root []byte) (d common.Trie, err error) {
			return dataTrie, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
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
			return &storageManager.StorageManagerStub{}
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
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			return nil, 0, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
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
	marshaller := &marshallerMock.MarshalizerMock{}

	trieStub := &trieMock.TrieStub{
		GetCalled: func(key []byte) ([]byte, uint32, error) {
			if bytes.Equal(key, acc.AddressBytes()) {
				serializedAcc, err := marshaller.Marshal(acc)
				return serializedAcc, 0, err
			}
			if bytes.Equal(key, codeHash) {
				serializedCodeEntry, err := marshaller.Marshal(state.CodeEntry{Code: code})
				return serializedCodeEntry, 0, err
			}
			return nil, 0, nil
		},
		RecreateCalled: func(root []byte) (d common.Trie, err error) {
			return dataTrie, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
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
			return &storageManager.StorageManagerStub{}
		},
	}
	adr, _, _ := generateAddressAccountAccountsDB(tr)

	// Step 1. Create an account + its DbAccount representation
	testAccount := stateMock.NewAccountWrapMock(adr)
	testAccount.MockValue = 45

	// Step 2. marshalize the DbAccount
	marshaller := &marshallerMock.MarshalizerMock{}
	buff, err := marshaller.Marshal(testAccount)
	assert.Nil(t, err)

	tr.GetCalled = func(_ []byte) ([]byte, uint32, error) {
		// whatever the key is, return the same marshalized DbAccount
		return buff, 0, nil
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
			return &storageManager.StorageManagerStub{}
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
			return &storageManager.StorageManagerStub{}
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
			return &storageManager.StorageManagerStub{}
		},
	}
	adr, account, _ := generateAddressAccountAccountsDB(tr)
	marshaller := &marshallerMock.MarshalizerMock{}

	trieStub := &trieMock.TrieStub{
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			// will return adr.Bytes() so its hash will correspond to adr.Hash()
			serializedCodeEntry, err := marshaller.Marshal(&state.CodeEntry{Code: adr})
			return serializedCodeEntry, 0, err
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
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
			return &storageManager.StorageManagerStub{}
		},
	}
	_, account, adb := generateAddressAccountAccountsDB(tr)

	// since root is nil, result should be nil and data trie should be nil
	err := adb.LoadDataTrieConcurrentSafe(account)
	assert.Nil(t, err)
	assert.Nil(t, account.DataTrie())
}

func TestAccountsDB_LoadDataBadLengthShouldErr(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB(&trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
		},
	})

	account.SetRootHash([]byte("12345"))

	// should return error
	err := adb.LoadDataTrieConcurrentSafe(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadDataMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	account := generateAccount()
	account.SetRootHash([]byte("12345"))

	mockTrie := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
		},
	}
	adb := generateAccountDBFromTrie(mockTrie)

	// should return error
	err := adb.LoadDataTrieConcurrentSafe(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadDataNotFoundRootShouldReturnErr(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB(&trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
		},
	})

	rootHash := make([]byte, (&hashingMocks.HasherMock{}).Size())
	rootHash[0] = 1
	account.SetRootHash(rootHash)

	// should return error
	err := adb.LoadDataTrieConcurrentSafe(account)
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
		GetCalled: func(key []byte) ([]byte, uint32, error) {
			if bytes.Equal(key, keyRequired) {
				return trieVal, 0, nil
			}

			return nil, 0, nil
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
			return &storageManager.StorageManagerStub{}
		},
	}
	adb := generateAccountDBFromTrie(mockTrie)

	account.SetRootHash(rootHash)

	// should not return error
	err := adb.LoadDataTrieConcurrentSafe(account)
	assert.Nil(t, err)

	// verify data
	dataRecov, _, err := account.RetrieveValue(keyRequired)
	assert.Nil(t, err)
	assert.Equal(t, val, dataRecov)
}

// ------- Commit

func TestAccountsDB_CommitShouldCallCommitFromTrie(t *testing.T) {
	t.Parallel()

	commitCalled := 0
	marshaller := &marshallerMock.MarshalizerMock{}
	serializedAccount, _ := marshaller.Marshal(stateMock.AccountWrapMock{})
	trieStub := trieMock.TrieStub{
		CommitCalled: func() error {
			commitCalled++

			return nil
		},
		RootCalled: func() (i []byte, e error) {
			return nil, nil
		},
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			return serializedAccount, 0, nil
		},
		RecreateCalled: func(root []byte) (trie common.Trie, err error) {
			return &trieMock.TrieStub{
				GetCalled: func(_ []byte) ([]byte, uint32, error) {
					return []byte("doge"), 0, nil
				},
				UpdateWithVersionCalled: func(key, value []byte, version core.TrieNodeVersion) error {
					return nil
				},
				CommitCalled: func() error {
					commitCalled++

					return nil
				},
				RootCalled: func() ([]byte, error) {
					return nil, nil
				},
			}, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
		},
	}

	adb := generateAccountDBFromTrie(&trieStub)

	accnt, _ := adb.LoadAccount(make([]byte, 32))
	_ = accnt.(state.UserAccountHandler).SaveKeyValue([]byte("dog"), []byte("puppy"))
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
			return &storageManager.StorageManagerStub{}
		},
	}
	trieStub.RecreateFromEpochCalled = func(_ common.RootHashHolder) (tree common.Trie, e error) {
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
			return &storageManager.StorageManagerStub{}
		},
	}
	trieStub.RecreateFromEpochCalled = func(_ common.RootHashHolder) (tree common.Trie, e error) {
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
			return &storageManager.StorageManagerStub{}
		},
		RecreateFromEpochCalled: func(_ common.RootHashHolder) (common.Trie, error) {
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
			return &storageManager.StorageManagerStub{
				TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, _ *common.TrieIteratorChannels, _ chan []byte, _ common.SnapshotStatisticsHandler, _ uint32) {
					snapshotMut.Lock()
					takeSnapshotWasCalled = true
					snapshotMut.Unlock()
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.SnapshotState([]byte("roothash"), 0)
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
			return &storageManager.StorageManagerStub{
				ShouldTakeSnapshotCalled: func() bool {
					return true
				},
				TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, iteratorChannels *common.TrieIteratorChannels, _ chan []byte, stats common.SnapshotStatisticsHandler, _ uint32) {
					close(iteratorChannels.LeavesChan)
					stats.SnapshotFinished()
				},
				IsClosedCalled: func() bool {
					return true
				},
				PutInEpochCalled: func(key []byte, val []byte, epoch uint32) error {
					mut.Lock()
					defer mut.Unlock()

					if string(key) == common.ActiveDBKey {
						activeDBWasPut = true
					}

					if string(key) == lastSnapshotMarker.LastSnapshot {
						lastSnapshotStartedWasPut = true
					}

					return nil
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.SnapshotState([]byte("roothash"), 0)
	time.Sleep(time.Second)

	mut.RLock()
	defer mut.RUnlock()
	assert.False(t, lastSnapshotStartedWasPut)
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
			return &storageManager.StorageManagerStub{
				ShouldTakeSnapshotCalled: func() bool {
					return true
				},
				TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, iteratorChannels *common.TrieIteratorChannels, _ chan []byte, stats common.SnapshotStatisticsHandler, _ uint32) {
					iteratorChannels.ErrChan.WriteInChanNonBlocking(expectedErr)
					close(iteratorChannels.LeavesChan)
					stats.SnapshotFinished()
				},
				IsClosedCalled: func() bool {
					return false
				},
				PutInEpochCalled: func(key []byte, val []byte, epoch uint32) error {
					mut.Lock()
					defer mut.Unlock()

					if string(key) == common.ActiveDBKey {
						activeDBWasPut = true
					}

					if string(key) == lastSnapshotMarker.LastSnapshot {
						lastSnapshotStartedWasPut = true
					}

					return nil
				},
				GetLatestStorageEpochCalled: func() (uint32, error) {
					return 1, nil
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.SnapshotState([]byte("roothash"), 1)
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
			return &storageManager.StorageManagerStub{
				GetLatestStorageEpochCalled: func() (uint32, error) {
					return 0, fmt.Errorf("new error")
				},
				TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, iteratorChannels *common.TrieIteratorChannels, _ chan []byte, _ common.SnapshotStatisticsHandler, _ uint32) {
					takeSnapshotCalled = true
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	adb.SnapshotState([]byte("roothash"), 0)
	time.Sleep(time.Second)

	assert.False(t, takeSnapshotCalled)
}

func TestAccountsDB_SnapshotStateSnapshotSameRootHash(t *testing.T) {
	t.Parallel()

	rootHash1 := []byte("rootHash1")
	rootHash2 := []byte("rootHash2")
	latestEpoch := atomic.Uint32{}
	snapshotMutex := sync.RWMutex{}
	takeSnapshotCalled := 0
	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{
				GetLatestStorageEpochCalled: func() (uint32, error) {
					return latestEpoch.Get(), nil
				},
				TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, iteratorChannels *common.TrieIteratorChannels, _ chan []byte, stats common.SnapshotStatisticsHandler, _ uint32) {
					snapshotMutex.Lock()
					takeSnapshotCalled++
					close(iteratorChannels.LeavesChan)
					stats.SnapshotFinished()
					snapshotMutex.Unlock()
				},
			}
		},
	}
	args := createMockAccountsDBArgs()
	args.Trie = trieStub

	adb, _ := state.NewAccountsDB(args)
	waitForOpToFinish := time.Millisecond * 10

	// snapshot rootHash1 and epoch 0
	adb.SnapshotState(rootHash1, 0)
	for adb.IsSnapshotInProgress() {
		time.Sleep(waitForOpToFinish)
	}
	snapshotMutex.Lock()
	assert.Equal(t, 1, takeSnapshotCalled)
	snapshotMutex.Unlock()

	// snapshot rootHash1 and epoch 1
	latestEpoch.Set(1)
	adb.SnapshotState(rootHash1, 1)
	for adb.IsSnapshotInProgress() {
		time.Sleep(waitForOpToFinish)
	}
	snapshotMutex.Lock()
	assert.Equal(t, 2, takeSnapshotCalled)
	snapshotMutex.Unlock()

	// snapshot rootHash1 and epoch 0 again
	latestEpoch.Set(0)
	adb.SnapshotState(rootHash1, 0)
	for adb.IsSnapshotInProgress() {
		time.Sleep(waitForOpToFinish)
	}
	snapshotMutex.Lock()
	assert.Equal(t, 3, takeSnapshotCalled)
	snapshotMutex.Unlock()

	// snapshot rootHash1 and epoch 0 again
	adb.SnapshotState(rootHash1, 0)
	for adb.IsSnapshotInProgress() {
		time.Sleep(waitForOpToFinish)
	}
	snapshotMutex.Lock()
	assert.Equal(t, 3, takeSnapshotCalled)
	snapshotMutex.Unlock()

	// snapshot rootHash2 and epoch 0
	adb.SnapshotState(rootHash2, 0)
	for adb.IsSnapshotInProgress() {
		time.Sleep(waitForOpToFinish)
	}
	snapshotMutex.Lock()
	assert.Equal(t, 4, takeSnapshotCalled)
	snapshotMutex.Unlock()

	// snapshot rootHash2 and epoch 1
	latestEpoch.Set(1)
	adb.SnapshotState(rootHash2, 1)
	for adb.IsSnapshotInProgress() {
		time.Sleep(waitForOpToFinish)
	}
	snapshotMutex.Lock()
	assert.Equal(t, 5, takeSnapshotCalled)
	snapshotMutex.Unlock()

	// snapshot rootHash2 and epoch 1 again
	latestEpoch.Set(1)
	adb.SnapshotState(rootHash2, 1)
	for adb.IsSnapshotInProgress() {
		time.Sleep(waitForOpToFinish)
	}
	snapshotMutex.Lock()
	assert.Equal(t, 5, takeSnapshotCalled)
	snapshotMutex.Unlock()
}

func TestAccountsDB_SnapshotStateSkipSnapshotIfSnapshotInProgress(t *testing.T) {
	t.Parallel()

	rootHashes := [][]byte{[]byte("rootHash1"), []byte("rootHash2"), []byte("rootHash3"), []byte("rootHash4")}
	snapshotMutex := sync.RWMutex{}
	takeSnapshotCalled := 0
	numPutInEpochCalled := atomic.Counter{}

	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{
				GetLatestStorageEpochCalled: func() (uint32, error) {
					return uint32(mathRand.Intn(5)), nil
				},
				TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, iteratorChannels *common.TrieIteratorChannels, _ chan []byte, stats common.SnapshotStatisticsHandler, _ uint32) {
					snapshotMutex.Lock()
					takeSnapshotCalled++
					close(iteratorChannels.LeavesChan)
					stats.SnapshotFinished()
					for numPutInEpochCalled.Get() != 4 {
						time.Sleep(time.Millisecond * 10)
					}
					snapshotMutex.Unlock()
				},
				PutInEpochCalled: func(key []byte, val []byte, epoch uint32) error {
					assert.Equal(t, []byte(lastSnapshotMarker.LastSnapshot), key)
					assert.Equal(t, rootHashes[epoch-1], val)

					numPutInEpochCalled.Add(1)
					return nil
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)

	for i, rootHash := range rootHashes {
		epoch := i + 1
		adb.SnapshotState(rootHash, uint32(epoch))
	}
	for adb.IsSnapshotInProgress() {
		time.Sleep(time.Millisecond * 10)
	}

	snapshotMutex.Lock()
	assert.Equal(t, 1, takeSnapshotCalled)
	snapshotMutex.Unlock()
	assert.Equal(t, len(rootHashes), int(numPutInEpochCalled.Get()))
}

func TestAccountsDB_SnapshotStateCallsRemoveFromAllActiveEpochs(t *testing.T) {
	t.Parallel()

	latestEpoch := uint32(0)
	removeFromAllActiveEpochsCalled := false

	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{
				GetLatestStorageEpochCalled: func() (uint32, error) {
					return latestEpoch, nil
				},
				TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, iteratorChannels *common.TrieIteratorChannels, _ chan []byte, stats common.SnapshotStatisticsHandler, _ uint32) {
					close(iteratorChannels.LeavesChan)
					stats.SnapshotFinished()
				},
				RemoveFromAllActiveEpochsCalled: func(hash []byte) error {
					removeFromAllActiveEpochsCalled = true
					assert.Equal(t, []byte(lastSnapshotMarker.LastSnapshot), hash)
					return nil
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(trieStub)
	_ = adb.SetSyncer(&mock.AccountsDBSyncerStub{})

	adb.SnapshotState([]byte("rootHash"), 0)
	for adb.IsSnapshotInProgress() {
		time.Sleep(time.Millisecond * 10)
	}
	assert.True(t, removeFromAllActiveEpochsCalled)
}

func TestAccountsDB_IsPruningEnabled(t *testing.T) {
	t.Parallel()

	trieStub := &trieMock.TrieStub{
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{
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
			return &storageManager.StorageManagerStub{}
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

	err = userAcc.SaveKeyValue(key, value)
	assert.Nil(t, err)
	err = adb.SaveAccount(userAcc)
	assert.Nil(t, err)
	_, err = adb.Commit()
	assert.Nil(t, err)

	err = userAcc.SaveKeyValue(key1, value)
	assert.Nil(t, err)
	err = adb.SaveAccount(userAcc)
	assert.Nil(t, err)
	_, err = adb.Commit()
	assert.Nil(t, err)

	account, err = adb.LoadAccount(address)
	assert.Nil(t, err)
	userAcc = account.(state.UserAccountHandler)

	returnedVal, _, err := userAcc.RetrieveValue(key)
	assert.Nil(t, err)
	assert.Equal(t, value, returnedVal)

	returnedVal, _, err = userAcc.RetrieveValue(key1)
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
	_ = acc.(state.UserAccountHandler).SaveKeyValue(key, value)
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
			return &storageManager.StorageManagerStub{}
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
		GetAllLeavesOnChannelCalled: func(channels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, builder common.KeyBuilder, _ common.TrieLeafParser) error {
			getAllLeavesCalled = true
			close(channels.LeavesChan)
			channels.ErrChan.Close()

			return nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
		},
	}

	adb := generateAccountDBFromTrie(trieStub)

	leavesChannel := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err := adb.GetAllLeaves(leavesChannel, context.Background(), []byte("root hash"), parsers.NewMainTrieLeafParser())
	assert.Nil(t, err)
	assert.True(t, getAllLeavesCalled)

	err = leavesChannel.ErrChan.ReadFromChanNonBlocking()
	assert.Nil(t, err)
}

func checkCodeEntry(
	codeHash []byte,
	expectedCode []byte,
	expectedNumReferences uint32,
	marshaller marshal.Marshalizer,
	tr common.Trie,
	t *testing.T,
) {
	val, _, err := tr.Get(codeHash)
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

	marshaller := &marshallerMock.MarshalizerMock{}
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

	marshaller := &marshallerMock.MarshalizerMock{}
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

	val, _, err := tr.Get(expectedCodeHash)
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestAccountsDB_saveCode_OldCodeIsNilAndNewCodeAlreadyExistsAndRevert(t *testing.T) {
	t.Parallel()

	marshaller := &marshallerMock.MarshalizerMock{}
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

	marshaller := &marshallerMock.MarshalizerMock{}
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

	val, _, err := tr.Get(oldCodeHash)
	assert.Nil(t, err)
	assert.Nil(t, val)

	err = adb.RevertToSnapshot(journalLen)
	assert.Nil(t, err)

	checkCodeEntry(oldCodeHash, code, 1, marshaller, tr, t)
}

func TestAccountsDB_saveCode_OldCodeExistsAndNewCodeExistsAndRevert(t *testing.T) {
	t.Parallel()

	marshaller := &marshallerMock.MarshalizerMock{}
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

	val, _, err := tr.Get(oldCodeHash)
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

	marshaller := &marshallerMock.MarshalizerMock{}
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

	val, _, err := tr.Get(oldCodeHash)
	assert.Nil(t, err)
	assert.NotNil(t, val)

	err = adb.RemoveAccount(addr)
	assert.Nil(t, err)

	val, _, err = tr.Get(oldCodeHash)
	assert.Nil(t, err)
	assert.Nil(t, val)

	_ = adb.RevertToSnapshot(snapshot)

	val, _, err = tr.Get(oldCodeHash)
	assert.Nil(t, err)
	assert.NotNil(t, val)
}

func TestAccountsDB_MainTrieAutomaticallyMarksCodeUpdatesForEviction(t *testing.T) {
	t.Parallel()

	marshaller := &marshallerMock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	ewl := stateMock.NewEvictionWaitingListMock(100)
	args := storage.GetStorageManagerArgs()
	tsm, _ := trie.NewTrieStorageManager(args)
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(tsm, marshaller, hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, maxTrieLevelInMemory)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 5)

	argsAccountsDB := createMockAccountsDBArgs()
	argsAccountsDB.Trie = tr
	argsAccountsDB.Hasher = hasher
	argsAccountsDB.Marshaller = marshaller
	argsAccCreator := factory.ArgsAccountCreator{
		Hasher:              hasher,
		Marshaller:          marshaller,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	argsAccountsDB.AccountFactory, _ = factory.NewAccountCreator(argsAccCreator)
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
	_ = userAcc.SaveKeyValue([]byte("key"), []byte("value"))

	_ = adb.SaveAccount(userAcc)
	_, _ = adb.Commit()

	acc, _ = adb.LoadAccount(addr)
	userAcc = acc.(state.UserAccountHandler)
	userAcc.SetCode([]byte("code"))
	snapshot := adb.JournalLen()
	hashes, _ := userAcc.DataTrie().(common.Trie).GetAllHashes()

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
	marshaller := &marshallerMock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	ewl := stateMock.NewEvictionWaitingListMock(100)
	args := storage.GetStorageManagerArgs()
	tsm, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(tsm, marshaller, hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, maxTrieLevelInMemory)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 5)

	argsAccountsDB := createMockAccountsDBArgs()
	argsAccountsDB.Trie = tr
	argsAccountsDB.Hasher = hasher
	argsAccountsDB.Marshaller = marshaller
	argsAccCreator := factory.ArgsAccountCreator{
		Hasher:              hasher,
		Marshaller:          marshaller,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	argsAccountsDB.AccountFactory, _ = factory.NewAccountCreator(argsAccCreator)
	argsAccountsDB.StoragePruningManager = spm

	adb, _ := state.NewAccountsDB(argsAccountsDB)

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)
	_ = userAcc.SaveKeyValue([]byte("key"), []byte("value"))
	_ = adb.SaveAccount(userAcc)

	addr1 := make([]byte, 32)
	addr1[0] = 1
	acc, _ = adb.LoadAccount(addr1)
	_ = adb.SaveAccount(acc)

	rootHash, _ := adb.Commit()
	hashes, _ := userAcc.DataTrie().(common.Trie).GetAllHashes()

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

func mergeMaps(map1 common.ModifiedHashes, map2 common.ModifiedHashes) {
	for key, val := range map2 {
		map1[key] = val
	}
}

func generateAccounts(t testing.TB, numAccounts int, adb state.AccountsAdapter) [][]byte {
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
	err := acc.(state.UserAccountHandler).SaveKeyValue([]byte("key1"), []byte("value1"))
	assert.Nil(t, err)
	err = acc.(state.UserAccountHandler).SaveKeyValue([]byte("key2"), []byte("value2"))
	assert.Nil(t, err)
	_ = adb.SaveAccount(acc)
	newHashes, _ := acc.(state.UserAccountHandler).DataTrie().(common.Trie).GetDirtyHashes()

	acc, _ = adb.LoadAccount(accountsAddresses[1])
	err = acc.(state.UserAccountHandler).SaveKeyValue([]byte("key2"), []byte("value2"))
	assert.Nil(t, err)
	_ = adb.SaveAccount(acc)
	newHashesDataTrie, _ := acc.(state.UserAccountHandler).DataTrie().(common.Trie).GetDirtyHashes()
	mergeMaps(newHashes, newHashesDataTrie)

	return newHashes
}

func TestAccountsDB_GetCode(t *testing.T) {
	t.Parallel()

	maxTrieLevelInMemory := uint(5)
	marshaller := &marshallerMock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	args := storage.GetStorageManagerArgs()
	tsm, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(tsm, marshaller, hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, maxTrieLevelInMemory)
	spm := disabled.NewDisabledStoragePruningManager()
	argsAccountsDB := createMockAccountsDBArgs()
	argsAccountsDB.Trie = tr
	argsAccountsDB.Hasher = hasher
	argsAccountsDB.Marshaller = marshaller
	argsAccCreator := factory.ArgsAccountCreator{
		Hasher:              hasher,
		Marshaller:          marshaller,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	argsAccountsDB.AccountFactory, _ = factory.NewAccountCreator(argsAccCreator)
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

	t.Run("error on getting all leaves", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()

		expectedErr := errors.New("expected error")
		args.Trie = &trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder common.KeyBuilder, _ common.TrieLeafParser) error {
				go func() {
					leavesChannels.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("key"), []byte("val"))
					leavesChannels.ErrChan.WriteInChanNonBlocking(expectedErr)

					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RecreateCalled: func(root []byte) (common.Trie, error) {
				return &trieMock.TrieStub{}, nil
			},
			GetStorageManagerCalled: func() common.StorageManager {
				return &storageManager.StorageManagerStub{}
			},
		}

		adb, _ := state.NewAccountsDB(args)

		tries, err := adb.RecreateAllTries([]byte{})
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 0, len(tries))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsDBArgs()

		args.Trie = &trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder common.KeyBuilder, _ common.TrieLeafParser) error {
				go func() {
					leavesChannels.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("key"), []byte("val"))

					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RecreateCalled: func(root []byte) (common.Trie, error) {
				return &trieMock.TrieStub{}, nil
			},
			GetStorageManagerCalled: func() common.StorageManager {
				return &storageManager.StorageManagerStub{}
			},
		}

		adb, _ := state.NewAccountsDB(args)

		tries, err := adb.RecreateAllTries([]byte{})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(tries))
	})
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
			return &storageManager.StorageManagerStub{}
		},
	}
	marshaller := &marshallerMock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	ewlArgs := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: 100,
		HashesSize:     10000,
	}
	ewl, _ := evictionWaitingList.NewMemoryEvictionWaitingList(ewlArgs)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)

	argsAccountsDB := createMockAccountsDBArgs()
	argsAccountsDB.Trie = tr
	argsAccountsDB.Hasher = hasher
	argsAccountsDB.Marshaller = marshaller
	argsAccCreator := factory.ArgsAccountCreator{
		Hasher:              hasher,
		Marshaller:          marshaller,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	argsAccountsDB.AccountFactory, _ = factory.NewAccountCreator(argsAccCreator)
	argsAccountsDB.StoragePruningManager = spm

	adb, _ := state.NewAccountsDB(argsAccountsDB)

	err := adb.Close()
	assert.Nil(t, err)
	assert.True(t, closeCalled)
}

func TestAccountsDB_GetAccountFromBytes(t *testing.T) {
	t.Parallel()

	t.Run("nil address", func(t *testing.T) {
		t.Parallel()

		adb, _ := state.NewAccountsDB(createMockAccountsDBArgs())

		acc, err := adb.GetAccountFromBytes(nil, []byte{})
		assert.Nil(t, acc)
		assert.True(t, strings.Contains(err.Error(), state.ErrNilAddress.Error()))
	})

	t.Run("accountFactory error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockAccountsDBArgs()
		args.AccountFactory = &stateMock.AccountsFactoryStub{
			CreateAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
				return nil, expectedErr
			},
		}
		adb, _ := state.NewAccountsDB(args)

		acc, err := adb.GetAccountFromBytes([]byte{1}, []byte{})
		assert.Nil(t, acc)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockAccountsDBArgs()
		args.Marshaller = &marshallerMock.MarshalizerStub{
			UnmarshalCalled: func(_ interface{}, _ []byte) error {
				return expectedErr
			},
		}
		adb, _ := state.NewAccountsDB(args)

		acc, err := adb.GetAccountFromBytes([]byte{1}, []byte{})
		assert.Nil(t, acc)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("return peer account directly", func(t *testing.T) {
		t.Parallel()

		expectedAccount := &stateMock.PeerAccountHandlerMock{}
		args := createMockAccountsDBArgs()
		args.AccountFactory = &stateMock.AccountsFactoryStub{
			CreateAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
				return expectedAccount, nil
			},
		}
		args.Marshaller = &marshallerMock.MarshalizerStub{
			UnmarshalCalled: func(_ interface{}, _ []byte) error {
				return nil
			},
		}
		adb, _ := state.NewAccountsDB(args)

		acc, err := adb.GetAccountFromBytes([]byte{1}, []byte{})
		assert.Nil(t, err)
		assert.Equal(t, expectedAccount, acc)
	})

	t.Run("loads data trie for user account", func(t *testing.T) {
		t.Parallel()

		rootHash := []byte("root hash")
		setDataTrieCalled := false
		expectedAccount := &stateMock.UserAccountStub{
			SetDataTrieCalled: func(_ common.Trie) {
				setDataTrieCalled = true
			},
			GetRootHashCalled: func() []byte {
				return rootHash
			},
		}

		args := createMockAccountsDBArgs()
		args.AccountFactory = &stateMock.AccountsFactoryStub{
			CreateAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
				return expectedAccount, nil
			},
		}
		args.Marshaller = &marshallerMock.MarshalizerStub{
			UnmarshalCalled: func(_ interface{}, _ []byte) error {
				return nil
			},
		}
		args.Trie = &trieMock.TrieStub{
			RecreateCalled: func(root []byte) (common.Trie, error) {
				assert.Equal(t, rootHash, root)
				return &trieMock.TrieStub{}, nil
			},
			GetStorageManagerCalled: func() common.StorageManager {
				return &storageManager.StorageManagerStub{}
			},
		}
		adb, _ := state.NewAccountsDB(args)

		acc, err := adb.GetAccountFromBytes([]byte{1}, []byte{})
		assert.Nil(t, err)
		assert.Equal(t, expectedAccount, acc)
		assert.True(t, setDataTrieCalled)
	})
}

func TestAccountsDB_GetAccountFromBytesShouldLoadDataTrie(t *testing.T) {
	t.Parallel()

	acc := generateAccount()
	acc.SetRootHash([]byte("root hash"))
	dataTrie := &trieMock.TrieStub{}
	marshaller := &marshallerMock.MarshalizerMock{}
	serializerAcc, _ := marshaller.Marshal(acc)

	trieStub := &trieMock.TrieStub{
		GetCalled: func(key []byte) ([]byte, uint32, error) {
			if bytes.Equal(key, acc.AddressBytes()) {
				return serializerAcc, 0, nil
			}
			return nil, 0, nil
		},
		RecreateCalled: func(root []byte) (d common.Trie, err error) {
			return dataTrie, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{}
		},
	}

	adb := generateAccountDBFromTrie(trieStub)
	retrievedAccount, err := adb.GetAccountFromBytes(acc.AddressBytes(), serializerAcc)
	assert.Nil(t, err)

	account, _ := retrievedAccount.(state.UserAccountHandler)
	assert.Equal(t, dataTrie, account.DataTrie())
}

func TestAccountsDB_SetSyncerAndStartSnapshotIfNeeded(t *testing.T) {
	t.Parallel()

	rootHash := []byte("rootHash")
	expectedErr := errors.New("expected error")

	t.Run("epoch 0, GetLatestStorageEpoch errors should not put", func(t *testing.T) {
		trieStub := &trieMock.TrieStub{
			RootCalled: func() ([]byte, error) {
				return rootHash, nil
			},
			GetStorageManagerCalled: func() common.StorageManager {
				return &storageManager.StorageManagerStub{
					ShouldTakeSnapshotCalled: func() bool {
						return true
					},
					GetLatestStorageEpochCalled: func() (uint32, error) {
						return 0, expectedErr
					},
					PutCalled: func(key []byte, val []byte) error {
						assert.Fail(t, "should have not called put")

						return nil
					},
				}
			},
		}

		adb := generateAccountDBFromTrie(trieStub)
		err := adb.SetSyncer(&mock.AccountsDBSyncerStub{})
		assert.Nil(t, err)
		err = adb.StartSnapshotIfNeeded()
		assert.Nil(t, err)
	})
	t.Run("in import DB mode", func(t *testing.T) {
		trieStub := &trieMock.TrieStub{
			RootCalled: func() ([]byte, error) {
				return rootHash, nil
			},
			GetStorageManagerCalled: func() common.StorageManager {
				return &storageManager.StorageManagerStub{
					ShouldTakeSnapshotCalled: func() bool {
						return true
					},
					GetLatestStorageEpochCalled: func() (uint32, error) {
						return 1, nil
					},
					GetFromCurrentEpochCalled: func(i []byte) ([]byte, error) {
						return nil, expectedErr
					},
				}
			},
		}

		args := createMockAccountsDBArgs()
		args.SnapshotsManager, _ = state.NewSnapshotsManager(state.ArgsNewSnapshotsManager{
			ProcessingMode:       common.ImportDb,
			Marshaller:           &marshallerMock.MarshalizerMock{},
			AddressConverter:     &testscommon.PubkeyConverterMock{},
			ProcessStatusHandler: &testscommon.ProcessStatusHandlerStub{},
			StateMetrics:         &stateMock.StateMetricsStub{},
			AccountFactory:       args.AccountFactory,
			ChannelsProvider:     iteratorChannelsProvider.NewUserStateIteratorChannelsProvider(),
			LastSnapshotMarker:   lastSnapshotMarker.NewLastSnapshotMarker(),
			StateStatsHandler:    statistics.NewStateStatistics(),
		})
		args.Trie = trieStub

		adb, _ := state.NewAccountsDB(args)
		err := adb.SetSyncer(&mock.AccountsDBSyncerStub{})
		assert.Nil(t, err)
		err = adb.StartSnapshotIfNeeded()
		assert.Nil(t, err)
	})
}

func TestAccountsDB_NewAccountsDbStartsSnapshotAfterRestart(t *testing.T) {
	t.Parallel()

	rootHash := []byte("rootHash")
	takeSnapshotCalled := atomic.Flag{}
	trieStub := &trieMock.TrieStub{
		RootCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		GetStorageManagerCalled: func() common.StorageManager {
			return &storageManager.StorageManagerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, []byte(common.ActiveDBKey)) {
						return nil, fmt.Errorf("key not found")
					}
					return []byte("rootHash"), nil
				},
				ShouldTakeSnapshotCalled: func() bool {
					return true
				},
				TakeSnapshotCalled: func(_ string, _ []byte, _ []byte, _ *common.TrieIteratorChannels, _ chan []byte, _ common.SnapshotStatisticsHandler, _ uint32) {
					takeSnapshotCalled.SetValue(true)
				},
				GetLatestStorageEpochCalled: func() (uint32, error) {
					return 1, nil
				},
			}
		},
	}

	adb := generateAccountDBFromTrie(trieStub)
	err := adb.SetSyncer(&mock.AccountsDBSyncerStub{})
	assert.Nil(t, err)
	err = adb.StartSnapshotIfNeeded()
	assert.Nil(t, err)
	time.Sleep(time.Second)
	assert.True(t, takeSnapshotCalled.IsSet())
}

func BenchmarkAccountsDb_GetCodeEntry(b *testing.B) {
	maxTrieLevelInMemory := uint(5)
	marshaller := &marshallerMock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	args := storage.GetStorageManagerArgs()
	tsm, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(tsm, marshaller, hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, maxTrieLevelInMemory)
	spm := disabled.NewDisabledStoragePruningManager()

	argsAccountsDB := createMockAccountsDBArgs()
	argsAccountsDB.Trie = tr
	argsAccountsDB.Hasher = hasher
	argsAccountsDB.Marshaller = marshaller
	argsAccCreator := factory.ArgsAccountCreator{
		Hasher:              hasher,
		Marshaller:          marshaller,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	argsAccountsDB.AccountFactory, _ = factory.NewAccountCreator(argsAccCreator)
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

func TestEmptyErrChanReturningHadContained(t *testing.T) {
	t.Parallel()

	t.Run("empty chan should return false", func(t *testing.T) {
		t.Parallel()

		t.Run("unbuffered chan", func(t *testing.T) {
			t.Parallel()

			errChannel := make(chan error)
			assert.False(t, state.EmptyErrChanReturningHadContained(errChannel))
			assert.Equal(t, 0, len(errChannel))
		})
		t.Run("buffered chan", func(t *testing.T) {
			t.Parallel()

			for i := 1; i < 10; i++ {
				errChannel := make(chan error, i)
				assert.False(t, state.EmptyErrChanReturningHadContained(errChannel))
				assert.Equal(t, 0, len(errChannel))
			}
		})
	})
	t.Run("chan containing elements should return true", func(t *testing.T) {
		t.Parallel()

		t.Run("unbuffered chan", func(t *testing.T) {
			t.Parallel()

			errChannel := make(chan error)
			go func() {
				errChannel <- errors.New("test")
			}()

			time.Sleep(time.Second) // allow the go routine to start

			assert.True(t, state.EmptyErrChanReturningHadContained(errChannel))
			assert.Equal(t, 0, len(errChannel))
		})
		t.Run("buffered chan", func(t *testing.T) {
			t.Parallel()

			for i := 1; i < 10; i++ {
				errChannel := make(chan error, i)
				for j := 0; j < i; j++ {
					errChannel <- errors.New("test")
				}

				assert.True(t, state.EmptyErrChanReturningHadContained(errChannel))
				assert.Equal(t, 0, len(errChannel))
			}
		})
	})
}

func TestAccountsDB_PrintStatsForRootHash(t *testing.T) {
	t.Parallel()
	_, adb := getDefaultTrieAndAccountsDb()

	numAccounts := 1000
	accountsAddresses := generateAccounts(t, numAccounts, adb)
	addDataTries(accountsAddresses, adb)
	rootHash, _ := adb.Commit()

	stats, err := adb.GetStatsForRootHash(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, stats)

	stats.Print()
}

func TestAccountsDB_SyncMissingSnapshotNodes(t *testing.T) {
	t.Parallel()

	t.Run("can not sync missing snapshot node should not put activeDbKey", func(t *testing.T) {
		t.Parallel()

		trieHashes := make([][]byte, 0)
		valuesMap := make(map[string][]byte)
		missingNodeError := errors.New("missing trie node")
		isMissingNodeCalled := false
		isSyncError := false

		memDbMock := testscommon.NewMemDbMock()
		memDbMock.PutCalled = func(key, val []byte) error {
			require.False(t, bytes.Equal(key, []byte(common.ActiveDBKey)), "should not have put active db key")

			valuesMap[string(key)] = val
			return nil
		}
		memDbMock.GetCalled = func(key []byte) ([]byte, error) {
			if bytes.Equal(key, []byte(common.ActiveDBKey)) {
				return []byte(common.ActiveDBVal), nil
			}

			if len(trieHashes) != 0 && bytes.Equal(key, trieHashes[0]) {
				isMissingNodeCalled = true
				return nil, missingNodeError
			}

			val, ok := valuesMap[string(key)]
			if !ok {
				return nil, errors.New("key not found")
			}
			return val, nil
		}

		tr, adb := getDefaultTrieAndAccountsDbWithCustomDB(&testscommon.SnapshotPruningStorerMock{MemDbMock: memDbMock})
		prepareTrie(tr, 3)

		rootHash, _ := tr.RootHash()
		trieHashes, _ = tr.GetAllHashes()

		syncer := &mock.AccountsDBSyncerStub{
			SyncAccountsCalled: func(rootHash []byte, _ common.StorageMarker) error {
				isSyncError = true
				return errors.New("sync error")
			},
		}
		_ = adb.SetSyncer(syncer)
		adb.SnapshotState(rootHash, 0)

		for adb.IsSnapshotInProgress() {
			time.Sleep(time.Millisecond * 100)
		}

		assert.True(t, isMissingNodeCalled)
		assert.True(t, isSyncError)
	})

	t.Run("nil syncer should not put activeDbKey", func(t *testing.T) {
		t.Parallel()

		trieHashes := make([][]byte, 0)
		valuesMap := make(map[string][]byte)
		missingNodeError := errors.New("missing trie node")
		isMissingNodeCalled := false

		memDbMock := testscommon.NewMemDbMock()
		memDbMock.PutCalled = func(key, val []byte) error {
			require.False(t, bytes.Equal(key, []byte(common.ActiveDBKey)), "should not have put active db key")

			valuesMap[string(key)] = val
			return nil
		}
		memDbMock.GetCalled = func(key []byte) ([]byte, error) {
			if bytes.Equal(key, []byte(common.ActiveDBKey)) {
				return []byte(common.ActiveDBVal), nil
			}

			if len(trieHashes) != 0 && bytes.Equal(key, trieHashes[0]) {
				isMissingNodeCalled = true
				return nil, missingNodeError
			}

			val, ok := valuesMap[string(key)]
			if !ok {
				return nil, errors.New("key not found")
			}
			return val, nil
		}

		tr, adb := getDefaultTrieAndAccountsDbWithCustomDB(&testscommon.SnapshotPruningStorerMock{MemDbMock: memDbMock})
		prepareTrie(tr, 3)

		rootHash, _ := tr.RootHash()
		trieHashes, _ = tr.GetAllHashes()

		adb.SnapshotState(rootHash, 0)

		for adb.IsSnapshotInProgress() {
			time.Sleep(time.Millisecond * 100)
		}

		assert.True(t, isMissingNodeCalled)
	})

	t.Run("should not deadlock if sync err after another err", func(t *testing.T) {
		t.Parallel()

		missingNodeError := errors.New("missing trie node")
		isMissingNodeCalled := false

		memDbMock := testscommon.NewMemDbMock()
		memDbMock.PutCalled = func(key, val []byte) error {
			return fmt.Errorf("put error")
		}
		memDbMock.GetCalled = func(key []byte) ([]byte, error) {
			if bytes.Equal(key, []byte(common.ActiveDBKey)) {
				return []byte(common.ActiveDBVal), nil
			}

			isMissingNodeCalled = true
			return nil, missingNodeError
		}

		tr, adb := getDefaultTrieAndAccountsDbWithCustomDB(&testscommon.SnapshotPruningStorerMock{MemDbMock: memDbMock})
		prepareTrie(tr, 3)

		rootHash, _ := tr.RootHash()

		adb.SnapshotState(rootHash, 0)

		for adb.IsSnapshotInProgress() {
			time.Sleep(time.Millisecond * 100)
		}

		assert.True(t, isMissingNodeCalled)
	})
}

func prepareTrie(tr common.Trie, numKeys int) {
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key%d", i)
		val := fmt.Sprintf("val%d", i)
		_ = tr.Update([]byte(key), []byte(val))
	}

	_ = tr.Commit()
}

func addDataTries(accountsAddresses [][]byte, adb *state.AccountsDB) {
	for i := range accountsAddresses {
		numVals := mathRand.Intn(100)
		acc, _ := adb.LoadAccount(accountsAddresses[i])
		userAcc := acc.(state.UserAccountHandler)
		for j := 0; j < numVals; j++ {
			_ = userAcc.SaveKeyValue(generateRandomByteArray(32), generateRandomByteArray(32))
		}
		_ = adb.SaveAccount(acc)
	}
}

func TestAccountsDb_Concurrent(t *testing.T) {
	_, adb := getDefaultTrieAndAccountsDb()

	numAccounts := 100000
	accountsAddresses := generateAccounts(t, numAccounts, adb)
	rootHash, _ := adb.Commit()

	testAccountMethodsConcurrency(t, adb, accountsAddresses, rootHash)
}

func TestAccountsDB_SaveKeyValAfterAccountIsReverted(t *testing.T) {
	t.Parallel()

	_, adb := getDefaultTrieAndAccountsDb()
	addr := generateRandomByteArray(32)

	acc, _ := adb.LoadAccount(addr)
	_ = adb.SaveAccount(acc)

	acc, _ = adb.LoadAccount(addr)
	acc.(state.UserAccountHandler).IncreaseNonce(1)
	_ = acc.(state.UserAccountHandler).SaveKeyValue([]byte("key"), []byte("value"))
	_ = adb.SaveAccount(acc)

	err := adb.RevertToSnapshot(1)
	require.Nil(t, err)

	acc, _ = adb.LoadAccount(addr)
	_ = acc.(state.UserAccountHandler).SaveKeyValue([]byte("key"), []byte("value"))
	_ = adb.SaveAccount(acc)

	_, err = adb.Commit()
	require.Nil(t, err)

	acc, err = adb.LoadAccount(addr)
	require.Nil(t, err)
	require.NotNil(t, acc)
}

func TestAccountsDB_RevertTxWhichMigratesDataRemovesMigratedData(t *testing.T) {
	t.Parallel()

	marshaller := &marshallerMock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
	tsm, _ := trie.NewTrieStorageManager(storage.GetStorageManagerArgs())
	tr, _ := trie.NewTrie(tsm, marshaller, hasher, enableEpochsHandler, uint(5))
	spm := &stateMock.StoragePruningManagerStub{}
	argsAccountsDB := createMockAccountsDBArgs()
	argsAccountsDB.Trie = tr
	argsAccountsDB.Hasher = hasher
	argsAccountsDB.Marshaller = marshaller
	argsAccCreator := factory.ArgsAccountCreator{
		Hasher:              hasher,
		Marshaller:          marshaller,
		EnableEpochsHandler: enableEpochsHandler,
	}
	argsAccountsDB.AccountFactory, _ = factory.NewAccountCreator(argsAccCreator)
	argsAccountsDB.StoragePruningManager = spm
	adb, _ := state.NewAccountsDB(argsAccountsDB)

	address := make([]byte, 32)
	acc, err := adb.LoadAccount(address)
	require.Nil(t, err)

	// save account with data trie that is not migrated
	userAcc := acc.(state.UserAccountHandler)
	key := []byte("key")
	err = userAcc.SaveKeyValue(key, []byte("value"))
	require.Nil(t, err)
	err = userAcc.SaveKeyValue([]byte("key1"), []byte("value"))
	require.Nil(t, err)

	err = adb.SaveAccount(userAcc)
	userAccRootHash := userAcc.GetRootHash()
	require.Nil(t, err)
	_, err = adb.Commit()
	require.Nil(t, err)

	enableEpochsHandler.AddActiveFlags(common.AutoBalanceDataTriesFlag)

	// a JournalEntry is needed so the revert can happen at snapshot 1. Creating a new account creates a new journal entry.
	newAcc, _ := adb.LoadAccount(generateRandomByteArray(32))
	_ = adb.SaveAccount(newAcc)
	assert.Equal(t, 1, adb.JournalLen())

	// change the account data trie. This will trigger the migration
	acc, err = adb.LoadAccount(address)
	require.Nil(t, err)
	userAcc = acc.(state.UserAccountHandler)
	value1 := []byte("value1")
	err = userAcc.SaveKeyValue(key, value1)
	require.Nil(t, err)
	err = adb.SaveAccount(userAcc)
	require.Nil(t, err)

	// revert the migration
	err = adb.RevertToSnapshot(1)
	require.Nil(t, err)

	// check that the data trie was completely reverted. The rootHash of the user account should be present
	// in both old and new hashes. This means that after the revert, the rootHash is the same as before
	markForEvictionCalled := false
	spm.MarkForEvictionCalled = func(_ []byte, _ []byte, oldHashes common.ModifiedHashes, newHashes common.ModifiedHashes) error {
		_, ok := oldHashes[string(userAccRootHash)]
		require.True(t, ok)
		_, ok = newHashes[string(userAccRootHash)]
		require.True(t, ok)
		markForEvictionCalled = true

		return nil
	}
	_, _ = adb.Commit()
	require.True(t, markForEvictionCalled)
}

func testAccountMethodsConcurrency(
	t *testing.T,
	adb state.AccountsAdapter,
	addresses [][]byte,
	rootHash []byte,
) {
	numOperations := 100
	marshaller := &marshallerMock.MarshalizerMock{}
	wg := sync.WaitGroup{}
	wg.Add(numOperations)

	numAccToLoad := 10
	accountsSlice := make([]vmcommon.AccountHandler, numAccToLoad)
	for i := 0; i < numAccToLoad; i++ {
		acc, err := adb.GetExistingAccount(addresses[i])
		assert.Nil(t, err)
		accountsSlice[i] = acc
	}
	accountBytes, err := marshaller.Marshal(accountsSlice[0])
	assert.Nil(t, err)
	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			switch idx % 22 {
			case 0:
				_, _ = adb.GetExistingAccount(addresses[idx])
			case 1:
				_, _ = adb.GetAccountFromBytes(accountsSlice[0].AddressBytes(), accountBytes)
			case 2:
				_, _ = adb.LoadAccount(addresses[idx])
			case 3:
				_ = adb.SaveAccount(accountsSlice[idx%numAccToLoad])
			case 4:
				_ = adb.RemoveAccount(rootHash)
			case 5:
				_, _ = adb.CommitInEpoch(0, 1)
			case 6:
				_, _ = adb.Commit()
			case 7:
				_ = adb.JournalLen()
			case 8:
				_ = adb.RevertToSnapshot(idx)
			case 9:
				_ = adb.GetCode(rootHash)
			case 10:
				_, _ = adb.RootHash()
			case 11:
				_ = adb.RecreateTrie(rootHash)
			case 12:
				_ = adb.RecreateTrieFromEpoch(holders.NewRootHashHolder(rootHash, core.OptionalUint32{}))
			case 13:
				adb.PruneTrie(rootHash, state.OldRoot, state.NewPruningHandler(state.DisableDataRemoval))
			case 14:
				adb.CancelPrune(rootHash, state.NewRoot)
			case 15:
				adb.SnapshotState(rootHash, 0)
			case 16:
				_ = adb.IsPruningEnabled()
			case 17:
				_ = adb.GetAllLeaves(&common.TrieIteratorChannels{}, context.Background(), rootHash, parsers.NewMainTrieLeafParser())
			case 18:
				_, _ = adb.RecreateAllTries(rootHash)
			case 19:
				_, _ = adb.GetTrie(rootHash)
			case 20:
				_ = adb.GetStackDebugFirstEntry()
			case 21:
				_ = adb.SetSyncer(&mock.AccountsDBSyncerStub{})
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestAccountsDB_MigrateDataTrieWithFunc(t *testing.T) {
	t.Parallel()

	checkpointHashesHolder := hashesHolder.NewCheckpointHashesHolder(10000000, testscommon.HashSize)
	enabeEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsAutoBalanceDataTriesEnabledField: false,
	}
	adb, _, _ := getDefaultStateComponents(checkpointHashesHolder, testscommon.NewSnapshotPruningStorerMock(), enabeEpochsHandler)

	addr := []byte("addr")
	acc, _ := adb.LoadAccount(addr)
	value := []byte("value")
	_ = acc.(state.UserAccountHandler).SaveKeyValue([]byte("key"), value)
	_ = acc.(state.UserAccountHandler).SaveKeyValue([]byte("key2"), value)
	_ = adb.SaveAccount(acc)

	enabeEpochsHandler.IsAutoBalanceDataTriesEnabledField = true
	acc, _ = adb.LoadAccount(addr)

	isMigrated, err := acc.(state.AccountHandlerWithDataTrieMigrationStatus).IsDataTrieMigrated()
	assert.Nil(t, err)
	assert.False(t, isMigrated)

	accWithMigrate := acc.(vmcommon.UserAccountHandler).AccountDataHandler()
	dataTrieMig := dataTrieMigrator.NewDataTrieMigrator(dataTrieMigrator.ArgsNewDataTrieMigrator{
		GasProvided: 100000000,
		DataTrieGasCost: dataTrieMigrator.DataTrieGasCost{
			TrieLoadPerNode:  1,
			TrieStorePerNode: 1,
		},
	})
	err = accWithMigrate.MigrateDataTrieLeaves(vmcommon.ArgsMigrateDataTrieLeaves{
		OldVersion:   core.NotSpecified,
		NewVersion:   core.AutoBalanceEnabled,
		TrieMigrator: dataTrieMig,
	})
	assert.Nil(t, err)
	_ = adb.SaveAccount(acc)

	acc, _ = adb.LoadAccount(addr)
	retrievedVal, _, err := acc.(state.UserAccountHandler).RetrieveValue([]byte("key"))
	assert.Equal(t, value, retrievedVal)
	assert.Nil(t, err)

	isMigrated, err = acc.(state.AccountHandlerWithDataTrieMigrationStatus).IsDataTrieMigrated()
	assert.Nil(t, err)
	assert.True(t, isMigrated)
}

func BenchmarkAccountsDB_GetMethodsInParallel(b *testing.B) {
	_, adb := getDefaultTrieAndAccountsDb()

	numAccounts := 100000
	accountsAddresses := generateAccounts(b, numAccounts, adb)
	rootHash, _ := adb.Commit()

	numOperations := 100000
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testAccountLoadInParallel(adb, numOperations, accountsAddresses, rootHash)
	}
}

func testAccountLoadInParallel(
	adb state.AccountsAdapter,
	numOperations int,
	addresses [][]byte,
	rootHash []byte,
) {
	wg := sync.WaitGroup{}
	wg.Add(numOperations)

	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			defer func() {
				wg.Done()
			}()
			switch idx % 3 {
			case 0:
				_, _ = adb.LoadAccount(addresses[idx])
			case 1:
				_, _ = adb.GetExistingAccount(addresses[idx])
			case 2:
				_ = adb.RecreateTrie(rootHash)
			}
		}(i)
	}

	wg.Wait()
}
