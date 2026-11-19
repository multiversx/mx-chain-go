package state

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/disabled"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/collapseManager"
	"github.com/stretchr/testify/require"
)

func TestAccountsAdapters_APIReadsReleaseEvictedDataTries(t *testing.T) {
	for _, view := range []string{"current", "final", "historical"} {
		for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
			t.Run(fmt.Sprintf("%s/version-%d", view, version), func(t *testing.T) {
				f := newAPIDataTrieCacheFixture(t, view, version)
				for i := byte(1); i <= 8; i++ {
					f.save(t, i, 600<<10, 1)
				}
				f.commit(t)
				read := func(id byte) state.UserAccountHandler {
					user := f.account(t, id)
					requireAPIDataTrieValue(t, user, id, 0, 600<<10)
					tr := requireAccountDataTrie(t, user).(common.Trie)
					require.GreaterOrEqual(t, tr.SizeInMemory(), 600<<10)
					require.Less(t, tr.SizeInMemory(), 1<<20, "two decoded data tries must exceed the holder budget")
					dirty, hashErr := tr.GetDirtyHashes()
					require.NoError(t, hashErr)
					require.Empty(t, dirty, "API reads must use persisted tries")
					return user
				}
				// Stay on the same committed root. Neither a block advance nor a
				// commit may be needed to release clean cache entries.
				for cycle := 0; cycle < 3; cycle++ {
					first := read(1)
					firstTrie := requireAccountDataTrie(t, first)
					require.Same(t, firstTrie, requireAccountDataTrie(t, read(1)), "a resident cache hit must remain reusable")
					for i := byte(2); i <= 8; i++ {
						read(i)
					}
					reloaded := read(1)
					require.NotSame(t, firstTrie, requireAccountDataTrie(t, reloaded), "an evicted clean trie must not survive in a pending-commit buffer")
					// Eviction releases only the cache's reference. A request that
					// already attached its trie must still be able to finish.
					requireAPIDataTrieValue(t, first, 1, 0, 600<<10)
					require.Same(t, firstTrie, requireAccountDataTrie(t, first))
				}
			})
		}
	}
}

func TestAccountsAdapters_APIReadsEvictGrowingDataTries(t *testing.T) {
	for _, view := range []string{"current", "final", "historical"} {
		for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
			t.Run(fmt.Sprintf("%s/version-%d", view, version), func(t *testing.T) {
				f := newAPIDataTrieCacheFixture(t, view, version)
				for i := byte(1); i <= 3; i++ {
					f.save(t, i, 225<<10, 2)
				}
				f.commit(t)

				first, firstTrie := f.attach(t, 1)
				second, secondTrie := f.attach(t, 2)
				require.Less(t, firstTrie.SizeInMemory()+secondTrie.SizeInMemory(), 225<<10, "only roots should be loaded")
				for k := 0; k < 2; k++ {
					requireAPIDataTrieValue(t, first, 1, k, 225<<10)
					requireAPIDataTrieValue(t, second, 2, k, 225<<10)
				}
				_, recent := f.attach(t, 1)
				require.Same(t, firstTrie, recent)

				third, thirdTrie := f.attach(t, 3)
				const budget = 1 << 20
				require.Less(t, firstTrie.SizeInMemory()+secondTrie.SizeInMemory()+thirdTrie.SizeInMemory(), budget)
				requireAPIDataTrieValue(t, third, 3, 0, 225<<10)
				require.Greater(t, firstTrie.SizeInMemory()+secondTrie.SizeInMemory()+thirdTrie.SizeInMemory(), budget)
				_, stillCached := f.attach(t, 1)
				require.Same(t, firstTrie, stillCached)
				_, stillCached = f.attach(t, 3)
				require.Same(t, thirdTrie, stillCached)
				reopened, freshTrie := f.attach(t, 2)
				require.NotSame(t, secondTrie, freshTrie, "point reads must evict the least-recently-used whole trie")
				requireAPIDataTrieValue(t, reopened, 2, 1, 225<<10)
				requireAPIDataTrieValue(t, second, 2, 0, 225<<10)
			})
		}
	}
}

