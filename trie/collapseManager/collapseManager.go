package collapseManager

import (
	"container/list"
	"fmt"

	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("trie/collapseManager")

const (
	defaultNumLeavesToCollapseSingleRun = 100
	minSizeInMemory                     = 1048576 // 1 MB
)

type collapseManager struct {
	accessedKeys map[string]*list.Element
	orderAccess  *list.List
	sizeInMemory int

	// Immutable after construction so CloneWithoutState can read these without the trie lock.
	maxSizeInMem                 uint64
	numLeavesToCollapseSingleRun int
}

// NewCollapseManager creates a new collapse manager
func NewCollapseManager(maxSize uint64, numLeavesToCollapseSingleRun uint32) (*collapseManager, error) {
	if maxSize < minSizeInMemory {
		return nil, fmt.Errorf("invalid max size provided: %d, minimum %d", maxSize, minSizeInMemory)
	}

	if numLeavesToCollapseSingleRun == 0 {
		numLeavesToCollapseSingleRun = defaultNumLeavesToCollapseSingleRun
	}

	return &collapseManager{
		accessedKeys:                 make(map[string]*list.Element),
		orderAccess:                  list.New(),
		sizeInMemory:                 0,
		maxSizeInMem:                 maxSize,
		numLeavesToCollapseSingleRun: int(numLeavesToCollapseSingleRun),
	}, nil
}

func (cm *collapseManager) addSizeInMemory(size int) {
	if cm.sizeInMemory+size < 0 {
		log.Warn("trie size in memory is negative after adding size, resetting to 0", "size", size, "currentSize", cm.sizeInMemory)
		cm.sizeInMemory = 0
		return
	}
	cm.sizeInMemory += size
}

// MarkKeyAsAccessed marks a key as accessed, updating its position in the access order
func (cm *collapseManager) MarkKeyAsAccessed(key []byte, sizeLoadedInMemory int) {
	defer cm.addSizeInMemory(sizeLoadedInMemory)

	entry, ok := cm.accessedKeys[string(key)]
	if !ok {
		e := cm.orderAccess.PushFront(key)
		cm.accessedKeys[string(key)] = e

		return
	}

	cm.orderAccess.MoveToFront(entry)
}

// RemoveKey removes a key from the accessed keys list and updates the size in memory
func (cm *collapseManager) RemoveKey(key []byte, sizeLoadedInMemory int) {
	defer cm.addSizeInMemory(sizeLoadedInMemory)

	entry, ok := cm.accessedKeys[string(key)]
	if !ok {
		return
	}

	cm.orderAccess.Remove(entry)
	delete(cm.accessedKeys, string(key))
}

// AddSizeInMemory adds size to the current size in memory
func (cm *collapseManager) AddSizeInMemory(size int) {
	cm.addSizeInMemory(size)
}

// GetSizeInMemory returns the current size in memory
func (cm *collapseManager) GetSizeInMemory() int {
	return cm.sizeInMemory
}

func (cm *collapseManager) ShouldCollapseTrie() bool {
	return uint64(cm.sizeInMemory) > cm.maxSizeInMem
}

// GetCollapsibleLeaves returns a list of keys that can be collapsed to free memory
func (cm *collapseManager) GetCollapsibleLeaves() ([][]byte, error) {
	if uint64(cm.sizeInMemory) < cm.maxSizeInMem {
		return make([][]byte, 0), nil
	}

	evictedKeys := make([][]byte, 0)
	for i := 0; i < cm.numLeavesToCollapseSingleRun; i++ {
		if cm.orderAccess.Len() == 0 {
			break
		}
		entry := cm.orderAccess.Back()
		if entry == nil {
			return nil, fmt.Errorf("unexpected nil entry in collapseManager orderAccess list")
		}
		cm.orderAccess.Remove(entry)
		keyBytes, ok := entry.Value.([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid key type in collapseManager orderAccess list: %T", entry.Value)
		}
		delete(cm.accessedKeys, string(keyBytes))

		evictedKeys = append(evictedKeys, keyBytes)
	}

	return evictedKeys, nil
}

// CloneWithoutState creates a new collapse manager with the same configuration but without the current state
func (cm *collapseManager) CloneWithoutState() common.TrieCollapseManager {
	return &collapseManager{
		accessedKeys:                 make(map[string]*list.Element),
		orderAccess:                  list.New(),
		sizeInMemory:                 0,
		maxSizeInMem:                 cm.maxSizeInMem,
		numLeavesToCollapseSingleRun: cm.numLeavesToCollapseSingleRun,
	}
}

func (cm *collapseManager) Reset() {
	// Replace the map to release its retained capacity.
	cm.accessedKeys = make(map[string]*list.Element)
	cm.orderAccess.Init()
	cm.sizeInMemory = 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (cm *collapseManager) IsInterfaceNil() bool {
	return cm == nil
}
