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
				scf, tries := newAccountsViewsFactory(t)
				scf.config.StateTriesConfig.DataTriesSizeInMemoryForCurrentAPI = 1 << 20
				scf.config.StateTriesConfig.DataTriesSizeInMemoryForFinalAPI = 1 << 20
				scf.config.StateTriesConfig.DataTriesSizeInMemoryForHistoricalAPI = 1 << 20
				flags := scf.core.EnableEpochsHandler().(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
				flags.AddActiveFlags(common.FixDataTrieLeafValidationFlag)
				if version == core.AutoBalanceEnabled {
					flags.AddActiveFlags(common.AutoBalanceDataTriesFlag)
				}
				// The shared factory fixture disables size tracking. Use the
				// production collapse manager so these reads really cause eviction.
				mainTrie := tries.Get([]byte(dataRetriever.UserAccountsUnit.String()))
				manager, err := collapseManager.NewCollapseManager(1<<20, 100)
				require.NoError(t, err)
				trackedTrie, err := trie.NewTrie(mainTrie.GetStorageManager(), scf.core.InternalMarshalizer(), scf.core.Hasher(), flags, manager)
				require.NoError(t, err)
				tries.Put([]byte(dataRetriever.UserAccountsUnit.String()), trackedTrie)
				var root []byte
				scf.chainHandler = &testscommon.ChainHandlerStub{
					GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler { return &block.Header{Nonce: 1} },
					GetCurrentBlockHeaderHashCalled:  func() []byte { return []byte("block") },
					GetCurrentBlockRootHashCalled:    func() []byte { return root },
					GetFinalBlockInfoCalled:          func() (uint64, []byte, []byte) { return 1, []byte("block"), root },
				}
				adapters, err := scf.createAccountsAdapters(tries, disabled.NewDisabledStateAccessesCollector())
				require.NoError(t, err)
				key := []byte("storage key")
				var addresses [][]byte
				for i := byte(1); i <= 8; i++ {
					address := bytes.Repeat([]byte{i}, 32)
					addresses = append(addresses, address)
					account, loadErr := adapters.accountsAdapter.LoadAccount(address)
					require.NoError(t, loadErr)
					user := account.(state.UserAccountHandler)
					require.NoError(t, user.SaveKeyValue(key, bytes.Repeat([]byte{i}, 600<<10)))
					require.NoError(t, adapters.accountsAdapter.SaveAccount(user))
				}
				root, err = adapters.accountsAdapter.Commit()
				require.NoError(t, err)
				options := api.AccountQueryOptions{OnFinalBlock: view == "final"}
				if view == "historical" {
					options.BlockRootHash = root
				}
				read := func(index int) state.UserAccountHandler {
					account, info, readErr := adapters.accountsRepository.GetAccountWithBlockInfo(addresses[index], options)
					require.NoError(t, readErr)
					require.Equal(t, root, info.GetRootHash())
					user := account.(state.UserAccountHandler)
					value, _, readErr := user.RetrieveValue(key)
					require.NoError(t, readErr)
					require.Equal(t, bytes.Repeat([]byte{byte(index + 1)}, 600<<10), value)
					tr := user.DataTrie().(common.Trie)
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
					first := read(0)
					firstTrie := first.DataTrie()
					require.Same(t, firstTrie, read(0).DataTrie(), "a resident cache hit must remain reusable")
					for i := 1; i < len(addresses); i++ {
						read(i)
					}
					reloaded := read(0)
					require.NotSame(t, firstTrie, reloaded.DataTrie(), "an evicted clean trie must not survive in a pending-commit buffer")
					// Eviction releases only the cache's reference. A request that
					// already attached its trie must still be able to finish.
					value, _, readErr := first.RetrieveValue(key)
					require.NoError(t, readErr)
					require.Equal(t, bytes.Repeat([]byte{1}, 600<<10), value)
					require.Same(t, firstTrie, first.DataTrie())
				}
			})
		}
	}
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