func TestAccountsAdapters_APIReadsDoNotCacheEmptyDataTries(t *testing.T) {
	for _, view := range []string{"current", "final", "historical"} {
		for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
			t.Run(fmt.Sprintf("%s/version-%d", view, version), func(t *testing.T) {
				f := newAPIDataTrieCacheFixture(t, view, version)
				f.save(t, 1, 16, 1)
				f.save(t, 2, 16, 1)
				f.commit(t)

				processing := f.adapters.accountsAdapter
				account, err := processing.LoadAccount(bytes.Repeat([]byte{1}, 32))
				require.NoError(t, err)
				user := account.(state.UserAccountHandler)
				require.NoError(t, user.SaveKeyValue([]byte("key-0"), nil))
				require.NoError(t, processing.SaveAccount(user))
				f.commit(t)

				_, nonempty := f.attach(t, 2)
				emptyAccount, emptyTrie := f.attach(t, 1)
				require.Equal(t, common.EmptyTrieHash, emptyAccount.GetRootHash())
				require.Zero(t, emptyTrie.SizeInMemory())
				value, _, err := emptyAccount.RetrieveValue([]byte("key-0"))
				require.NoError(t, err)
				require.Empty(t, value)
				require.Same(t, emptyTrie, requireAccountDataTrie(t, emptyAccount))

				_, reopened := f.attach(t, 1)
				require.NotSame(t, emptyTrie, reopened)
				_, cached := f.attach(t, 2)
				require.Same(t, nonempty, cached)
			})
		}
	}
}

func TestAccountsAdapters_APIReadsReleaseOversizedDataTries(t *testing.T) {
	for _, view := range []string{"current", "final", "historical"} {
		for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
			t.Run(fmt.Sprintf("%s/version-%d", view, version), func(t *testing.T) {
				const budget = 1 << 20
				t.Run("growing-singleton", func(t *testing.T) {
					f := newAPIDataTrieCacheFixture(t, view, version)
					f.save(t, 1, 600<<10, 2)
					f.commit(t)

					account, original := f.attach(t, 1)
					requireAPIDataTrieValue(t, account, 1, 0, 600<<10)
					require.Less(t, original.SizeInMemory(), budget)
					_, stillCached := f.attach(t, 1)
					require.Same(t, original, stillCached)
					requireAPIDataTrieValue(t, account, 1, 1, 600<<10)
					require.Greater(t, original.SizeInMemory(), budget)
					reopened, freshTrie := f.attach(t, 1)
					require.NotSame(t, original, freshTrie)
					requireAPIDataTrieValue(t, reopened, 1, 1, 600<<10)
					requireAPIDataTrieValue(t, account, 1, 0, 600<<10)
				})
				t.Run("oversized-root", func(t *testing.T) {
					f := newAPIDataTrieCacheFixture(t, view, version)
					f.save(t, 1, budget, 1)
					f.save(t, 2, 64<<10, 1)
					f.commit(t)

					_, small := f.attach(t, 2)
					account, original := f.attach(t, 1)
					require.Greater(t, original.SizeInMemory(), budget)
					requireAPIDataTrieValue(t, account, 1, 0, budget)
					_, cached := f.attach(t, 2)
					require.Same(t, small, cached, "an oversized trie must not displace cacheable state")
					reopened, freshTrie := f.attach(t, 1)
					require.NotSame(t, original, freshTrie)
					requireAPIDataTrieValue(t, reopened, 1, 0, budget)
				})
			})
		}
	}
}

type apiDataTrieCacheFixture struct {
	adapters *accountsAdapterCreationResult
	view     string
	root     []byte
}

