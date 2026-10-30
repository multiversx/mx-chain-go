package stateAnalysis

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime/pprof"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
)

func readGzipCsv(t *testing.T, path string) [][]string {
	f, err := os.Open(path)
	require.Nil(t, err)
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	require.Nil(t, err)
	rows, err := csv.NewReader(gz).ReadAll()
	require.Nil(t, err)
	return rows
}

func column(header []string, name string) int {
	for i, h := range header {
		if h == name {
			return i
		}
	}
	return -1
}

func TestRegistry_NamesAreUnique(t *testing.T) {
	seen := make(map[string]struct{})
	for _, n := range reg.counterNames {
		_, found := seen[n]
		assert.False(t, found, n)
		seen[n] = struct{}{}
	}
	for _, n := range reg.gaugeNames {
		_, found := seen[n]
		assert.False(t, found, n)
		seen[n] = struct{}{}
	}
	assert.Equal(t, len(reg.counterNames), len(reg.counterMax))
}

func TestDistanceBucket(t *testing.T) {
	assert.Equal(t, 0, distanceBucket(0))
	assert.Equal(t, 1, distanceBucket(1))
	assert.Equal(t, 2, distanceBucket(2))
	assert.Equal(t, 2, distanceBucket(3))
	assert.Equal(t, 3, distanceBucket(4))
	assert.Equal(t, 11, distanceBucket(1024))
	assert.Equal(t, numReuseBuckets-2, distanceBucket(1<<40))
}

func TestReuseTracker(t *testing.T) {
	rt := newReuseTracker(2)
	rt.touch("a", 10, 0)
	rt.touch("a", 10, 0)
	rt.touch("a", 10, 3)
	rt.touch("b", 5, 3)
	rt.touch("c", 1, 4) // map is full -> reset, c is new

	counts, bytes, tracked, resets, touches := rt.flush()
	assert.Equal(t, int64(5), touches)
	assert.Equal(t, int64(3), counts[reuseBucketNew])
	assert.Equal(t, int64(1), counts[0])
	assert.Equal(t, int64(1), counts[distanceBucket(3)])
	assert.Equal(t, int64(16), bytes[reuseBucketNew])
	assert.Equal(t, 1, tracked)
	assert.Equal(t, int64(1), resets)

	counts, _, _, resets, _ = rt.flush()
	assert.Equal(t, int64(0), counts[0])
	assert.Equal(t, int64(0), resets)
}

func TestHotSet(t *testing.T) {
	hs := newHotSet(2)
	hs.touch("a")
	hs.touch("a")
	hs.touch("b")
	hs.touch("c") // dropped

	top, distinct, total, dropped := hs.flush(1)
	require.Len(t, top, 1)
	assert.Equal(t, "a", top[0].key)
	assert.Equal(t, int64(2), top[0].count)
	assert.Equal(t, 2, distinct)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, int64(1), dropped)
}

func TestBlockSet(t *testing.T) {
	bs := newBlockSet()
	bs.touch([]byte("a"), false)
	bs.touch([]byte("a"), true)
	bs.touch([]byte("a"), true)
	bs.touch([]byte("b"), true)
	unique, seenPrev, repeat := bs.rotate()
	assert.Equal(t, int64(2), unique)
	assert.Equal(t, int64(0), seenPrev)
	assert.Equal(t, int64(1), repeat)

	bs.touch([]byte("b"), false)
	bs.touch([]byte("c"), false)
	unique, seenPrev, repeat = bs.rotate()
	assert.Equal(t, int64(2), unique)
	assert.Equal(t, int64(1), seenPrev)
	assert.Equal(t, int64(0), repeat)
}

func TestDisabledCollectorIsNoop(t *testing.T) {
	assert.False(t, Enabled())
	Inc(CBlocks)
	Add(CBlockWallNs, 5)
	Max(MStDbGetNs, 5)
	SetGauge(GMemRSS, 5)
	assert.True(t, Now().IsZero())
	assert.Equal(t, KindMain, SetTrieKind(KindData))
	assert.Equal(t, KindMain, TrieKind())
	BeginBlock(0, 0, 0, 0, 0, false)
	EndBlock()
	TouchAccount(AdapterUser, []byte("a"), 1, true)
	TouchDataTrie(AdapterUser, []byte("a"), true)
	TouchDataTrieKey([]byte("a"), []byte("k"), 1)
	assert.False(t, ShouldSampleMemory())
	SnapshotCompleted("AccountsTrie", 0, nil, 0, 0, 0, false)
	Close()
	require.Nil(t, Init(Config{Enabled: false}, 0, t.TempDir()))
	assert.False(t, Enabled())
}

