package state_test

import (
	"compress/gzip"
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/common/stateAnalysis"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/dataTrieValue"
	"github.com/multiversx/mx-chain-go/state/trackableDataTrie"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
)

func readBlocksCsv(t *testing.T, path string) (map[string]int, []string) {
	f, err := os.Open(path)
	require.Nil(t, err)
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	require.Nil(t, err)
	rows, err := csv.NewReader(gz).ReadAll()
	require.Nil(t, err)
	require.GreaterOrEqual(t, len(rows), 2)

	columns := make(map[string]int)
	for i, name := range rows[0] {
		columns[name] = i
	}
	return columns, rows[len(rows)-1]
}

func TestStateAnalysisHooks_AccountsDBFlow(t *testing.T) {
	// this test enables the global collector, so it must not run in parallel with other tests

	dir := t.TempDir()
	cfg := stateAnalysis.Config{
		Enabled:                  true,
		Label:                    "hooks",
		BlocksPerRow:             1,
		MemorySampleEveryNBlocks: 1,
	}
	require.Nil(t, stateAnalysis.Init(cfg, 0, dir))
	defer stateAnalysis.Close()

	_, adb := getDefaultTrieAndAccountsDb()
	addr := []byte("address-with-data-trie-000000000")

	stateAnalysis.BeginBlock(0, 0, 1, 1, 2, false)

	acc, err := adb.LoadAccount(addr)
	require.Nil(t, err)
	userAcc := acc.(state.UserAccountHandler)
	require.Nil(t, userAcc.SaveKeyValue([]byte("key"), []byte("value")))
	require.Nil(t, adb.SaveAccount(userAcc))

	// second save of the same account, with an identical data trie value (no-op write)
	acc, err = adb.LoadAccount(addr)
	require.Nil(t, err)
	userAcc = acc.(state.UserAccountHandler)
	require.Nil(t, userAcc.SaveKeyValue([]byte("key"), []byte("value")))
	require.Nil(t, adb.SaveAccount(userAcc))

	_, err = adb.Commit()
	require.Nil(t, err)

	stateAnalysis.EndBlock()
	stateAnalysis.Close()

	columns, row := readBlocksCsv(t, filepath.Join(dir, "state-analysis", "shard_0", "blocks.csv.gz"))
	get := func(name string) int64 {
		idx, found := columns[name]
		require.True(t, found, name)
		v, errConv := strconv.ParseInt(row[idx], 10, 64)
		require.Nil(t, errConv)
		return v
	}

	assert.Equal(t, int64(1), get("blk_count"))
	assert.Equal(t, int64(2), get("blk_txs"))
	assert.Equal(t, int64(2), get("adb_load_account_user"))
	assert.Equal(t, int64(2), get("adb_save_account_user"))
	assert.Equal(t, int64(1), get("adb_save_new_account_user"))
	assert.Equal(t, int64(1), get("adb_save_repeat_addr_user"))
	assert.Equal(t, int64(1), get("adb_unique_accounts_block_user"))
	assert.Equal(t, int64(1), get("adb_datatrie_holder_hit_user"))
	assert.Equal(t, int64(2), get("adb_save_datatrie_dirty_user"))
	assert.Equal(t, int64(1), get("adb_commit_user"))
	assert.Equal(t, int64(1), get("adb_commit_datatries_user"))
	assert.Greater(t, get("adb_journal_len_at_commit_user"), int64(0))
	assert.Equal(t, int64(2), get("dt_update_trie_calls"))
	assert.Equal(t, int64(1), get("dt_update_insert"))
	assert.Equal(t, int64(1), get("dt_update_noop"))
	assert.Greater(t, get("trie_get_main"), int64(0))
	assert.Greater(t, get("trie_get_data"), int64(0))
	assert.Greater(t, get("trie_update_main"), int64(0))
	assert.Greater(t, get("trie_update_data"), int64(0))
	assert.Equal(t, int64(1), get("trie_commit_main"))
	assert.Equal(t, int64(1), get("trie_commit_data"))
	assert.Greater(t, get("trie_dirty_committed_main_leaf"), int64(0))
	assert.Greater(t, get("trie_dirty_committed_data_leaf"), int64(0))
	assert.Greater(t, get("trie_encode_main_leaf"), int64(0))
	assert.Greater(t, get("trie_roothash_calls_data"), int64(0))
	assert.Greater(t, get("st_tsm_put_accounts"), int64(0))
	assert.Greater(t, get("mem_heap_alloc"), int64(0))
	assert.Greater(t, get("trie_resident_nodes_main_leaf"), int64(0))
	assert.Greater(t, get("adb_holder_datatries_user"), int64(0))

}

