package integrationtests

import (
	"github.com/ElrondNetwork/elrond-go-core/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/database"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageunit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
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
	persisterFactory := factory.NewPersisterFactory(dbConfig)
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
	return CreateAccountsDB(CreateMemUnit())
}

// CreateAccountsDB -
func CreateAccountsDB(db storage.Storer) *state.AccountsDB {
	ewlArgs := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: 100,
		HashesSize:     10000,
	}
	ewl, _ := evictionWaitingList.NewMemoryEvictionWaitingList(ewlArgs)
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:      1000,
		SnapshotsBufferLen:    10,
		SnapshotsGoroutineNum: 1,
	}
	args := trie.NewTrieStorageManagerArgs{
		MainStorer:             db,
		CheckpointsStorer:      CreateMemUnit(),
		Marshalizer:            TestMarshalizer,
		Hasher:                 TestHasher,
		GeneralConfig:          generalCfg,
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10000000, uint64(TestHasher.Size())),
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
	}
	trieStorage, _ := trie.NewTrieStorageManager(args)

	tr, _ := trie.NewTrie(trieStorage, TestMarshalizer, TestHasher, MaxTrieLevelInMemory)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)

	argsAccountsDB := state.ArgsAccountsDB{
		Trie:                  tr,
		Hasher:                TestHasher,
		Marshaller:            TestMarshalizer,
		AccountFactory:        &TestAccountFactory{},
		StoragePruningManager: spm,
		ProcessingMode:        common.Normal,
		ProcessStatusHandler:  &testscommon.ProcessStatusHandlerStub{},
		AppStatusHandler:      &statusHandler.AppStatusHandlerStub{},
		AddressConverter:      &testscommon.PubkeyConverterMock{},
	}
	adb, _ := state.NewAccountsDB(argsAccountsDB)

	return adb
}