func TestCollector_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		Enabled:                  true,
		Label:                    "test",
		OutputDir:                "out",
		BlocksPerRow:             2,
		MaxTrackedKeys:           100,
		TopKPerEpoch:             5,
		MemorySampleEveryNBlocks: 1,
		OutlierCommitTimeMs:      1,
		OutlierStorageReads:      1000,
		OutlierJournalLen:        1000,
	}
	require.Nil(t, Init(cfg, 1, dir))
	require.True(t, Enabled())
	defer func() {
		Close()
		global = nil
	}()

	RegisterCacheProbe("AccountsTrie", func() (int, int, int) { return 7, 700, 3 })

	// block 1, epoch 0
	BeginBlock(1, 0, 10, 100, 3, false)
	assert.Equal(t, KindMain, SetTrieKind(KindData))
	assert.Equal(t, KindData, TrieKind())
	SetTrieKind(KindMain)
	Inc(CAdbSaveAccount.Adapter(AdapterUser))
	Add(CStDbGet.Unit(UnitAccounts), 10)
	Max(MStDbGetNs, 40)
	Max(MStDbGetNs, 20)
	TouchAccount(AdapterUser, []byte("addr1"), 100, true)
	TouchAccount(AdapterUser, []byte("addr1"), 100, true)
	TouchAccount(AdapterUser, []byte("addr2"), 100, false)
	TouchDataTrie(AdapterUser, []byte("addr1"), true)
	TouchDataTrieKey([]byte("addr1"), []byte("k"), 8)
	assert.True(t, ShouldSampleMemory())
	SetResidentStats(KindMain, ResidentStats{Nodes: [numNodeTypes]int64{1, 2, 3}, Bytes: [numNodeTypes]int64{10, 20, 30}})
	EndBlock()

	// block 2, epoch 0: outlier (commit time above 1 ms)
	BeginBlock(1, 0, 11, 101, 0, false)
	Add(CBlockCommitNs, 5_000_000)
	TouchAccount(AdapterUser, []byte("addr1"), 100, false)
	EndBlock()

	// block 3, epoch 1: triggers the epoch flush of epoch 0
	BeginBlock(1, 1, 12, 102, 0, true)
	// an aborted block: BeginBlock again without EndBlock
	BeginBlock(1, 1, 13, 103, 0, false)
	EndBlock()

	Close()

	outDir := filepath.Join(dir, "out", "shard_1")
	rows := readGzipCsv(t, filepath.Join(outDir, "blocks.csv.gz"))
	require.GreaterOrEqual(t, len(rows), 3) // header + epoch 0 bucket + epoch 1 bucket
	header := rows[0]
	assert.Equal(t, "shard", header[0])
	first := rows[1]
	assert.Equal(t, "1", first[column(header, "shard")])
	assert.Equal(t, "test", first[column(header, "label")])
	assert.Equal(t, "0", first[column(header, "epoch")])
	assert.Equal(t, "100", first[column(header, "nonce_first")])
	assert.Equal(t, "101", first[column(header, "nonce_last")])
	assert.Equal(t, "2", first[column(header, "blocks")])
	assert.Equal(t, "2", first[column(header, "blk_count")])
	assert.Equal(t, "3", first[column(header, "blk_txs")])
	assert.Equal(t, "1", first[column(header, "adb_save_account_user")])
	assert.Equal(t, "10", first[column(header, "st_db_get_accounts")])
	assert.Equal(t, "5000000", first[column(header, "blk_commit_state_ns")])
	assert.Equal(t, "5000000", first[column(header, "max_blk_commit_state_ns")])
	assert.Equal(t, "3", first[column(header, "adb_unique_accounts_block_user")])
	assert.Equal(t, "2", first[column(header, "max_adb_unique_accounts_block_user")])
	assert.Equal(t, "1", first[column(header, "adb_unique_accounts_seen_prev_block_user")])
	assert.Equal(t, "1", first[column(header, "adb_save_repeat_addr_user")])
	assert.Equal(t, "1", first[column(header, "adb_unique_datatries_block_user")])
	assert.Equal(t, "7", first[column(header, "st_cache_entries_accounts")])
	assert.Equal(t, "700", first[column(header, "st_cache_bytes_accounts")])
	assert.Equal(t, "3", first[column(header, "st_active_persisters_accounts")])
	assert.Equal(t, "2", first[column(header, "trie_resident_nodes_main_ext")])
	assert.Equal(t, "30", first[column(header, "trie_resident_bytes_main_leaf")])
	assert.NotEqual(t, "0", first[column(header, "mem_heap_alloc")])
	assert.Equal(t, "40", first[column(header, "st_db_get_ns_max")]) // largest per-block running max inside the row

	second := rows[2]
	assert.Equal(t, "1", second[column(header, "epoch")])
	assert.Equal(t, "2", second[column(header, "blk_count")])
	assert.Equal(t, "1", second[column(header, "blk_aborted")])
	assert.Equal(t, "1", second[column(header, "blk_start_of_epoch")])

	outliers := readGzipCsv(t, filepath.Join(outDir, "outliers.csv.gz"))
	require.Len(t, outliers, 2)
	assert.Equal(t, "101", outliers[1][column(header, "nonce_first")])
	assert.Equal(t, "1", outliers[1][column(header, "blocks")])
	assert.Equal(t, "0", outliers[1][column(header, "st_db_get_ns_max")]) // running max is reset at every block

	reuse := readGzipCsv(t, filepath.Join(outDir, "reuse.csv.gz"))
	require.GreaterOrEqual(t, len(reuse), 1+2*numReuse)
	reuseHeader := reuse[0]
	accountRow := reuse[1]
	assert.Equal(t, "0", accountRow[column(reuseHeader, "epoch")])
	assert.Equal(t, "account", accountRow[column(reuseHeader, "kind")])
	assert.Equal(t, "4", accountRow[column(reuseHeader, "touches")])
	assert.Equal(t, "2", accountRow[column(reuseHeader, "count_b"+itoa(reuseBucketNew))])
	assert.Equal(t, "1", accountRow[column(reuseHeader, "count_b0")])
	assert.Equal(t, "1", accountRow[column(reuseHeader, "count_b1")])
	assert.Equal(t, "200", accountRow[column(reuseHeader, "bytes_b"+itoa(reuseBucketNew))])

	hot := readGzipCsv(t, filepath.Join(outDir, "hotset.csv.gz"))
	require.GreaterOrEqual(t, len(hot), 2)
	hotHeader := hot[0]
	assert.Equal(t, "account", hot[1][column(hotHeader, "kind")])
	assert.Equal(t, "1", hot[1][column(hotHeader, "rank")])
	assert.Equal(t, "3", hot[1][column(hotHeader, "count")])
	assert.Equal(t, "6164647231", hot[1][column(hotHeader, "key")]) // hex("addr1")
}

