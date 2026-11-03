package snapshotAnalysis

import (
	"sync"
	"sync/atomic"
	"time"
)

// latency histogram buckets: <100us, <1ms, <10ms, <100ms, >=100ms
const numLatencyBuckets = 5

var latencyBucketNames = [numLatencyBuckets]string{"le100us", "le1ms", "le10ms", "le100ms", "gt100ms"}

func latencyBucket(d time.Duration) int {
	switch {
	case d < 100*time.Microsecond:
		return 0
	case d < time.Millisecond:
		return 1
	case d < 10*time.Millisecond:
		return 2
	case d < 100*time.Millisecond:
		return 3
	default:
		return 4
	}
}

func atomicMax(a *atomic.Int64, v int64) {
	for {
		cur := a.Load()
		if v <= cur {
			return
		}
		if a.CompareAndSwap(cur, v) {
			return
		}
	}
}

type histogram [numLatencyBuckets]atomic.Int64

func (h *histogram) observe(d time.Duration) {
	h[latencyBucket(d)].Add(1)
}

func (h *histogram) load() [numLatencyBuckets]int64 {
	var out [numLatencyBuckets]int64
	for i := range h {
		out[i] = h[i].Load()
	}
	return out
}

// StorerMeter collects the storage level measurements of one trie pruning storer. The processing path
// (Get / Put) and the snapshot path (GetWithoutAddingToCache / PutInEpochWithoutCache) are kept apart,
// so the two can be compared while a snapshot runs. The hooks are safe to call on a nil meter.
type StorerMeter struct {
	identifier string

	procGets      atomic.Int64
	procCacheHits atomic.Int64
	procGetNs     atomic.Int64
	procGetMaxNs  atomic.Int64
	procGetHist   histogram
	procPuts      atomic.Int64
	procPutBytes  atomic.Int64
	procPutNs     atomic.Int64
	procPutMaxNs  atomic.Int64
	procPutHist   histogram
	procRemoves   atomic.Int64
	procRemoveNs  atomic.Int64

	snapGets        atomic.Int64
	snapCacheHits   atomic.Int64
	snapGetRLockNs  atomic.Int64
	snapGetProbes   atomic.Int64
	snapGetProbeNs  atomic.Int64
	snapGetMaxNs    atomic.Int64
	snapGetHist     histogram
	snapPuts        atomic.Int64
	snapPutBytes    atomic.Int64
	snapPutLockNs   atomic.Int64
	snapPutLockMax  atomic.Int64
	snapPutNs       atomic.Int64
	snapPutMaxNs    atomic.Int64
	snapPutHist     histogram
	snapPutNoEpoch  atomic.Int64
	epochChanges    atomic.Int64
	epochChangeNs   atomic.Int64
	epochChangeLock atomic.Int64

	mutProbe   sync.RWMutex
	cacheProbe func() (numEntries int, numBytes int64, numActivePersisters int)
}

// NewStorerMeter creates a meter for the trie pruning storer with the given identifier
func NewStorerMeter(identifier string) *StorerMeter {
	m := &StorerMeter{identifier: identifier}
	registerStorerMeter(m)
	return m
}

// SetCacheProbe registers a function that reports the size of the storer's cache and the number of
// active persisters; it is sampled on the timeline while a snapshot runs
func (m *StorerMeter) SetCacheProbe(probe func() (numEntries int, numBytes int64, numActivePersisters int)) {
	if m == nil {
		return
	}
	m.mutProbe.Lock()
	m.cacheProbe = probe
	m.mutProbe.Unlock()
}

func (m *StorerMeter) probeCache() (int, int64, int) {
	if m == nil {
		return 0, 0, 0
	}
	m.mutProbe.RLock()
	probe := m.cacheProbe
	m.mutProbe.RUnlock()
	if probe == nil {
		return 0, 0, 0
	}
	return probe()
}

// Active returns true when the collector is enabled and the hooks should be timed
func (m *StorerMeter) Active() bool {
	return m != nil && enabled.Load()
}

