package triesHolder

import (
	"fmt"
	"math"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const maxTrieSizeMinValue = 1 * 1024 * 1024 // 1 MB

var log = logger.GetOrCreate("state/dataTriesHolder")

// dataTriesHolder is a structure that holds a map of tries and manages their memory usage
// It uses a doubly linked list to keep track of the order in which the tries were used
// and evicts the oldest used tries when the total size exceeds a maximum limit.
type dataTriesHolder struct {
	cacher        storage.AdaptedSizedLRUCache
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

	c, err := cache.NewCapacityLRU(math.MaxInt, int64(maxTriesSize))
	if err != nil {
		return nil, err
	}

	return &dataTriesHolder{
		cacher:        c,
		dirtyTries:    make(map[string]struct{}),
		touchedTries:  make(map[string]struct{}),
		evictedBuffer: make(map[string]common.Trie),
		mutex:         sync.RWMutex{},
	}, nil
}

// Put adds a trie pointer to the tries map
func (dth *dataTriesHolder) Put(key []byte, tr common.Trie) {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	dth.putUnprotected(key, tr)
}

func (dth *dataTriesHolder) putUnprotected(key []byte, tr common.Trie) {
	keyString := string(key)

	if len(dth.evictedBuffer) > 0 {
		if _, ok := dth.evictedBuffer[keyString]; ok {
			// this means that this trie was evicted while being dirty
			delete(dth.evictedBuffer, keyString)
		}
	}

	evicted := dth.cacher.AddSizedAndReturnEvicted(keyString, tr, int64(tr.SizeInMemory()))
	dth.dirtyTries[keyString] = struct{}{}
	dth.touchedTries[keyString] = struct{}{}

	if len(evicted) == 0 {
		return
	}

	for evictedKey, evictedValue := range evicted {
		evictedKeyString, ok := evictedKey.(string)
		if !ok {
			log.Warn("invalid data in dataTriesHolder cache", "entry type", fmt.Sprintf("%T", evictedKey))
			continue
		}
		_, ok = dth.dirtyTries[evictedKeyString]
		if !ok {
			continue
		}

		evictedTrie, ok := evictedValue.(common.Trie)
		if !ok {
			log.Warn("invalid data in dataTriesHolder cache", "entry type", fmt.Sprintf("%T", evictedValue))
			continue
		}
		dth.evictedBuffer[evictedKeyString] = evictedTrie
	}
}

// Get returns the trie pointer that is stored in the map at the given key
func (dth *dataTriesHolder) Get(key []byte) common.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()
	keyString := string(key)

	val, ok := dth.cacher.Get(keyString)
	if !ok {
		// maybe it was evicted while being dirty
		evictedTr, exists := dth.evictedBuffer[keyString]
		if !exists {
			return nil
		}

		delete(dth.evictedBuffer, keyString)
		dth.putUnprotected(key, evictedTr)
		return evictedTr
	}

	dth.touchedTries[keyString] = struct{}{}
	tr, ok := val.(common.Trie)
	if !ok {
		log.Warn("invalid data in dataTriesHolder cache", "entry type", fmt.Sprintf("%T", val))
		return nil
	}

	return tr
}

// GetAll returns all the tries that are marked as dirty for this implementation.
// It also resets their dirty flag and recomputes the total size.
func (dth *dataTriesHolder) GetAll() []common.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	tries := make([]common.Trie, 0)
	for keyString := range dth.dirtyTries {
		tr := dth.getDirtyTrie(keyString)
		if check.IfNil(tr) {
			continue
		}
		tries = append(tries, tr)
	}
	dth.dirtyTries = make(map[string]struct{})
	dth.evictedBuffer = make(map[string]common.Trie)
	dth.recomputeTotalSize()
	return tries
}

func (dth *dataTriesHolder) getDirtyTrie(key string) common.Trie {
	entry, exists := dth.cacher.Get(key)
	if exists {
		tr, ok := entry.(common.Trie)
		if !ok {
			log.Warn("invalid data in dataTriesHolder cache", "entry type", fmt.Sprintf("%T", entry))
			return nil
		}

		return tr
	}

	tr, ok := dth.evictedBuffer[key]
	if !ok {
		return nil
	}
	return tr
}

func (dth *dataTriesHolder) recomputeTotalSize() {
	for keyString := range dth.touchedTries {
		entry, exists := dth.cacher.Get(keyString)
		if !exists {
			continue
		}
		tr, ok := entry.(common.Trie)
		if !ok {
			log.Warn("invalid data in dataTriesHolder cache", "entry type", fmt.Sprintf("%T", entry))
			continue
		}

		evicted := dth.cacher.AddSized(keyString, tr, int64(tr.SizeInMemory()))
		if evicted {
			log.Warn("unexpected eviction while recomputing total size in dataTriesHolder")
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
	log.Trace("reset data tries holder")

	dth.cacher.Purge()
	dth.dirtyTries = make(map[string]struct{})
	dth.touchedTries = make(map[string]struct{})
	dth.evictedBuffer = make(map[string]common.Trie)
}

// IsInterfaceNil returns true if underlying object is nil
func (dth *dataTriesHolder) IsInterfaceNil() bool {
	return dth == nil
}
