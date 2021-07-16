package state_test

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/mock"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/factory"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/disabled"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const trieDbOperationDelay = time.Second

func generateAccountDBFromTrie(trie temporary.Trie) *state.AccountsDB {
	accnt, _ := state.NewAccountsDB(
		trie,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{
			CreateAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return mock.NewAccountWrapMock(address), nil
			},
		},
		disabled.NewDisabledStoragePruningManager(),
	)
	return accnt
}

func generateAccount() *mock.AccountWrapMock {
	return mock.NewAccountWrapMock(make([]byte, 32))
}

func generateAddressAccountAccountsDB(trie temporary.Trie) ([]byte, *mock.AccountWrapMock, *state.AccountsDB) {
	adr := make([]byte, 32)
	account := mock.NewAccountWrapMock(adr)

	adb := generateAccountDBFromTrie(trie)

	return adr, account, adb
}

func getDefaultTrieAndAccountsDb() (temporary.Trie, *state.AccountsDB) {
	checkpointHashesHolder := hashesHolder.NewCheckpointHashesHolder(10000000, testscommon.HashSize)
	adb, tr, _ := getDefaultStateComponents(checkpointHashesHolder)
	return tr, adb
}

func getDefaultStateComponents(
	hashesHolder temporary.CheckpointHashesHolder,
) (*state.AccountsDB, temporary.Trie, temporary.StorageManager) {
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}
	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	args := trie.NewTrieStorageManagerArgs{
		DB:          mock.NewMemDbMock(),
		Marshalizer: marshalizer,
		Hasher:      hsh,
		SnapshotDbConfig: config.DBConfig{
			Type: "MemoryDB",
		},
		GeneralConfig:          generalCfg,
		CheckpointHashesHolder: hashesHolder,
	}
	trieStorage, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(trieStorage, marshalizer, hsh, 5)
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(100, mock.NewMemDbMock(), marshalizer)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, generalCfg.PruningBufferLen)
	adb, _ := state.NewAccountsDB(tr, hsh, marshalizer, factory.NewAccountCreator(), spm)

	return adb, tr, trieStorage
}

//------- NewAccountsDB

func TestNewAccountsDB_WithNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewAccountsDB(
		nil,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilTrie, err)
}

func TestNewAccountsDB_WithNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewAccountsDB(
		&testscommon.TrieStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilHasher, err)
}

func TestNewAccountsDB_WithNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewAccountsDB(
		&testscommon.TrieStub{},
		&mock.HasherMock{},
		nil,
		&mock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilMarshalizer, err)
}

func TestNewAccountsDB_WithNilAddressFactoryShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewAccountsDB(
		&testscommon.TrieStub{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilAccountFactory, err)
}

func TestNewAccountsDB_WithNilStoragePruningManagerShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewAccountsDB(
		&testscommon.TrieStub{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
		nil,
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilStoragePruningManager, err)
}

func TestNewAccountsDB_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	adb, err := state.NewAccountsDB(
		&testscommon.TrieStub{
			GetStorageManagerCalled: func() temporary.StorageManager {
				return &testscommon.StorageManagerStub{
					DatabaseCalled: func() temporary.DBWriteCacher {
						return mock.NewMemDbMock()
					},
				}
			},
		},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
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
		&testscommon.TrieStub{
			GetStorageManagerCalled: func() temporary.StorageManager {
				return &testscommon.StorageManagerStub{
					DatabaseCalled: func() temporary.DBWriteCacher {
						return db
					},
				}
			},
		},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
		disabled.NewDisabledStoragePruningManager(),
	)

	assert.Equal(t, numCheckpoints, adb.GetNumCheckpoints())
}