// ProcCacheHit records a processing Get served from the cache
func (m *StorerMeter) ProcCacheHit() {
	if !m.Active() {
		return
	}
	m.procGets.Add(1)
	m.procCacheHits.Add(1)
}

// ProcGet records a processing Get that went to the persisters
func (m *StorerMeter) ProcGet(d time.Duration) {
	if !m.Active() {
		return
	}
	m.procGets.Add(1)
	m.procGetNs.Add(int64(d))
	atomicMax(&m.procGetMaxNs, int64(d))
	m.procGetHist.observe(d)
}

// ProcPut records a processing Put into a persister
func (m *StorerMeter) ProcPut(d time.Duration, size int) {
	if !m.Active() {
		return
	}
	m.procPuts.Add(1)
	m.procPutBytes.Add(int64(size))
	m.procPutNs.Add(int64(d))
	atomicMax(&m.procPutMaxNs, int64(d))
	m.procPutHist.observe(d)
}

// ProcRemove records a removal (pruning) from a persister
func (m *StorerMeter) ProcRemove(d time.Duration) {
	if !m.Active() {
		return
	}
	m.procRemoves.Add(1)
	m.procRemoveNs.Add(int64(d))
}

// SnapCacheHit records a snapshot read served from the cache
func (m *StorerMeter) SnapCacheHit() {
	if !m.Active() {
		return
	}
	m.snapGets.Add(1)
	m.snapCacheHits.Add(1)
}

// SnapGet records a snapshot read that scanned the persisters: the time spent acquiring the storer
// read lock, the number of persisters probed and the time spent inside the persister calls
func (m *StorerMeter) SnapGet(rlockWait time.Duration, probes int, probeTime time.Duration) {
	if !m.Active() {
		return
	}
	m.snapGets.Add(1)
	m.snapGetRLockNs.Add(int64(rlockWait))
	m.snapGetProbes.Add(int64(probes))
	m.snapGetProbeNs.Add(int64(probeTime))
	total := rlockWait + probeTime
	atomicMax(&m.snapGetMaxNs, int64(total))
	m.snapGetHist.observe(total)
}

// SnapPut records a snapshot write: the time spent acquiring the storer write lock and the persister Put
func (m *StorerMeter) SnapPut(lockWait time.Duration, putTime time.Duration, size int) {
	if !m.Active() {
		return
	}
	m.snapPuts.Add(1)
	m.snapPutBytes.Add(int64(size))
	m.snapPutLockNs.Add(int64(lockWait))
	atomicMax(&m.snapPutLockMax, int64(lockWait))
	m.snapPutNs.Add(int64(putTime))
	atomicMax(&m.snapPutMaxNs, int64(putTime))
	m.snapPutHist.observe(lockWait + putTime)
}

// SnapPutNoEpoch records a snapshot write that failed because the epoch persister does not exist
func (m *StorerMeter) SnapPutNoEpoch() {
	if !m.Active() {
		return
	}
	m.snapPutNoEpoch.Add(1)
}

// EpochChange records the duration of a storer epoch change and how long it waited for the lock
func (m *StorerMeter) EpochChange(lockWait time.Duration, total time.Duration) {
	if !m.Active() {
		return
	}
	m.epochChanges.Add(1)
	m.epochChangeLock.Add(int64(lockWait))
	m.epochChangeNs.Add(int64(total))
}

// storerValues is a point in time copy of the meter counters
type storerValues struct {
	procGets, procCacheHits, procGetNs, procGetMaxNs       int64
	procPuts, procPutBytes, procPutNs, procPutMaxNs        int64
	procRemoves, procRemoveNs                              int64
	procGetHist, procPutHist                               [numLatencyBuckets]int64
	snapGets, snapCacheHits, snapGetRLockNs, snapGetProbes int64
	snapGetProbeNs, snapGetMaxNs                           int64
	snapGetHist, snapPutHist                               [numLatencyBuckets]int64
	snapPuts, snapPutBytes, snapPutLockNs, snapPutLockMax  int64
	snapPutNs, snapPutMaxNs, snapPutNoEpoch                int64
	epochChanges, epochChangeNs, epochChangeLock           int64
}

