package snapshotAnalysis

import (
	"encoding/hex"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
)

// rowBuilder collects (name, value) pairs so that the header and the row are always aligned
type rowBuilder struct {
	names []string
	vals  []string
}

func (b *rowBuilder) str(name string, v string) {
	b.names = append(b.names, name)
	b.vals = append(b.vals, v)
}

func (b *rowBuilder) i64(name string, v int64) {
	b.str(name, strconv.FormatInt(v, 10))
}

func (b *rowBuilder) u64(name string, v uint64) {
	b.str(name, strconv.FormatUint(v, 10))
}

func (b *rowBuilder) f(name string, v float64) {
	b.str(name, strconv.FormatFloat(v, 'f', 3, 64))
}

func (b *rowBuilder) dur(name string, d time.Duration) {
	b.i64(name, int64(d))
}

func (b *rowBuilder) hist(prefix string, h [numLatencyBuckets]int64) {
	for i := range h {
		b.i64(prefix+"_"+latencyBucketNames[i], h[i])
	}
}

func (b *rowBuilder) storerValues(prefix string, v storerValues) {
	b.i64(prefix+"snap_gets", v.snapGets)
	b.i64(prefix+"snap_cache_hits", v.snapCacheHits)
	b.i64(prefix+"snap_get_rlock_ns", v.snapGetRLockNs)
	b.i64(prefix+"snap_get_probes", v.snapGetProbes)
	b.i64(prefix+"snap_get_probe_ns", v.snapGetProbeNs)
	b.i64(prefix+"snap_get_max_ns", v.snapGetMaxNs)
	b.hist(prefix+"snap_get", v.snapGetHist)
	b.i64(prefix+"snap_puts", v.snapPuts)
	b.i64(prefix+"snap_put_bytes", v.snapPutBytes)
	b.i64(prefix+"snap_put_lock_ns", v.snapPutLockNs)
	b.i64(prefix+"snap_put_lock_max_ns", v.snapPutLockMax)
	b.i64(prefix+"snap_put_ns", v.snapPutNs)
	b.i64(prefix+"snap_put_max_ns", v.snapPutMaxNs)
	b.hist(prefix+"snap_put", v.snapPutHist)
	b.i64(prefix+"snap_put_no_epoch", v.snapPutNoEpoch)
	b.i64(prefix+"proc_gets", v.procGets)
	b.i64(prefix+"proc_cache_hits", v.procCacheHits)
	b.i64(prefix+"proc_get_ns", v.procGetNs)
	b.i64(prefix+"proc_get_max_ns", v.procGetMaxNs)
	b.hist(prefix+"proc_get", v.procGetHist)
	b.i64(prefix+"proc_puts", v.procPuts)
	b.i64(prefix+"proc_put_bytes", v.procPutBytes)
	b.i64(prefix+"proc_put_ns", v.procPutNs)
	b.i64(prefix+"proc_put_max_ns", v.procPutMaxNs)
	b.hist(prefix+"proc_put", v.procPutHist)
	b.i64(prefix+"proc_removes", v.procRemoves)
	b.i64(prefix+"proc_remove_ns", v.procRemoveNs)
	b.i64(prefix+"epoch_changes", v.epochChanges)
	b.i64(prefix+"epoch_change_ns", v.epochChangeNs)
	b.i64(prefix+"epoch_change_lock_ns", v.epochChangeLock)
}

func (c *collector) identity(b *rowBuilder, s *Snapshot) {
	shard := uint32(0)
	label := ""
	if c != nil {
		shard = c.shardID
		label = c.cfg.Label
	}
	b.u64("shard", uint64(shard))
	b.str("label", label)
	b.str("identifier", s.identifier)
	b.u64("epoch", uint64(s.epoch))
	b.u64("seq", s.seq)
}

func (s *Snapshot) phase(now time.Time) string {
	switch {
	case s.skipped.Load():
		return "skipped"
	case s.completed.Load():
		return "after"
	case s.marks[markTriesDone].Load() != 0:
		return "finishing"
	case s.marks[markMainDone].Load() != 0:
		return "data_tries"
	case s.marks[markStarted].Load() != 0:
		return "main_trie"
	case s.marks[markChecked].Load() != 0:
		return "starting"
	case s.marks[markEpochReady].Load() != 0:
		return "checking"
	default:
		_ = now
		return "waiting_epoch"
	}
}

