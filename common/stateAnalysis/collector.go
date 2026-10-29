package stateAnalysis

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("stateAnalysis")

// ID helpers -----------------------------------------------------------------

func (id CounterID) unit(u int) CounterID         { return id + CounterID(u) }
func (id CounterID) kind(k int) CounterID         { return id + CounterID(k) }
func (id CounterID) adapter(a int) CounterID      { return id + CounterID(a) }
func (id CounterID) node(k, t int) CounterID      { return id + CounterID(k*numNodeTypes+t) }
func (id CounterID) bucket(k, n, b int) CounterID { return id + CounterID(k*n+b) }
func (id GaugeID) node(k, t int) GaugeID          { return id + GaugeID(k*numNodeTypes+t) }
func (id GaugeID) at(i int) GaugeID               { return id + GaugeID(i) }

// Unit returns the counter for the given storage unit
func (id CounterID) Unit(u int) CounterID { return id.unit(u) }

// Kind returns the counter for the given trie kind
func (id CounterID) Kind(k int) CounterID { return id.kind(k) }

// Adapter returns the counter for the given accounts adapter
func (id CounterID) Adapter(a int) CounterID { return id.adapter(a) }

// Node returns the counter for the given trie kind and node type
func (id CounterID) Node(k, t int) CounterID { return id.node(k, t) }

// DepthBucket returns the counter for the given trie kind and depth
func (id CounterID) DepthBucket(k int, depth uint32) CounterID {
	d := int(depth)
	if d >= numDepthBuckets {
		d = numDepthBuckets - 1
	}
	return id.bucket(k, numDepthBuckets, d)
}

// PersisterBucket returns the counter for the given unit and persister index
func (id CounterID) PersisterBucket(u int, idx int) CounterID {
	if idx >= numPersisterBuckets {
		idx = numPersisterBuckets - 1
	}
	return id.bucket(u, numPersisterBuckets, idx)
}

// At returns the gauge at the given offset
func (id GaugeID) At(i int) GaugeID { return id.at(i) }

// Node returns the gauge for the given trie kind and node type
func (id GaugeID) Node(k, t int) GaugeID { return id.node(k, t) }

// collector ------------------------------------------------------------------

type collector struct {
	cfg     Config
	shardID uint32
	dir     string

	counters []atomic.Int64
	gauges   []atomic.Int64
	maxes    []atomic.Int64

	// trie kind context, set by the processing goroutine around trie operations
	trieKind atomic.Int32

	mut        sync.Mutex
	blockOpen  atomic.Bool
	epoch      uint32
	round      uint64
	nonce      uint64
	blockIndex atomic.Uint64
	blockStart time.Time
	blockCPU   time.Duration
	epochKnown bool

	prev             []int64
	bucketSum        []int64
	bucketMax        []int64
	bucketRunMax     []int64
	bucketBlocks     int64
	bucketFirstNonce uint64
	bucketFirstRound uint64
	bucketWallStart  time.Time

	blocksFile    *csvFile
	outliersFile  *csvFile
	snapshotsFile *csvFile
	reuseFile     *csvFile
	hotFile       *csvFile

	accounts  [numAdapters]*blockSet
	dataTries [numAdapters]*blockSet
	reuse     [numReuse]*reuseTracker
	hot       [numReuse]*hotSet

	cacheProbes [numUnits]func() (entries int, bytes int, persisters int)
	probesMut   sync.RWMutex

	structure [numUnits]*structureStats
	prof      *profiler
}

var (
	global  *collector
	enabled atomic.Bool
)

// Enabled returns true when the collector is active
func Enabled() bool {
	return enabled.Load()
}