func (m *StorerMeter) values() storerValues {
	if m == nil {
		return storerValues{}
	}
	return storerValues{
		procGets:        m.procGets.Load(),
		procCacheHits:   m.procCacheHits.Load(),
		procGetNs:       m.procGetNs.Load(),
		procGetMaxNs:    m.procGetMaxNs.Load(),
		procPuts:        m.procPuts.Load(),
		procPutBytes:    m.procPutBytes.Load(),
		procPutNs:       m.procPutNs.Load(),
		procPutMaxNs:    m.procPutMaxNs.Load(),
		procRemoves:     m.procRemoves.Load(),
		procRemoveNs:    m.procRemoveNs.Load(),
		procGetHist:     m.procGetHist.load(),
		procPutHist:     m.procPutHist.load(),
		snapGets:        m.snapGets.Load(),
		snapCacheHits:   m.snapCacheHits.Load(),
		snapGetRLockNs:  m.snapGetRLockNs.Load(),
		snapGetProbes:   m.snapGetProbes.Load(),
		snapGetProbeNs:  m.snapGetProbeNs.Load(),
		snapGetMaxNs:    m.snapGetMaxNs.Load(),
		snapGetHist:     m.snapGetHist.load(),
		snapPutHist:     m.snapPutHist.load(),
		snapPuts:        m.snapPuts.Load(),
		snapPutBytes:    m.snapPutBytes.Load(),
		snapPutLockNs:   m.snapPutLockNs.Load(),
		snapPutLockMax:  m.snapPutLockMax.Load(),
		snapPutNs:       m.snapPutNs.Load(),
		snapPutMaxNs:    m.snapPutMaxNs.Load(),
		snapPutNoEpoch:  m.snapPutNoEpoch.Load(),
		epochChanges:    m.epochChanges.Load(),
		epochChangeNs:   m.epochChangeNs.Load(),
		epochChangeLock: m.epochChangeLock.Load(),
	}
}

func subHist(a, b [numLatencyBuckets]int64) [numLatencyBuckets]int64 {
	var out [numLatencyBuckets]int64
	for i := range a {
		out[i] = a[i] - b[i]
	}
	return out
}

// sub returns the difference of two copies (maxima are kept from the newer copy)
func (v storerValues) sub(o storerValues) storerValues {
	return storerValues{
		procGets:        v.procGets - o.procGets,
		procCacheHits:   v.procCacheHits - o.procCacheHits,
		procGetNs:       v.procGetNs - o.procGetNs,
		procGetMaxNs:    v.procGetMaxNs,
		procPuts:        v.procPuts - o.procPuts,
		procPutBytes:    v.procPutBytes - o.procPutBytes,
		procPutNs:       v.procPutNs - o.procPutNs,
		procPutMaxNs:    v.procPutMaxNs,
		procRemoves:     v.procRemoves - o.procRemoves,
		procRemoveNs:    v.procRemoveNs - o.procRemoveNs,
		procGetHist:     subHist(v.procGetHist, o.procGetHist),
		procPutHist:     subHist(v.procPutHist, o.procPutHist),
		snapGets:        v.snapGets - o.snapGets,
		snapCacheHits:   v.snapCacheHits - o.snapCacheHits,
		snapGetRLockNs:  v.snapGetRLockNs - o.snapGetRLockNs,
		snapGetProbes:   v.snapGetProbes - o.snapGetProbes,
		snapGetProbeNs:  v.snapGetProbeNs - o.snapGetProbeNs,
		snapGetMaxNs:    v.snapGetMaxNs,
		snapGetHist:     subHist(v.snapGetHist, o.snapGetHist),
		snapPutHist:     subHist(v.snapPutHist, o.snapPutHist),
		snapPuts:        v.snapPuts - o.snapPuts,
		snapPutBytes:    v.snapPutBytes - o.snapPutBytes,
		snapPutLockNs:   v.snapPutLockNs - o.snapPutLockNs,
		snapPutLockMax:  v.snapPutLockMax,
		snapPutNs:       v.snapPutNs - o.snapPutNs,
		snapPutMaxNs:    v.snapPutMaxNs,
		snapPutNoEpoch:  v.snapPutNoEpoch - o.snapPutNoEpoch,
		epochChanges:    v.epochChanges - o.epochChanges,
		epochChangeNs:   v.epochChangeNs - o.epochChangeNs,
		epochChangeLock: v.epochChangeLock - o.epochChangeLock,
	}
}