// snapshotRow builds the one-line summary of a snapshot, written when its after-window ends
func (c *collector) snapshotRow(s *Snapshot, now time.Time) *rowBuilder {
	b := &rowBuilder{}
	c.identity(b, s)
	b.str("root_hash", s.rootHashHex())
	b.i64("requested_unix_ms", s.created.UnixMilli())
	b.str("import_db", boolStr(s.importDb))
	b.str("skipped", boolStr(s.skipped.Load()))
	reason := s.skipReason.Load()
	if reason == nil {
		b.str("skip_reason", "")
	} else {
		b.str("skip_reason", *reason)
	}
	b.str("error", boolStr(s.errored.Load()))

	// phases
	b.dur("wait_epoch_ns", s.between(markRequested, markEpochReady))
	b.dur("check_ns", s.between(markEpochReady, markChecked))
	b.dur("to_first_goroutine_ns", s.between(markStarted, markFirstGoroutine))
	b.dur("main_trie_ns", s.between(markFirstGoroutine, markMainDone))
	b.dur("leaves_consumed_at_ns", s.between(markStarted, markLeavesConsumed))
	b.dur("data_tries_tail_ns", s.between(markMainDone, markTriesDone))
	b.dur("traversal_ns", s.between(markStarted, markTriesDone))
	b.dur("sync_wait_ns", s.between(markTriesDone, markSyncDone))
	b.dur("completion_ns", s.between(markSyncDone, markCompleted))
	b.dur("total_ns", s.between(markRequested, markCompleted))
	b.dur("stall_ns", time.Duration(s.stalledNs.Load()))

	// traversal
	mainNodes := s.mainNodes.Load()
	dataNodes := s.dataNodes.Load()
	nodes := mainNodes + dataNodes
	b.i64("nodes", nodes)
	b.i64("bytes", s.mainBytes.Load()+s.dataBytes.Load())
	b.i64("main_nodes", mainNodes)
	b.i64("main_bytes", s.mainBytes.Load())
	b.i64("main_max_depth", s.maxDepthMain.Load())
	b.i64("data_nodes", dataNodes)
	b.i64("data_bytes", s.dataBytes.Load())
	b.i64("data_max_depth", s.maxDepthData.Load())
	b.i64("branch_nodes", s.nodeTypes[NodeBranch].Load())
	b.i64("extension_nodes", s.nodeTypes[NodeExtension].Load())
	b.i64("leaf_nodes", s.nodeTypes[NodeLeaf].Load())
	b.i64("tries", s.tries.Load())
	b.i64("data_tries", s.triesData.Load())
	b.i64("empty_tries_skipped", s.triesEmpty.Load())

	// leaves consumer
	b.i64("main_leaves", s.leaves.Load())
	b.i64("code_leaves", s.leavesCode.Load())
	b.i64("accounts_with_data_trie", s.leavesWithDataTrie.Load())
	b.i64("leaves_recv_wait_ns", s.leavesRecvWaitNs.Load())
	b.i64("leaves_unmarshal_ns", s.leavesUnmarshalNs.Load())
	b.i64("leaves_send_wait_ns", s.leavesSendNs.Load())

	// duplicates
	s.mutSets.Lock()
	sampledTotal, sampledDistinct := s.sampledTotal, s.sampledDistinct
	dupTries, dupNodes, dupBytes, capped := s.dupTries, s.dupTrieNodes, s.dupTrieBytes, s.rootsCapped
	s.mutSets.Unlock()
	b.i64("sampled_nodes", sampledTotal)
	b.i64("sampled_distinct", sampledDistinct)
	dupRatio := 0.0
	if sampledTotal > 0 {
		dupRatio = float64(sampledTotal-sampledDistinct) / float64(sampledTotal)
	}
	b.f("dup_node_ratio", dupRatio)
	b.i64("dup_data_tries", dupTries)
	b.i64("dup_data_trie_nodes", dupNodes)
	b.i64("dup_data_trie_bytes", dupBytes)
	b.str("data_roots_capped", boolStr(capped))

	// reads (trie level)
	b.i64("get_ns", s.getNs.Load())
	b.i64("get_max_ns", s.getMaxNs.Load())
	b.i64("get_bytes", s.getBytes.Load())
	b.i64("get_cache_hits", s.getCacheHits.Load())
	b.i64("get_found_epoch0", s.getDist[0].Load())
	b.i64("get_found_epoch1", s.getDist[1].Load())
	b.i64("get_found_epoch2", s.getDist[2].Load())
	b.i64("get_found_older", s.getDist[3].Load())
	b.i64("get_missing", s.getMiss.Load())
	b.hist("get", s.getHist.load())

	// writes (trie level)
	b.i64("puts", s.puts.Load())
	b.i64("put_ns", s.putNs.Load())
	b.i64("put_max_ns", s.putMaxNs.Load())
	b.i64("put_bytes", s.putBytes.Load())
	// a node found in the snapshot epoch persister is written into it again; cache hits are of unknown epoch
	b.i64("put_redundant_same_epoch", s.getDist[0].Load())
	b.hist("put", s.putHist.load())
	b.i64("prev_epoch_puts", s.prevPuts.Load())
	b.i64("prev_epoch_put_bytes", s.prevPutBytes.Load())
	b.i64("prev_epoch_put_ns", s.prevPutNs.Load())
	b.i64("decode_ns", s.decodeNs.Load())

	// goroutines
	grWall := s.grWallNs.Load()
	b.i64("goroutines", s.grStarted.Load())
	b.i64("goroutine_wall_ns", grWall)
	b.i64("goroutine_max_active", s.grMaxActive.Load())
	b.i64("parked_ns", s.parkedNs.Load())
	b.i64("parks", s.parks.Load())
	b.i64("queue_latency_ns", s.queueLatencyNs.Load())
	b.i64("queue_latency_max_ns", s.queueLatencyMax.Load())
	b.i64("enqueue_waits", s.enqueueWaits.Load())
	b.i64("enqueue_wait_ns", s.enqueueWaitNs.Load())
	b.i64("throttler_waits", s.throttlerWaits.Load())
	b.i64("throttler_wait_ns", s.throttlerWaitNs.Load())
	ioNs := s.getNs.Load() + s.putNs.Load() + s.prevPutNs.Load()
	b.i64("goroutine_io_ns", ioNs)
	other := grWall - ioNs - s.decodeNs.Load() - s.parkedNs.Load() - s.leavesSendNs.Load()
	b.i64("goroutine_other_ns", other)

	// missing nodes
	b.i64("missing_nodes", s.missing.Load())
	b.i64("missing_sync_ns", s.syncNs.Load())

	// processing side during the whole window (snapshot + after-window)
	busy := time.Duration(0)
	cpu := time.Duration(0)
	var commit commitValues
	var prune pruneValues
	var storer storerValues
	walks, walkNs := int64(0), int64(0)
	if c != nil {
		busy = c.busy.totalAt(now) - s.baseBusy
		cpu = processCPUTime() - s.baseCPU
		commit = c.commitMeter(s.identifier).values().sub(s.baseCommit)
		prune = c.pruneMeter(s.identifier).values().sub(s.basePrune)
		storer = s.storerDelta()
		walks = c.epochRootHashWalks.Load() - s.baseWalks
		walkNs = c.epochRootHashWalkNs.Load() - s.baseWalkNs
	}
	window := now.Sub(s.created)
	b.dur("window_ns", window)
	b.dur("node_busy_ns", busy)
	b.dur("process_cpu_ns", cpu)
	b.i64("commits", commit.count)
	b.i64("commit_ns", commit.ns)
	b.i64("commit_max_ns", commit.maxNs)
	b.i64("epoch_root_hash_walks", walks)
	b.i64("epoch_root_hash_walk_ns", walkNs)
	b.i64("prune_buffered", prune.buffered)
	b.i64("prune_resolves", prune.resolves)
	b.i64("prune_resolve_ops", prune.resolveOps)
	b.i64("prune_resolve_ns", prune.resolveNs)
	b.i64("prunes", prune.prunes)
	b.i64("prune_removed", prune.removed)
	b.i64("prune_remove_ns", prune.removeNs)
	b.storerValues("st_", storer)
	entries, cacheBytes, persisters := s.storer.Load().probeCache()
	b.i64("cache_entries_end", int64(entries))
	b.i64("cache_bytes_end", cacheBytes)
	b.i64("active_persisters_end", int64(persisters))

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	b.u64("heap_alloc_start", s.baseHeap)
	b.u64("heap_alloc_end", ms.HeapAlloc)
	b.u64("heap_sys_end", ms.HeapSys)
	b.u64("gc_count", uint64(ms.NumGC-s.baseGC))

	return b
}