// Init creates the global collector and opens the output files. It is a no-op when the config is disabled.
func Init(cfg Config, shardID uint32, workingDir string) error {
	if !cfg.Enabled {
		return nil
	}
	cfg = cfg.withDefaults()

	dir := cfg.OutputDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(workingDir, dir)
	}
	dir = filepath.Join(dir, fmt.Sprintf("shard_%d", shardID))
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}

	c := &collector{
		cfg:          cfg,
		shardID:      shardID,
		dir:          dir,
		counters:     make([]atomic.Int64, len(reg.counterNames)),
		gauges:       make([]atomic.Int64, len(reg.gaugeNames)),
		maxes:        make([]atomic.Int64, len(reg.maxNames)),
		prev:         make([]int64, len(reg.counterNames)),
		bucketSum:    make([]int64, len(reg.counterNames)),
		bucketMax:    make([]int64, len(reg.counterNames)),
		bucketRunMax: make([]int64, len(reg.maxNames)),
		prof: &profiler{
			dir:          filepath.Join(dir, "profiles"),
			everyNEpochs: cfg.ProfileEveryNEpochs,
			numBlocks:    cfg.ProfileNumBlocks,
			lastEpoch:    -1,
		},
	}
	for i := range c.accounts {
		c.accounts[i] = newBlockSet()
		c.dataTries[i] = newBlockSet()
	}
	for i := range c.reuse {
		c.reuse[i] = newReuseTracker(cfg.MaxTrackedKeys)
		c.hot[i] = newHotSet(cfg.MaxTrackedKeys)
	}
	for i := range c.structure {
		c.structure[i] = newStructureStats()
	}

	err = c.openFiles()
	if err != nil {
		return err
	}

	global = c
	enabled.Store(true)
	log.Info("state analysis collector enabled", "dir", dir, "label", cfg.Label, "columns", len(reg.counterNames))
	return nil
}

// Close flushes the open bucket and closes all files
func Close() {
	if !enabled.Load() {
		return
	}
	c := global
	c.mut.Lock()
	defer c.mut.Unlock()

	enabled.Store(false)
	c.prof.stop()
	c.flushBucketUnprotected()
	c.flushEpochUnprotected(c.epoch)
	for _, f := range []*csvFile{c.blocksFile, c.outliersFile, c.snapshotsFile, c.reuseFile, c.hotFile} {
		if f != nil {
			_ = f.close()
		}
	}
	log.Info("state analysis collector closed")
}

func (c *collector) rowHeader() []string {
	header := []string{"shard", "label", "epoch", "round_first", "nonce_first", "nonce_last", "blocks", "elapsed_ns"}
	header = append(header, reg.counterNames...)
	for i, name := range reg.counterNames {
		if reg.counterMax[i] {
			header = append(header, "max_"+name)
		}
	}
	for _, name := range reg.maxNames {
		header = append(header, name)
	}
	header = append(header, reg.gaugeNames...)
	return header
}

func (c *collector) openFiles() error {
	var err error
	c.blocksFile, err = newCsvFile(filepath.Join(c.dir, "blocks.csv.gz"), c.rowHeader())
	if err != nil {
		return err
	}
	c.outliersFile, err = newCsvFile(filepath.Join(c.dir, "outliers.csv.gz"), c.rowHeader())
	if err != nil {
		return err
	}
	c.snapshotsFile, err = newCsvFile(filepath.Join(c.dir, "snapshots.csv.gz"), snapshotHeader())
	if err != nil {
		return err
	}
	reuseHeader := []string{"shard", "label", "epoch", "kind", "touches", "tracked_keys", "resets"}
	for b := 0; b < numReuseBuckets; b++ {
		reuseHeader = append(reuseHeader, "count_b"+itoa(b))
	}
	for b := 0; b < numReuseBuckets; b++ {
		reuseHeader = append(reuseHeader, "bytes_b"+itoa(b))
	}
	c.reuseFile, err = newCsvFile(filepath.Join(c.dir, "reuse.csv.gz"), reuseHeader)
	if err != nil {
		return err
	}
	c.hotFile, err = newCsvFile(filepath.Join(c.dir, "hotset.csv.gz"),
		[]string{"shard", "label", "epoch", "kind", "rank", "key", "count", "distinct_keys", "total_touches", "dropped"})
	return err
}

// Counter API ----------------------------------------------------------------

// Inc increments the counter by one
func Inc(id CounterID) {
	if !enabled.Load() {
		return
	}
	global.counters[id].Add(1)
}

// Add adds the value to the counter
func Add(id CounterID, v int64) {
	if !enabled.Load() {
		return
	}
	global.counters[id].Add(v)
}

// AddSince adds the elapsed time since start to the counter
func AddSince(id CounterID, start time.Time) {
	if !enabled.Load() {
		return
	}
	global.counters[id].Add(int64(time.Since(start)))
}