func TestStateAnalysisHooks_PeerRootHashCosts(t *testing.T) {
	tr, _ := getDefaultTrieAndAccountsDb()
	require.NoError(t, tr.Update([]byte("peer-key"), []byte("peer-value")))
	args := createMockAccountsDBArgs()
	args.Trie = tr
	adb, err := state.NewPeerAccountsDB(args)
	require.NoError(t, err)

	dir := t.TempDir()
	require.NoError(t, stateAnalysis.Init(stateAnalysis.Config{Enabled: true}, 0, dir))
	t.Cleanup(stateAnalysis.Close)
	stateAnalysis.BeginBlock(0, 0, 1, 1, 0, false)
	stateAnalysis.SetTrieKind(stateAnalysis.KindMain)
	root, err := adb.RootHash()
	require.NoError(t, err)
	assert.NotEmpty(t, root)
	assert.Equal(t, stateAnalysis.KindMain, stateAnalysis.TrieKind())
	stateAnalysis.EndBlock()
	stateAnalysis.Close()

	columns, row := readBlocksCsv(t, filepath.Join(dir, "state-analysis", "shard_0", "blocks.csv.gz"))
	for _, names := range [][2]string{
		{"trie_roothash_calls_peer", "trie_roothash_calls_main"},
		{"trie_roothash_computed_peer", "trie_roothash_computed_main"},
		{"trie_encode_peer_leaf", "trie_encode_main_leaf"},
		{"trie_hash_compute_peer", "trie_hash_compute_main"},
	} {
		require.Contains(t, columns, names[0])
		require.Contains(t, columns, names[1])
		assert.Equal(t, "1", row[columns[names[0]]], names[0])
		assert.Equal(t, "0", row[columns[names[1]]], names[1])
	}
}

func TestStateAnalysisHooks_RootHashContextRestored(t *testing.T) {
	// These calls share a helper so commit can hash while holding mutOp without relocking it.
	for _, peer := range []bool{false, true} {
		for _, entrypoint := range []string{"root", "commit", "recreate"} {
			for _, fail := range []bool{false, true} {
				name := strconv.FormatBool(peer) + "/" + entrypoint + "/error_" + strconv.FormatBool(fail)
				t.Run(name, func(t *testing.T) {
					dir := t.TempDir()
					require.NoError(t, stateAnalysis.Init(stateAnalysis.Config{Enabled: true}, 0, dir))
					t.Cleanup(stateAnalysis.Close)
					wantKind := stateAnalysis.KindMain
					if peer {
						wantKind = stateAnalysis.KindPeer
					}
					var wantErr error
					if fail {
						wantErr = errors.New("root hash failed")
					}
					args := createMockAccountsDBArgs()
					called := false
					args.Trie.(*trieMock.TrieStub).CommitCalled = func() error { return nil }
					args.Trie.(*trieMock.TrieStub).RootCalled = func() ([]byte, error) {
						called = true
						assert.Equal(t, wantKind, stateAnalysis.TrieKind())
						return []byte("root"), wantErr
					}
					var adb state.AccountsAdapter
					var err error
					if peer {
						adb, err = state.NewPeerAccountsDB(args)
					} else {
						adb, err = state.NewAccountsDB(args)
					}
					require.NoError(t, err)
					stateAnalysis.SetTrieKind(stateAnalysis.KindData)
					switch entrypoint {
					case "root":
						_, err = adb.RootHash()
					case "commit":
						_, err = adb.Commit()
					case "recreate":
						err = adb.RecreateTrieIfNeeded(holders.NewDefaultRootHashesHolder([]byte("root")))
					}
					assert.Equal(t, wantErr, err)
					assert.True(t, called)
					assert.Equal(t, stateAnalysis.KindData, stateAnalysis.TrieKind())
				})
			}
		}
	}
}

