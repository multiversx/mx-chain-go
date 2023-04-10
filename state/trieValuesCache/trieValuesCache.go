package trieValuesCache

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("state/trieValuesCache")

// trieValuesCache is a cache that holds the values of the trie when they are retrieved.
// It is used to avoid multiple trie lookups for the same key. This struct is not concurrent safe.
type trieValuesCache struct {
	cache         map[string]core.TrieData
	numEntries    int
	maxNumEntries int
}

const minNumEntries = 10

// NewTrieValuesCache creates a new instance of trieValuesCache
func NewTrieValuesCache(maxNumEntries int) (*trieValuesCache, error) {
	if maxNumEntries < minNumEntries {
		return nil, fmt.Errorf("maxNumEntries should be at least %d, given num entries %d", minNumEntries, maxNumEntries)
	}

	return &trieValuesCache{
		cache:         make(map[string]core.TrieData),
		numEntries:    0,
		maxNumEntries: maxNumEntries,
	}, nil
}

// Put adds a new entry in the cache
func (tvc *trieValuesCache) Put(key []byte, value core.TrieData) {
	if tvc.numEntries >= tvc.maxNumEntries {
		log.Warn("trieValuesCache is full", "numEntries", tvc.numEntries, "maxNumEntries", tvc.maxNumEntries, "key", key)
		return
	}

	tvc.cache[string(key)] = value
	tvc.numEntries++
}

// Get returns the value associated with the given key
func (tvc *trieValuesCache) Get(key []byte) (core.TrieData, bool) {
	value, ok := tvc.cache[string(key)]
	return value, ok
}

// Clean removes all entries from the cache
func (tvc *trieValuesCache) Clean() {
	tvc.cache = make(map[string]core.TrieData)
	tvc.numEntries = 0

	log.Trace("trieValuesCache cleaned", "numEntries", tvc.numEntries)
}

// IsInterfaceNil returns true if there is no value under the interface
func (tvc *trieValuesCache) IsInterfaceNil() bool {
	return tvc == nil
}
