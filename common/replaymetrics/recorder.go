package replaymetrics

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
)

// Kind identifies a processing trie. The zero value leaves the trie untracked.
type Kind uint8

const (
	Untracked Kind = iota
	User
	Peer
	Data
	kindCount
)

var kindNames = [kindCount]string{User: "main", Peer: "peer", Data: "data"}

type Counter uint8

const (
	CollapseCommits Counter = iota
	DepthCollapses
	LeafBatches
	LeavesCollapsed
	RootCollapses
	CollapseNs
	DepthCollapseNs
	LeafCollapseNs
	RootCollapseNs
	NodesResolved
	ResolvedBytes
	ResolveErrors
	NodeWrites
	NodeWriteBytes
	RootOpens
	RootOpenErrors
	RootOpenNs
	EmptyTries
	CacheHits
	AttachedHits
	CacheEvictions
	DirtyEvictions
	CacheResets
	counterCount
)

var counterNames = [counterCount]string{
	"collapse_commits", "depth_collapses", "leaf_batches", "leaves_collapsed", "root_collapses",
	"collapse_ns", "depth_collapse_ns", "leaf_collapse_ns", "root_collapse_ns",
	"nodes_resolved", "resolved_bytes", "resolve_errors", "node_writes", "node_write_bytes",
	"root_opens", "root_open_errors", "root_open_ns", "empty_tries", "cache_hits", "attached_hits",
	"cache_evictions", "dirty_evictions", "cache_resets",
}

type Item uint8

const (
	Transaction Item = iota
	ScheduledTransaction
	SmartContractResult
	RewardTransaction
	itemCount
)

var itemNames = [itemCount]string{
	"transaction_calls", "scheduled_transaction_calls", "smart_contract_result_calls", "reward_transaction_calls",
}

func RecordItem(item Item) {
	if r := active.Load(); r != nil {
		if a := r.current.Load(); a != nil {
			a.items[item].Add(1)
		}
	}
}

func RecordStateCommit(header data.HeaderHandler, duration time.Duration) {
	if r := active.Load(); r != nil {
		if a := r.current.Load(); a != nil && a.Shard == header.GetShardID() && a.Nonce == header.GetNonce() {
			a.stateCommitNs.Add(uint64(duration))
		}
	}
}

type gauge uint8

const (
	mainBytes gauge = iota
	peerBytes
	dataCacheBytes
	pendingDirtyBytes
	dataCacheEntries
	pendingDirtyEntries
	gaugeCount
)

var gaugeNames = [gaugeCount]string{
	"main_modeled_bytes", "peer_modeled_bytes", "data_cache_modeled_bytes", "evicted_dirty_modeled_bytes_at_precommit",
	"data_cache_entries", "evicted_dirty_entries",
}

type counts [kindCount][counterCount]uint64

func (c counts) fields() map[string]map[string]uint64 {
	fields := make(map[string]map[string]uint64, kindCount-1)
	for kind := User; kind < kindCount; kind++ {
		group := make(map[string]uint64, counterCount)
		for id, name := range counterNames {
			if value := c[kind][id]; value != 0 {
				group[name] = value
			}
		}
		fields[kindNames[kind]] = group
	}
	return fields
}

type progress struct {
	Shard uint32 `json:"shard"`
	Epoch uint32 `json:"epoch"`
	Nonce uint64 `json:"nonce"`
}

type Attempt struct {
	recorder *Recorder
	progress
	headerHash    string
	path          string
	headerTxs     uint32
	start         time.Time
	counters      [kindCount][counterCount]atomic.Uint64
	items         [itemCount]atomic.Uint64
	stateCommitNs atomic.Uint64
	peaks         [gaugeCount]atomic.Uint64
	execution     time.Duration
	scheduled     time.Duration
	commit        time.Duration
}

type epochKey struct {
	shard uint32
	epoch uint32
}

type epochStats struct {
	counts        counts
	peaks         [gaugeCount]uint64
	attempts      uint64
	executed      uint64
	firstNonce    uint64
	lastNonce     uint64
	headerTxs     uint64
	executionNs   uint64
	stateCommitNs uint64
	items         [itemCount]uint64
	scheduledNs   uint64
	commitNs      uint64
	peakRSS       uint64
	start         time.Time
	end           time.Time
}