// Max records the value if it is above the current maximum
func Max(id MaxID, v int64) {
	if !enabled.Load() {
		return
	}
	m := &global.maxes[id]
	for {
		cur := m.Load()
		if v <= cur || m.CompareAndSwap(cur, v) {
			return
		}
	}
}

// SetGauge sets the gauge value
func SetGauge(id GaugeID, v int64) {
	if !enabled.Load() {
		return
	}
	global.gauges[id].Store(v)
}

func (c *collector) SetGauge(id GaugeID, v int64) {
	c.gauges[id].Store(v)
}

// Now returns the current time when enabled and the zero time otherwise, so that hot paths skip the clock read
func Now() time.Time {
	if !enabled.Load() {
		return time.Time{}
	}
	return time.Now()
}

// Trie kind context ----------------------------------------------------------

// SetTrieKind sets the trie kind used to attribute node-level counters and returns the previous one
func SetTrieKind(kind int) int {
	if !enabled.Load() {
		return KindMain
	}
	return int(global.trieKind.Swap(int32(kind)))
}

// TrieKind returns the current trie kind
func TrieKind() int {
	if !enabled.Load() {
		return KindMain
	}
	return int(global.trieKind.Load())
}

// KindForAdapter returns the main trie kind of the given adapter
func KindForAdapter(adapter int) int {
	if adapter == AdapterPeer {
		return KindPeer
	}
	return KindMain
}

// Working set API ------------------------------------------------------------

// TouchAccount records an access to a main trie leaf
func TouchAccount(adapter int, address []byte, size int, isWrite bool) {
	if !enabled.Load() {
		return
	}
	c := global
	c.accounts[adapter].touch(address, isWrite)
	if adapter != AdapterUser {
		return
	}
	blockIndex := c.blockIndex.Load()
	key := string(address)
	c.reuse[ReuseAccount].touch(key, size, blockIndex)
	c.hot[ReuseAccount].touch(key)
}

// TouchDataTrie records an access to a data trie
func TouchDataTrie(adapter int, address []byte, isWrite bool) {
	if !enabled.Load() {
		return
	}
	c := global
	c.dataTries[adapter].touch(address, isWrite)
	if adapter != AdapterUser {
		return
	}
	blockIndex := c.blockIndex.Load()
	key := string(address)
	c.reuse[ReuseDataTrie].touch(key, 0, blockIndex)
	c.hot[ReuseDataTrie].touch(key)
}

// TouchDataTrieKey records an access to a data trie leaf
func TouchDataTrieKey(address []byte, key []byte, size int) {
	if !enabled.Load() {
		return
	}
	c := global
	full := string(address) + string(key)
	blockIndex := c.blockIndex.Load()
	c.reuse[ReuseDataTrieKey].touch(full, size, blockIndex)
	c.hot[ReuseDataTrieKey].touch(full)
}

// RegisterCacheProbe registers a function reporting cache size for the given storage unit
func RegisterCacheProbe(identifier string, probe func() (entries int, bytes int, persisters int)) {
	if !enabled.Load() {
		return
	}
	unit := UnitFromIdentifier(identifier)
	if unit == UnitOther {
		return // only the trie storers are probed, every other storer would overwrite the same slot
	}
	c := global
	c.probesMut.Lock()
	c.cacheProbes[unit] = probe
	c.probesMut.Unlock()
}

// ExcludeFromBlockCommit subtracts instrumentation time from the block commit timer, only while a block is open
// (commits done outside block processing, e.g. genesis, have no matching addition)
func ExcludeFromBlockCommit(ns int64) {
	if !enabled.Load() || !global.blockOpen.Load() {
		return
	}
	Add(CBlockCommitNs, -ns)
}

// ShouldSampleMemory returns true when the current block is a memory sampling block
func ShouldSampleMemory() bool {
	if !enabled.Load() {
		return false
	}
	return global.blockIndex.Load()%uint64(global.cfg.MemorySampleEveryNBlocks) == 0
}

// Block lifecycle ------------------------------------------------------------

