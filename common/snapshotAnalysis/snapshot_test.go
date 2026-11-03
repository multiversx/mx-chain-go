package snapshotAnalysis

import (
	"compress/gzip"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readCsvGz(t *testing.T, path string) [][]string {
	f, err := os.Open(path)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	require.NoError(t, err)
	gz.Multistream(true)
	data, err := io.ReadAll(gz)
	require.NoError(t, err)
	rows, err := csv.NewReader(stringsReader(string(data))).ReadAll()
	require.NoError(t, err)
	return rows
}

type stringReader struct {
	s string
	i int
}

func stringsReader(s string) io.Reader { return &stringReader{s: s} }

func (r *stringReader) Read(p []byte) (int, error) {
	if r.i >= len(r.s) {
		return 0, io.EOF
	}
	n := copy(p, r.s[r.i:])
	r.i += n
	return n, nil
}

func TestHeadersAndRowsAreAligned(t *testing.T) {
	t.Parallel()

	s := newSnapshot(nil, 1, "UserAccountsUnit", 3, []byte("root"), true, nil)
	row := (*collector)(nil).snapshotRow(s, time.Now())
	assert.Equal(t, len(row.names), len(row.vals))
	assert.Equal(t, snapshotHeader(), row.names)

	var st timelineState
	tl := (*collector)(nil).timelineRow(s, st, st, 0)
	assert.Equal(t, len(tl.names), len(tl.vals))
	assert.Equal(t, timelineHeader(), tl.names)

	g := &GoroutineMeter{snap: s}
	tr := (*collector)(nil).trieRow(s, g, nil, 0)
	assert.Equal(t, len(tr.names), len(tr.vals))
	assert.Equal(t, triesHeader(), tr.names)
}

func TestDisabledHooksAreNoOps(t *testing.T) {
	// not parallel: touches the global state
	Close()

	assert.False(t, Enabled())
	assert.Nil(t, SnapshotRequested("UserAccountsUnit", 1, []byte("root"), false))
	assert.Nil(t, Current("UserAccountsUnit"))

	var s *Snapshot
	s.EpochReady()
	s.Started()
	s.Leaf(time.Second, time.Second, LeafAccount)
	s.EnqueueWait(time.Second)
	s.Completed(false)
	s.StallReleased(time.Second)
	m := s.GoroutineStarted("", 0)
	assert.Nil(t, m)
	m.Get(time.Second, 10, 0)
	m.Put(time.Second, 10)
	m.Node([]byte("hash"), NodeLeaf, 10, 1)
	m.Finish(nil)

	var sm *StorerMeter
	sm.ProcGet(time.Second)
	sm.SnapPut(time.Second, time.Second, 10)
	assert.False(t, sm.Active())

	NodeIdleChanged(false)
	AccountsCommitted("UserAccountsUnit", time.Second)
	PruneBuffered("UserAccountsUnit")
	PruneResolved("UserAccountsUnit", 1, time.Second)
	PruneRemoved("UserAccountsUnit", 1, time.Second)
	EpochRootHashWalk(time.Second)
}

func TestSnapshotLifecycleWritesRows(t *testing.T) {
	// not parallel: touches the global state
	dir := t.TempDir()
	cfg := Config{
		Enabled:              true,
		Label:                "test",
		OutputDir:            dir,
		SampleIntervalMs:     20,
		AfterSnapshotSeconds: 0,
		NodeHashSampleShift:  1,
	}
	require.NoError(t, Init(cfg, 1, dir))
	defer Close()
	assert.True(t, Enabled())

	storer := NewStorerMeter("AccountsTrie")
	storer.SetCacheProbe(func() (int, int64, int) { return 5, 500, 2 })
	LinkStorer("UserAccountsUnit", storer)

	s := SnapshotRequested("UserAccountsUnit", 7, []byte("root hash of the trie"), true)
	require.NotNil(t, s)
	assert.Equal(t, s, Current("UserAccountsUnit"))

	NodeIdleChanged(false)
	time.Sleep(5 * time.Millisecond)
	NodeIdleChanged(true)

	s.EpochReady()
	s.Checked()
	s.SetChannelProbes(func() int { return 3 }, func() int { return 0 })
	s.SetQueueProbe(func() int { return 1 })
	s.Started()

	storer.SnapGet(time.Microsecond, 2, 50*time.Microsecond)
	storer.SnapPut(time.Microsecond, 200*time.Microsecond, 100)
	storer.ProcCacheHit()
	storer.ProcGet(10 * time.Microsecond)
	storer.ProcPut(10*time.Microsecond, 50)
	AccountsCommitted("UserAccountsUnit", 3*time.Millisecond)
	PruneBuffered("UserAccountsUnit")

	main := s.GoroutineStarted("", time.Millisecond)
	require.NotNil(t, main)
	hashA := []byte("0123456789abcdef0123456789abcdef")
	hashB := []byte("fedcba9876543210fedcba9876543210")
	for i := 0; i < 3*flushEveryNodes; i++ {
		h := hashA
		if i%2 == 0 {
			h = hashB
		}
		main.Get(10*time.Microsecond, 100, i%3)
		main.Decode(time.Microsecond)
		main.Node(h, NodeBranch, 100, i%5)
		main.Put(20*time.Microsecond, 100)
	}
	main.Get(time.Microsecond, 100, DistCache)
	main.GetMiss(time.Microsecond)
	main.PrevPut(time.Microsecond, 100)
	main.ParkEnter()
	main.ParkExit(time.Millisecond)
	main.LeafSent(time.Millisecond)
	s.Leaf(time.Microsecond, time.Microsecond, LeafAccountWithDataTrie)
	s.Leaf(time.Microsecond, time.Microsecond, LeafCode)
	s.EnqueueWait(time.Microsecond)
	s.ThrottlerWait(time.Microsecond)
	s.EmptyTrieSkipped()
	main.Finish([]byte("root hash of the trie"))

	dataRoot := []byte("data trie root hash 0123456789ab")
	for i := 0; i < 2; i++ {
		data := s.GoroutineStarted("erd1address", 0)
		for j := 0; j < largeTrieMinNodes+1; j++ {
			data.Node(hashA, NodeLeaf, 10, 2)
		}
		data.Finish(dataRoot)
	}
	s.LeavesConsumed()
	s.MissingNode()
	s.MissingSynced(time.Millisecond)

	time.Sleep(60 * time.Millisecond) // a few timeline samples
	s.TriesDone()
	s.SyncDone()
	s.StallReleased(2 * time.Millisecond)
	s.Completed(false)
	assert.Nil(t, Current("UserAccountsUnit"))

	PruneResolved("UserAccountsUnit", 1, time.Millisecond)
	PruneRemoved("UserAccountsUnit", 10, time.Millisecond)

	// the after-window is 0 so the next tick writes the snapshot row; a new snapshot for the same identifier
	// must also finish the previous one
	time.Sleep(60 * time.Millisecond)
	s2 := SnapshotRequested("UserAccountsUnit", 8, []byte("another root"), false)
	require.NotNil(t, s2)
	s2.Skipped("trie already synced")
	Close()

	shardDir := filepath.Join(dir, "shard_1")
	snapshots := readCsvGz(t, filepath.Join(shardDir, "snapshots.csv.gz"))
	require.GreaterOrEqual(t, len(snapshots), 3, "header + 2 rows")
	header := snapshots[0]
	assert.Equal(t, snapshotHeader(), header)
	first := map[string]string{}
	for i, name := range header {
		first[name] = snapshots[1][i]
	}
	assert.Equal(t, "UserAccountsUnit", first["identifier"])
	assert.Equal(t, "7", first["epoch"])
	assert.Equal(t, "1", first["import_db"])
	assert.Equal(t, "0", first["skipped"])
	assert.Equal(t, "3072", first["main_nodes"])
	assert.Equal(t, "2002", first["data_nodes"])
	assert.Equal(t, "3", first["tries"])
	assert.Equal(t, "2", first["data_tries"])
	assert.Equal(t, "1", first["dup_data_tries"])
	assert.Equal(t, "1001", first["dup_data_trie_nodes"])
	assert.Equal(t, "1024", first["put_redundant_same_epoch"])
	assert.Equal(t, "1", first["get_cache_hits"])
	assert.Equal(t, "1", first["get_missing"])
	assert.Equal(t, "1", first["prev_epoch_puts"])
	assert.Equal(t, "1", first["parks"])
	assert.Equal(t, "1", first["missing_nodes"])
	assert.Equal(t, "2", first["main_leaves"])
	assert.Equal(t, "1", first["code_leaves"])
	assert.Equal(t, "1", first["accounts_with_data_trie"])
	assert.Equal(t, "1", first["empty_tries_skipped"])
	assert.Equal(t, "1", first["commits"])
	assert.Equal(t, "1", first["prune_buffered"])
	assert.Equal(t, "10", first["prune_removed"])
	assert.Equal(t, "1", first["st_snap_gets"])
	assert.Equal(t, "1", first["st_snap_puts"])
	assert.Equal(t, "2", first["st_proc_gets"])
	assert.Equal(t, "1", first["st_proc_cache_hits"])
	assert.Equal(t, "5", first["cache_entries_end"])
	assert.NotEqual(t, "0", first["stall_ns"])
	assert.NotEqual(t, "0", first["main_trie_ns"])
	assert.NotEqual(t, "0", first["node_busy_ns"])
	// every sampled hash is one of two values, so the sampled duplicate ratio is high
	assert.NotEqual(t, "0", first["sampled_nodes"])
	assert.Equal(t, "2", first["sampled_distinct"])

	second := map[string]string{}
	for i, name := range header {
		second[name] = snapshots[2][i]
	}
	assert.Equal(t, "8", second["epoch"])
	assert.Equal(t, "1", second["skipped"])
	assert.Equal(t, "trie already synced", second["skip_reason"])

	timeline := readCsvGz(t, filepath.Join(shardDir, "timeline.csv.gz"))
	require.Greater(t, len(timeline), 2)
	assert.Equal(t, timelineHeader(), timeline[0])
	phases := map[string]bool{}
	for _, r := range timeline[1:] {
		phases[r[indexOf(timeline[0], "phase")]] = true
	}
	assert.True(t, phases["after"] || phases["data_tries"] || phases["finishing"], phases)

	tries := readCsvGz(t, filepath.Join(shardDir, "tries.csv.gz"))
	require.Equal(t, 4, len(tries), "header + main trie + 2 large data tries")
	assert.Equal(t, "1", tries[1][indexOf(tries[0], "main")])
	assert.Equal(t, "erd1address", tries[2][indexOf(tries[0], "address")])
}

func indexOf(header []string, name string) int {
	for i, h := range header {
		if h == name {
			return i
		}
	}
	return -1
}

func TestLatencyBucket(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, latencyBucket(50*time.Microsecond))
	assert.Equal(t, 1, latencyBucket(500*time.Microsecond))
	assert.Equal(t, 2, latencyBucket(5*time.Millisecond))
	assert.Equal(t, 3, latencyBucket(50*time.Millisecond))
	assert.Equal(t, 4, latencyBucket(time.Second))
}

func TestBusyTracker(t *testing.T) {
	t.Parallel()

	var b busyTracker
	start := time.Now()
	b.set(true, start)
	b.set(true, start.Add(time.Second)) // no-op, same state
	assert.Equal(t, 2*time.Second, b.totalAt(start.Add(2*time.Second)))
	b.set(false, start.Add(3*time.Second))
	assert.Equal(t, 3*time.Second, b.totalAt(start.Add(10*time.Second)))
}