func TestCollector_ReinitAppendsToExistingFiles(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{Enabled: true, Label: "test", OutputDir: "out", BlocksPerRow: 1}
	defer func() {
		Close()
		global = nil
	}()

	for cycle := 0; cycle < 2; cycle++ {
		require.Nil(t, Init(cfg, 0, dir))
		BeginBlock(0, 0, uint64(cycle), uint64(cycle), 0, false)
		EndBlock()
		Close()
	}

	rows := readGzipCsv(t, filepath.Join(dir, "out", "shard_0", "blocks.csv.gz"))
	require.Len(t, rows, 3) // one header + one row per cycle
	assert.Equal(t, "shard", rows[0][0])
	assert.Equal(t, "0", rows[1][column(rows[0], "nonce_first")])
	assert.Equal(t, "1", rows[2][column(rows[0], "nonce_first")])
}

func TestCollector_FirstBlockExcludesStartup(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Init(Config{
		Enabled: true, BlocksPerRow: 1, OutlierCommitTimeMs: 1, OutlierStorageReads: 4,
	}, 0, dir))
	t.Cleanup(func() { Close(); global = nil })

	touch := func(address string) {
		for a := 0; a < numAdapters; a++ {
			TouchAccount(a, []byte(address), 10, true)
			TouchDataTrie(a, []byte(address), true)
		}
		TouchDataTrieKey([]byte(address), []byte("key"), 10)
	}
	Add(CBlockCommitNs, 10_000_000)
	Add(CStDbGet.Unit(UnitAccounts), 100)
	Max(MStDbGetNs, 999)
	touch("startup-only")
	touch("shared")
	touch("shared")
	SetGauge(GHolderDataTries.At(AdapterUser), 23)

	for block := uint64(1); block <= 2; block++ {
		BeginBlock(0, 7, block, block, 3, false)
		Add(CStDbGet.Unit(UnitAccounts), int64(block))
		Max(MStDbGetNs, 10)
		touch("shared")
		EndBlock()
	}
	Close()

	outDir := filepath.Join(dir, "state-analysis", "shard_0")
	rows := readGzipCsv(t, filepath.Join(outDir, "blocks.csv.gz"))
	require.Len(t, rows, 3)
	header := rows[0]
	for block, row := range rows[1:] {
		assert.Equal(t, itoa(block+1), row[column(header, "st_db_get_accounts")])
		assert.Equal(t, "0", row[column(header, "blk_commit_state_ns")])
		assert.Equal(t, "3", row[column(header, "blk_txs")])
		assert.Equal(t, "10", row[column(header, "st_db_get_ns_max")])
		assert.Equal(t, "23", row[column(header, "adb_holder_datatries_user")])
		for _, adapter := range adapterNames {
			assert.Equal(t, "1", row[column(header, "adb_unique_accounts_block_"+adapter)])
			assert.Equal(t, "1", row[column(header, "adb_unique_datatries_block_"+adapter)])
			assert.Equal(t, itoa(block), row[column(header, "adb_unique_accounts_seen_prev_block_"+adapter)])
			assert.Equal(t, itoa(block), row[column(header, "adb_unique_datatries_seen_prev_block_"+adapter)])
			assert.Equal(t, "0", row[column(header, "adb_save_repeat_addr_"+adapter)])
		}
	}
	assert.Len(t, readGzipCsv(t, filepath.Join(outDir, "outliers.csv.gz")), 1)

	reuse := readGzipCsv(t, filepath.Join(outDir, "reuse.csv.gz"))
	require.Len(t, reuse, 1+numReuse)
	for _, row := range reuse[1:] {
		assert.Equal(t, "2", row[column(reuse[0], "touches")])
		assert.Equal(t, "1", row[column(reuse[0], "tracked_keys")])
		assert.Equal(t, "1", row[column(reuse[0], "count_b"+itoa(reuseBucketNew))])
		assert.Equal(t, "0", row[column(reuse[0], "count_b0")])
		assert.Equal(t, "1", row[column(reuse[0], "count_b1")])
	}
	hot := readGzipCsv(t, filepath.Join(outDir, "hotset.csv.gz"))
	require.Len(t, hot, 1+numReuse)
	for _, row := range hot[1:] {
		assert.Equal(t, "1", row[column(hot[0], "distinct_keys")])
		assert.Equal(t, "2", row[column(hot[0], "count")])
		assert.Equal(t, "2", row[column(hot[0], "total_touches")])
	}
}