// BeginBlock marks the start of the processing of a block
func BeginBlock(shardID uint32, epoch uint32, round uint64, nonce uint64, numTxs uint32, isStartOfEpoch bool) {
	if !enabled.Load() {
		return
	}
	c := global
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.blockOpen.Load() {
		Inc(CBlocksAborted)
		c.endBlockUnprotected()
	}

	if c.epochKnown && epoch != c.epoch {
		c.flushBucketUnprotected()
		c.flushEpochUnprotected(c.epoch)
	}
	c.epochKnown = true
	c.epoch = epoch
	c.round = round
	c.nonce = nonce
	c.blockOpen.Store(true)
	c.blockStart = time.Now()
	c.blockCPU = processCPUTime()
	if c.bucketBlocks == 0 {
		c.bucketFirstNonce = nonce
		c.bucketFirstRound = round
		c.bucketWallStart = c.blockStart
	}
	Add(CBlockTxs, int64(numTxs))
	if isStartOfEpoch {
		Inc(CBlockStartOfEpoch)
	}
	c.prof.maybeStart(epoch, nonce)
}

// EndBlock marks the end of the commit of the current block
func EndBlock() {
	if !enabled.Load() {
		return
	}
	c := global
	c.mut.Lock()
	defer c.mut.Unlock()

	if !c.blockOpen.Load() {
		return
	}
	c.endBlockUnprotected()
}

func (c *collector) endBlockUnprotected() {
	c.blockOpen.Store(false)
	Inc(CBlocks)
	Add(CBlockWallNs, int64(time.Since(c.blockStart)))
	Add(CBlockCPUNs, int64(processCPUTime()-c.blockCPU))

	for a := 0; a < numAdapters; a++ {
		unique, seenPrev, repeat := c.accounts[a].touchRotate()
		Add(CAdbUniqueAccounts.adapter(a), unique)
		Add(CAdbUniqueAccountsPrev.adapter(a), seenPrev)
		Add(CAdbSaveRepeatAddr.adapter(a), repeat)
		unique, seenPrev, _ = c.dataTries[a].touchRotate()
		Add(CAdbUniqueDataTries.adapter(a), unique)
		Add(CAdbUniqueDataTriesPrev.adapter(a), seenPrev)
	}

	if c.blockIndex.Load()%uint64(c.cfg.MemorySampleEveryNBlocks) == 0 {
		c.sampleMemory()
		c.sampleProbes()
	}

	// per block deltas
	delta := make([]int64, len(c.counters))
	for i := range c.counters {
		cur := c.counters[i].Load()
		delta[i] = cur - c.prev[i]
		c.prev[i] = cur
		c.bucketSum[i] += delta[i]
		if delta[i] > c.bucketMax[i] {
			c.bucketMax[i] = delta[i]
		}
	}
	c.bucketBlocks++

	// running maxima are reset at every block; the bucket keeps the largest per-block value seen in the row
	runMax := make([]int64, len(c.maxes))
	for i := range c.maxes {
		runMax[i] = c.maxes[i].Swap(0)
		if runMax[i] > c.bucketRunMax[i] {
			c.bucketRunMax[i] = runMax[i]
		}
	}

	if c.isOutlier(delta) {
		row := c.buildRow(delta, delta, runMax, 1, c.nonce, c.nonce, c.round, int64(time.Since(c.blockStart)))
		_ = c.outliersFile.write(row)
	}

	c.prof.blockDone(c.epoch, c.nonce)
	c.blockIndex.Add(1)

	if c.bucketBlocks >= int64(c.cfg.BlocksPerRow) {
		c.flushBucketUnprotected()
	}
}

func (c *collector) isOutlier(delta []int64) bool {
	if delta[CBlockCommitNs] > c.cfg.OutlierCommitTimeMs*int64(time.Millisecond) {
		return true
	}
	// only the trie units count: import-db reads headers, mini blocks and transactions from other storers
	dbGets := delta[CStDbGet.unit(UnitAccounts)] + delta[CStDbGet.unit(UnitPeer)]
	if dbGets > c.cfg.OutlierStorageReads {
		return true
	}
	for a := 0; a < numAdapters; a++ {
		if delta[CAdbJournalLenAtCommit.adapter(a)] > c.cfg.OutlierJournalLen {
			return true
		}
	}
	return false
}

func (c *collector) sampleProbes() {
	c.probesMut.RLock()
	defer c.probesMut.RUnlock()
	for u, probe := range c.cacheProbes {
		if probe == nil {
			continue
		}
		entries, bytes, persisters := probe()
		c.SetGauge(GStCacheEntries.at(u), int64(entries))
		c.SetGauge(GStCacheBytes.at(u), int64(bytes))
		c.SetGauge(GStActivePersisters.at(u), int64(persisters))
	}
	for k := 0; k < numReuse; k++ {
		c.SetGauge(GTrackedKeys.at(k), int64(c.reuse[k].size()))
	}
}