func TestAccountsDB_SetStateCheckpointSavesNumCheckpoints(t *testing.T) {
	t.Parallel()

	numCheckpointsKey := []byte("state checkpoint")
	numCheckpoints := 50
	wg := sync.WaitGroup{}
	wg.Add(numCheckpoints)
	db := mock.NewMemDbMock()
	numExitPruningBufferingModeCalled := uint32(0)
	db.PutCalled = func(key, val []byte) error {
		wg.Done()

		return nil
	}
	adb, _ := state.NewAccountsDB(
		&testscommon.TrieStub{
			GetStorageManagerCalled: func() temporary.StorageManager {
				return &testscommon.StorageManagerStub{
					DatabaseCalled: func() temporary.DBWriteCacher {
						return db
					},
					SetCheckpointCalled: func(_ []byte, leavesChan chan core.KeyValueHolder) {
						close(leavesChan)
					},
					ExitPruningBufferingModeCalled: func() {
						atomic.AddUint32(&numExitPruningBufferingModeCalled, 1)
					},
				}
			},
			RecreateCalled: func(root []byte) (temporary.Trie, error) {
				return &testscommon.TrieStub{
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
		disabled.NewDisabledStoragePruningManager(),
	)

	for i := 0; i < numCheckpoints; i++ {
		adb.SetStateCheckpoint([]byte("rootHash"))
	}

	wg.Wait()

	val, err := db.Get(numCheckpointsKey)
	assert.Nil(t, err)

	numCheckpointsRecovered := binary.BigEndian.Uint32(val)
	assert.Equal(t, uint32(numCheckpoints), numCheckpointsRecovered)
	assert.Equal(t, uint32(numCheckpoints), adb.GetNumCheckpoints())
	assert.Equal(t, uint32(numCheckpoints), atomic.LoadUint32(&numExitPruningBufferingModeCalled))
}

//------- SaveAccount

func TestAccountsDB_SaveAccountNilAccountShouldErr(t *testing.T) {
	t.Parallel()

	adb := generateAccountDBFromTrie(&testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	})

	err := adb.SaveAccount(nil)
	assert.True(t, errors.Is(err, state.ErrNilAccountHandler))
}

func TestAccountsDB_SaveAccountErrWhenGettingOldAccountShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("trie get err")
	adb := generateAccountDBFromTrie(&testscommon.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, expectedErr
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	})

	err := adb.SaveAccount(generateAccount())
	assert.Equal(t, expectedErr, err)
}

func TestAccountsDB_SaveAccountNilOldAccount(t *testing.T) {
	t.Parallel()

	adb := generateAccountDBFromTrie(&testscommon.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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

	adb := generateAccountDBFromTrie(&testscommon.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return (&mock.MarshalizerMock{}).Marshal(acc)
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	})

	err := adb.SaveAccount(acc)
	assert.Nil(t, err)
	assert.Equal(t, 1, adb.JournalLen())
}

func TestAccountsDB_SaveAccountSavesCodeAndDataTrieForUserAccount(t *testing.T) {
	t.Parallel()

	updateCalled := 0
	trieStub := &testscommon.TrieStub{
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

	adb := generateAccountDBFromTrie(&testscommon.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			updateCalled++
			return nil
		},
		RecreateCalled: func(root []byte) (d temporary.Trie, err error) {
			return trieStub, nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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

func TestAccountsDB_SaveAccountMalfunctionMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	account := generateAccount()
	mockTrie := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	adb, _ := state.NewAccountsDB(
		mockTrie,
		&mock.HasherMock{},
		marshalizer,
		&mock.AccountsFactoryStub{
			CreateAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return mock.NewAccountWrapMock(address), nil
			},
		},
		disabled.NewDisabledStoragePruningManager(),
	)

	marshalizer.Fail = true

	//should return error
	err := adb.SaveAccount(account)

	assert.NotNil(t, err)
}

func TestAccountsDB_SaveAccountWithSomeValuesShouldWork(t *testing.T) {
	t.Parallel()

	ts := &testscommon.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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
	trieStub := &testscommon.TrieStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return marshalizer.Marshal(mock.AccountWrapMock{})
		},
		UpdateCalled: func(key, value []byte) error {
			wasCalled = true
			return nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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

	trieStub := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}
	adr := make([]byte, 32)
	adb := generateAccountDBFromTrie(trieStub)

	_, err := adb.LoadAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadAccountNotFoundShouldCreateEmpty(t *testing.T) {
	t.Parallel()

	trieStub := &testscommon.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, nil
		},
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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
	dataTrie := &testscommon.TrieStub{}
	marshalizer := &mock.MarshalizerMock{}

	trieStub := &testscommon.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, acc.AddressBytes()) {
				return marshalizer.Marshal(acc)
			}
			if bytes.Equal(key, codeHash) {
				return marshalizer.Marshal(state.CodeEntry{Code: code})
			}
			return nil, nil
		},
		RecreateCalled: func(root []byte) (d temporary.Trie, err error) {
			return dataTrie, nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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

	trieStub := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}
	adr := make([]byte, 32)
	adb := generateAccountDBFromTrie(trieStub)

	_, err := adb.GetExistingAccount(adr)
	assert.NotNil(t, err)
}

