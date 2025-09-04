package state

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const maxTrieSizeMinValue = 1 * 1024 * 1024 // 1 MB

type trieEntry struct {
	trie      common.Trie
	trieSize  int
	key       []byte
	nextEntry *trieEntry
	prevEntry *trieEntry
}

// dataTriesHolder is a structure that holds a map of tries and manages their memory usage
// It uses a doubly linked list to keep track of the order in which the tries were used
// and evicts the oldest used tries when the total size exceeds a maximum limit.
type dataTriesHolder struct {
	tries          map[string]*trieEntry
	dirtyTries     map[string]struct{} // These are the tries that have been modified and need to be persisted
	touchedTries   map[string]struct{} // These are needed to compute an accurate totalTriesSize
	oldestUsed     *trieEntry
	newestUsed     *trieEntry
	totalTriesSize uint64
	maxTriesSize   uint64

	mutex sync.RWMutex
}

// NewDataTriesHolder creates a new instance of dataTriesHolder
func NewDataTriesHolder(maxTriesSize uint64) (*dataTriesHolder, error) {
	if maxTriesSize < maxTrieSizeMinValue {
		return nil, fmt.Errorf("%w, provided %d, minimum %d", ErrInvalidMaxTrieSizeValue, maxTriesSize, maxTrieSizeMinValue)
	}

	return &dataTriesHolder{
		tries:          make(map[string]*trieEntry),
		dirtyTries:     make(map[string]struct{}),
		touchedTries:   make(map[string]struct{}),
		oldestUsed:     nil,
		newestUsed:     nil,
		totalTriesSize: 0,
		maxTriesSize:   maxTriesSize,
		mutex:          sync.RWMutex{},
	}, nil
}

// Put adds a trie pointer to the tries map
func (dth *dataTriesHolder) Put(key []byte, tr common.Trie) {
	dth.mutex.Lock()
	defer func() {
		// If the total size of the tries exceeds the maximum size, we need to evict the oldest used tries
		dth.evictIfNeeded()
		dth.mutex.Unlock()
	}()

	log.Trace("put trie in data tries holder", "key", key)

	if len(dth.tries) == 0 {
		// If the tries map is empty, we create a new entry
		entry := &trieEntry{
			trie:      tr,
			trieSize:  tr.SizeInMemory(),
			key:       key,
			nextEntry: nil,
			prevEntry: nil,
		}
		dth.tries[string(key)] = entry
		dth.dirtyTries[string(key)] = struct{}{}
		dth.oldestUsed = entry
		dth.newestUsed = entry
		dth.totalTriesSize += uint64(entry.trieSize)
		return
	}

	entry, exists := dth.tries[string(key)]
	if !exists {
		// If the entry does not exist, we create a new one
		entry = &trieEntry{
			trie:      tr,
			trieSize:  tr.SizeInMemory(),
			key:       key,
			nextEntry: nil,
			prevEntry: dth.newestUsed,
		}
		dth.newestUsed.nextEntry = entry

		dth.newestUsed = entry
		dth.tries[string(key)] = entry
		dth.dirtyTries[string(key)] = struct{}{}
		dth.totalTriesSize += uint64(entry.trieSize)

		return
	}

	dth.touchedTries[string(key)] = struct{}{}
	dth.dirtyTries[string(key)] = struct{}{}
	dth.moveEntryToNewestUsed(entry)
}

func (dth *dataTriesHolder) evictIfNeeded() {
	if dth.totalTriesSize <= dth.maxTriesSize {
		return
	}

	if dth.oldestUsed == nil {
		// This should never happen, but we check it just in case
		log.Error("data tries holder is in an invalid state: totalTriesSize exceeds maxTriesSize but oldestUsed is nil")
		dth.reset()
		return
	}

	// We evict the oldest used entries until the total size is less than or equal to the maximum size
	entryForEviction := dth.oldestUsed
	for dth.totalTriesSize > dth.maxTriesSize && entryForEviction != nil {
		_, isDirty := dth.dirtyTries[string(entryForEviction.key)]
		if isDirty {
			entryForEviction = entryForEviction.nextEntry
			continue
		}

		log.Trace("evicting trie from data tries holder", "key", entryForEviction.key)

		// Remove entry from the map and update the links between the entries
		delete(dth.tries, string(entryForEviction.key))
		delete(dth.touchedTries, string(entryForEviction.key))
		dth.totalTriesSize -= uint64(entryForEviction.trieSize)

		if bytes.Equal(entryForEviction.key, dth.oldestUsed.key) {
			if bytes.Equal(entryForEviction.key, dth.newestUsed.key) {
				// If the entry to be evicted is the only entry, we need to reset the oldest and newest used pointers
				dth.oldestUsed = nil
				dth.newestUsed = nil
				return
			}

			// If the entry to be evicted is the oldest used, we need to update the oldest used pointer
			dth.oldestUsed = dth.oldestUsed.nextEntry
			dth.oldestUsed.prevEntry = nil
			entryForEviction = dth.oldestUsed
			continue
		}

		if bytes.Equal(entryForEviction.key, dth.newestUsed.key) {
			// If the entry to be evicted is the newest used, we need to update the newest used pointer and return
			dth.newestUsed = dth.newestUsed.prevEntry
			dth.newestUsed.nextEntry = nil
			return
		}

		// If the entry to be evicted is neither the oldest nor the newest used, we need to update the links between the neighbors
		entryForEviction.prevEntry.nextEntry = entryForEviction.nextEntry
		entryForEviction.nextEntry.prevEntry = entryForEviction.prevEntry
		entryForEviction = entryForEviction.nextEntry
	}
}

