package snapshotAnalysis

import (
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("snapshotAnalysis")

// collector is the process wide state of the snapshot analysis. There is one snapshot in progress per
// trie storage manager identifier at most, so snapshots are keyed by identifier.
type collector struct {
	cfg     Config
	shardID uint32
	dir     string

	snapshotsFile *csvFile
	timelineFile  *csvFile
	triesFile     *csvFile

	mut      sync.Mutex
	current  map[string]*Snapshot // the snapshot in progress or in its after-window, per identifier
	seq      uint64
	closed   bool
	samplers sync.WaitGroup

	storers sync.Map                // identifier -> *StorerMeter
	links   map[string]*StorerMeter // trie storage manager identifier -> storer meter
	commits sync.Map                // identifier -> *commitMeter
	prunes  sync.Map                // identifier -> *pruneMeter

	busy busyTracker

	epochRootHashWalks  atomic.Int64
	epochRootHashWalkNs atomic.Int64
}

var (
	global  *collector
	enabled atomic.Bool
	mutInit sync.Mutex
)

// Init creates the global collector. It is a no-op when the config is disabled.
func Init(cfg Config, shardID uint32, workingDir string) error {
	mutInit.Lock()
	defer mutInit.Unlock()

	if global != nil {
		Close()
	}
	if !cfg.Enabled {
		return nil
	}

	cfg = cfg.withDefaults()
	dir := cfg.OutputDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(workingDir, dir)
	}
	dir = filepath.Join(dir, fmt.Sprintf("shard_%d", shardID))

	c := &collector{
		cfg:     cfg,
		shardID: shardID,
		dir:     dir,
		current: make(map[string]*Snapshot),
	}

	var err error
	c.snapshotsFile, err = newCsvFile(filepath.Join(dir, "snapshots.csv.gz"), snapshotHeader())
	if err != nil {
		return err
	}
	c.timelineFile, err = newCsvFile(filepath.Join(dir, "timeline.csv.gz"), timelineHeader())
	if err != nil {
		_ = c.snapshotsFile.close()
		return err
	}
	c.triesFile, err = newCsvFile(filepath.Join(dir, "tries.csv.gz"), triesHeader())
	if err != nil {
		_ = c.snapshotsFile.close()
		_ = c.timelineFile.close()
		return err
	}

	global = c
	enabled.Store(true)
	log.Info("snapshot analysis enabled", "output dir", dir, "label", cfg.Label,
		"sample interval", cfg.sampleInterval(), "after window", cfg.afterWindow())

	return nil
}

// Close finishes the snapshots still open, stops the samplers and closes the files
func Close() {
	if !enabled.Load() {
		return
	}
	enabled.Store(false)

	c := global
	if c == nil {
		return
	}

	c.mut.Lock()
	c.closed = true
	open := make([]*Snapshot, 0, len(c.current))
	for _, s := range c.current {
		open = append(open, s)
	}
	c.mut.Unlock()

	for _, s := range open {
		s.stopSampler()
	}
	c.samplers.Wait()

	_ = c.snapshotsFile.close()
	_ = c.timelineFile.close()
	_ = c.triesFile.close()
	global = nil
}

// Enabled returns true when the collector is active
func Enabled() bool {
	return enabled.Load()
}

func registerStorerMeter(m *StorerMeter) {
	if global == nil {
		// storers are created after Init; a meter created without a collector is still usable because
		// every hook checks the enabled flag first
		return
	}
	global.storers.Store(m.identifier, m)
}

func storerMeterFor(identifier string) *StorerMeter {
	if global == nil {
		return nil
	}
	v, ok := global.storers.Load(identifier)
	if !ok {
		return nil
	}
	return v.(*StorerMeter)
}

func (c *collector) commitMeter(identifier string) *commitMeter {
	v, ok := c.commits.Load(identifier)
	if ok {
		return v.(*commitMeter)
	}
	v, _ = c.commits.LoadOrStore(identifier, &commitMeter{})
	return v.(*commitMeter)
}

func (c *collector) pruneMeter(identifier string) *pruneMeter {
	v, ok := c.prunes.Load(identifier)
	if ok {
		return v.(*pruneMeter)
	}
	v, _ = c.prunes.LoadOrStore(identifier, &pruneMeter{})
	return v.(*pruneMeter)
}