func TestAccountsDB_GetExistingAccountNotFoundShouldRetNil(t *testing.T) {
	t.Parallel()

	trieStub := &testscommon.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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
	dataTrie := &testscommon.TrieStub{}
	marshalizer := &mock.MarshalizerMock{}

	trieStub := &testscommon.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, acc.AddressBytes()) {
				return marshalizer.Marshal(acc)
			}
			if bytes.Equal(key, codeHash) {
				return marshalizer.Marshal(state.CodeEntry{Code: code})
			}
			return nil, nil
		},
		RecreateCalled: func(root []byte) (d temporary.Trie, err error) {
			return dataTrie, nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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

	trieMock := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}
	adr, _, _ := generateAddressAccountAccountsDB(trieMock)

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

	adb, _ := state.NewAccountsDB(
		trieMock,
		&mock.HasherMock{},
		&marshalizer,
		&mock.AccountsFactoryStub{
			CreateAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return mock.NewAccountWrapMock(address), nil
			},
		},
		disabled.NewDisabledStoragePruningManager(),
	)

	//Step 3. call get, should return a copy of DbAccount, recover an Account object
	recoveredAccount, err := adb.GetAccount(adr)
	assert.Nil(t, err)

	//Step 4. Let's test
	assert.Equal(t, testAccount.MockValue, recoveredAccount.(*mock.AccountWrapMock).MockValue)
}

//------- loadCode

func TestAccountsDB_LoadCodeWrongHashLengthShouldErr(t *testing.T) {
	t.Parallel()

	tr := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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
	mockTrie := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(mockTrie)

	//just search a hash. Any hash will do
	account.SetCodeHash(adr)

	err := adb.LoadCode(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadCodeOkValsShouldWork(t *testing.T) {
	t.Parallel()

	tr := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}
	adr, account, _ := generateAddressAccountAccountsDB(tr)
	marshalizer := mock.MarshalizerMock{}

	trieStub := testscommon.TrieStub{
		GetCalled: func(key []byte) (bytes []byte, e error) {
			//will return adr.Bytes() so its hash will correspond to adr.Hash()
			return marshalizer.Marshal(&state.CodeEntry{Code: adr})
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}

	adb, _ := state.NewAccountsDB(
		&trieStub,
		&mock.HasherMock{},
		&marshalizer,
		&mock.AccountsFactoryStub{
			CreateAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return mock.NewAccountWrapMock(address), nil
			},
		},
		disabled.NewDisabledStoragePruningManager(),
	)

	//just search a hash. Any hash will do
	account.SetCodeHash(adr)

	err := adb.LoadCode(account)
	assert.Nil(t, err)
	assert.Equal(t, adr, state.GetCode(account))
}

//------- RetrieveData

func TestAccountsDB_LoadDataNilRootShouldRetNil(t *testing.T) {
	t.Parallel()

	tr := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}
	_, account, adb := generateAddressAccountAccountsDB(tr)

	//since root is nil, result should be nil and data trie should be nil
	err := adb.LoadDataTrie(account)
	assert.Nil(t, err)
	assert.Nil(t, account.DataTrie())
}

func TestAccountsDB_LoadDataBadLengthShouldErr(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB(&testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	})

	account.SetRootHash([]byte("12345"))

	//should return error
	err := adb.LoadDataTrie(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadDataMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	account := generateAccount()
	account.SetRootHash([]byte("12345"))

	mockTrie := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}
	adb := generateAccountDBFromTrie(mockTrie)

	//should return error
	err := adb.LoadDataTrie(account)
	assert.NotNil(t, err)
}

