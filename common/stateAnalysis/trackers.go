package stateAnalysis

import (
	"encoding/hex"
	"math/bits"
	"sort"
	"sync"
)

const (
	numReuseBuckets = 23 // 0: same block, 1..21: log2 distance, 22: never seen before
	reuseBucketNew  = numReuseBuckets - 1
)

// reuseTracker records, for every touched key, the distance in blocks since the key was last touched.
// Keys are kept in a bounded map; when the map is full it is dropped and the number of resets is recorded.
type reuseTracker struct {
	mut     sync.Mutex
	last    map[string]uint64
	maxKeys int
	resets  int64
	counts  [numReuseBuckets]int64
	bytes   [numReuseBuckets]int64
	touches int64
}

func newReuseTracker(maxKeys int) *reuseTracker {
	return &reuseTracker{
		last:    make(map[string]uint64),
		maxKeys: maxKeys,
	}
}

func distanceBucket(d uint64) int {
	if d == 0 {
		return 0
	}
	b := bits.Len64(d) // 1 for d=1, 2 for d in [2,3], ...
	if b > numReuseBuckets-2 {
		b = numReuseBuckets - 2
	}
	return b
}

func (rt *reuseTracker) touch(key []byte, size int, currentBlock uint64) {
	rt.mut.Lock()
	defer rt.mut.Unlock()

	rt.touches++
	lastBlock, found := rt.last[string(key)]
	bucket := reuseBucketNew
	if found {
		bucket = distanceBucket(currentBlock - lastBlock)
	}
	rt.counts[bucket]++
	rt.bytes[bucket] += int64(size)

	if !found && len(rt.last) >= rt.maxKeys {
		rt.last = make(map[string]uint64)
		rt.resets++
	}
	rt.last[string(key)] = currentBlock
}

func (rt *reuseTracker) flush() (counts, bytes [numReuseBuckets]int64, tracked int, resets int64, touches int64) {
	rt.mut.Lock()
	defer rt.mut.Unlock()

	counts = rt.counts
	bytes = rt.bytes
	tracked = len(rt.last)
	resets = rt.resets
	touches = rt.touches
	rt.counts = [numReuseBuckets]int64{}
	rt.bytes = [numReuseBuckets]int64{}
	rt.resets = 0
	rt.touches = 0
	return
}

func (rt *reuseTracker) size() int {
	rt.mut.Lock()
	defer rt.mut.Unlock()
	return len(rt.last)
}

// hotSet counts accesses per key within an epoch and reports the top K keys
type hotSet struct {
	mut     sync.Mutex
	counts  map[string]int64
	maxKeys int
	dropped int64
}

func newHotSet(maxKeys int) *hotSet {
	return &hotSet{
		counts:  make(map[string]int64),
		maxKeys: maxKeys,
	}
}

func (hs *hotSet) touch(key []byte) {
	hs.mut.Lock()
	defer hs.mut.Unlock()

	cnt, found := hs.counts[string(key)]
	if !found && len(hs.counts) >= hs.maxKeys {
		hs.dropped++
		return
	}
	hs.counts[string(key)] = cnt + 1
}

type hotEntry struct {
	key   string
	count int64
}

func (hs *hotSet) flush(topK int) (top []hotEntry, distinct int, total int64, dropped int64) {
	hs.mut.Lock()
	counts := hs.counts
	dropped = hs.dropped
	hs.counts = make(map[string]int64)
	hs.dropped = 0
	hs.mut.Unlock()

	distinct = len(counts)
	entries := make([]hotEntry, 0, len(counts))
	for k, c := range counts {
		total += c
		entries = append(entries, hotEntry{key: k, count: c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count == entries[j].count {
			return entries[i].key < entries[j].key
		}
		return entries[i].count > entries[j].count
	})
	if len(entries) > topK {
		entries = entries[:topK]
	}
	return entries, distinct, total, dropped
}

func hexKey(k string) string {
	return hex.EncodeToString([]byte(k))
}

// blockSet tracks the distinct keys touched in the current block and in the previous one
type blockSet struct {
	mut  sync.Mutex
	cur  map[string]int32
	prev map[string]int32
}

func newBlockSet() *blockSet {
	return &blockSet{
		cur:  make(map[string]int32),
		prev: make(map[string]int32),
	}
}

func (bs *blockSet) touch(key []byte, isWrite bool) {
	bs.mut.Lock()
	defer bs.mut.Unlock()

	v, found := bs.cur[string(key)]
	if !found {
		bs.cur[string(key)] = 0
		v = 0
	}
	if isWrite {
		bs.cur[string(key)] = v + 1
	}
}

// rotate returns the number of distinct keys, how many of them were touched in the previous block,
// and how many extra writes (beyond the first) were made to the same key in the block
func (bs *blockSet) rotate() (unique int64, seenPrev int64, repeatWrites int64) {
	bs.mut.Lock()
	defer bs.mut.Unlock()

	unique = int64(len(bs.cur))
	for k, writes := range bs.cur {
		if _, ok := bs.prev[k]; ok {
			seenPrev++
		}
		if writes > 1 {
			repeatWrites += int64(writes - 1)
		}
	}
	bs.prev = bs.cur
	bs.cur = make(map[string]int32, len(bs.prev))
	return
}
