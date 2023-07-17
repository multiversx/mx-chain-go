package integrationtests

import (
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/evictionWaitingList"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	testcommonStorage "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/trie"
)

// TestMarshalizer -
var TestMarshalizer = &marshal.GogoProtoMarshalizer{}

// TestHasher -
var TestHasher = sha256.NewSha256()

// MaxTrieLevelInMemory -
const MaxTrieLevelInMemory = uint(5)

// CreateMemUnit -
func CreateMemUnit() storage.Storer {
	capacity := uint32(10)
	shards := uint32(1)
	sizeInBytes := uint64(0)
	cache, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: capacity, Shards: shards, SizeInBytes: sizeInBytes})

	unit, _ := storageunit.NewStorageUnit(cache, database.NewMemDB())
	return unit
}

// CreateStorer -
func CreateStorer(parentDir string) storage.Storer {
	cacheConfig := storageunit.CacheConfig{
		Name:        "trie",
		Type:        "SizeLRU",
		SizeInBytes: 314572800, // 300MB
		Capacity:    500000,
	}
	trieCache, err := storageunit.NewCache(cacheConfig)
	if err != nil {
		return nil
	}

	dbConfig := config.DBConfig{
		FilePath:          "trie",
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 2,
		MaxBatchSize:      45000,
		MaxOpenFiles:      10,
	}
	dbConfigHandler := factory.NewDBConfigHandler(dbConfig)
	persisterFactory, err := factory.NewPersisterFactory(dbConfigHandler)
	if err != nil {
		return nil
	}

	triePersister, err := persisterFactory.Create(parentDir)
	if err != nil {
		return nil
	}

	trieStorage, err := storageunit.NewStorageUnit(trieCache, triePersister)
	if err != nil {
		return nil
	}

	return trieStorage
}

// CreateInMemoryShardAccountsDB -
func CreateInMemoryShardAccountsDB() *state.AccountsDB {
	return CreateAccountsDB(testscommon.CreateStorerWithStats(), &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
}

// CreateAccountsDB -
func CreateAccountsDB(db storage.Storer, enableEpochs common.EnableEpochsHandler) *state.AccountsDB {
	ewlArgs := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: 100,
		HashesSize:     10000,
	}
	ewl, _ := evictionWaitingList.NewMemoryEvictionWaitingList(ewlArgs)

	args := testcommonStorage.GetStorageManagerArgs()
	args.MainStorer = db
	args.Marshalizer = TestMarshalizer
	args.Hasher = TestHasher

	trieStorage, _ := trie.NewTrieStorageManager(args)

	tr, _ := trie.NewTrie(trieStorage, TestMarshalizer, TestHasher, enableEpochs, MaxTrieLevelInMemory)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)

	argsAccountsDB := state.ArgsAccountsDB{
		Trie:       tr,
		Hasher:     TestHasher,
		Marshaller: TestMarshalizer,
		AccountFactory: &TestAccountFactory{
			args: state.ArgsAccountCreation{
				Hasher:              TestHasher,
				Marshaller:          TestMarshalizer,
				EnableEpochsHandler: enableEpochs,
			},
		},
		StoragePruningManager: spm,
		ProcessingMode:        common.Normal,
		ProcessStatusHandler:  &testscommon.ProcessStatusHandlerStub{},
		AppStatusHandler:      &statusHandler.AppStatusHandlerStub{},
		AddressConverter:      &testscommon.PubkeyConverterMock{},
	}
	adb, _ := state.NewAccountsDB(argsAccountsDB)

	return adb
}