func TestCollector_CloseFinalizesOpenBlock(t *testing.T) {
	for _, blocksPerRow := range []int{1, 100} {
		for _, completed := range []int{0, 1} {
			t.Run(itoa(blocksPerRow)+"_completed_"+itoa(completed), func(t *testing.T) {
				dir := t.TempDir()
				require.NoError(t, Init(Config{
					Enabled: true, BlocksPerRow: blocksPerRow, OutlierCommitTimeMs: 1,
				}, 0, dir))
				t.Cleanup(func() { Close(); global = nil })
				if completed != 0 {
					BeginBlock(0, 7, 10, 100, 0, false)
					EndBlock()
				}
				BeginBlock(0, 7, 11, 101, 3, false)
				Add(CBlockCommitNs, 5_000_000)
				Add(CStDbGet.Unit(UnitAccounts), 9)
				Max(MStDbGetNs, 40)
				TouchAccount(AdapterUser, []byte("address"), 10, true)
				Close()
				Close() // finalization must be idempotent
				EndBlock()
				assert.False(t, Enabled())
				assert.False(t, global.blockOpen.Load())

				outDir := filepath.Join(dir, "state-analysis", "shard_0")
				rows := readGzipCsv(t, filepath.Join(outDir, "blocks.csv.gz"))
				wantRows, wantBlocks := 2, completed+1
				if blocksPerRow == 1 {
					wantRows += completed
					wantBlocks = 1
				}
				require.Len(t, rows, wantRows)
				header, last := rows[0], rows[len(rows)-1]
				assert.Equal(t, itoa(wantBlocks), last[column(header, "blocks")])
				assert.Equal(t, itoa(wantBlocks), last[column(header, "blk_count")])
				assert.Equal(t, "1", last[column(header, "blk_aborted")])
				assert.Equal(t, "101", last[column(header, "nonce_last")])
				assert.Equal(t, "3", last[column(header, "blk_txs")])
				assert.Equal(t, "9", last[column(header, "st_db_get_accounts")])
				assert.Equal(t, "40", last[column(header, "st_db_get_ns_max")])
				assert.Equal(t, "1", last[column(header, "adb_unique_accounts_block_user")])
				outliers := readGzipCsv(t, filepath.Join(outDir, "outliers.csv.gz"))
				require.Len(t, outliers, 2)
				assert.Equal(t, "101", outliers[1][column(header, "nonce_first")])
				assert.Equal(t, "1", outliers[1][column(header, "blk_aborted")])
				assert.Equal(t, "5000000", outliers[1][column(header, "blk_commit_state_ns")])
			})
		}
	}
}

