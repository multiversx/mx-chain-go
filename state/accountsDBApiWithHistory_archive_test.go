package state_test

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	commonDisabled "github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/common/holders"
	statsDisabled "github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/disabled"
	stateFactory "github.com/multiversx/mx-chain-go/state/factory"
	pruningDisabled "github.com/multiversx/mx-chain-go/state/storagePruningManager/disabled"
	"github.com/multiversx/mx-chain-go/state/triesHolder"
	"github.com/multiversx/mx-chain-go/storage"
	storageMock "github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	storageTest "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/collapseManager"
	"github.com/stretchr/testify/require"
)

type archivedAccountsFixture struct {
	history       state.AccountsAdapterAPI
	roots         map[uint32][]byte
	addresses     [][]byte
	dataRootReads *atomic.Uint64
}

func newArchivedAccountsFixture(t *testing.T, version core.TrieNodeVersion) archivedAccountsFixture {
	t.Helper()
	msh, hsh := &marshal.GogoProtoMarshalizer{}, blake2b.NewBlake2b()
	flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.FixDataTrieLeafValidationFlag)
	if version == core.AutoBalanceEnabled {
		flags.AddActiveFlags(common.AutoBalanceDataTriesFlag)
	}
	newAccountsDB := func(storer common.BaseStorer) *state.AccountsDB {
		tsm, err := trie.NewTrieStorageManager(trie.NewTrieStorageManagerArgs{
			MainStorer: storer, Marshalizer: msh, Hasher: hsh, Identifier: "historical regression",
			IdleProvider: commonDisabled.NewProcessStatusHandler(), StatsCollector: statsDisabled.NewStateStatistics(),
			GeneralConfig: config.TrieStorageManagerConfig{SnapshotsGoroutinesPerCore: 1},
		})
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, tsm.Close()) })
		mainTrie, err := trie.NewTrie(tsm, msh, hsh, flags, collapseManager.NewDisabledCollapseManager())
		require.NoError(t, err)
		holder, err := triesHolder.NewDataTriesHolder(1 << 20)
		require.NoError(t, err)
		collector := disabled.NewDisabledStateAccessesCollector()
		factory, err := stateFactory.NewAccountCreator(stateFactory.ArgsAccountCreator{
			Hasher: hsh, Marshaller: msh, EnableEpochsHandler: flags, StateAccessesCollector: collector,
			DataTriesHolder: holder, DataTrieCreator: mainTrie,
		})
		require.NoError(t, err)
		adb, err := state.NewAccountsDB(state.ArgsAccountsDB{
			Trie: mainTrie, Hasher: hsh, Marshaller: msh, AccountFactory: factory, DataTriesHolder: holder,
			StateAccessesCollector: collector, StoragePruningManager: pruningDisabled.NewDisabledStoragePruningManager(),
			AddressConverter: &testscommon.PubkeyConverterStub{}, SnapshotsManager: disabled.NewDisabledSnapshotsManager(),
		})
		require.NoError(t, err)
		return adb
	}
	persisters := make(map[string]*testscommon.MemDbMock)
	dataRoots := make(map[string]struct{})
	reads := &atomic.Uint64{}
	getPersister := func(path string) *testscommon.MemDbMock {
		db := persisters[path]
		if db == nil {
			db = testscommon.NewMemDbMock()
			db.GetCalledNoReturn = func(key []byte) {
				if _, ok := dataRoots[string(key)]; ok {
					reads.Add(1)
				}
			}
			persisters[path] = db
		}
		return db
	}
	epoch := uint32(7)
	writerStore := &storageTest.SnapshotPruningStorerStub{
		MemDbMock:                   getPersister("epoch-7"),
		GetLatestStorageEpochCalled: func() (uint32, error) { return epoch, nil },
	}
	writer := newAccountsDB(writerStore)
	f := archivedAccountsFixture{roots: make(map[uint32][]byte), dataRootReads: reads}
	for i := byte(1); i <= 2; i++ {
		address := make([]byte, 32)
		address[8], address[31] = 5, i
		f.addresses = append(f.addresses, address)
	}
	// Advance the same accounts through two committed states. Writes move to the
	// next epoch's persister; every account and storage value changes in state B.
	for _, epoch = range []uint32{7, 12} {
		writerStore.MemDbMock = getPersister(fmt.Sprintf("epoch-%d", epoch))
		for _, address := range f.addresses {
			account, err := writer.LoadAccount(address)
			require.NoError(t, err)
			user := account.(state.UserAccountHandler)
			user.IncreaseNonce(1)
			for key := 0; key < 16; key++ {
				require.NoError(t, user.SaveKeyValue([]byte(fmt.Sprintf("key-%d", key)), archivedValue(epoch, address, key)))
			}
			require.NoError(t, writer.SaveAccount(user))
			dataRoots[string(user.GetRootHash())] = struct{}{}
		}
		root, err := writer.Commit()
		require.NoError(t, err)
		f.roots[epoch] = root
	}
	require.False(t, bytes.Equal(f.roots[7], f.roots[12]))
	// Reopen as a full-history reader at epoch 20. Only 20 and 19 are active;
	// the existing epoch-7 and epoch-12 persisters remain available as archives.
	epochs := pruning.EpochArgs{StartingEpoch: 20, NumOfEpochsToKeep: 3, NumOfActivePersisters: 2}
	fullStore, err := pruning.NewFullHistoryTriePruningStorer(pruning.FullHistoryStorerArgs{
		NumOfOldActivePersisters: 2,
		StorerArgs: pruning.StorerArgs{
			Identifier: "historical regression", PruningEnabled: true, EpochsData: epochs,
			CacheConf:        storageunit.CacheConfig{Capacity: 100, Type: "LRU", Shards: 1},
			ShardCoordinator: storageMock.NewShardCoordinatorMock(0, 2),
			PathManager:      &testscommon.PathManagerStub{PathForEpochCalled: func(_ string, e uint32, _ string) string { return fmt.Sprintf("epoch-%d", e) }},
			PersisterFactory: &storageMock.PersisterFactoryStub{
				CreateCalled:          func(path string) (storage.Persister, error) { return getPersister(path), nil },
				PersisterExistsCalled: func(path string) (bool, error) { _, ok := persisters[path]; return ok, nil },
			},
			Notifier:               &storageMock.EpochStartNotifierStub{},
			OldDataCleanerProvider: &testscommon.OldDataCleanerProviderStub{},
			CustomDatabaseRemover:  &testscommon.CustomDatabaseRemoverStub{},
			MaxBatchSize:           10, PersistersTracker: pruning.NewPersistersTracker(epochs),
			StateStatsHandler: statsDisabled.NewStateStatistics(),
		},
	})
	require.NoError(t, err)
	for _, root := range f.roots {
		_, err = fullStore.Get(root)
		require.Error(t, err, "historical roots must be absent from active persisters")
	}
	fullStore.ClearCache()
	history, err := state.NewAccountsDBApiWithHistory(newAccountsDB(fullStore))
	require.NoError(t, err)
	f.history = history
	reads.Store(0)
	return f
}