func TestStateAnalysisHooks_DataTrieKeyHashes(t *testing.T) {
	for _, tc := range []struct {
		name       string
		balanced   bool
		present    bool
		version    core.TrieNodeVersion
		read       bool
		value      string
		wantHashes int
	}{
		{name: "legacy write", present: true, value: "new-value", wantHashes: 0},
		{name: "balanced write", balanced: true, present: true, version: core.AutoBalanceEnabled, value: "new-value", wantHashes: 2},
		{name: "balanced insert", balanced: true, value: "new-value", wantHashes: 3},
		{name: "balanced delete", balanced: true, present: true, version: core.AutoBalanceEnabled, wantHashes: 3},
		{name: "migration", balanced: true, present: true, value: "new-value", wantHashes: 2},
		{name: "balanced read", balanced: true, present: true, version: core.AutoBalanceEnabled, read: true, wantHashes: 1},
		{name: "balanced missing read", balanced: true, read: true, wantHashes: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, identifier := []byte("key"), []byte("address")
			hashCalls := 0
			hasher := &testscommon.HasherStub{ComputeCalled: func(input string) []byte {
				assert.Equal(t, string(key), input)
				hashCalls++
				return []byte("hashed-key")
			}}
			marshaller := &marshallerMock.MarshalizerMock{}
			tdt, err := trackableDataTrie.NewTrackableDataTrie(identifier, hasher, marshaller,
				&enableEpochsHandlerMock.EnableEpochsHandlerStub{IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return tc.balanced && flag == common.AutoBalanceDataTriesFlag
				}}, &stateMock.StateAccessesCollectorStub{})
			require.NoError(t, err)
			storedKey, storedValue := "key", []byte("old-valuekeyaddress")
			if tc.version == core.AutoBalanceEnabled {
				storedKey = "hashed-key"
				storedValue, err = marshaller.Marshal(&dataTrieValue.TrieLeafData{
					Key: key, Address: identifier, Value: []byte("old-value"),
				})
				require.NoError(t, err)
			}
			tr := &trieMock.TrieStub{
				GetWithVersionCalled: func(k []byte) ([]byte, uint32, uint32, error) {
					if tc.present && string(k) == storedKey {
						return storedValue, 0, uint32(tc.version), nil
					}
					return nil, 0, 0, nil
				},
				UpdateWithVersionCalled: func(_ []byte, _ []byte, _ core.TrieNodeVersion) error { return nil },
				DeleteCalled:            func(_ []byte) error { return nil },
			}
			tdt.SetDataTrie(tr)
			dir := t.TempDir()
			require.NoError(t, stateAnalysis.Init(stateAnalysis.Config{Enabled: true}, 0, dir))
			t.Cleanup(stateAnalysis.Close)
			stateAnalysis.BeginBlock(0, 0, 1, 1, 0, false)
			if tc.read {
				_, _, err = tdt.RetrieveValue(key)
			} else {
				require.NoError(t, tdt.SaveKeyValue(key, []byte(tc.value)))
				_, _, err = tdt.SaveDirtyData(tr)
			}
			require.NoError(t, err)
			stateAnalysis.EndBlock()
			stateAnalysis.Close()
			assert.Equal(t, tc.wantHashes, hashCalls)
			columns, row := readBlocksCsv(t, filepath.Join(dir, "state-analysis", "shard_0", "blocks.csv.gz"))
			require.Contains(t, columns, "dt_key_hash_computed")
			assert.Equal(t, strconv.Itoa(hashCalls), row[columns["dt_key_hash_computed"]])
		})
	}
}