func (c *collector) buildRow(sum []int64, maxes []int64, runMax []int64, blocks int64, nonceFirst, nonceLast, roundFirst uint64, elapsedNs int64) []string {
	row := make([]string, 0, 8+2*len(sum)+len(c.maxes)+len(c.gauges))
	row = append(row,
		itoa(int(c.shardID)), c.cfg.Label, itoa(int(c.epoch)),
		u64s(roundFirst), u64s(nonceFirst), u64s(nonceLast), i64s(blocks), i64s(elapsedNs))
	for _, v := range sum {
		row = append(row, i64s(v))
	}
	for i, v := range maxes {
		if reg.counterMax[i] {
			row = append(row, i64s(v))
		}
	}
	for _, v := range runMax {
		row = append(row, i64s(v))
	}
	for i := range c.gauges {
		row = append(row, i64s(c.gauges[i].Load()))
	}
	return row
}

func (c *collector) flushBucketUnprotected() {
	if c.bucketBlocks == 0 {
		return
	}
	row := c.buildRow(c.bucketSum, c.bucketMax, c.bucketRunMax, c.bucketBlocks, c.bucketFirstNonce, c.nonce, c.bucketFirstRound, int64(time.Since(c.bucketWallStart)))
	err := c.blocksFile.write(row)
	if err != nil {
		log.Warn("stateAnalysis: cannot write blocks row", "err", err)
	}
	for i := range c.bucketSum {
		c.bucketSum[i] = 0
		c.bucketMax[i] = 0
	}
	for i := range c.bucketRunMax {
		c.bucketRunMax[i] = 0
	}
	c.bucketBlocks = 0
}

func (c *collector) flushEpochUnprotected(epoch uint32) {
	for k := 0; k < numReuse; k++ {
		counts, bytes, tracked, resets, touches := c.reuse[k].flush()
		row := []string{itoa(int(c.shardID)), c.cfg.Label, itoa(int(epoch)), reuseKindNames[k], i64s(touches), itoa(tracked), i64s(resets)}
		for _, v := range counts {
			row = append(row, i64s(v))
		}
		for _, v := range bytes {
			row = append(row, i64s(v))
		}
		_ = c.reuseFile.write(row)

		top, distinct, total, dropped := c.hot[k].flush(c.cfg.TopKPerEpoch)
		for rank, e := range top {
			_ = c.hotFile.write([]string{
				itoa(int(c.shardID)), c.cfg.Label, itoa(int(epoch)), reuseKindNames[k], itoa(rank + 1),
				hexKey(e.key), i64s(e.count), itoa(distinct), i64s(total), i64s(dropped),
			})
		}
	}
}

// the reuse trackers are indexed by kind but hold different entities: main trie leaves, data trie leaves, data tries

func (bs *blockSet) touchRotate() (int64, int64, int64) {
	return bs.rotate()
}

// ResidentStats holds the number of nodes and their estimated size, by node type, that a trie keeps in memory
type ResidentStats struct {
	Nodes [numNodeTypes]int64
	Bytes [numNodeTypes]int64
}

// Add accumulates the given stats
func (rs *ResidentStats) Add(other ResidentStats) {
	for i := 0; i < numNodeTypes; i++ {
		rs.Nodes[i] += other.Nodes[i]
		rs.Bytes[i] += other.Bytes[i]
	}
}

// AddNode records one resident node of the given type
func (rs *ResidentStats) AddNode(nodeType int, size int) {
	rs.Nodes[nodeType]++
	rs.Bytes[nodeType] += int64(size)
}

// ResidentStatsProvider is implemented by tries that can report their resident nodes
type ResidentStatsProvider interface {
	ResidentStats() ResidentStats
}

// SetResidentStats stores the resident stats gauges for the given trie kind
func SetResidentStats(kind int, stats ResidentStats) {
	if !enabled.Load() {
		return
	}
	for t := 0; t < numNodeTypes; t++ {
		global.gauges[GTrieResidentNodes.node(kind, t)].Store(stats.Nodes[t])
		global.gauges[GTrieResidentBytes.node(kind, t)].Store(stats.Bytes[t])
	}
}