func TestCollector_FamilyTimerPeaks(t *testing.T) {
	timers := []struct {
		id   CounterID
		name string
	}{
		{CAdbCommitNs.Adapter(AdapterUser), "adb_commit_ns_user"},
		{CAdbCommitNs.Adapter(AdapterPeer), "adb_commit_ns_peer"},
		{CTrieGetNs.Kind(KindData), "trie_get_ns_data"},
		{CStDbGetNs.Unit(UnitAccounts), "st_db_get_ns_accounts"},
	}
	dir := t.TempDir()
	require.NoError(t, Init(Config{Enabled: true, BlocksPerRow: 2}, 0, dir))
	t.Cleanup(func() { Close(); global = nil })
	for block, ns := range []int64{10, 30} {
		BeginBlock(0, 0, uint64(block), uint64(block), 0, false)
		for _, timer := range timers {
			Add(timer.id, ns)
		}
		EndBlock()
	}
	Close()
	rows := readGzipCsv(t, filepath.Join(dir, "state-analysis", "shard_0", "blocks.csv.gz"))
	require.Len(t, rows, 2)
	for _, timer := range timers {
		maxCol := column(rows[0], "max_"+timer.name)
		require.NotEqual(t, -1, maxCol, timer.name)
		assert.Equal(t, "30", rows[1][maxCol], timer.name)
		assert.Equal(t, "40", rows[1][column(rows[0], timer.name)], timer.name)
	}
	assert.Equal(t, -1, column(rows[0], "max_adb_load_account_user"))
}