func newAPIDataTrieCacheFixture(t *testing.T, view string, version core.TrieNodeVersion) *apiDataTrieCacheFixture {
	t.Helper()
	const budget = 1 << 20
	scf, tries := newAccountsViewsFactory(t)
	scf.config.StateTriesConfig.DataTriesSizeInMemoryForCurrentAPI = budget
	scf.config.StateTriesConfig.DataTriesSizeInMemoryForFinalAPI = budget
	scf.config.StateTriesConfig.DataTriesSizeInMemoryForHistoricalAPI = budget
	flags := scf.core.EnableEpochsHandler().(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	flags.AddActiveFlags(common.FixDataTrieLeafValidationFlag)
	if version == core.AutoBalanceEnabled {
		flags.AddActiveFlags(common.AutoBalanceDataTriesFlag)
	}
	trieKey := []byte(dataRetriever.UserAccountsUnit.String())
	manager, err := collapseManager.NewCollapseManager(budget, 100)
	require.NoError(t, err)
	trackedTrie, err := trie.NewTrie(tries.Get(trieKey).GetStorageManager(), scf.core.InternalMarshalizer(), scf.core.Hasher(), flags, manager)
	require.NoError(t, err)
	tries.Put(trieKey, trackedTrie)
	f := &apiDataTrieCacheFixture{view: view}
	scf.chainHandler = &testscommon.ChainHandlerStub{
		GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler { return &block.Header{Nonce: 1} },
		GetCurrentBlockHeaderHashCalled:  func() []byte { return []byte("block") },
		GetCurrentBlockRootHashCalled:    func() []byte { return f.root },
		GetFinalBlockInfoCalled:          func() (uint64, []byte, []byte) { return 1, []byte("block"), f.root },
	}
	f.adapters, err = scf.createAccountsAdapters(tries, disabled.NewDisabledStateAccessesCollector())
	require.NoError(t, err)
	return f
}

func (f *apiDataTrieCacheFixture) save(t *testing.T, id byte, leafSize int, numKeys int) {
	t.Helper()
	account, err := f.adapters.accountsAdapter.LoadAccount(bytes.Repeat([]byte{id}, 32))
	require.NoError(t, err)
	user := account.(state.UserAccountHandler)
	for k := 0; k < numKeys; k++ {
		key := []byte(fmt.Sprintf("key-%d", k))
		require.NoError(t, user.SaveKeyValue(key, bytes.Repeat([]byte{id + byte(k)}, leafSize)))
	}
	require.NoError(t, f.adapters.accountsAdapter.SaveAccount(user))
}

func (f *apiDataTrieCacheFixture) commit(t *testing.T) {
	t.Helper()
	root, err := f.adapters.accountsAdapter.Commit()
	require.NoError(t, err)
	f.root = root
}

func (f *apiDataTrieCacheFixture) account(t *testing.T, id byte) state.UserAccountHandler {
	t.Helper()
	options := api.AccountQueryOptions{OnFinalBlock: f.view == "final"}
	if f.view == "historical" {
		options.BlockRootHash = f.root
	}
	account, blockInfo, err := f.adapters.accountsRepository.GetAccountWithBlockInfo(bytes.Repeat([]byte{id}, 32), options)
	require.NoError(t, err)
	require.Equal(t, f.root, blockInfo.GetRootHash())
	return account.(state.UserAccountHandler)
}

func (f *apiDataTrieCacheFixture) attach(t *testing.T, id byte) (state.UserAccountHandler, common.Trie) {
	t.Helper()
	user := f.account(t, id)
	dataTrie, err := user.DataTrie()
	require.NoError(t, err)
	return user, dataTrie.(common.Trie)
}

func requireAPIDataTrieValue(t *testing.T, user state.UserAccountHandler, id byte, keyIndex int, size int) {
	t.Helper()
	value, _, err := user.RetrieveValue([]byte(fmt.Sprintf("key-%d", keyIndex)))
	require.NoError(t, err)
	require.Equal(t, bytes.Repeat([]byte{id + byte(keyIndex)}, size), value)
}

func TestAccountsAdapters_ProposalPersistsPendingDataTries(t *testing.T) {
	for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
		t.Run(fmt.Sprintf("version-%d", version), func(t *testing.T) {
			scf, tries := newAccountsViewsFactory(t)
			flags := scf.core.EnableEpochsHandler().(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
			flags.AddActiveFlags(common.FixDataTrieLeafValidationFlag)
			if version == core.AutoBalanceEnabled {
				flags.AddActiveFlags(common.AutoBalanceDataTriesFlag)
			}
			adapters, err := scf.createAccountsAdapters(tries, disabled.NewDisabledStateAccessesCollector())
			require.NoError(t, err)
			proposal := adapters.accountsAdapterProposal
			require.NoError(t, proposal.RecreateTrie(holders.NewDefaultRootHashesHolder(nil)))
			address, key := bytes.Repeat([]byte{1}, 32), []byte("key")
			for cycle := 0; cycle < 3; cycle++ {
				account, loadErr := proposal.LoadAccount(address)
				require.NoError(t, loadErr)
				user := account.(state.UserAccountHandler)
				value := []byte(fmt.Sprintf("proposal value %d", cycle))
				require.NoError(t, user.SaveKeyValue(key, value))
				require.NoError(t, proposal.SaveAccount(user))
				root, commitErr := proposal.Commit()
				require.NoError(t, commitErr)
				require.NoError(t, proposal.RecreateTrie(holders.NewDefaultRootHashesHolder(root)))
				persisted, readErr := proposal.GetExistingAccount(address)
				require.NoError(t, readErr)
				actual, _, readErr := persisted.(state.UserAccountHandler).RetrieveValue(key)
				require.NoError(t, readErr)
				require.Equal(t, value, actual)
			}
		})
	}
}