func (dth *dataTriesHolder) moveEntryToNewestUsed(entry *trieEntry) {
	if bytes.Equal(dth.newestUsed.key, entry.key) {
		return
	}

	if bytes.Equal(dth.oldestUsed.key, entry.key) {
		// If the entry is the oldest used, we need to update the oldest used pointer and move the entry to the
		// newest used position
		dth.oldestUsed = entry.nextEntry
		if dth.oldestUsed != nil {
			dth.oldestUsed.prevEntry = nil
		}

		entry.prevEntry = dth.newestUsed
		entry.nextEntry = nil
		dth.newestUsed.nextEntry = entry
		dth.newestUsed = entry
		return
	}

	// If the entry is neither the oldest nor the newest used, we need to move it to the newest used position
	// and update the neighbors

	entry.prevEntry.nextEntry = entry.nextEntry
	entry.nextEntry.prevEntry = entry.prevEntry
	entry.prevEntry = dth.newestUsed
	entry.nextEntry = nil
	dth.newestUsed.nextEntry = entry
	dth.newestUsed = entry
}

// Get returns the trie pointer that is stored in the map at the given key
func (dth *dataTriesHolder) Get(key []byte) common.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	entry, exists := dth.tries[string(key)]
	if !exists {
		return nil
	}

	dth.touchedTries[string(key)] = struct{}{}
	dth.moveEntryToNewestUsed(entry)
	return entry.trie
}

// GetAllDirtyAndResetFlag returns all the tries that are marked as dirty. It also resets their dirty flag and recomputes the total size.
func (dth *dataTriesHolder) GetAllDirtyAndResetFlag() []common.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	tries := make([]common.Trie, 0)
	for key := range dth.dirtyTries {
		entry, exists := dth.tries[key]
		if !exists {
			log.Warn("data tries holder is in an invalid state: dirty trie not found in tries map", "key", key)
			continue
		}
		tries = append(tries, entry.trie)
	}
	dth.dirtyTries = make(map[string]struct{})
	dth.recomputeTotalSize()
	dth.evictIfNeeded()
	return tries
}

func (dth *dataTriesHolder) recomputeTotalSize() {
	for key := range dth.touchedTries {
		entry, exists := dth.tries[key]
		if !exists {
			log.Warn("data tries holder is in an invalid state: touched trie not found in tries map", "key", key)
			continue
		}
		oldSize := entry.trieSize
		newSize := entry.trie.SizeInMemory()
		if oldSize != newSize {
			dth.totalTriesSize = dth.totalTriesSize - uint64(oldSize) + uint64(newSize)
			entry.trieSize = newSize
		}
	}
	dth.touchedTries = make(map[string]struct{})
}

// Reset clears the tries map
func (dth *dataTriesHolder) Reset() {
	dth.mutex.Lock()
	dth.reset()
	dth.mutex.Unlock()
}

func (dth *dataTriesHolder) reset() {
	if log.GetLevel() == logger.LogTrace {
		for key := range dth.tries {
			log.Trace("reset data tries holder", "key", key)
		}
	}

	dth.tries = make(map[string]*trieEntry)
	dth.dirtyTries = make(map[string]struct{})
	dth.oldestUsed = nil
	dth.newestUsed = nil
	dth.totalTriesSize = 0
}

// IsInterfaceNil returns true if underlying object is nil
func (dth *dataTriesHolder) IsInterfaceNil() bool {
	return dth == nil
}