func TestSnapshotStructure(t *testing.T) {
	dir := t.TempDir()
	require.Nil(t, Init(Config{Enabled: true, Label: "s"}, 0, dir))
	defer func() {
		Close()
		global = nil
	}()

	identifier := "AccountsTrie"
	SnapshotStarted(identifier)

	mainStats := WrapSnapshotStats(newStatsStub(), identifier, false)
	sc, ok := mainStats.(StructureCollector)
	require.True(t, ok)
	sc.AddBranchStructure(0, 100, 5)
	sc.AddExtensionStructure(1, 40, 3)
	sc.AddLeafStructure(2, 200, 60, 120, core.NotSpecified)
	ObserveSnapshotTrie(mainStats)
	assert.NotNil(t, UnwrapSnapshotStats(mainStats))
	_, isWrapper := UnwrapSnapshotStats(mainStats).(*snapshotStats)
	assert.False(t, isWrapper)

	dataStats := WrapSnapshotStats(newStatsStub(), identifier, true)
	dsc := dataStats.(StructureCollector)
	dsc.AddLeafStructure(1, 100, 10, 90, core.AutoBalanceEnabled)
	ObserveSnapshotTrie(dataStats)

	ObserveSnapshotAccount(identifier, 50, true, true, true)
	ObserveSnapshotAccount(identifier, 50, true, false, false)
	ObserveSnapshotAccount(identifier, 500, false, false, false)
	ObserveSnapshotCodeEntry(identifier, 450, 2)

	// unrelated unit is untouched
	peerStats := WrapSnapshotStats(newStatsStub(), "PeerAccountsTrie", false)
	peerStats.(StructureCollector).AddLeafStructure(0, 10, 1, 1, core.NotSpecified)

	SnapshotCompleted(identifier, 7, []byte{1, 2}, 12, 3, 1, false)

	path := filepath.Join(dir, "state-analysis", "shard_0", "structure", "epoch_7_accounts.json")
	data, err := os.ReadFile(path)
	require.Nil(t, err)
	var out structureJSON
	require.Nil(t, json.Unmarshal(data, &out))
	assert.Equal(t, uint32(7), out.Epoch)
	assert.Equal(t, "accounts", out.Unit)
	main := out.Tries["main"]
	assert.Equal(t, int64(1), main.Nodes["branch"])
	assert.Equal(t, int64(1), main.Nodes["ext"])
	assert.Equal(t, int64(1), main.Nodes["leaf"])
	assert.Equal(t, int64(100), main.Bytes["branch"])
	assert.Equal(t, int64(1), main.BranchFanOut[5])
	assert.Equal(t, int64(1), main.ExtensionKeyLen[3])
	assert.Equal(t, int64(1), main.LeafKeyLen[60])
	assert.Equal(t, int64(120), main.LeafValueBytes)
	assert.Equal(t, int64(1), main.LeafVersions[0])
	assert.Equal(t, int64(0), main.LeafMetadataEst)
	assert.Equal(t, int64(1), main.DepthNodes["leaf"][2])
	assert.Equal(t, int64(1), main.Tries)
	dataTrie := out.Tries["data"]
	assert.Equal(t, int64(1), dataTrie.Nodes["leaf"])
	assert.Equal(t, int64(autoBalanceMeta), dataTrie.LeafMetadataEst)
	assert.Equal(t, int64(1), dataTrie.LeafVersions[1])
	_, hasPeer := out.Tries["peer"]
	assert.False(t, hasPeer)
	assert.Equal(t, int64(2), out.Accounts["accounts"])
	assert.Equal(t, int64(1), out.Accounts["withCode"])
	assert.Equal(t, int64(1), out.Accounts["withDataTrie"])
	assert.Equal(t, int64(1), out.Accounts["otherLeaves"])
	assert.Equal(t, int64(1), out.Accounts["codeEntries"])
	assert.Equal(t, int64(450), out.Accounts["codeBytes"])
	assert.Equal(t, int64(2), out.Accounts["codeReferences"])

	Close()
	rows := readGzipCsv(t, filepath.Join(dir, "state-analysis", "shard_0", "snapshots.csv.gz"))
	require.Len(t, rows, 2)
	header := rows[0]
	assert.Equal(t, "7", rows[1][column(header, "epoch")])
	assert.Equal(t, "0102", rows[1][column(header, "root_hash")])
	assert.Equal(t, "12", rows[1][column(header, "duration_s")])
	assert.Equal(t, "3", rows[1][column(header, "main_nodes")])
	assert.Equal(t, "1", rows[1][column(header, "data_tries")])
	assert.Equal(t, "false", rows[1][column(header, "had_error")])

	// structure stats are reset after completion
	assert.Equal(t, int64(0), sumNodes(global.structure[UnitAccounts].kinds[KindMain]))
	// peer stats are still pending
	assert.Equal(t, int64(1), sumNodes(global.structure[UnitPeer].kinds[KindPeer]))
}

func TestUnitFromIdentifier(t *testing.T) {
	assert.Equal(t, UnitAccounts, UnitFromIdentifier("AccountsTrie"))
	assert.Equal(t, UnitAccounts, UnitFromIdentifier("/db/1/Epoch_3/Shard_0/AccountsTrie"))
	assert.Equal(t, UnitPeer, UnitFromIdentifier("PeerAccountsTrie_0"))
	assert.Equal(t, UnitAccounts, UnitFromIdentifier("UserAccountsUnit"))
	assert.Equal(t, UnitPeer, UnitFromIdentifier("PeerAccountsUnit"))
	assert.Equal(t, UnitOther, UnitFromIdentifier("/db/1/Epoch_3/Shard_0/MiniBlocks"))
}

