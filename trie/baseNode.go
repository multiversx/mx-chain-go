package trie

import (
	"sync"
)

type baseNode struct {
	mutex sync.RWMutex
	dirty bool
}

func (bn *baseNode) isDirty() bool {
	bn.mutex.RLock()
	defer bn.mutex.RUnlock()

	return bn.dirty
}

func (bn *baseNode) setDirty(dirty bool) {
	bn.mutex.Lock()
	defer bn.mutex.Unlock()

	bn.dirty = dirty
}
