package triesHolder

import (
	"fmt"
	"github.com/multiversx/mx-chain-go/common/replaymetrics"
	"math"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
	storageCommon "github.com/multiversx/mx-chain-storage-go/common"
)

const maxTrieSizeMinValue = 1 * 1024 * 1024 // 1 MB

var log = logger.GetOrCreate("state/dataTriesHolder")

// dataTriesHolder is a structure that holds a map of tries and manages their memory usage
// It uses a doubly linked list to keep track of the order in which the tries were used
// and evicts the oldest used tries when the total size exceeds a maximum limit.
type dataTriesHolder struct {
	replayMetrics bool
	cacher        dataTrieCache
	dirtyTries    map[string]struct{}    // These are the tries that have been modified and need to be persisted
	touchedTries  map[string]struct{}    // These are needed to compute an accurate totalTriesSize
	evictedBuffer map[string]common.Trie // in case eviction happens for a dirty trie, we keep it here until we commit it

	mutex sync.RWMutex
}

// NewDataTriesHolder creates a new instance of dataTriesHolder
func NewDataTriesHolder(maxTriesSize uint64) (*dataTriesHolder, error) {
	if maxTriesSize < maxTrieSizeMinValue {
		return nil, fmt.Errorf("%w, provided %d, minimum %d", ErrInvalidMaxTrieSizeValue, maxTriesSize, maxTrieSizeMinValue)
	}
	log.Trace("creating new data tries holder", "maxTriesSize", maxTriesSize)

	if maxTriesSize > math.MaxInt64 {
		return nil, storageCommon.ErrCacheCapacityInvalid
	}

	return &dataTriesHolder{
		cacher:        newDataTrieCache(maxTriesSize),
		dirtyTries:    make(map[string]struct{}),
		touchedTries:  make(map[string]struct{}),
		evictedBuffer: make(map[string]common.Trie),
		mutex:         sync.RWMutex{},
	}, nil
}

// SetReplayMetrics enables measurements for the processing holder.
func (dth *dataTriesHolder) SetReplayMetrics(enabled bool) {
	dth.replayMetrics = enabled
}

// Put adds a trie pointer to the tries map
func (dth *dataTriesHolder) Put(key []byte, tr common.Trie) {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	dth.putNoLock(key, tr)
}

func (dth *dataTriesHolder) putNoLock(key []byte, tr common.Trie) {
	if check.IfNil(tr) || len(key) == 0 {
		log.Warn("trying to put nil trie or empty key in dataTriesHolder", "key", key, "trie", tr)
		return
	}

	size := tr.SizeInMemory()
	if size < 0 {
		log.Warn("invalid data trie size", "key", key, "size", size)
		return
	}

	keyString := string(key)
	if _, wasEvicted := dth.evictedBuffer[keyString]; wasEvicted {
		delete(dth.evictedBuffer, keyString)
		log.Trace("removed trie from evicted buffer", "key", key)
	}
	dth.cacher.put(keyString, tr, uint64(size))
	dth.dirtyTries[keyString] = struct{}{}
	dth.touchedTries[keyString] = struct{}{}
	dth.evictOverBudget()

	if log.GetLevel() == logger.LogTrace {
		log.Trace("put trie in data tries holder", "key", key, "trieSize", size, "totalCacheSize", dth.cacher.size)
	}
}

func (dth *dataTriesHolder) evictOverBudget() {
	var replay replaymetrics.Scope
	if dth.replayMetrics {
		replay = replaymetrics.Capture(replaymetrics.Data)
	}
	for evicted := dth.cacher.evictOldestOverBudget(); evicted != nil; evicted = dth.cacher.evictOldestOverBudget() {
		replay.Add(replaymetrics.CacheEvictions, 1)
		log.Trace("trie evicted from dataTriesHolder cache", "key", []byte(evicted.key))
		if _, dirty := dth.dirtyTries[evicted.key]; dirty {
			log.Trace("storing evicted dirty trie in evicted buffer", "key", []byte(evicted.key))
			dth.evictedBuffer[evicted.key] = evicted.trie
			replay.Add(replaymetrics.DirtyEvictions, 1)
		}
	}
	replay.CacheSize(dth.cacher.order.Len(), dth.cacher.size, len(dth.evictedBuffer), 0)
}

// Get returns the trie pointer that is stored in the map at the given key
func (dth *dataTriesHolder) Get(key []byte) common.Trie {
	if len(key) == 0 {
		return nil
	}

	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	keyString := string(key)

	tr := dth.cacher.get(keyString)
	if tr == nil {
		// maybe it was evicted while being dirty
		evictedTr, exists := dth.evictedBuffer[keyString]
		if !exists {
			log.Trace("trie not found in data tries holder", "key", key)
			return nil
		}

		dth.putNoLock(key, evictedTr)
		return evictedTr
	}

	dth.touchedTries[keyString] = struct{}{}
	log.Trace("trie found in data tries holder cache", "key", key)

	return tr
}

