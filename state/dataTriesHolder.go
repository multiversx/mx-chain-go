package state

import (
	"bytes"
	"sync"

	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// TODO: add size cap, eviction policy and unit tests

type trieEntry struct {
	trie      common.Trie
	key       []byte
	isDirty   bool
	nextEntry *trieEntry
	prevEntry *trieEntry
}

type dataTriesHolder struct {
	tries          map[string]*trieEntry
	oldestUsed     *trieEntry
	newestUsed     *trieEntry
	totalTriesSize uint64

	mutex sync.RWMutex
}

// NewDataTriesHolder creates a new instance of dataTriesHolder
func NewDataTriesHolder() *dataTriesHolder {
	return &dataTriesHolder{
		tries:          make(map[string]*trieEntry),
		oldestUsed:     nil,
		newestUsed:     nil,
		totalTriesSize: 0,
		mutex:          sync.RWMutex{},
	}
}

// Put adds a trie pointer to the tries map
func (dth *dataTriesHolder) Put(key []byte, tr common.Trie) {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	log.Trace("put trie in data tries holder", "key", key)

	if len(dth.tries) == 0 {
		// If the tries map is empty, we create a new entry
		entry := &trieEntry{
			trie:      tr,
			key:       key,
			isDirty:   true,
			nextEntry: nil,
			prevEntry: nil,
		}
		dth.tries[string(key)] = entry
		dth.oldestUsed = entry
		dth.newestUsed = entry
		dth.totalTriesSize += uint64(tr.SizeInMemory())
		return
	}

	entry, exists := dth.tries[string(key)]
	if !exists {
		// If the entry does not exist, we create a new one
		entry = &trieEntry{
			trie:      tr,
			key:       key,
			isDirty:   true,
			nextEntry: nil,
			prevEntry: dth.newestUsed,
		}
		dth.newestUsed.nextEntry = entry

		dth.newestUsed = entry
		dth.tries[string(key)] = entry
		dth.totalTriesSize += uint64(tr.SizeInMemory())

		return
	}

	if bytes.Equal(dth.newestUsed.key, key) {
		// If the entry is already the newest used, mark it as dirty and return
		entry.isDirty = true
		return
	}

	dth.moveEntryToNewestUsed(entry)
}

func (dth *dataTriesHolder) moveEntryToNewestUsed(entry *trieEntry) {
	if bytes.Equal(dth.oldestUsed.key, entry.key) {
		if bytes.Equal(dth.newestUsed.key, entry.key) {
			// If the entry is both the oldest and newest used, we just mark it as dirty
			entry.isDirty = true
			return
		}

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
		entry.isDirty = true
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
	entry.isDirty = true
}

// Get returns the trie pointer that is stored in the map at the given key
func (dth *dataTriesHolder) Get(key []byte) common.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	entry, exists := dth.tries[string(key)]
	if !exists {
		return nil
	}

	dth.moveEntryToNewestUsed(entry)
	return entry.trie
}

// GetAllDirty returns all the tries that are marked as dirty
func (dth *dataTriesHolder) GetAllDirty() []common.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	tries := make([]common.Trie, 0)
	shouldStop := false
	entry := dth.newestUsed
	for entry != nil && !shouldStop {
		if !entry.isDirty {
			shouldStop = true
			continue
		}
		tries = append(tries, entry.trie)
		entry.isDirty = false // Reset the dirty flag after retrieving
		entry = entry.prevEntry
	}

	return tries
}

// Reset clears the tries map
func (dth *dataTriesHolder) Reset() {
	dth.mutex.Lock()

	if log.GetLevel() == logger.LogTrace {
		for key := range dth.tries {
			log.Trace("reset data tries holder", "key", key)
		}
	}

	dth.tries = make(map[string]*trieEntry)
	dth.oldestUsed = nil
	dth.newestUsed = nil
	dth.totalTriesSize = 0

	dth.mutex.Unlock()
}

// IsInterfaceNil returns true if underlying object is nil
func (dth *dataTriesHolder) IsInterfaceNil() bool {
	return dth == nil
}
