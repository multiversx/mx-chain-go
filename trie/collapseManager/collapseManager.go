package collapseManager

import (
	"container/list"
	"fmt"

	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("trie/collapseManager")

const (
	// TODO calibrate these values
	minNumLeavesToCollapseTrie = 1000
	minSizeInMemory            = 1048576 // 1 MB
)

type collapseManager struct {
	accessedKeys map[string]*list.Element
	orderAccess  *list.List
	sizeInMemory int
	maxSizeInMem uint64
}

// NewCollapseManager creates a new collapse manager
func NewCollapseManager(maxSize uint64) (*collapseManager, error) {
	if maxSize < minSizeInMemory {
		return nil, fmt.Errorf("invalid max size provided: %d, minimum %d", maxSize, minSizeInMemory)
	}

	return &collapseManager{
		accessedKeys: make(map[string]*list.Element),
		orderAccess:  list.New(),
		sizeInMemory: 0,
		maxSizeInMem: maxSize,
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

func (cm *collapseManager) RemoveKey(key []byte, sizeLoadedInMemory int) {
	defer cm.addSizeInMemory(sizeLoadedInMemory)

	entry, ok := cm.accessedKeys[string(key)]
	if !ok {
		return
	}

	cm.orderAccess.Remove(entry)
	delete(cm.accessedKeys, string(key))
}

func (cm *collapseManager) AddSizeInMemory(size int) {
	cm.addSizeInMemory(size)
}
func (cm *collapseManager) GetSizeInMemory() int {
	return cm.sizeInMemory
}

func (cm *collapseManager) ShouldCollapseTrie() bool {
	if uint64(cm.sizeInMemory) > cm.maxSizeInMem && len(cm.accessedKeys) < minNumLeavesToCollapseTrie {
		return true
	}

	return false
}

func (cm *collapseManager) GetCollapsibleLeaves() ([][]byte, error) {
	if uint64(cm.sizeInMemory) < cm.maxSizeInMem {
		return nil, nil
	}

	evictedKeys := make([][]byte, 0)
	for cm.sizeInMemory > int(cm.maxSizeInMem) && cm.orderAccess.Len() > 0 {
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

func (cm *collapseManager) CloneWithoutState() common.TrieCollapseManager {
	return &collapseManager{}
}

func (cm *collapseManager) IsCollapseEnabled() bool {
	return true
}

func (cm *collapseManager) IsInterfaceNil() bool {
	return cm == nil
}
