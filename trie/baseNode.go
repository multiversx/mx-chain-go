package trie

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

type baseNode struct {
	mutex  sync.RWMutex
	dirty  bool
	marsh  marshal.Marshalizer
	hasher hashing.Hasher
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

func (bn *baseNode) getMarshalizer() marshal.Marshalizer {
	return bn.marsh
}

func (bn *baseNode) setMarshalizer(marshalizer marshal.Marshalizer) {
	bn.marsh = marshalizer
}

func (bn *baseNode) getHasher() hashing.Hasher {
	return bn.hasher
}

func (bn *baseNode) setHasher(hasher hashing.Hasher) {
	bn.hasher = hasher
}