type Recorder struct {
	file    *os.File
	buffer  *bufio.Writer
	encoder *json.Encoder

	mu          sync.Mutex
	err         error
	closed      bool
	epochs      map[epochKey]*epochStats
	attemptID   uint64
	overlapping atomic.Uint64
	current     atomic.Pointer[Attempt]
	last        atomic.Pointer[Attempt]
	start       time.Time
	stop        chan struct{}
	done        chan struct{}
}

var active atomic.Pointer[Recorder]

func Enabled() bool {
	return active.Load() != nil
}

// Start creates a replay report. An empty path uses a unique file in workingDir/stats.
func Start(workingDir string, path string, build string) (*Recorder, error) {
	file, err := openResultFile(workingDir, path)
	if err != nil {
		return nil, err
	}
	buffer := bufio.NewWriterSize(file, 64<<10)
	r := &Recorder{
		file: file, buffer: buffer, encoder: json.NewEncoder(buffer),
		epochs: make(map[epochKey]*epochStats), start: time.Now(),
		stop: make(chan struct{}), done: make(chan struct{}),
	}
	if !active.CompareAndSwap(nil, r) {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return nil, fmt.Errorf("import replay metrics already recording")
	}
	r.write(map[string]any{"type": "run_start", "schema": 1, "build": build, "time": r.start.UTC()})
	go r.sampleResources()
	return r, nil
}

func openResultFile(workingDir string, path string) (*os.File, error) {
	if path != "" {
		return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	}

	dir, err := filepath.Abs(filepath.Join(workingDir, "stats"))
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	prefix := "import-db-" + time.Now().UTC().Format("20060102T150405Z")
	return os.CreateTemp(dir, prefix+"-*.jsonl")
}

// Path returns the path to the open report.
func (r *Recorder) Path() string {
	return r.file.Name()
}

// Begin tracks the execution attempt that owns the processing tries.
func Begin(header data.HeaderHandler, hash []byte, path string) *Attempt {
	r := active.Load()
	if r == nil {
		return nil
	}
	a := &Attempt{
		recorder: r, progress: progress{Shard: header.GetShardID(), Epoch: header.GetEpoch(), Nonce: header.GetNonce()},
		headerHash: hex.EncodeToString(hash), path: path, headerTxs: header.GetTxCount(), start: time.Now(),
	}
	if !r.current.CompareAndSwap(nil, a) {
		r.overlapping.Add(1)
		return nil
	}
	return a
}

func (a *Attempt) Execution(d time.Duration) {
	if a != nil {
		a.execution = d
	}
}

func (a *Attempt) Scheduled(d time.Duration) {
	if a != nil {
		a.scheduled = d
	}
}

func (a *Attempt) Commit(d time.Duration) {
	if a != nil {
		a.commit = d
	}
}