func TestWrapPersister(t *testing.T) {
	assert.Nil(t, WrapPersister(nil, "AccountsTrie"))
	var _ common.TrieStatisticsHandler = &snapshotStats{}

	p := &persisterStub{}
	assert.Equal(t, p, WrapPersister(p, "AccountsTrie")) // disabled: unchanged

	require.Nil(t, Init(Config{Enabled: true, OutputDir: "out"}, 0, t.TempDir()))
	defer func() {
		Close()
		global = nil
	}()
	assert.Equal(t, p, WrapPersister(p, "/db/1/Epoch_3/Shard_0/MiniBlocks")) // other units are not timed
	wrapped := WrapPersister(p, "/db/1/Epoch_3/Shard_0/AccountsTrie")
	assert.NotEqual(t, p, wrapped)
	_, _ = wrapped.Get([]byte("k"))
	assert.Equal(t, int64(1), global.counters[CStDbGet.Unit(UnitAccounts)].Load())

	// commit time exclusion only applies while a block is open
	ExcludeFromBlockCommit(10)
	assert.Equal(t, int64(0), global.counters[CBlockCommitNs].Load())
	BeginBlock(0, 0, 1, 1, 0, false)
	Add(CBlockCommitNs, 100)
	ExcludeFromBlockCommit(10)
	assert.Equal(t, int64(90), global.counters[CBlockCommitNs].Load())
	EndBlock()
}

func TestProfiler_EpochZeroAndNoRetry(t *testing.T) {
	dir := t.TempDir()
	p := &profiler{dir: dir, everyNEpochs: 50, numBlocks: 1, lastEpoch: -1}

	p.maybeStart(0, 1)
	require.NotNil(t, p.cpuFile, "epoch 0 must be profiled")
	p.blockDone(0, 1)
	assert.Nil(t, p.cpuFile)
	p.maybeStart(0, 2)
	assert.Nil(t, p.cpuFile, "epoch 0 must not be profiled twice")
	p.maybeStart(1, 3)
	assert.Nil(t, p.cpuFile)

	// a failed start (profiler already running elsewhere) is not retried and leaves no empty file
	other, err := os.CreateTemp(dir, "other")
	require.Nil(t, err)
	require.Nil(t, pprof.StartCPUProfile(other))
	p.maybeStart(50, 4)
	assert.Nil(t, p.cpuFile)
	assert.Equal(t, int64(50), p.lastEpoch)
	pprof.StopCPUProfile()
	_ = other.Close()
	p.maybeStart(50, 5)
	assert.Nil(t, p.cpuFile)
	entries, err := os.ReadDir(dir)
	require.Nil(t, err)
	for _, e := range entries {
		assert.NotContains(t, e.Name(), "cpu_epoch_50")
	}
}

type persisterStub struct{}

func (p *persisterStub) Put(_, _ []byte) error                         { return nil }
func (p *persisterStub) Get(_ []byte) ([]byte, error)                  { return nil, nil }
func (p *persisterStub) Has(_ []byte) error                            { return nil }
func (p *persisterStub) Close() error                                  { return nil }
func (p *persisterStub) Remove(_ []byte) error                         { return nil }
func (p *persisterStub) Destroy() error                                { return nil }
func (p *persisterStub) DestroyClosed() error                          { return nil }
func (p *persisterStub) RangeKeys(_ func(key []byte, val []byte) bool) {}
func (p *persisterStub) IsInterfaceNil() bool                          { return p == nil }

type statsStub struct {
	common.TrieStatisticsHandler
	numNodes uint64
	size     uint64
	depth    uint32
}

func newStatsStub() *statsStub {
	return &statsStub{numNodes: 3, size: 340, depth: 2}
}

// GetTotalNodesSize -
func (s *statsStub) GetTotalNodesSize() uint64 { return s.size }

// GetTotalNumNodes -
func (s *statsStub) GetTotalNumNodes() uint64 { return s.numNodes }

// GetMaxTrieDepth -
func (s *statsStub) GetMaxTrieDepth() uint32 { return s.depth }

// IsInterfaceNil -
func (s *statsStub) IsInterfaceNil() bool { return s == nil }