func archivedValue(epoch uint32, address []byte, key int) []byte {
	return []byte(fmt.Sprintf("epoch-%d/account-%d/value-%d", epoch, address[31], key))
}

func TestAccountsDBApiWithHistory_ReadArchivedDataTrie(t *testing.T) {
	for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
		for _, query := range []struct{ epoch, hint uint32 }{{7, 7}, {7, 8}, {12, 12}} {
			t.Run(fmt.Sprintf("version-%d/hint-%d", version, query.hint), func(t *testing.T) {
				f := newArchivedAccountsFixture(t, version)
				options := holders.NewRootHashHolder(f.roots[query.epoch], core.OptionalUint32{Value: query.hint, HasValue: true})
				for _, address := range f.addresses {
					readsBefore := f.dataRootReads.Load()
					account, info, err := f.history.GetAccountWithBlockInfo(address, options)
					require.NoError(t, err)
					require.Equal(t, f.roots[query.epoch], info.GetRootHash())
					user := account.(state.UserAccountHandler)
					expectedNonce := uint64(1)
					if query.epoch == 12 {
						expectedNonce = 2
					}
					require.Equal(t, expectedNonce, user.GetNonce())
					require.Equal(t, readsBefore, f.dataRootReads.Load(), "account loading must remain lazy")
					for key := 0; key < 16; key++ {
						value, _, readErr := user.RetrieveValue([]byte(fmt.Sprintf("key-%d", key)))
						require.NoError(t, readErr)
						require.Equal(t, archivedValue(query.epoch, address, key), value)
					}
					require.Greater(t, f.dataRootReads.Load(), readsBefore)
				}
			})
		}
	}
}

