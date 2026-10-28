package state

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	commonDisabled "github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/common/holders"
	statisticsDisabled "github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/disabled"
	"github.com/multiversx/mx-chain-go/state/triesHolder"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	factoryMock "github.com/multiversx/mx-chain-go/testscommon/factory"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/collapseManager"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

type accountsViewsFixture struct {
	adapters *accountsAdapterCreationResult
	address  []byte
	key      []byte
	oldRoot  []byte
	newRoot  []byte
}

func newAccountsViewsFactory(t *testing.T) (*stateComponentsFactory, common.TriesHolder) {
	t.Helper()
	marshaller := &marshal.GogoProtoMarshalizer{}
	hasher := blake2b.NewBlake2b()
	flags := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	storageManager, err := trie.NewTrieStorageManager(trie.NewTrieStorageManagerArgs{
		MainStorer: testscommon.NewMemDbMock(), Marshalizer: marshaller, Hasher: hasher,
		IdleProvider:   commonDisabled.NewProcessStatusHandler(),
		StatsCollector: statisticsDisabled.NewStateStatistics(),
		GeneralConfig:  config.TrieStorageManagerConfig{SnapshotsGoroutinesPerCore: 1},
		Identifier:     "accounts views",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storageManager.Close()) })
	mainTrie, err := trie.NewTrie(storageManager, marshaller, hasher, flags, collapseManager.NewDisabledCollapseManager())
	require.NoError(t, err)
	tries := triesHolder.NewTriesHolder()
	tries.Put([]byte(dataRetriever.UserAccountsUnit.String()), mainTrie)
	cfg := testscommon.GetGeneralConfig()
	cfg.StateTriesConfig.SnapshotsEnabled = false
	scf := &stateComponentsFactory{
		config: cfg,
		core: &factoryMock.CoreComponentsHolderStub{
			InternalMarshalizerCalled:    func() marshal.Marshalizer { return marshaller },
			HasherCalled:                 func() hashing.Hasher { return hasher },
			EnableEpochsHandlerCalled:    func() common.EnableEpochsHandler { return flags },
			AddressPubKeyConverterCalled: func() core.PubkeyConverter { return &testscommon.PubkeyConverterStub{} },
		},
		statusCore:   &factoryMock.StatusCoreComponentsStub{AppStatusHandlerField: &statusHandler.AppStatusHandlerStub{}},
		chainHandler: &testscommon.ChainHandlerStub{},
	}
	return scf, tries
}

func newAccountsViewsFixture(t *testing.T) accountsViewsFixture {
	t.Helper()
	scf, tries := newAccountsViewsFactory(t)
	var currentRoot []byte
	scf.chainHandler = &testscommon.ChainHandlerStub{
		GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler { return &block.Header{Nonce: 2} },
		GetCurrentBlockHeaderHashCalled:  func() []byte { return []byte("block B") },
		GetCurrentBlockRootHashCalled:    func() []byte { return currentRoot },
		GetFinalBlockInfoCalled:          func() (uint64, []byte, []byte) { return 2, []byte("block B"), currentRoot },
	}
	adapters, err := scf.createAccountsAdapters(tries, disabled.NewDisabledStateAccessesCollector())
	require.NoError(t, err)
	f := accountsViewsFixture{adapters: adapters, address: bytes.Repeat([]byte{1}, 32), key: []byte("storage key")}
	commitValue := func(value string) []byte {
		account, loadErr := adapters.accountsAdapter.LoadAccount(f.address)
		require.NoError(t, loadErr)
		userAccount, ok := account.(state.UserAccountHandler)
		require.True(t, ok)
		require.NoError(t, userAccount.SaveKeyValue(f.key, []byte(value)))
		require.NoError(t, adapters.accountsAdapter.SaveAccount(userAccount))
		root, commitErr := adapters.accountsAdapter.Commit()
		require.NoError(t, commitErr)
		return root
	}
	f.oldRoot = commitValue("old value")
	f.newRoot = commitValue("new value")
	require.NotEqual(t, f.oldRoot, f.newRoot)
	currentRoot = f.newRoot
	require.NoError(t, adapters.accountsAdapterProposal.RecreateTrie(holders.NewDefaultRootHashesHolder(f.newRoot)))
	// Initialize both wrappers at B before historical reads. Otherwise their first
	// recreation clears the shared cache and can conceal the bug.
	for _, options := range []api.AccountQueryOptions{{}, {OnFinalBlock: true}} {
		_, info, getErr := adapters.accountsRepository.GetAccountWithBlockInfo(f.address, options)
		require.NoError(t, getErr)
		require.Equal(t, f.newRoot, info.GetRootHash())
	}
	return f
}

func (f accountsViewsFixture) readValue(view string, root []byte) ([]byte, error) {
	var account vmcommon.AccountHandler
	var err error
	if view == "proposal" {
		actualRoot, rootErr := f.adapters.accountsAdapterProposal.RootHash()
		if rootErr != nil {
			return nil, rootErr
		}
		if !bytes.Equal(root, actualRoot) {
			return nil, fmt.Errorf("proposal root: expected %x, got %x", root, actualRoot)
		}
		account, err = f.adapters.accountsAdapterProposal.GetExistingAccount(f.address)
	} else {
		options := api.AccountQueryOptions{OnFinalBlock: view == "final"}
		if view == "historical" {
			options.BlockRootHash = root
		}
		var info common.BlockInfo
		account, info, err = f.adapters.accountsRepository.GetAccountWithBlockInfo(f.address, options)
		if err == nil && !bytes.Equal(root, info.GetRootHash()) {
			return nil, fmt.Errorf("%s root: expected %x, got %x", view, root, info.GetRootHash())
		}
	}
	if err != nil {
		return nil, err
	}
	value, _, err := account.(state.UserAccountHandler).RetrieveValue(f.key)
	return value, err
}

