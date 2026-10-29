package state_test

import (
	"compress/gzip"
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common/stateAnalysis"
	"github.com/multiversx/mx-chain-go/state"
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
