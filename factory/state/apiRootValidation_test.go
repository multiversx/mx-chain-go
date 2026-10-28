package state

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/disabled"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/require"
)

type apiRootValidationFixture struct {
	adapters *accountsAdapterCreationResult
	blocks   []common.BlockInfo
	publish  func(index int)
	address  []byte
	key      []byte
}

func newAPIRootValidationFixture(t *testing.T, version core.TrieNodeVersion, headerV3 bool, numBlocks int) apiRootValidationFixture {
	t.Helper()
	scf, tries := newAccountsViewsFactory(t)
	flags := scf.core.EnableEpochsHandler().(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	flags.AddActiveFlags(common.FixDataTrieLeafValidationFlag)
	if version == core.AutoBalanceEnabled {
		flags.AddActiveFlags(common.AutoBalanceDataTriesFlag)
	}
	var published atomic.Value
	published.Store(holders.NewBlockInfo(nil, 0, nil))
	selected := func() common.BlockInfo { return published.Load().(common.BlockInfo) }
	scf.chainHandler = &testscommon.ChainHandlerStub{
		GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
			if headerV3 {
				return &block.HeaderV3{Nonce: selected().GetNonce()}
			}
			return &block.Header{Nonce: selected().GetNonce()}
		},
		GetCurrentBlockHeaderHashCalled: func() []byte { return selected().GetHash() },
		GetCurrentBlockRootHashCalled:   func() []byte { return selected().GetRootHash() },
		GetLastExecutedBlockInfoCalled: func() (uint64, []byte, []byte) {
			info := selected()
			return info.GetNonce(), info.GetHash(), info.GetRootHash()
		},
		GetFinalBlockInfoCalled: func() (uint64, []byte, []byte) {
			info := selected()
			return info.GetNonce(), info.GetHash(), info.GetRootHash()
		},
	}
	adapters, err := scf.createAccountsAdapters(tries, disabled.NewDisabledStateAccessesCollector())
	require.NoError(t, err)
	f := apiRootValidationFixture{adapters: adapters, address: bytes.Repeat([]byte{1}, 32), key: []byte("key")}
	// Advance one processing AccountsDB through committed states. Both API views
	// get their own production account factories and caches from the state factory.
	for nonce := uint64(1); nonce <= uint64(numBlocks); nonce++ {
		account, loadErr := adapters.accountsAdapter.LoadAccount(f.address)
		require.NoError(t, loadErr)
		user := account.(state.UserAccountHandler)
		user.IncreaseNonce(1)
		require.NoError(t, user.SaveKeyValue(f.key, []byte(fmt.Sprintf("value-%d", nonce))))
		require.NoError(t, adapters.accountsAdapter.SaveAccount(user))
		root, commitErr := adapters.accountsAdapter.Commit()
		require.NoError(t, commitErr)
		f.blocks = append(f.blocks, holders.NewBlockInfo([]byte(fmt.Sprintf("block-%d", nonce)), nonce, root))
	}
	f.publish = func(index int) { published.Store(f.blocks[index]) }
	return f
}

func (f apiRootValidationFixture) account(t *testing.T, view string, index int) state.UserAccountHandler {
	t.Helper()
	account, info, err := f.adapters.accountsRepository.GetAccountWithBlockInfo(f.address, api.AccountQueryOptions{OnFinalBlock: view == "final"})
	require.NoError(t, err)
	require.True(t, f.blocks[index].Equal(info))
	user := account.(state.UserAccountHandler)
	require.Equal(t, f.blocks[index].GetNonce(), user.GetNonce())
	return user
}

func (f apiRootValidationFixture) read(t *testing.T, account state.UserAccountHandler, index int) {
	t.Helper()
	value, _, err := account.RetrieveValue(f.key)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("value-%d", f.blocks[index].GetNonce()), string(value))
}

func TestAccountsAdapters_CurrentAndFinalKeepAccountStorageAtSelectedRoot(t *testing.T) {
	for _, view := range []string{"current", "final"} {
		for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
			for _, headerV3 := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/version-%d/headerV3-%t", view, version, headerV3), func(t *testing.T) {
					f := newAPIRootValidationFixture(t, version, headerV3, 6)
					for cycle := 0; cycle < 2; cycle++ {
						base := cycle * 3
						f.publish(base)
						older := f.account(t, view, base)
						// The first API request pauses after account selection, before
						// its lazy storage read. Another request warms the next state.
						f.publish(base + 1)
						newer := f.account(t, view, base+1)
						f.read(t, newer, base+1)
						f.read(t, older, base)
						attached := older.DataTrie()
						f.publish(base + 2)
						latest := f.account(t, view, base+2)
						f.read(t, latest, base+2)
						f.read(t, older, base)
						require.Same(t, attached, older.DataTrie(), "root changes must not invalidate an attached trie")
						matching := f.account(t, view, base+2)
						require.Same(t, latest.DataTrie(), matching.DataTrie(), "matching-root cache hits must stay reusable")
					}
				})
			}
		}
	}
}

func TestAccountsAdapters_CurrentAndFinalDelayedReadsAcrossAdvancingRoots(t *testing.T) {
	for _, view := range []string{"current", "final"} {
		t.Run(view, func(t *testing.T) {
			f := newAPIRootValidationFixture(t, core.AutoBalanceEnabled, true, 34)
			for cycle := 0; cycle < 2; cycle++ {
				base := cycle * 17
				f.publish(base)
				older := f.account(t, view, base)
				for next := base + 1; next <= base+16; next++ {
					f.publish(next)
					f.read(t, f.account(t, view, next), next)
				}
				f.read(t, older, base)
				// The delayed request can replace the address-cache entry with its
				// older trie. A fresh request at the unchanged head must reject it.
				f.read(t, f.account(t, view, base+16), base+16)
				f.read(t, older, base)
			}
		})
	}
}
