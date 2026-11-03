package snapshotAnalysis

import (
	"encoding/binary"
	"encoding/hex"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// marks are the moments of the snapshot life cycle, stored as unix nanoseconds (0 = not reached)
const (
	markRequested      = iota // SnapshotState passed the dedup / in-progress checks
	markEpochReady            // the storage epoch matches the snapshot epoch (WaitForStorageEpochChange)
	markChecked               // ShouldTakeSnapshot returned
	markStarted               // the snapshot goroutines were launched
	markFirstGoroutine        // the first trie traversal goroutine started
	markMainDone              // the main trie traversal finished
	markLeavesConsumed        // the main trie leaves consumer finished enqueuing data tries
	markTriesDone             // every trie traversal finished
	markSyncDone              // the missing nodes syncer finished
	markCompleted             // the snapshot was marked complete (active db written)
	markStallReleased         // the caller that waited for the snapshot (import-db / serialized) resumed
	numMarks
)

// node types reported by the trie traversal
const (
	NodeBranch = iota
	NodeExtension
	NodeLeaf
	numNodeTypes
)

// found-epoch distance buckets: target epoch, epoch-1, epoch-2, older; DistCache marks a cache hit, whose
// epoch is unknown
const (
	numDistBuckets = 4
	// DistCache is the distance reported for a node served from the storer cache
	DistCache = -1
)

// leaf kinds of the main trie reported by the leaves consumer
const (
	LeafCode = iota
	LeafAccount
	LeafAccountWithDataTrie
)

const (
	flushEveryNodes   = 1024
	largeTrieMinNodes = 1000
)

// Snapshot is the measurement of one state snapshot (one trie storage manager, one epoch). Every method is
// safe to call on a nil receiver, so callers can keep the value returned by SnapshotRequested without checks.
type Snapshot struct {
	c          *collector
	seq        uint64
	identifier string
	epoch      uint32
	rootHash   []byte
	importDb   bool
	created    time.Time

	marks      [numMarks]atomic.Int64
	skipped    atomic.Bool
	skipReason atomic.Pointer[string]
	errored    atomic.Bool
	completed  atomic.Bool
	stalledNs  atomic.Int64

	mainNodes    atomic.Int64
	mainBytes    atomic.Int64
	dataNodes    atomic.Int64
	dataBytes    atomic.Int64
	nodeTypes    [numNodeTypes]atomic.Int64
	tries        atomic.Int64
	triesData    atomic.Int64
	triesEmpty   atomic.Int64
	maxDepthMain atomic.Int64
	maxDepthData atomic.Int64

	getNs        atomic.Int64
	getMaxNs     atomic.Int64
	getBytes     atomic.Int64
	getCacheHits atomic.Int64
	getDist      [numDistBuckets]atomic.Int64
	getMiss      atomic.Int64
	getHist      histogram
	putNs        atomic.Int64
	putMaxNs     atomic.Int64
	puts         atomic.Int64
	putBytes     atomic.Int64
	putHist      histogram
	prevPuts     atomic.Int64
	prevPutBytes atomic.Int64
	prevPutNs    atomic.Int64
	decodeNs     atomic.Int64
	parkedNs     atomic.Int64
	parks        atomic.Int64
	leavesSendNs atomic.Int64

	grStarted       atomic.Int64
	grWallNs        atomic.Int64
	grActive        atomic.Int64
	grMaxActive     atomic.Int64
	grParked        atomic.Int64
	queueLatencyNs  atomic.Int64
	queueLatencyMax atomic.Int64
	enqueueWaits    atomic.Int64
	enqueueWaitNs   atomic.Int64
	throttlerWaits  atomic.Int64
	throttlerWaitNs atomic.Int64

	leaves             atomic.Int64
	leavesCode         atomic.Int64
	leavesWithDataTrie atomic.Int64
	leavesRecvWaitNs   atomic.Int64
	leavesUnmarshalNs  atomic.Int64

	missing atomic.Int64
	syncNs  atomic.Int64

	mutSets         sync.Mutex
	sampled         map[uint64]struct{}
	sampledTotal    int64
	sampledDistinct int64
	roots           map[uint64]struct{}
	rootsCapped     bool
	dupTries        int64
	dupTrieNodes    int64
	dupTrieBytes    int64

	storer atomic.Pointer[StorerMeter]

	mutProbes  sync.RWMutex
	leavesLen  func() int
	missingLen func() int
	queueLen   func() int

	baseBusy   time.Duration
	baseCPU    time.Duration
	baseCommit commitValues
	basePrune  pruneValues
	baseStorer storerValues
	baseHeap   uint64
	baseGC     uint32
	baseWalks  int64
	baseWalkNs int64

	stopCh        chan struct{}
	stopOnce      sync.Once
	samplerDone   chan struct{}
	afterDeadline atomic.Int64
}

func newSnapshot(c *collector, seq uint64, identifier string, epoch uint32, rootHash []byte, importDb bool, storer *StorerMeter) *Snapshot {
	now := time.Now()
	s := &Snapshot{
		c:           c,
		seq:         seq,
		identifier:  identifier,
		epoch:       epoch,
		rootHash:    append([]byte{}, rootHash...),
		importDb:    importDb,
		created:     now,
		sampled:     make(map[uint64]struct{}),
		roots:       make(map[uint64]struct{}),
		stopCh:      make(chan struct{}),
		samplerDone: make(chan struct{}),
	}
	s.marks[markRequested].Store(now.UnixNano())
	if c == nil {
		return s
	}

	s.storer.Store(storer)
	s.baseBusy = c.busy.totalAt(now)
	s.baseCPU = processCPUTime()
	s.baseCommit = c.commitMeter(identifier).values()
	s.basePrune = c.pruneMeter(identifier).values()
	s.baseStorer = s.storer.Load().values()
	s.baseWalks = c.epochRootHashWalks.Load()
	s.baseWalkNs = c.epochRootHashWalkNs.Load()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	s.baseHeap = ms.HeapAlloc
	s.baseGC = ms.NumGC

	return s
}

func (s *Snapshot) mark(idx int) {
	if s == nil {
		return
	}
	s.marks[idx].CompareAndSwap(0, time.Now().UnixNano())
}

func (s *Snapshot) markTime(idx int) time.Time {
	v := s.marks[idx].Load()
	if v == 0 {
		return time.Time{}
	}
	return time.Unix(0, v)
}

// between returns the duration between two marks, 0 when either is not set
func (s *Snapshot) between(from int, to int) time.Duration {
	a := s.marks[from].Load()
	b := s.marks[to].Load()
	if a == 0 || b == 0 || b < a {
		return 0
	}
	return time.Duration(b - a)
}

func (s *Snapshot) isCompleted() bool {
	return s.completed.Load()
}

// EpochReady marks that the trie storage reached the snapshot epoch
func (s *Snapshot) EpochReady() { s.mark(markEpochReady) }

// Checked marks that ShouldTakeSnapshot returned
func (s *Snapshot) Checked() { s.mark(markChecked) }

// Started marks that the snapshot goroutines were launched
func (s *Snapshot) Started() { s.mark(markStarted) }

// LeavesConsumed marks that the main trie leaves consumer enqueued its last data trie
func (s *Snapshot) LeavesConsumed() { s.mark(markLeavesConsumed) }

// TriesDone marks that every trie traversal finished
func (s *Snapshot) TriesDone() { s.mark(markTriesDone) }

// SyncDone marks that the missing nodes syncer finished
func (s *Snapshot) SyncDone() { s.mark(markSyncDone) }

// Skipped marks a snapshot that ended early (storage closed, epoch change timeout, trie already synced)
func (s *Snapshot) Skipped(reason string) {
	if s == nil {
		return
	}
	s.skipped.Store(true)
	s.skipReason.Store(&reason)
	s.Completed(false)
}

// Errored marks that an error was written on the snapshot error channel
func (s *Snapshot) Errored() {
	if s == nil {
		return
	}
	s.errored.Store(true)
}

// StallReleased records how long the caller of SnapshotState was blocked waiting for the snapshot
// (import-db and serialized snapshots), which is a processing stall of that length
func (s *Snapshot) StallReleased(d time.Duration) {
	if s == nil {
		return
	}
	s.stalledNs.Store(int64(d))
	s.mark(markStallReleased)
}

// Completed marks the end of the snapshot; the row is written once the after-window elapsed
func (s *Snapshot) Completed(hadError bool) {
	if s == nil {
		return
	}
	if hadError {
		s.errored.Store(true)
	}
	s.mark(markCompleted)
	if !s.completed.CompareAndSwap(false, true) {
		return
	}
	if s.c == nil {
		return
	}
	s.afterDeadline.Store(time.Now().Add(s.c.cfg.afterWindow()).UnixNano())
	s.logSummary()
}

// Leaf records one main trie leaf handled by the leaves consumer: the time it waited for the leaf, the time
// spent unmarshalling it and what it turned out to be
func (s *Snapshot) Leaf(recvWait time.Duration, unmarshal time.Duration, kind int) {
	if s == nil {
		return
	}
	s.leaves.Add(1)
	s.leavesRecvWaitNs.Add(int64(recvWait))
	s.leavesUnmarshalNs.Add(int64(unmarshal))
	switch kind {
	case LeafCode:
		s.leavesCode.Add(1)
	case LeafAccountWithDataTrie:
		s.leavesWithDataTrie.Add(1)
	}
}

// EnqueueWait records a TakeSnapshot call blocked on the snapshot requests channel
func (s *Snapshot) EnqueueWait(d time.Duration) {
	if s == nil {
		return
	}
	s.enqueueWaits.Add(1)
	s.enqueueWaitNs.Add(int64(d))
}

// ThrottlerWait records the time the snapshot dispatcher waited for a free goroutine slot
func (s *Snapshot) ThrottlerWait(d time.Duration) {
	if s == nil {
		return
	}
	s.throttlerWaits.Add(1)
	s.throttlerWaitNs.Add(int64(d))
}

// EmptyTrieSkipped records a TakeSnapshot request for an empty trie root
func (s *Snapshot) EmptyTrieSkipped() {
	if s == nil {
		return
	}
	s.triesEmpty.Add(1)
}

// MissingNode records a node that was not found in storage and was sent to the syncer
func (s *Snapshot) MissingNode() {
	if s == nil {
		return
	}
	s.missing.Add(1)
}

// MissingSynced records one syncer call for a missing node
func (s *Snapshot) MissingSynced(d time.Duration) {
	if s == nil {
		return
	}
	s.syncNs.Add(int64(d))
}

// SetChannelProbes registers the length probes of the leaves and missing nodes channels
func (s *Snapshot) SetChannelProbes(leavesLen func() int, missingLen func() int) {
	if s == nil {
		return
	}
	s.mutProbes.Lock()
	s.leavesLen = leavesLen
	s.missingLen = missingLen
	s.mutProbes.Unlock()
}

// SetQueueProbe registers the length probe of the trie storage manager snapshot requests queue
func (s *Snapshot) SetQueueProbe(queueLen func() int) {
	if s == nil {
		return
	}
	s.mutProbes.Lock()
	s.queueLen = queueLen
	s.mutProbes.Unlock()
}

// AttachStorer links the storer meter of the trie storage manager taking this snapshot
func (s *Snapshot) AttachStorer(m *StorerMeter) {
	if s == nil || m == nil {
		return
	}
	if s.storer.CompareAndSwap(nil, m) {
		s.mutSets.Lock()
		s.baseStorer = m.values()
		s.mutSets.Unlock()
	}
}

func (s *Snapshot) storerDelta() storerValues {
	s.mutSets.Lock()
	base := s.baseStorer
	s.mutSets.Unlock()
	return s.storer.Load().values().sub(base)
}

func (s *Snapshot) probe(get func() int) int {
	if get == nil {
		return 0
	}
	return get()
}

func (s *Snapshot) probes() (leaves int, missing int, queue int) {
	s.mutProbes.RLock()
	defer s.mutProbes.RUnlock()

	return s.probe(s.leavesLen), s.probe(s.missingLen), s.probe(s.queueLen)
}

// GoroutineStarted records the start of a trie traversal goroutine and returns its meter. The address is
// empty for the main trie. queueLatency is the time the request spent in the snapshot queue.
func (s *Snapshot) GoroutineStarted(address string, queueLatency time.Duration) *GoroutineMeter {
	if s == nil {
		return nil
	}
	s.mark(markFirstGoroutine)
	s.grStarted.Add(1)
	active := s.grActive.Add(1)
	atomicMax(&s.grMaxActive, active)
	s.queueLatencyNs.Add(int64(queueLatency))
	atomicMax(&s.queueLatencyMax, int64(queueLatency))

	return &GoroutineMeter{
		snap:    s,
		isMain:  len(address) == 0,
		address: address,
		start:   time.Now(),
	}
}

// GoroutineMeter accumulates the measurements of one trie traversal goroutine locally and flushes them
// into the snapshot every flushEveryNodes nodes, so the hot path does not touch shared memory
type GoroutineMeter struct {
	snap    *Snapshot
	isMain  bool
	address string
	start   time.Time

	nodes      int64
	bytes      int64
	nodeTypes  [numNodeTypes]int64
	maxDepth   int64
	sinceFlush int

	getNs        int64
	getMaxNs     int64
	getBytes     int64
	getCacheHits int64
	getDist      [numDistBuckets]int64
	getMiss      int64
	getHist      [numLatencyBuckets]int64
	puts         int64
	putNs        int64
	putMaxNs     int64
	putBytes     int64
	putHist      [numLatencyBuckets]int64
	prevPuts     int64
	prevPutBytes int64
	prevPutNs    int64
	decodeNs     int64
	parkedNs     int64
	parks        int64
	leavesSendNs int64

	totalNodes   int64
	totalBytes   int64
	totalGetNs   int64
	totalPutNs   int64
	totalDecode  int64
	totalParked  int64
	totalLeafSnd int64
	totalDepth   int64

	sampled []uint64
}

// Get records a storage read of a node found in storage. dist is the distance between the snapshot epoch
// and the epoch the node was found in (0 = found in the snapshot epoch), or distCache for a cache hit
func (g *GoroutineMeter) Get(d time.Duration, size int, dist int) {
	if g == nil {
		return
	}
	g.getNs += int64(d)
	if int64(d) > g.getMaxNs {
		g.getMaxNs = int64(d)
	}
	g.getBytes += int64(size)
	g.getHist[latencyBucket(d)]++
	if dist == DistCache {
		g.getCacheHits++
		return
	}
	if dist >= numDistBuckets {
		dist = numDistBuckets - 1
	}
	if dist < 0 {
		dist = 0
	}
	g.getDist[dist]++
}

// GetMiss records a storage read that did not find the node
func (g *GoroutineMeter) GetMiss(d time.Duration) {
	if g == nil {
		return
	}
	g.getNs += int64(d)
	g.getMiss++
	g.getHist[latencyBucket(d)]++
}

// Put records the write of a node into the snapshot epoch
func (g *GoroutineMeter) Put(d time.Duration, size int) {
	if g == nil {
		return
	}
	g.puts++
	g.putNs += int64(d)
	if int64(d) > g.putMaxNs {
		g.putMaxNs = int64(d)
	}
	g.putBytes += int64(size)
	g.putHist[latencyBucket(d)]++
}

// PrevPut records the extra write of a node into the previous epoch (putInPreviousStorerIfAbsent)
func (g *GoroutineMeter) PrevPut(d time.Duration, size int) {
	if g == nil {
		return
	}
	g.prevPuts++
	g.prevPutBytes += int64(size)
	g.prevPutNs += int64(d)
}

// Decode records the time spent decoding a node
func (g *GoroutineMeter) Decode(d time.Duration) {
	if g == nil {
		return
	}
	g.decodeNs += int64(d)
}

// ParkEnter marks the goroutine as parked because the node is busy
func (g *GoroutineMeter) ParkEnter() {
	if g == nil {
		return
	}
	g.snap.grParked.Add(1)
}

// ParkExit records the time the goroutine stayed parked
func (g *GoroutineMeter) ParkExit(d time.Duration) {
	if g == nil {
		return
	}
	g.snap.grParked.Add(-1)
	g.parks++
	g.parkedNs += int64(d)
}

// LeafSent records the time a main trie leaf waited to be written on the leaves channel
func (g *GoroutineMeter) LeafSent(d time.Duration) {
	if g == nil {
		return
	}
	g.leavesSendNs += int64(d)
}

// Node records one traversed node. The hash is sampled for the duplicate writes estimate.
func (g *GoroutineMeter) Node(hash []byte, nodeType int, size int, depth int) {
	if g == nil {
		return
	}
	g.nodes++
	g.bytes += int64(size)
	if nodeType >= 0 && nodeType < numNodeTypes {
		g.nodeTypes[nodeType]++
	}
	if int64(depth) > g.maxDepth {
		g.maxDepth = int64(depth)
	}
	if len(hash) >= 16 && g.snap.c != nil {
		shift := g.snap.c.cfg.NodeHashSampleShift
		tail := binary.LittleEndian.Uint16(hash[len(hash)-2:])
		if tail>>(16-shift) == 0 {
			g.sampled = append(g.sampled, binary.LittleEndian.Uint64(hash[:8]))
		}
	}
	g.sinceFlush++
	if g.sinceFlush >= flushEveryNodes {
		g.flush()
	}
}

func (g *GoroutineMeter) flush() {
	s := g.snap
	if g.isMain {
		s.mainNodes.Add(g.nodes)
		s.mainBytes.Add(g.bytes)
		atomicMax(&s.maxDepthMain, g.maxDepth)
	} else {
		s.dataNodes.Add(g.nodes)
		s.dataBytes.Add(g.bytes)
		atomicMax(&s.maxDepthData, g.maxDepth)
	}
	for i := range g.nodeTypes {
		s.nodeTypes[i].Add(g.nodeTypes[i])
		g.nodeTypes[i] = 0
	}
	s.getNs.Add(g.getNs)
	atomicMax(&s.getMaxNs, g.getMaxNs)
	s.getBytes.Add(g.getBytes)
	s.getCacheHits.Add(g.getCacheHits)
	for i := range g.getDist {
		s.getDist[i].Add(g.getDist[i])
		g.getDist[i] = 0
	}
	s.getMiss.Add(g.getMiss)
	for i := range g.getHist {
		s.getHist[i].Add(g.getHist[i])
		s.putHist[i].Add(g.putHist[i])
		g.getHist[i] = 0
		g.putHist[i] = 0
	}
	s.puts.Add(g.puts)
	s.putNs.Add(g.putNs)
	atomicMax(&s.putMaxNs, g.putMaxNs)
	s.putBytes.Add(g.putBytes)
	s.prevPuts.Add(g.prevPuts)
	s.prevPutBytes.Add(g.prevPutBytes)
	s.prevPutNs.Add(g.prevPutNs)
	s.decodeNs.Add(g.decodeNs)
	s.parkedNs.Add(g.parkedNs)
	s.parks.Add(g.parks)
	s.leavesSendNs.Add(g.leavesSendNs)

	g.totalNodes += g.nodes
	g.totalBytes += g.bytes
	g.totalGetNs += g.getNs
	g.totalPutNs += g.putNs
	g.totalDecode += g.decodeNs
	g.totalParked += g.parkedNs
	g.totalLeafSnd += g.leavesSendNs
	if g.maxDepth > g.totalDepth {
		g.totalDepth = g.maxDepth
	}

	if len(g.sampled) > 0 {
		s.mutSets.Lock()
		for _, h := range g.sampled {
			s.sampledTotal++
			_, seen := s.sampled[h]
			if !seen {
				s.sampled[h] = struct{}{}
				s.sampledDistinct++
			}
		}
		s.mutSets.Unlock()
		g.sampled = g.sampled[:0]
	}

	g.nodes, g.bytes, g.maxDepth = 0, 0, 0
	g.getNs, g.getMaxNs, g.getBytes, g.getCacheHits, g.getMiss = 0, 0, 0, 0, 0
	g.puts, g.putNs, g.putMaxNs, g.putBytes = 0, 0, 0, 0
	g.prevPuts, g.prevPutBytes, g.prevPutNs = 0, 0, 0
	g.decodeNs, g.parkedNs, g.parks, g.leavesSendNs = 0, 0, 0, 0
	g.sinceFlush = 0
}

// Finish flushes the meter, records the goroutine wall time and, for data tries, whether the same root
// hash was already snapshotted during this snapshot (identical data tries are traversed and written again)
func (g *GoroutineMeter) Finish(rootHash []byte) {
	if g == nil {
		return
	}
	g.flush()
	s := g.snap
	wall := time.Since(g.start)
	s.grWallNs.Add(int64(wall))
	s.grActive.Add(-1)
	s.tries.Add(1)
	if g.isMain {
		s.mark(markMainDone)
	} else {
		s.triesData.Add(1)
		g.recordDataTrieRoot(rootHash)
	}

	if s.c != nil && (g.isMain || g.totalNodes >= largeTrieMinNodes) {
		s.c.writeTrieRow(s, g, rootHash, wall)
	}
}

func (g *GoroutineMeter) recordDataTrieRoot(rootHash []byte) {
	if len(rootHash) < 8 || g.snap.c == nil {
		return
	}
	s := g.snap
	key := binary.LittleEndian.Uint64(rootHash[:8])

	s.mutSets.Lock()
	defer s.mutSets.Unlock()

	_, seen := s.roots[key]
	if seen {
		s.dupTries++
		s.dupTrieNodes += g.totalNodes
		s.dupTrieBytes += g.totalBytes
		return
	}
	if len(s.roots) >= s.c.cfg.MaxTrackedDataTrieRoots {
		s.rootsCapped = true
		return
	}
	s.roots[key] = struct{}{}
}

func (s *Snapshot) rootHashHex() string {
	return hex.EncodeToString(s.rootHash)
}