// commitMeter counts the accounts adapter commits of one trie
type commitMeter struct {
	count atomic.Int64
	ns    atomic.Int64
	maxNs atomic.Int64
}

type commitValues struct {
	count, ns, maxNs int64
}

func (c *commitMeter) values() commitValues {
	if c == nil {
		return commitValues{}
	}
	return commitValues{count: c.count.Load(), ns: c.ns.Load(), maxNs: c.maxNs.Load()}
}

func (v commitValues) sub(o commitValues) commitValues {
	return commitValues{count: v.count - o.count, ns: v.ns - o.ns, maxNs: v.maxNs}
}

// pruneMeter counts the pruning activity of one trie storage manager
type pruneMeter struct {
	buffered   atomic.Int64
	resolves   atomic.Int64
	resolveOps atomic.Int64
	resolveNs  atomic.Int64
	prunes     atomic.Int64
	removed    atomic.Int64
	removeNs   atomic.Int64

	mutProbe sync.RWMutex
	probe    func() (bufferLen int, evictionWaitingListLen int)
}

type pruneValues struct {
	buffered, resolves, resolveOps, resolveNs, prunes, removed, removeNs int64
}

func (p *pruneMeter) values() pruneValues {
	if p == nil {
		return pruneValues{}
	}
	return pruneValues{
		buffered:   p.buffered.Load(),
		resolves:   p.resolves.Load(),
		resolveOps: p.resolveOps.Load(),
		resolveNs:  p.resolveNs.Load(),
		prunes:     p.prunes.Load(),
		removed:    p.removed.Load(),
		removeNs:   p.removeNs.Load(),
	}
}

func (v pruneValues) sub(o pruneValues) pruneValues {
	return pruneValues{
		buffered:   v.buffered - o.buffered,
		resolves:   v.resolves - o.resolves,
		resolveOps: v.resolveOps - o.resolveOps,
		resolveNs:  v.resolveNs - o.resolveNs,
		prunes:     v.prunes - o.prunes,
		removed:    v.removed - o.removed,
		removeNs:   v.removeNs - o.removeNs,
	}
}

func (p *pruneMeter) probeLens() (int, int) {
	if p == nil {
		return 0, 0
	}
	p.mutProbe.RLock()
	probe := p.probe
	p.mutProbe.RUnlock()
	if probe == nil {
		return 0, 0
	}
	return probe()
}

// busyTracker accumulates the time the process status handler reported the node as not idle,
// which is the time the snapshot goroutines are forced to park
type busyTracker struct {
	mut   sync.Mutex
	busy  bool
	since time.Time
	total time.Duration
}

func (b *busyTracker) set(busy bool, now time.Time) {
	b.mut.Lock()
	defer b.mut.Unlock()

	if b.busy == busy {
		return
	}
	if b.busy {
		b.total += now.Sub(b.since)
	}
	b.busy = busy
	b.since = now
}

func (b *busyTracker) totalAt(now time.Time) time.Duration {
	b.mut.Lock()
	defer b.mut.Unlock()

	total := b.total
	if b.busy {
		total += now.Sub(b.since)
	}
	return total
}