// GetAll returns all the tries that are marked as dirty for this implementation.
// Pending tries remain retained until MarkAllAsCommitted or an explicit reset.
func (dth *dataTriesHolder) GetAll() []common.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	tries := make([]common.Trie, 0)
	for keyString := range dth.dirtyTries {
		tr := dth.getDirtyTrieNoLock(keyString)
		if check.IfNil(tr) {
			continue
		}
		log.Trace("getting dirty trie from data tries holder", "key", []byte(keyString))
		tries = append(tries, tr)
	}
	if dth.replayMetrics {
		replay := replaymetrics.Capture(replaymetrics.Data)
		if replay.Active() {
			var pendingBytes uint64
			for _, trie := range dth.evictedBuffer {
				pendingBytes += uint64(max(trie.SizeInMemory(), 0))
			}
			replay.CacheSize(dth.cacher.order.Len(), dth.cacher.size, len(dth.evictedBuffer), pendingBytes)
		}
	}
	log.Trace("data tries holder returned all dirty tries", "numTries", len(tries), "totalCacheSize", dth.cacher.size)
	return tries
}

// MarkAllAsCommitted releases evicted tries and reconciles cache sizes only after
// all pending data tries have been successfully persisted.
func (dth *dataTriesHolder) MarkAllAsCommitted() {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	dth.dirtyTries = make(map[string]struct{})
	dth.evictedBuffer = make(map[string]common.Trie)
	dth.recomputeTotalSize()
}

func (dth *dataTriesHolder) getDirtyTrieNoLock(key string) common.Trie {
	entry := dth.cacher.peek(key)
	if entry != nil {
		return entry.trie
	}
	return dth.evictedBuffer[key]
}

func (dth *dataTriesHolder) recomputeTotalSize() {
	for keyString := range dth.touchedTries {
		entry := dth.cacher.peek(keyString)
		if entry == nil {
			continue
		}

		// Only committed empty tries can be dropped from the cache.
		rootHash, err := entry.trie.RootHash()
		if err == nil && common.IsEmptyTrie(rootHash) {
			dth.cacher.remove(keyString)
			continue
		}

		size := entry.trie.SizeInMemory()
		if size < 0 {
			log.Warn("invalid data trie size", "key", []byte(keyString), "size", size)
			continue
		}
		dth.cacher.setSize(entry, uint64(size))
	}
	// Apply every size change before eviction so growing tries can use space
	// released by tries that shrank or became empty in the same commit.
	dth.evictOverBudget()
	dth.touchedTries = make(map[string]struct{})
}

// Remove evicts the trie associated with the given key from the holder.
// This must be called when an account is deleted so that a subsequent recreation
// of the same account does not inherit the stale data trie from the previous incarnation.
func (dth *dataTriesHolder) Remove(key []byte) {
	if len(key) == 0 {
		return
	}

	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	keyString := string(key)
	dth.cacher.remove(keyString)
	delete(dth.dirtyTries, keyString)
	delete(dth.touchedTries, keyString)
	delete(dth.evictedBuffer, keyString)
	log.Trace("removed trie from data tries holder", "key", key)
}

// Reset clears the tries map
func (dth *dataTriesHolder) Reset() {
	dth.mutex.Lock()
	dth.reset()
	if dth.replayMetrics {
		replaymetrics.Capture(replaymetrics.Data).Add(replaymetrics.CacheResets, 1)
	}
	log.Trace("data tries holder reset")
	dth.mutex.Unlock()
}

// MarkAsDirty marks the trie at the given key as dirty
func (dth *dataTriesHolder) MarkAsDirty(key []byte) {
	dth.mutex.Lock()
	stringKey := string(key)
	dth.dirtyTries[stringKey] = struct{}{}
	dth.touchedTries[stringKey] = struct{}{}
	log.Trace("marked trie as dirty in data tries holder", "key", key)
	dth.mutex.Unlock()
}

func (dth *dataTriesHolder) reset() {
	log.Trace("reset data tries holder")

	dth.cacher.reset()
	dth.dirtyTries = make(map[string]struct{})
	dth.touchedTries = make(map[string]struct{})
	dth.evictedBuffer = make(map[string]common.Trie)
}

// IsInterfaceNil returns true if underlying object is nil
func (dth *dataTriesHolder) IsInterfaceNil() bool {
	return dth == nil
}