func (a *Attempt) Finish(outcome string, root []byte, err error) {
	if a == nil {
		return
	}
	r := a.recorder
	end := time.Now()
	record := map[string]any{
		"type": "block", "shard": a.Shard, "epoch": a.Epoch, "nonce": a.Nonce, "hash": a.headerHash,
		"path": a.path, "outcome": outcome, "header_tx_count": a.headerTxs, "state_root": hex.EncodeToString(root),
		"started_at": a.start.UTC(), "finished_at": end.UTC(), "elapsed_ns": end.Sub(a.start).Nanoseconds(),
		"execution_ns": a.execution.Nanoseconds(), "scheduled_execution_ns": a.scheduled.Nanoseconds(), "commit_ns": a.commit.Nanoseconds(),
	}
	if err != nil {
		record["error"] = err.Error()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	r.last.Store(a)
	r.current.CompareAndSwap(a, nil)
	epoch := r.epoch(a.progress)
	epoch.attempts++
	if epoch.start.IsZero() || a.start.Before(epoch.start) {
		epoch.start = a.start
	}
	if end.After(epoch.end) {
		epoch.end = end
	}
	if outcome == "executed" {
		if epoch.executed == 0 || a.Nonce < epoch.firstNonce {
			epoch.firstNonce = a.Nonce
		}
		if a.Nonce > epoch.lastNonce {
			epoch.lastNonce = a.Nonce
		}
		epoch.executed++
		epoch.headerTxs += uint64(a.headerTxs)
	}
	stateCommitNs := a.stateCommitNs.Load()
	record["state_commit_ns"] = stateCommitNs
	epoch.stateCommitNs += stateCommitNs
	items := make(map[string]uint64, itemCount)
	for id, name := range itemNames {
		value := a.items[id].Load()
		items[name] = value
		if outcome == "executed" {
			epoch.items[id] += value
		}
	}
	record["execution_items"] = items
	epoch.executionNs += uint64(a.execution)
	epoch.scheduledNs += uint64(a.scheduled)
	epoch.commitNs += uint64(a.commit)
	var measured counts
	for kind := User; kind < kindCount; kind++ {
		for id := Counter(0); id < counterCount; id++ {
			value := a.counters[kind][id].Load()
			measured[kind][id] = value
			epoch.counts[kind][id] += value
		}
	}
	peaks := make(map[string]uint64, gaugeCount)
	for id, name := range gaugeNames {
		value := a.peaks[id].Load()
		peaks[name] = value
		epoch.peaks[id] = max(epoch.peaks[id], value)
	}
	r.attemptID++
	record["attempt"] = r.attemptID
	record["tries"] = measured.fields()
	record["peaks"] = peaks
	r.write(record)
}

func (r *Recorder) epoch(p progress) *epochStats {
	key := epochKey{shard: p.Shard, epoch: p.Epoch}
	s := r.epochs[key]
	if s == nil {
		s = &epochStats{}
		r.epochs[key] = s
	}
	return s
}

func (r *Recorder) write(record any) {
	if r.err == nil {
		r.err = r.encoder.Encode(record)
	}
}

func (r *Recorder) flush() {
	if r.err == nil {
		r.err = r.buffer.Flush()
	}
}

func (r *Recorder) Close(runErr error) error {
	active.CompareAndSwap(r, nil)
	close(r.stop)
	<-r.done
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	keys := make([]epochKey, 0, len(r.epochs))
	for key, stats := range r.epochs {
		if stats.attempts != 0 {
			keys = append(keys, key)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].shard < keys[j].shard || keys[i].shard == keys[j].shard && keys[i].epoch < keys[j].epoch
	})
	for _, key := range keys {
		stats := r.epochs[key]
		peaks := make(map[string]uint64, gaugeCount)
		for id, name := range gaugeNames {
			peaks[name] = stats.peaks[id]
		}
		items := make(map[string]uint64, itemCount)
		for id, name := range itemNames {
			items[name] = stats.items[id]
		}
		r.write(map[string]any{
			"type": "epoch", "shard": key.shard, "epoch": key.epoch, "attempts": stats.attempts, "executed_blocks": stats.executed,
			"first_executed_nonce": stats.firstNonce, "last_executed_nonce": stats.lastNonce, "header_tx_count": stats.headerTxs,
			"execution_ns": stats.executionNs, "scheduled_execution_ns": stats.scheduledNs, "commit_ns": stats.commitNs,
			"state_commit_ns": stats.stateCommitNs, "execution_items": items,
			"wall_ns": stats.end.Sub(stats.start).Nanoseconds(), "peak_sampled_rss_bytes": stats.peakRSS,
			"tries": stats.counts.fields(), "peaks": peaks,
		})
	}
	end := map[string]any{"type": "run_end", "time": time.Now().UTC(), "wall_ns": time.Since(r.start).Nanoseconds(), "overlapping_attempts": r.overlapping.Load()}
	if unfinished := r.current.Load(); unfinished != nil {
		end["unfinished_block"] = unfinished.progress
	}
	if runErr != nil {
		end["error"] = runErr.Error()
	}
	r.write(end)
	r.flush()
	closeErr := r.file.Close()
	if r.err != nil {
		return r.err
	}
	return closeErr
}