func TestAccountsDB_LoadDataNotFoundRootShouldReturnErr(t *testing.T) {
	t.Parallel()

	_, account, adb := generateAddressAccountAccountsDB(&testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	})

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

	dataTrie := &testscommon.TrieStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, keyRequired) {
				return trieVal, nil
			}

			return nil, nil
		},
	}

	account := generateAccount()
	mockTrie := &testscommon.TrieStub{
		RecreateCalled: func(root []byte) (trie temporary.Trie, e error) {
			if !bytes.Equal(root, rootHash) {
				return nil, errors.New("bad root hash")
			}

			return dataTrie, nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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
	trieStub := testscommon.TrieStub{
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
		RecreateCalled: func(root []byte) (trie temporary.Trie, err error) {
			return &testscommon.TrieStub{
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
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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
	trieStub := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}
	trieStub.RecreateCalled = func(root []byte) (tree temporary.Trie, e error) {
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

	trieStub := testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}
	trieStub.RecreateCalled = func(root []byte) (tree temporary.Trie, e error) {
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

	trieStub := testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
		RecreateCalled: func(root []byte) (temporary.Trie, error) {
			wasCalled = true
			return &testscommon.TrieStub{}, nil
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
	trieStub := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				TakeSnapshotCalled: func(rootHash []byte, _ bool, _ chan core.KeyValueHolder) {
					snapshotMut.Lock()
					takeSnapshotWasCalled = true
					snapshotMut.Unlock()
				},
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
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

func TestAccountsDB_SnapshotStateWithDataTries(t *testing.T) {
	t.Parallel()

	tr, adb := getDefaultTrieAndAccountsDb()

	accountsAddresses := generateAccounts(t, 3, adb)
	newHashes := modifyDataTries(t, accountsAddresses, adb)
	rootHash, _ := adb.Commit()

	adb.SnapshotState(rootHash)
	time.Sleep(time.Second)

	trieDb := tr.GetStorageManager().Database()
	err := trieDb.Remove(rootHash)
	assert.Nil(t, err)
	for hash := range newHashes {
		err = trieDb.Remove([]byte(hash))
		assert.Nil(t, err)
	}

	snapshotDb := tr.GetStorageManager().GetSnapshotThatContainsHash(rootHash)
	assert.NotNil(t, snapshotDb)

	val, err := trieDb.Get(rootHash)
	assert.Nil(t, val)
	assert.NotNil(t, err)

	val, err = snapshotDb.Get(rootHash)
	assert.NotNil(t, val)
	assert.Nil(t, err)

	for hash := range newHashes {
		val, err = trieDb.Get([]byte(hash))
		assert.Nil(t, val)
		assert.NotNil(t, err)

		val, err = snapshotDb.Get([]byte(hash))
		assert.NotNil(t, val)
		assert.Nil(t, err)
	}
}

func TestAccountsDB_SetStateCheckpoint(t *testing.T) {
	t.Parallel()

	setCheckPointWasCalled := false
	snapshotMut := sync.Mutex{}
	trieStub := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				SetCheckpointCalled: func(rootHash []byte, _ chan core.KeyValueHolder) {
					snapshotMut.Lock()
					setCheckPointWasCalled = true
					snapshotMut.Unlock()
				},
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
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

func TestAccountsDB_SetStateCheckpointWithDataTries(t *testing.T) {
	t.Parallel()

	tr, adb := getDefaultTrieAndAccountsDb()

	accountsAddresses := generateAccounts(t, 3, adb)
	newHashes := modifyDataTries(t, accountsAddresses, adb)
	rootHash, _ := adb.Commit()

	adb.SetStateCheckpoint(rootHash)
	time.Sleep(time.Second)

	trieDb := tr.GetStorageManager().Database()
	err := trieDb.Remove(rootHash)
	assert.Nil(t, err)
	for hash := range newHashes {
		err = trieDb.Remove([]byte(hash))
		assert.Nil(t, err)
	}

	snapshotDb := tr.GetStorageManager().GetSnapshotThatContainsHash(rootHash)
	assert.NotNil(t, snapshotDb)

	val, err := trieDb.Get(rootHash)
	assert.Nil(t, val)
	assert.NotNil(t, err)

	val, err = snapshotDb.Get(rootHash)
	assert.NotNil(t, val)
	assert.Nil(t, err)

	for hash := range newHashes {
		val, err = trieDb.Get([]byte(hash))
		assert.Nil(t, val)
		assert.NotNil(t, err)

		val, err = snapshotDb.Get([]byte(hash))
		assert.NotNil(t, val)
		assert.Nil(t, err)
	}
}

func TestAccountsDB_IsPruningEnabled(t *testing.T) {
	t.Parallel()

	trieStub := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				IsPruningEnabledCalled: func() bool {
					return true
				},
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
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

	trieStub := &testscommon.TrieStub{
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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
	trieStub := &testscommon.TrieStub{
		RootCalled: func() (i []byte, err error) {
			rootHashCalled = true
			return rootHash, nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
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
	trieStub := &testscommon.TrieStub{
		GetAllLeavesOnChannelCalled: func(rootHash []byte) (chan core.KeyValueHolder, error) {
			getAllLeavesCalled = true

			ch := make(chan core.KeyValueHolder)
			close(ch)

			return ch, nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}

	adb := generateAccountDBFromTrie(trieStub)
	_, err := adb.GetAllLeaves([]byte("root hash"))
	assert.Nil(t, err)
	assert.True(t, getAllLeavesCalled)
}

func checkCodeEntry(
	codeHash []byte,
	expectedCode []byte,
	expectedNumReferences uint32,
	marshalizer marshal.Marshalizer,
	tr temporary.Trie,
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
	tr, adb := getDefaultTrieAndAccountsDb()

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

	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}
	tr, adb := getDefaultTrieAndAccountsDb()

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
	tr, adb := getDefaultTrieAndAccountsDb()

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
	tr, adb := getDefaultTrieAndAccountsDb()

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
	tr, adb := getDefaultTrieAndAccountsDb()

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
	tr, adb := getDefaultTrieAndAccountsDb()

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

	hsh := mock.HasherMock{}
	tr, adb := getDefaultTrieAndAccountsDb()

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
	ewl := mock.NewEvictionWaitingList(100, mock.NewMemDbMock(), marshalizer)
	args := trie.NewTrieStorageManagerArgs{
		DB:                     mock.NewMemDbMock(),
		Marshalizer:            marshalizer,
		Hasher:                 hsh,
		SnapshotDbConfig:       config.DBConfig{},
		GeneralConfig:          config.TrieStorageManagerConfig{},
		CheckpointHashesHolder: &testscommon.CheckpointHashesHolderStub{},
	}
	storageManager, _ := trie.NewTrieStorageManager(args)
	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(storageManager, marshalizer, hsh, maxTrieLevelInMemory)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 5)
	adb, _ := state.NewAccountsDB(tr, hsh, marshalizer, factory.NewAccountCreator(), spm)

	addr := make([]byte, 32)
	acc, _ := adb.LoadAccount(addr)
	userAcc := acc.(state.UserAccountHandler)

	code := []byte("code")
	userAcc.SetCode(code)
	_ = adb.SaveAccount(userAcc)
	rootHash, _ := adb.Commit()

	rootHash1 := append(rootHash, byte(temporary.NewRoot))
	hashesForEviction := ewl.Cache[string(rootHash1)]
	assert.Equal(t, 3, len(hashesForEviction))

	acc, _ = adb.LoadAccount(addr)
	userAcc = acc.(state.UserAccountHandler)
	userAcc.SetCode(nil)
	_ = adb.SaveAccount(userAcc)
	_, _ = adb.Commit()

	rootHash2 := append(rootHash, byte(temporary.OldRoot))
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
	marshalizer := &mock.MarshalizerMock{}
	hsh := mock.HasherMock{}

	ewl := mock.NewEvictionWaitingList(100, mock.NewMemDbMock(), marshalizer)
	args := trie.NewTrieStorageManagerArgs{
		DB:                     mock.NewMemDbMock(),
		Marshalizer:            marshalizer,
		Hasher:                 hsh,
		SnapshotDbConfig:       config.DBConfig{},
		GeneralConfig:          config.TrieStorageManagerConfig{},
		CheckpointHashesHolder: &testscommon.CheckpointHashesHolderStub{},
	}
	storageManager, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(storageManager, marshalizer, hsh, maxTrieLevelInMemory)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 5)
	adb, _ := state.NewAccountsDB(tr, hsh, marshalizer, factory.NewAccountCreator(), spm)

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
	rootHash = append(rootHash, byte(temporary.OldRoot))
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

	adb.CancelPrune(rootHash, temporary.NewRoot)
	adb.PruneTrie(rootHash, temporary.OldRoot)
	time.Sleep(trieDbOperationDelay)

	for i := range oldHashes {
		encNode, errGet := tr.GetStorageManager().Database().Get(oldHashes[i])
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

	adb.CancelPrune(rootHashes[0], temporary.NewRoot)
	finalizeTrieState(t, 1, tr, adb, rootHashes)
	finalizeTrieState(t, 2, tr, adb, rootHashes)
	rollbackTrieState(t, 3, tr, adb, rootHashes)

	_, err := tr.Recreate(rootHashes[2])
	assert.Nil(t, err)
}

func finalizeTrieState(t *testing.T, index int, tr temporary.Trie, adb state.AccountsAdapter, rootHashes [][]byte) {
	adb.PruneTrie(rootHashes[index-1], temporary.OldRoot)
	adb.CancelPrune(rootHashes[index], temporary.NewRoot)
	time.Sleep(trieDbOperationDelay)

	_, err := tr.Recreate(rootHashes[index-1])
	assert.NotNil(t, err)
}

func rollbackTrieState(t *testing.T, index int, tr temporary.Trie, adb state.AccountsAdapter, rootHashes [][]byte) {
	adb.PruneTrie(rootHashes[index], temporary.NewRoot)
	adb.CancelPrune(rootHashes[index-1], temporary.OldRoot)
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

	adb.CancelPrune(rootHash, temporary.NewRoot)
	adb.PruneTrie(rootHash, temporary.OldRoot)

	expectedErr := fmt.Errorf("key: %s not found", base64.StdEncoding.EncodeToString(rootHash))
	val, err := tr.GetStorageManager().Database().Get(rootHash)
	assert.Nil(t, val)
	assert.Equal(t, expectedErr, err)
}

func TestAccountsDB_CommitAddsDirtyHashesToCheckpointHashesHolder(t *testing.T) {
	t.Parallel()

	newHashes := make(temporary.ModifiedHashes)
	var rootHash []byte
	checkpointHashesHolder := &testscommon.CheckpointHashesHolderStub{
		PutCalled: func(rH []byte, hashes temporary.ModifiedHashes) bool {
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

func mergeMaps(map1 temporary.ModifiedHashes, map2 temporary.ModifiedHashes) {
	for key, val := range map2 {
		map1[key] = val
	}
}

func TestAccountsDB_CommitSetsStateCheckpointIfCheckpointHashesHolderIsFull(t *testing.T) {
	t.Parallel()

	newHashes := make(temporary.ModifiedHashes)
	numRemoveCalls := 0
	checkpointHashesHolder := &testscommon.CheckpointHashesHolderStub{
		PutCalled: func(_ []byte, _ temporary.ModifiedHashes) bool {
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

func TestAccountsDB_SnapshotStateCommitsAllStateInOneDbAndCleansCheckpointHashesHolder(t *testing.T) {
	t.Parallel()

	removeCommitedCalled := false
	checkpointHashesHolder := &testscommon.CheckpointHashesHolderStub{
		PutCalled: func(_ []byte, _ temporary.ModifiedHashes) bool {
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
	snapshotDb := trieStorage.GetSnapshotThatContainsHash(rootHash)
	for key := range newHashes {
		val, err := snapshotDb.Get([]byte(key))
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}
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
	snapshotDb := trieStorage.GetSnapshotThatContainsHash(rootHash)
	mainTrieHashes, _ := tr.GetAllHashes()
	allStateHashes = append(allStateHashes, mainTrieHashes...)

	acc, _ := adb.LoadAccount(accountsAddresses[0])
	dataTrie1Hashes, _ := acc.(state.UserAccountHandler).DataTrie().GetAllHashes()
	allStateHashes = append(allStateHashes, dataTrie1Hashes...)

	acc, _ = adb.LoadAccount(accountsAddresses[1])
	dataTrie2Hashes, _ := acc.(state.UserAccountHandler).DataTrie().GetAllHashes()
	allStateHashes = append(allStateHashes, dataTrie2Hashes...)

	numPresent := 0
	numAbsent := 0
	for _, hash := range allStateHashes {
		_, ok := newHashes[string(hash)]
		if ok {
			val, errGet := snapshotDb.Get(hash)
			assert.Nil(t, errGet)
			assert.NotNil(t, val)
			numPresent++
			continue
		}

		val, errGet := snapshotDb.Get(hash)
		assert.Nil(t, val)
		assert.NotNil(t, errGet)
		numAbsent++
	}

	assert.Equal(t, len(newHashes), numPresent)
	assert.True(t, numAbsent > 0)
}

func TestAccountsDB_CheckpointHashesHolderReceivesOnly32BytesData(t *testing.T) {
	t.Parallel()

	putCalled := false
	checkpointHashesHolder := &testscommon.CheckpointHashesHolderStub{
		PutCalled: func(rootHash []byte, hashes temporary.ModifiedHashes) bool {
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

	newHashes := make(temporary.ModifiedHashes)
	removeCalled := 0
	checkpointHashesHolder := &testscommon.CheckpointHashesHolderStub{
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

	adb.CancelPrune(rootHash, temporary.NewRoot)
	adb.PruneTrie(rootHash, temporary.OldRoot)
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

func modifyDataTries(t *testing.T, accountsAddresses [][]byte, adb *state.AccountsDB) temporary.ModifiedHashes {
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
	marshalizer := &mock.MarshalizerMock{}
	hasher := mock.HasherMock{}

	args := trie.NewTrieStorageManagerArgs{
		DB:                     mock.NewMemDbMock(),
		Marshalizer:            marshalizer,
		Hasher:                 hasher,
		SnapshotDbConfig:       config.DBConfig{},
		GeneralConfig:          config.TrieStorageManagerConfig{},
		CheckpointHashesHolder: &testscommon.CheckpointHashesHolderStub{},
	}
	storageManager, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(storageManager, marshalizer, hasher, maxTrieLevelInMemory)
	spm := disabled.NewDisabledStoragePruningManager()
	adb, _ := state.NewAccountsDB(tr, hasher, marshalizer, factory.NewAccountCreator(), spm)

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
	tr := &testscommon.TrieStub{
		CloseCalled: func() error {
			closeCalled = true
			return nil
		},
		GetStorageManagerCalled: func() temporary.StorageManager {
			return &testscommon.StorageManagerStub{
				DatabaseCalled: func() temporary.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			}
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	hsh := &mock.HasherMock{}
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(100, mock.NewMemDbMock(), marshalizer)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)
	adb, _ := state.NewAccountsDB(tr, hsh, marshalizer, factory.NewAccountCreator(), spm)

	err := adb.Close()
	assert.Nil(t, err)
	assert.True(t, closeCalled)
}

func BenchmarkAccountsDb_GetCodeEntry(b *testing.B) {
	maxTrieLevelInMemory := uint(5)
	marshalizer := &mock.MarshalizerMock{}
	hasher := mock.HasherMock{}

	args := trie.NewTrieStorageManagerArgs{
		DB:                     mock.NewMemDbMock(),
		Marshalizer:            marshalizer,
		Hasher:                 hasher,
		SnapshotDbConfig:       config.DBConfig{},
		GeneralConfig:          config.TrieStorageManagerConfig{},
		CheckpointHashesHolder: &testscommon.CheckpointHashesHolderStub{},
	}
	storageManager, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(storageManager, marshalizer, hasher, maxTrieLevelInMemory)
	spm := disabled.NewDisabledStoragePruningManager()
	adb, _ := state.NewAccountsDB(tr, hasher, marshalizer, factory.NewAccountCreator(), spm)

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