func TestAccountsAdapters_IsolateHistoricalDataTries(t *testing.T) {
	for _, view := range []string{"current", "final", "proposal"} {
		t.Run(view, func(t *testing.T) {
			f := newAccountsViewsFixture(t)
			for cycle := 0; cycle < 3; cycle++ {
				for _, historical := range []struct {
					root  []byte
					value string
				}{
					{f.oldRoot, "old value"}, {f.newRoot, "new value"}, {f.oldRoot, "old value"},
				} {
					value, err := f.readValue("historical", historical.root)
					require.NoError(t, err)
					require.Equal(t, historical.value, string(value))
					value, err = f.readValue(view, f.newRoot)
					require.NoError(t, err)
					require.Equal(t, "new value", string(value), "cycle %d: %s must stay at B", cycle, view)
				}
			}
		})
	}
}

func TestAccountsAdapters_ConcurrentFixedStateViews(t *testing.T) {
	f := newAccountsViewsFixture(t)
	value, err := f.readValue("historical", f.oldRoot)
	require.NoError(t, err)
	require.Equal(t, "old value", string(value))
	start := make(chan struct{})
	errors := make(chan error, 4)
	var workers sync.WaitGroup
	for _, view := range []string{"current", "final", "historical", "proposal"} {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			root, expected := f.newRoot, "new value"
			if view == "historical" {
				root, expected = f.oldRoot, "old value"
			}
			for i := 0; i < 20; i++ {
				actual, readErr := f.readValue(view, root)
				if readErr != nil {
					errors <- fmt.Errorf("%s: %w", view, readErr)
					return
				}
				if string(actual) != expected {
					errors <- fmt.Errorf("%s: expected %q, got %q", view, expected, actual)
					return
				}
			}
		}()
	}
	close(start)
	workers.Wait()
	close(errors)
	for readErr := range errors {
		require.NoError(t, readErr)
	}
}

func TestCreateAccountsAdapters_InvalidCacheSizes(t *testing.T) {
	for _, test := range []struct {
		field string
		set   func(*config.StateTriesConfig, uint64)
	}{
		{"DataTriesSizeInMemoryForCurrentAPI", func(cfg *config.StateTriesConfig, size uint64) { cfg.DataTriesSizeInMemoryForCurrentAPI = size }},
		{"DataTriesSizeInMemoryForFinalAPI", func(cfg *config.StateTriesConfig, size uint64) { cfg.DataTriesSizeInMemoryForFinalAPI = size }},
		{"DataTriesSizeInMemoryForHistoricalAPI", func(cfg *config.StateTriesConfig, size uint64) { cfg.DataTriesSizeInMemoryForHistoricalAPI = size }},
		{"DataTriesSizeInMemoryForProposal", func(cfg *config.StateTriesConfig, size uint64) { cfg.DataTriesSizeInMemoryForProposal = size }},
	} {
		for _, size := range []uint64{0, (1 << 20) - 1} {
			t.Run(fmt.Sprintf("%s/%d", test.field, size), func(t *testing.T) {
				scf, tries := newAccountsViewsFactory(t)
				test.set(&scf.config.StateTriesConfig, size)
				adapters, err := scf.createAccountsAdapters(tries, disabled.NewDisabledStateAccessesCollector())
				require.Nil(t, adapters)
				require.ErrorIs(t, err, triesHolder.ErrInvalidMaxTrieSizeValue)
				require.ErrorContains(t, err, test.field)
			})
		}
	}
}

func TestCreateStateViewAccountsDBArgs_IndependentCacheBudgets(t *testing.T) {
	scf, tries := newAccountsViewsFactory(t)
	mainTrie := tries.Get([]byte(dataRetriever.UserAccountsUnit.String()))
	var holders []common.TriesHolder
	for i := 1; i <= 4; i++ {
		size := i << 20
		args, err := scf.createStateViewAccountsDBArgs(mainTrie, &stateMock.StoragePruningManagerStub{}, uint64(size), "cache size", false)
		require.NoError(t, err)
		holder := args.DataTriesHolder
		// Use the same address in every cache, with independently sized entries.
		fullTrie := &trieMock.TrieStub{SizeInMemoryCalled: func() int { return size }}
		holder.Put([]byte("account"), fullTrie)
		require.Equal(t, []common.Trie{fullTrie}, holder.GetAll())
		holder.MarkAllAsCommitted()
		require.Same(t, fullTrie, holder.Get([]byte("account")), "an entry at the exact budget must fit")
		holder.Put([]byte("other account"), &trieMock.TrieStub{SizeInMemoryCalled: func() int { return 1 }})
		require.Nil(t, holder.Get([]byte("account")), "one extra byte must evict the clean entry")
		holders = append(holders, holder)
	}
	for i, holder := range holders {
		holder.Put([]byte("retained"), &trieMock.TrieStub{})
		for j, other := range holders {
			if j != i {
				require.Nil(t, other.Get([]byte("retained")))
			}
		}
		holder.Reset()
	}
}