func TestAccountsDBApiWithHistory_AccountKeepsSelectedStorageContext(t *testing.T) {
	f := newArchivedAccountsFixture(t, core.AutoBalanceEnabled)
	account, _, err := f.history.GetAccountWithBlockInfo(f.addresses[0], holders.NewRootHashHolder(f.roots[7], core.OptionalUint32{Value: 7, HasValue: true}))
	require.NoError(t, err)
	// Select a different archived epoch without reading either data trie. This
	// isolates context capture from the separate address-only cache-hit issue.
	_, info, err := f.history.GetAccountWithBlockInfo(f.addresses[1], holders.NewRootHashHolder(f.roots[12], core.OptionalUint32{Value: 12, HasValue: true}))
	require.NoError(t, err)
	require.Equal(t, f.roots[12], info.GetRootHash())
	require.Zero(t, f.dataRootReads.Load(), "loading accounts must leave data tries lazy")
	value, _, err := account.(state.UserAccountHandler).RetrieveValue([]byte("key-0"))
	require.NoError(t, err)
	require.Equal(t, archivedValue(7, f.addresses[0], 0), value)
}

func TestAccountsDBApiWithHistory_RepeatedArchivedSelections(t *testing.T) {
	f := newArchivedAccountsFixture(t, core.AutoBalanceEnabled)
	for cycle := 0; cycle < 3; cycle++ {
		for _, epoch := range []uint32{7, 12} {
			account, info, err := f.history.GetAccountWithBlockInfo(f.addresses[0], holders.NewRootHashHolder(f.roots[epoch], core.OptionalUint32{Value: epoch, HasValue: true}))
			require.NoError(t, err)
			require.Equal(t, f.roots[epoch], info.GetRootHash())
			value, _, err := account.(state.UserAccountHandler).RetrieveValue([]byte("key-0"))
			require.NoError(t, err)
			require.Equal(t, archivedValue(epoch, f.addresses[0], 0), value, "cycle %d", cycle)
		}
	}
}

func TestAccountsDBApiWithHistory_ConcurrentArchivedReads(t *testing.T) {
	f := newArchivedAccountsFixture(t, core.AutoBalanceEnabled)
	options := holders.NewRootHashHolder(f.roots[7], core.OptionalUint32{Value: 7, HasValue: true})
	start := make(chan struct{})
	failures := make(chan error, 4)
	var workers sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			address := f.addresses[worker%len(f.addresses)]
			for iteration := 0; iteration < 8; iteration++ {
				account, info, err := f.history.GetAccountWithBlockInfo(address, options)
				if err != nil {
					failures <- err
					return
				}
				if !bytes.Equal(info.GetRootHash(), f.roots[7]) {
					failures <- fmt.Errorf("unexpected historical root %x", info.GetRootHash())
					return
				}
				for key := 0; key < 16; key++ {
					value, _, readErr := account.(state.UserAccountHandler).RetrieveValue([]byte(fmt.Sprintf("key-%d", key)))
					if readErr != nil {
						failures <- readErr
						return
					}
					if !bytes.Equal(value, archivedValue(7, address, key)) {
						failures <- fmt.Errorf("unexpected storage value %q", value)
						return
					}
				}
			}
		}()
	}
	close(start)
	workers.Wait()
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
}