// SnapshotRequested registers a snapshot that passed the "should start" checks of the snapshots manager.
// It returns nil when the collector is disabled. A snapshot still in its after-window for the same
// identifier is finished first.
func SnapshotRequested(identifier string, epoch uint32, rootHash []byte, importDb bool) *Snapshot {
	if !enabled.Load() {
		return nil
	}
	c := global
	if c == nil {
		return nil
	}

	storer := c.linkedStorer(identifier)

	c.mut.Lock()
	if c.closed {
		c.mut.Unlock()
		return nil
	}
	previous := c.current[identifier]
	c.seq++
	s := newSnapshot(c, c.seq, identifier, epoch, rootHash, importDb, storer)
	c.current[identifier] = s
	c.mut.Unlock()

	if previous != nil {
		previous.stopSampler()
	}
	s.startSampler()

	return s
}

// Current returns the snapshot in progress for the given trie storage manager identifier, or nil
func Current(identifier string) *Snapshot {
	if !enabled.Load() {
		return nil
	}
	c := global
	if c == nil {
		return nil
	}

	c.mut.Lock()
	s := c.current[identifier]
	c.mut.Unlock()

	if s == nil || s.isCompleted() {
		return nil
	}
	return s
}

func (c *collector) forget(s *Snapshot) {
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.current[s.identifier] == s {
		delete(c.current, s.identifier)
	}
}

// NodeIdleChanged is called by the process status handler whenever its effective idle state flips.
// Snapshot goroutines park while the node is busy, so the busy time bounds how long a snapshot can run.
func NodeIdleChanged(idle bool) {
	if !enabled.Load() {
		return
	}
	c := global
	if c == nil {
		return
	}
	c.busy.set(!idle, time.Now())
}

// AccountsCommitted records one accounts adapter commit (the trie commit of a block) for the trie with the
// given storage manager identifier
func AccountsCommitted(identifier string, d time.Duration) {
	if !enabled.Load() {
		return
	}
	c := global
	if c == nil {
		return
	}
	m := c.commitMeter(identifier)
	m.count.Add(1)
	m.ns.Add(int64(d))
	atomicMax(&m.maxNs, int64(d))
}

// PruneBuffered records a prune or cancel-prune request that was buffered because pruning is blocked
// (a snapshot is in progress)
func PruneBuffered(identifier string) {
	if !enabled.Load() {
		return
	}
	c := global
	if c == nil {
		return
	}
	c.pruneMeter(identifier).buffered.Add(1)
}

// PruneResolved records the resolution of the buffered pruning operations once pruning is unblocked
func PruneResolved(identifier string, numOps int, d time.Duration) {
	if !enabled.Load() {
		return
	}
	c := global
	if c == nil {
		return
	}
	m := c.pruneMeter(identifier)
	m.resolves.Add(1)
	m.resolveOps.Add(int64(numOps))
	m.resolveNs.Add(int64(d))
}

// PruneRemoved records one prune of a root hash: the number of hashes removed from storage and the time
func PruneRemoved(identifier string, numRemoved int, d time.Duration) {
	if !enabled.Load() {
		return
	}
	c := global
	if c == nil {
		return
	}
	m := c.pruneMeter(identifier)
	m.prunes.Add(1)
	m.removed.Add(int64(numRemoved))
	m.removeNs.Add(int64(d))
}

// RegisterPruneProbe registers a function that reports the pruning buffer length and the eviction waiting
// list length of the storage pruning manager working with the given trie storage manager
func RegisterPruneProbe(identifier string, probe func() (bufferLen int, evictionWaitingListLen int)) {
	if !enabled.Load() {
		return
	}
	c := global
	if c == nil {
		return
	}
	m := c.pruneMeter(identifier)
	m.mutProbe.Lock()
	m.probe = probe
	m.mutProbe.Unlock()
}

// EpochRootHashWalk records the duration of the epoch start leaves walk done by commitTrieEpochRootHashIfNeeded,
// which reads the same trie through the processing path while the snapshot runs
func EpochRootHashWalk(d time.Duration) {
	if !enabled.Load() {
		return
	}
	c := global
	if c == nil {
		return
	}
	c.epochRootHashWalks.Add(1)
	c.epochRootHashWalkNs.Add(int64(d))
}