func snapshotHeader() []string {
	return (*collector)(nil).snapshotRow(newSnapshot(nil, 0, "", 0, nil, false, nil), time.Now()).names
}

// timelineState is what the sampler keeps between two samples to compute deltas
type timelineState struct {
	at         time.Time
	mainNodes  int64
	dataNodes  int64
	bytes      int64
	tries      int64
	leaves     int64
	missing    int64
	getNs      int64
	putNs      int64
	decodeNs   int64
	parkedNs   int64
	parks      int64
	leafSendNs int64
	leafRecvNs int64
	enqueueNs  int64
	throttleNs int64
	queueLatNs int64
	grStarted  int64
	grWallNs   int64
	busy       time.Duration
	cpu        time.Duration
	commit     commitValues
	prune      pruneValues
	storer     storerValues
	gc         uint32
	gcPause    uint64
}

func (s *Snapshot) timelineState(now time.Time) timelineState {
	st := timelineState{
		at:         now,
		mainNodes:  s.mainNodes.Load(),
		dataNodes:  s.dataNodes.Load(),
		bytes:      s.mainBytes.Load() + s.dataBytes.Load(),
		tries:      s.tries.Load(),
		leaves:     s.leaves.Load(),
		missing:    s.missing.Load(),
		getNs:      s.getNs.Load(),
		putNs:      s.putNs.Load() + s.prevPutNs.Load(),
		decodeNs:   s.decodeNs.Load(),
		parkedNs:   s.parkedNs.Load(),
		parks:      s.parks.Load(),
		leafSendNs: s.leavesSendNs.Load(),
		leafRecvNs: s.leavesRecvWaitNs.Load(),
		enqueueNs:  s.enqueueWaitNs.Load(),
		throttleNs: s.throttlerWaitNs.Load(),
		queueLatNs: s.queueLatencyNs.Load(),
		grStarted:  s.grStarted.Load(),
		grWallNs:   s.grWallNs.Load(),
	}
	if s.c != nil {
		st.busy = s.c.busy.totalAt(now)
		st.cpu = processCPUTime()
		st.commit = s.c.commitMeter(s.identifier).values()
		st.prune = s.c.pruneMeter(s.identifier).values()
		st.storer = s.storer.Load().values()
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	st.gc = ms.NumGC
	st.gcPause = ms.PauseTotalNs
	return st
}

func pct(part time.Duration, whole time.Duration) float64 {
	if whole <= 0 {
		return 0
	}
	return 100 * float64(part) / float64(whole)
}

// timelineRow builds one sample of the snapshot progress; every *_d column is the delta since the previous sample
func (c *collector) timelineRow(s *Snapshot, prev timelineState, cur timelineState, heapAlloc uint64) *rowBuilder {
	b := &rowBuilder{}
	c.identity(b, s)
	interval := cur.at.Sub(prev.at)
	b.i64("ts_unix_ms", cur.at.UnixMilli())
	b.i64("elapsed_ms", cur.at.Sub(s.created).Milliseconds())
	b.str("phase", s.phase(cur.at))
	b.i64("interval_ms", interval.Milliseconds())

	b.i64("main_nodes_d", cur.mainNodes-prev.mainNodes)
	b.i64("data_nodes_d", cur.dataNodes-prev.dataNodes)
	b.i64("bytes_d", cur.bytes-prev.bytes)
	b.i64("tries_d", cur.tries-prev.tries)
	b.i64("goroutines_started_d", cur.grStarted-prev.grStarted)
	b.i64("main_leaves_d", cur.leaves-prev.leaves)
	b.i64("missing_d", cur.missing-prev.missing)

	leavesLen, missingLen, queueLen := s.probes()
	b.i64("goroutines_active", s.grActive.Load())
	b.i64("goroutines_parked", s.grParked.Load())
	b.i64("queue_len", int64(queueLen))
	b.i64("leaves_chan_len", int64(leavesLen))
	b.i64("missing_chan_len", int64(missingLen))

	b.f("node_busy_pct", pct(cur.busy-prev.busy, interval))
	b.f("process_cpu_cores", float64(cur.cpu-prev.cpu)/float64(max(interval, time.Nanosecond)))
	b.u64("heap_alloc_mb", heapAlloc/1024/1024)
	b.i64("goroutines_total", int64(runtime.NumGoroutine()))
	b.i64("gc_d", int64(cur.gc-prev.gc))
	b.i64("gc_pause_ns_d", int64(cur.gcPause-prev.gcPause))

	b.i64("goroutine_wall_ns_d", cur.grWallNs-prev.grWallNs)
	b.i64("get_ns_d", cur.getNs-prev.getNs)
	b.i64("put_ns_d", cur.putNs-prev.putNs)
	b.i64("decode_ns_d", cur.decodeNs-prev.decodeNs)
	b.i64("parked_ns_d", cur.parkedNs-prev.parkedNs)
	b.i64("parks_d", cur.parks-prev.parks)
	b.i64("leaves_send_wait_ns_d", cur.leafSendNs-prev.leafSendNs)
	b.i64("leaves_recv_wait_ns_d", cur.leafRecvNs-prev.leafRecvNs)
	b.i64("enqueue_wait_ns_d", cur.enqueueNs-prev.enqueueNs)
	b.i64("throttler_wait_ns_d", cur.throttleNs-prev.throttleNs)
	b.i64("queue_latency_ns_d", cur.queueLatNs-prev.queueLatNs)

	commit := cur.commit.sub(prev.commit)
	b.i64("commits_d", commit.count)
	b.i64("commit_ns_d", commit.ns)
	prune := cur.prune.sub(prev.prune)
	b.i64("prune_buffered_d", prune.buffered)
	b.i64("prune_resolve_ops_d", prune.resolveOps)
	b.i64("prune_resolve_ns_d", prune.resolveNs)
	b.i64("prune_removed_d", prune.removed)
	b.i64("prune_remove_ns_d", prune.removeNs)
	bufferLen, ewlLen := int64(0), int64(0)
	if c != nil {
		bl, el := c.pruneMeter(s.identifier).probeLens()
		bufferLen, ewlLen = int64(bl), int64(el)
	}
	b.i64("prune_buffer_len", bufferLen)
	b.i64("eviction_waiting_list_len", ewlLen)

	st := cur.storer.sub(prev.storer)
	b.i64("st_snap_gets_d", st.snapGets)
	b.i64("st_snap_cache_hits_d", st.snapCacheHits)
	b.i64("st_snap_get_rlock_ns_d", st.snapGetRLockNs)
	b.i64("st_snap_get_probes_d", st.snapGetProbes)
	b.i64("st_snap_get_probe_ns_d", st.snapGetProbeNs)
	b.i64("st_snap_puts_d", st.snapPuts)
	b.i64("st_snap_put_lock_ns_d", st.snapPutLockNs)
	b.i64("st_snap_put_ns_d", st.snapPutNs)
	b.i64("st_snap_put_gt100ms_d", st.snapPutHist[numLatencyBuckets-1])
	b.i64("st_snap_get_gt100ms_d", st.snapGetHist[numLatencyBuckets-1])
	b.i64("st_proc_gets_d", st.procGets)
	b.i64("st_proc_cache_hits_d", st.procCacheHits)
	b.i64("st_proc_get_ns_d", st.procGetNs)
	b.i64("st_proc_puts_d", st.procPuts)
	b.i64("st_proc_put_ns_d", st.procPutNs)
	b.i64("st_proc_removes_d", st.procRemoves)
	b.i64("st_proc_remove_ns_d", st.procRemoveNs)
	b.i64("st_epoch_changes_d", st.epochChanges)
	entries, cacheBytes, persisters := s.storer.Load().probeCache()
	b.i64("cache_entries", int64(entries))
	b.i64("cache_bytes", cacheBytes)
	b.i64("active_persisters", int64(persisters))

	return b
}

func timelineHeader() []string {
	s := newSnapshot(nil, 0, "", 0, nil, false, nil)
	var st timelineState
	return (*collector)(nil).timelineRow(s, st, st, 0).names
}

// trieRow describes one traversed trie: the main trie always, data tries above largeTrieMinNodes
func (c *collector) trieRow(s *Snapshot, g *GoroutineMeter, rootHash []byte, wall time.Duration) *rowBuilder {
	b := &rowBuilder{}
	c.identity(b, s)
	b.str("address", g.address)
	b.str("root_hash", hex.EncodeToString(rootHash))
	b.str("main", boolStr(g.isMain))
	b.i64("started_at_ms", g.start.Sub(s.created).Milliseconds())
	b.dur("wall_ns", wall)
	b.i64("nodes", g.totalNodes)
	b.i64("bytes", g.totalBytes)
	b.i64("max_depth", g.totalDepth)
	b.i64("get_ns", g.totalGetNs)
	b.i64("put_ns", g.totalPutNs)
	b.i64("decode_ns", g.totalDecode)
	b.i64("parked_ns", g.totalParked)
	b.i64("leaves_send_wait_ns", g.totalLeafSnd)
	return b
}

func triesHeader() []string {
	s := newSnapshot(nil, 0, "", 0, nil, false, nil)
	g := &GoroutineMeter{snap: s}
	return (*collector)(nil).trieRow(s, g, nil, 0).names
}

func (c *collector) writeTrieRow(s *Snapshot, g *GoroutineMeter, rootHash []byte, wall time.Duration) {
	if c == nil {
		return
	}
	err := c.triesFile.write(c.trieRow(s, g, rootHash, wall).vals)
	if err != nil {
		log.Warn("snapshotAnalysis: cannot write trie row", "error", err)
	}
}

func boolStr(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

// logSummary prints the snapshot measurements that matter most at completion, so the log alone tells
// where the time went
func (s *Snapshot) logSummary() {
	now := time.Now()
	grWall := time.Duration(s.grWallNs.Load())
	getNs := time.Duration(s.getNs.Load())
	putNs := time.Duration(s.putNs.Load() + s.prevPutNs.Load())
	decodeNs := time.Duration(s.decodeNs.Load())
	parkedNs := time.Duration(s.parkedNs.Load())
	traversal := s.between(markStarted, markTriesDone)
	maxActive := s.grMaxActive.Load()
	utilization := 0.0
	if traversal > 0 && maxActive > 0 {
		utilization = float64(grWall) / float64(traversal) / float64(maxActive)
	}
	s.mutSets.Lock()
	dupRatio := 0.0
	if s.sampledTotal > 0 {
		dupRatio = float64(s.sampledTotal-s.sampledDistinct) / float64(s.sampledTotal)
	}
	dupTries := s.dupTries
	dupNodes := s.dupTrieNodes
	s.mutSets.Unlock()

	storer := s.storerDelta()
	busy := time.Duration(0)
	if s.c != nil {
		busy = s.c.busy.totalAt(now) - s.baseBusy
	}

	log.Info("snapshot analysis summary",
		"identifier", s.identifier,
		"epoch", s.epoch,
		"skipped", s.skipped.Load(),
		"error", s.errored.Load(),
		"total", s.between(markRequested, markCompleted).Truncate(time.Millisecond),
		"wait epoch", s.between(markRequested, markEpochReady).Truncate(time.Millisecond),
		"main trie", s.between(markFirstGoroutine, markMainDone).Truncate(time.Millisecond),
		"data tries tail", s.between(markMainDone, markTriesDone).Truncate(time.Millisecond),
		"stall", time.Duration(s.stalledNs.Load()).Truncate(time.Millisecond),
		"nodes", s.mainNodes.Load()+s.dataNodes.Load(),
		"bytes", core.ConvertBytes(uint64(s.mainBytes.Load()+s.dataBytes.Load())),
		"data tries", s.triesData.Load(),
		"dup data tries", dupTries,
		"dup data trie nodes", dupNodes,
		"dup node ratio (sampled)", fmt.Sprintf("%.4f", dupRatio),
		"redundant puts (found in snapshot epoch)", s.getDist[0].Load(),
		"prev epoch puts", s.prevPuts.Load(),
		"cache hits", s.getCacheHits.Load(),
		"missing nodes", s.missing.Load(),
		"goroutines", s.grStarted.Load(),
		"max active", maxActive,
		"utilization", fmt.Sprintf("%.3f", utilization),
		"goroutine wall", grWall.Truncate(time.Millisecond),
		"get", getNs.Truncate(time.Millisecond),
		"put", putNs.Truncate(time.Millisecond),
		"decode", decodeNs.Truncate(time.Millisecond),
		"parked", parkedNs.Truncate(time.Millisecond),
		"leaves send wait", time.Duration(s.leavesSendNs.Load()).Truncate(time.Millisecond),
		"leaves recv wait", time.Duration(s.leavesRecvWaitNs.Load()).Truncate(time.Millisecond),
		"throttler wait", time.Duration(s.throttlerWaitNs.Load()).Truncate(time.Millisecond),
		"storer get rlock wait", time.Duration(storer.snapGetRLockNs).Truncate(time.Millisecond),
		"storer put lock wait", time.Duration(storer.snapPutLockNs).Truncate(time.Millisecond),
		"storer put", time.Duration(storer.snapPutNs).Truncate(time.Millisecond),
		"node busy", busy.Truncate(time.Millisecond),
	)
}
